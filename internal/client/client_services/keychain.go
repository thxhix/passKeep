package client_services

import (
	"context"
	"github.com/thxhix/passKeeper/internal/client/api"
	"github.com/thxhix/passKeeper/internal/transport/client_http"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
)

// KeychainClientService is a thin service that exposes keychain-related
// operations to the CLI layer.
type KeychainClientService struct {
	API    *api.KeychainAPI
	Client *client_http.Client
}

// NewKeychainClientService constructs a new KeychainClientService.
func NewKeychainClientService(api *api.KeychainAPI, httpClient *client_http.Client) *KeychainClientService {
	return &KeychainClientService{
		API:    api,
		Client: httpClient,
	}
}

// AddCredential sends a credential entry to the server.
// On success returns nil, otherwise returns an error returned by the API.
func (s *KeychainClientService) AddCredential(ctx context.Context, title, login, password, site, note string) error {
	in := &dto.AddCredentialsDTO{
		Title:    title,
		Login:    login,
		Password: password,
		Site:     site,
		Note:     note,
	}

	_, err := s.API.AddCredential(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

// AddCard sends a card entry to the server.
func (s *KeychainClientService) AddCard(ctx context.Context, title, number, expDate, cvv, holder, bank, note string) error {
	in := &dto.AddCardDTO{
		Title:   title,
		Number:  number,
		ExpDate: expDate,
		CVV:     cvv,
		Holder:  holder,
		Bank:    bank,
		Note:    note,
	}

	_, err := s.API.AddCard(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

// AddText sends a text entry to the server.
func (s *KeychainClientService) AddText(ctx context.Context, title, text, note string) error {
	in := &dto.AddTextDTO{
		Title: title,
		Text:  text,
		Note:  note,
	}

	_, err := s.API.AddText(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

// AddFile uploads a file to the server together with optional title and note.
// filePath must point to a readable file. The method streams the file via the
// underlying API's multipart endpoint.
func (s *KeychainClientService) AddFile(ctx context.Context, title, filePath, note string) error {
	in := &dto.AddFileDTO{
		Title: title,
		Note:  note,
	}

	_, err := s.API.AddFile(ctx, in, filePath)
	if err != nil {
		return err
	}

	return nil
}

// GetList requests a list of keys. If keyType is empty, server should return all keys.
func (s *KeychainClientService) GetList(ctx context.Context, keyType string) (dto.GetKeysResponse, error) {
	return s.API.GetKeysList(ctx, keyType)
}

// Get fetches a single key payload by UUID.
func (s *KeychainClientService) Get(ctx context.Context, keyUUID string) (dto.GetKeyResponse, error) {
	return s.API.GetKey(ctx, keyUUID)
}

// Delete removes a key by UUID.
func (s *KeychainClientService) Delete(ctx context.Context, keyUUID string) error {
	return s.API.DeleteKey(ctx, keyUUID)
}
