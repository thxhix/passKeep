package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/thxhix/passKeeper/internal/transport/client_http"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"net/http"
)

// AuthAPI defines HTTP methods for authentication and token management.
//
// It acts as a lightweight layer over client_http.Client — it performs HTTP requests,
// parses responses into DTO structs, and returns typed errors.
type AuthAPI struct {
	c *client_http.Client
}

// NewAuthAPI creates a new AuthAPI instance using the provided HTTP client.
func NewAuthAPI(client *client_http.Client) *AuthAPI {
	return &AuthAPI{
		c: client,
	}
}

// Register sends a registration request to the backend.
//
// It expects a non-nil RegisterRequest containing login and password fields.
// On success, it returns a TokenResponse with issued access and refresh tokens.
// If the server responds with a non-2xx code, an error of type *client_http.HTTPError
// will be returned.
func (a *AuthAPI) Register(ctx context.Context, req *dto.RegisterRequest) (dto.TokenResponse, error) {
	var out dto.TokenResponse
	if err := a.c.Do(ctx, http.MethodPost, "/api/auth/register", req, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.TokenResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.TokenResponse{}, err
	}
	return out, nil
}

// Login performs a login request using the provided credentials.
//
// On success, it returns a TokenResponse with access and refresh tokens.
// On error, it returns either a *client_http.HTTPError or another wrapped error
// if the HTTP call failed before receiving a response.
func (a *AuthAPI) Login(ctx context.Context, req *dto.LoginRequest) (dto.TokenResponse, error) {
	var out dto.TokenResponse
	if err := a.c.Do(ctx, http.MethodPost, "/api/auth/login", req, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.TokenResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.TokenResponse{}, err
	}
	return out, nil
}

// RefreshToken exchanges a valid refresh token for a new pair of tokens.
//
// The returned RefreshedTokenResponse contains the new access and refresh tokens.
// If the refresh token is invalid or expired, the server will respond with
// an error which is returned as *client_http.HTTPError.
func (a *AuthAPI) RefreshToken(ctx context.Context, req *dto.RefreshRequest) (dto.RefreshedTokenResponse, error) {
	var out dto.RefreshedTokenResponse
	if err := a.c.Do(ctx, http.MethodPost, "/api/auth/refresh", req, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.RefreshedTokenResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.RefreshedTokenResponse{}, err
	}
	return out, nil
}
