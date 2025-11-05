package client_services

import (
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/thxhix/passKeeper/internal/client/api"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestKeychainClientService_AddCredentialAndTextAndCard(t *testing.T) {
	// server stub
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
		_ = json.NewEncoder(w).Encode(dto.AddSuccessResponse{UUID: "cred-1"})
	})
	mux.HandleFunc("/api/keychain/card", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var in dto.AddCardDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid json"})
			return
		}
		_ = json.NewEncoder(w).Encode(dto.AddSuccessResponse{UUID: "card-1"})
	})
	mux.HandleFunc("/api/keychain/text", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		var in dto.AddTextDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid json"})
			return
		}
		_ = json.NewEncoder(w).Encode(dto.AddSuccessResponse{UUID: "text-1"})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := newTestClient(t, ts.URL)
	api := api.NewKeychainAPI(client)
	svc := NewKeychainClientService(api, client)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// AddCredential
	if err := svc.AddCredential(ctx, "title", "login", "pass", "site", "note"); err != nil {
		t.Fatalf("AddCredential failed: %v", err)
	}

	// AddCard
	if err := svc.AddCard(ctx, "t", "4111111111111111", "12/30", "123", "Holder", "Bank", "note"); err != nil {
		t.Fatalf("AddCard failed: %v", err)
	}

	// AddText
	if err := svc.AddText(ctx, "t", "some text", "note"); err != nil {
		t.Fatalf("AddText failed: %v", err)
	}
}

func TestKeychainClientService_AddFile_GetList_Get_Delete(t *testing.T) {
	// server stub
	mux := http.NewServeMux()

	uuidTest := uuid.New()

	// list handler
	mux.HandleFunc("/api/keychain", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(dto.GetKeysResponse{
			Keys: []*dto.GetKeysRecord{
				{
					KeyUUID:   uuidTest,
					KeyType:   keychain.KeyCredential,
					Title:     "t1",
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		})
	})

	// get key
	mux.HandleFunc("/api/keychain/"+uuidTest.String(), func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(dto.GetKeyResponse{
			KeyUUID: uuidTest,
			Title:   "t1",
		})
	})

	// delete existing -> 204, not found -> 404
	mux.HandleFunc("/api/keychain/exists-uuid", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/keychain/missing-uuid", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "not found"})
	})

	// file upload handler
	mux.HandleFunc("/api/keychain/file", func(w http.ResponseWriter, r *http.Request) {
		// parse multipart
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		f, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer f.Close()
		// read first bytes
		buf := make([]byte, 8)
		_, _ = f.Read(buf)

		_ = json.NewEncoder(w).Encode(dto.AddSuccessResponse{
			UUID: "file-uuid",
		})
	})

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := newTestClient(t, ts.URL)
	api := api.NewKeychainAPI(client)
	svc := NewKeychainClientService(api, client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// prepare temp file
	tmp := filepath.Join(os.TempDir(), "test-upload.txt")
	if err := os.WriteFile(tmp, []byte("hello-file"), 0o600); err != nil {
		t.Fatalf("write tmp failed: %v", err)
	}
	defer os.Remove(tmp)

	// AddFile
	if err := svc.AddFile(ctx, "mytitle", tmp, "mynote"); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}

	// GetList
	if _, err := svc.GetList(ctx, ""); err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	// Get
	if _, err := svc.Get(ctx, uuidTest.String()); err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	// Delete exists -> expect nil
	if err := svc.Delete(ctx, "exists-uuid"); err != nil {
		t.Fatalf("Delete (exists) failed: %v", err)
	}

	// Delete missing -> expect error
	if err := svc.Delete(ctx, "missing-uuid"); err == nil {
		t.Fatalf("Delete (missing) expected error, got nil")
	}
}
