package client_services

import (
	"context"
	"github.com/thxhix/passKeeper/internal/client/api"
	"github.com/thxhix/passKeeper/internal/client/token"
	"github.com/thxhix/passKeeper/internal/transport/client_http"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
)

// AuthClientService provides service-level authentication flows for the CLI client.
//
// It wraps the low-level api.AuthAPI and performs actions such as registering,
// logging in, and persisting tokens to the local token store.
type AuthClientService struct {
	API    *api.AuthAPI
	Client *client_http.Client
}

// NewAuthClientService creates a new AuthClientService using the provided
// AuthAPI wrapper and HTTP client.
func NewAuthClientService(api *api.AuthAPI, httpClient *client_http.Client) *AuthClientService {
	return &AuthClientService{
		API:    api,
		Client: httpClient,
	}
}

// Register registers a new user on the server and persists returned tokens
// using the token package. On success the tokens are stored; on failure the
// returned error is propagated.
func (s *AuthClientService) Register(ctx context.Context, login, password string) error {
	in := &dto.RegisterRequest{
		Login:    login,
		Password: password,
	}

	tokens, err := s.API.Register(ctx, in)
	if err != nil {
		return err
	}

	keyRingTokens := token.Tokens{
		Access:  tokens.AccessToken,
		Refresh: tokens.RefreshToken,
	}

	err = token.SaveTokens(keyRingTokens)
	if err != nil {
		return err
	}

	return nil
}

// Login authenticates the user and persists received tokens
// using the token package. On success the tokens are stored; on failure the
// returned error is propagated.
func (s *AuthClientService) Login(ctx context.Context, login, password string) error {
	in := &dto.LoginRequest{
		Login:    login,
		Password: password,
	}

	tokens, err := s.API.Login(ctx, in)
	if err != nil {
		return err
	}

	keyRingTokens := token.Tokens{
		Access:  tokens.AccessToken,
		Refresh: tokens.RefreshToken,
	}

	err = token.SaveTokens(keyRingTokens)
	if err != nil {
		return err
	}

	return nil
}

// RefreshToken reads the stored refresh token and exchanges it for a new pair
// of tokens via the API, persisting the new values on success.
func (s *AuthClientService) RefreshToken(ctx context.Context) error {
	keyRingTokensStorage, err := token.LoadTokens()
	if err != nil {
		return err
	}

	in := &dto.RefreshRequest{
		RefreshToken: keyRingTokensStorage.Refresh,
	}
	tokens, err := s.API.RefreshToken(ctx, in)
	if err != nil {
		return err
	}

	keyRingTokens := token.Tokens{
		Access:  tokens.AccessToken,
		Refresh: tokens.RefreshToken,
	}

	err = token.SaveTokens(keyRingTokens)
	if err != nil {
		return err
	}

	return nil
}
