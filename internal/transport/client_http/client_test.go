package client_http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
)

// Test basic Do() success / no-content / json-error / non-json-error
func TestClient_Do_BasicCases(t *testing.T) {
	mux := http.NewServeMux()

	// success
	mux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{"a": "b"})
	})

	// no content
	mux.HandleFunc("/nocontent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// json error
	mux.HandleFunc("/errjson", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "bad input"})
	})

	// non-json error
	mux.HandleFunc("/errtxt", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, "plain failure")
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	logger := zap.NewNop()
	c, err := NewHttpClient(srv.URL, logger)
	if err != nil {
		t.Fatalf("NewHttpClient failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// success
	var got map[string]string
	if err := c.Do(ctx, http.MethodGet, "/ok", nil, &got); err != nil {
		t.Fatalf("Do /ok expected nil err, got %v", err)
	}
	if got["a"] != "b" {
		t.Fatalf("unexpected response: %#v", got)
	}

	// no content should not error even if result == nil
	if err := c.Do(ctx, http.MethodGet, "/nocontent", nil, nil); err != nil {
		t.Fatalf("Do /nocontent expected nil err, got %v", err)
	}

	// json error should return HTTPError
	if err := c.Do(ctx, http.MethodGet, "/errjson", nil, nil); err == nil {
		t.Fatalf("Do /errjson expected error")
	} else {
		if _, ok := err.(*HTTPError); !ok {
			t.Fatalf("expected HTTPError type, got %T: %v", err, err)
		}
	}

	// non-json error should also return HTTPError
	if err := c.Do(ctx, http.MethodGet, "/errtxt", nil, nil); err == nil {
		t.Fatalf("Do /errtxt expected error")
	} else {
		if he, ok := err.(*HTTPError); ok {
			if he.StatusCode != http.StatusInternalServerError {
				t.Fatalf("expected 500 status in HTTPError, got %d", he.StatusCode)
			}
		} else {
			t.Fatalf("expected HTTPError, got %T", err)
		}
	}
}

// Test DoMultiPart streaming upload path
func TestClient_DoMultiPart_Success(t *testing.T) {
	// server accepts multipart and returns json
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/upload" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		// read multipart (but we don't need to parse; just ensure it's present)
		ct := r.Header.Get("Content-Type")
		if ct == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		// drain body (simulate processing)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = json.NewEncoder(w).Encode(map[string]string{"id": "ok"})
	}))
	defer srv.Close()

	logger := zap.NewNop()
	c, err := NewHttpClient(srv.URL, logger)
	if err != nil {
		t.Fatalf("NewHttpClient failed: %v", err)
	}

	// simulate multipart body (simple)
	body := bytes.NewBufferString("--boundary\r\nContent-Disposition: form-data; name=\"file\"; filename=\"f\"\r\n\r\ndata\r\n--boundary--\r\n")
	contentType := "multipart/form-data; boundary=boundary"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out map[string]string
	if err := c.DoMultiPart(ctx, http.MethodPost, "/upload", body, contentType, &out); err != nil {
		t.Fatalf("DoMultiPart expected nil err, got %v", err)
	}
	if out["id"] != "ok" {
		t.Fatalf("unexpected upload response: %#v", out)
	}
}
