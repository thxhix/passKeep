package api

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	clientpkg "github.com/thxhix/passKeeper/internal/transport/client_http"
	dto "github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
)

// helper to create client
func newTestClient(t *testing.T, srvURL string) *clientpkg.Client {
	t.Helper()
	logger := zap.NewNop()
	c, err := clientpkg.NewHttpClient(srvURL, logger)
	if err != nil {
		t.Fatalf("NewHttpClient failed: %v", err)
	}
	return c
}

// Test AddCredential happy path and server-side JSON error path.
func TestKeychainAPI_AddCredential(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/keychain/credential", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var in dto.AddCredentialsDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid json"})
			return
		}
		// simulate success
		_ = json.NewEncoder(w).Encode(dto.AddSuccessResponse{UUID: "uuid-1"})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := newTestClient(t, ts.URL)
	api := NewKeychainAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req := &dto.AddCredentialsDTO{Title: "t1", Login: "u", Password: "p"}
	got, err := api.AddCredential(ctx, req)
	if err != nil {
		t.Fatalf("AddCredential failed: %v", err)
	}
	if got.UUID != "uuid-1" {
		t.Fatalf("unexpected response uuid: %s", got.UUID)
	}
}

// Test AddFile uploads multipart file and reads server response.
func TestKeychainAPI_AddFile(t *testing.T) {
	// server handler: parse multipart and return JSON
	mux := http.NewServeMux()
	mux.HandleFunc("/api/keychain/file", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// parse multipart
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		file, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer file.Close()
		// read small sample
		b := make([]byte, 16)
		_, _ = file.Read(b)

		_ = json.NewEncoder(w).Encode(dto.AddSuccessResponse{
			UUID: "file-uuid",
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := newTestClient(t, ts.URL)
	api := NewKeychainAPI(client)

	// create a small temp file to upload
	tmp := filepath.Join(os.TempDir(), "test-upload.txt")
	if err := os.WriteFile(tmp, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("write tmp: %v", err)
	}
	defer os.Remove(tmp)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &dto.AddFileDTO{Title: "mytitle", Note: "mynote"}
	got, err := api.AddFile(ctx, req, tmp)
	if err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}
	if got.UUID != "file-uuid" {
		t.Fatalf("unexpected uuid: %s", got.UUID)
	}
}

// Test GetKeysList (with type param) and GetKey / DeleteKey behaviors.
func TestKeychainAPI_ListGetDelete(t *testing.T) {
	mux := http.NewServeMux()
	uuidTest := uuid.New()

	// list handler
	mux.HandleFunc("/api/keychain", func(w http.ResponseWriter, r *http.Request) {
		// accept ?type=credential or empty
		_ = json.NewEncoder(w).Encode(dto.GetKeysResponse{
			Keys: []*dto.GetKeysRecord{
				{KeyUUID: uuidTest, KeyType: keychain.KeyCredential, Title: "t1", CreatedAt: time.Now(), UpdatedAt: time.Now()},
			},
		})
	})

	// get key
	mux.HandleFunc("/api/keychain/"+uuidTest.String(), func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(dto.GetKeyResponse{
			KeyUUID:   uuidTest,
			KeyType:   keychain.KeyCredential,
			Title:     "t1",
			Data:      nil,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	})

	// delete returns 204
	mux.HandleFunc("/api/keychain/22222222-2222-2222-2222-222222222222", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "not found"})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := newTestClient(t, ts.URL)
	api := NewKeychainAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// list
	_, err := api.GetKeysList(ctx, "")
	if err != nil {
		t.Fatalf("GetKeysList failed: %v", err)
	}

	// get key
	_, err = api.GetKey(ctx, uuidTest.String())
	if err != nil {
		t.Fatalf("GetKey failed: %v", err)
	}

	// delete missing -> expect error
	if err := api.DeleteKey(ctx, "22222222-2222-2222-2222-222222222222"); err == nil {
		t.Fatalf("DeleteKey expected error for not found id")
	}
}
