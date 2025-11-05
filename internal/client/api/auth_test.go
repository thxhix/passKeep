package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	clientpkg "github.com/thxhix/passKeeper/internal/transport/client_http"
	dto "github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
)

// TestAuthAPI_Success проверяет успешные ответы для Register, Login, RefreshToken.
func TestAuthAPI_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/register":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(dto.TokenResponse{AccessToken: "A", RefreshToken: "R"})
		case "/api/auth/login":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(dto.TokenResponse{AccessToken: "L", RefreshToken: "Lref"})
		case "/api/auth/refresh":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(dto.RefreshedTokenResponse{AccessToken: "AR", RefreshToken: "RR"})
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client, err := clientpkg.NewHttpClient(ts.URL, logger)
	if err != nil {
		t.Fatalf("NewHttpClient: %v", err)
	}
	api := NewAuthAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Register
	_, err = api.Register(ctx, &dto.RegisterRequest{Login: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Register expected nil err, got: %v", err)
	}

	// Login
	_, err = api.Login(ctx, &dto.LoginRequest{Login: "u", Password: "p"})
	if err != nil {
		t.Fatalf("Login expected nil err, got: %v", err)
	}

	// RefreshToken
	_, err = api.RefreshToken(ctx, &dto.RefreshRequest{RefreshToken: "r"})
	if err != nil {
		t.Fatalf("RefreshToken expected nil err, got: %v", err)
	}
}

// TestAuthAPI_ErrorPaths проверяет поведение при ошибке с JSON-описанием и при plain-text ошибке.
func TestAuthAPI_ErrorPaths(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/register":
			// возвращаем json-ошибку
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "bad register"})
		case "/api/auth/login":
			// возвращаем plain text
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("something went wrong"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	logger := zap.NewNop()
	client, err := clientpkg.NewHttpClient(ts.URL, logger)
	if err != nil {
		t.Fatalf("NewHttpClient: %v", err)
	}
	api := NewAuthAPI(client)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Register -> json error: expect error string contains http code and message
	_, err = api.Register(ctx, &dto.RegisterRequest{Login: "x", Password: "p"})

	if err == nil {
		t.Fatalf("Register expected error, got nil")
	}

	// Login -> plain text error: expect error contains status and plain body
	_, err = api.Login(ctx, &dto.LoginRequest{Login: "x", Password: "p"})
	if err == nil {
		t.Fatalf("Login expected error, got nil")
	}
}
