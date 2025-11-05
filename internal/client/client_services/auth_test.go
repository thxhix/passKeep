package client_services

import (
	"context"
	"encoding/json"
	"github.com/thxhix/passKeeper/internal/client/api"
	"github.com/thxhix/passKeeper/internal/client/token"
	clientpkg "github.com/thxhix/passKeeper/internal/transport/client_http"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// helper: create http client pointing to test server
func newTestClient(t *testing.T, baseURL string) *clientpkg.Client {
	t.Helper()
	logger := zap.NewNop()
	c, err := clientpkg.NewHttpClient(baseURL, logger)
	if err != nil {
		t.Fatalf("NewHttpClient failed: %v", err)
	}
	return c
}

func TestAuthClientService_RegisterAndLoginAndRefresh(t *testing.T) {
	// Clean any existing tokens before/after test to avoid interference.
	_ = token.DeleteTokens()

	// test server that handles register, login, refresh endpoints
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/auth/register":
			// read request body to ensure it's valid JSON
			var in dto.RegisterRequest
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid json"})
				return
			}
			// respond with tokens
			_ = json.NewEncoder(w).Encode(dto.TokenResponse{
				AccessToken:  "access-register",
				RefreshToken: "refresh-register",
			})
			return

		case "/api/auth/login":
			var in dto.LoginRequest
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid json"})
				return
			}
			_ = json.NewEncoder(w).Encode(dto.TokenResponse{
				AccessToken:  "access-login",
				RefreshToken: "refresh-login",
			})
			return

		case "/api/auth/refresh":
			var in dto.RefreshRequest
			if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid json"})
				return
			}
			// if refresh token matches expected, return new tokens
			if in.RefreshToken == "refresh-login" || in.RefreshToken == "refresh-register" {
				_ = json.NewEncoder(w).Encode(dto.RefreshedTokenResponse{
					AccessToken:  "access-refreshed",
					RefreshToken: "refresh-refreshed",
				})
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(dto.ErrorResponse{ErrorText: "invalid refresh"})
			return

		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	defer token.DeleteTokens() // cleanup

	// build client / api / service
	client := newTestClient(t, ts.URL)
	authAPI := api.NewAuthAPI(client)
	authSvc := NewAuthClientService(authAPI, client)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 1) Register: should save tokens returned by server
	if err := authSvc.Register(ctx, "user-reg", "pass"); err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	tok, err := token.LoadTokens()
	if err != nil {
		t.Fatalf("token.LoadTokens after Register: %v", err)
	}
	if tok.Access != "access-register" || tok.Refresh != "refresh-register" {
		t.Fatalf("unexpected tokens after register: %+v", tok)
	}

	// 2) Login: should overwrite tokens with login's tokens
	if err := authSvc.Login(ctx, "user-log", "pass"); err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	tok, err = token.LoadTokens()
	if err != nil {
		t.Fatalf("token.LoadTokens after Login: %v", err)
	}
	if tok.Access != "access-login" || tok.Refresh != "refresh-login" {
		t.Fatalf("unexpected tokens after login: %+v", tok)
	}

	// 3) RefreshToken: uses stored refresh token and updates tokens
	if err := authSvc.RefreshToken(ctx); err != nil {
		t.Fatalf("RefreshToken failed: %v", err)
	}
	tok, err = token.LoadTokens()
	if err != nil {
		t.Fatalf("token.LoadTokens after Refresh: %v", err)
	}
	if tok.Access != "access-refreshed" || tok.Refresh != "refresh-refreshed" {
		t.Fatalf("unexpected tokens after refresh: %+v", tok)
	}
}
