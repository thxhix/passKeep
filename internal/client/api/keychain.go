package api

import (
	"context"
	"errors"
	"fmt"
	"github.com/thxhix/passKeeper/internal/transport/client_http"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// KeychainAPI provides low-level methods to operate on the user's keychain
// resources on the server side (credentials, cards, texts, files).
//
// KeychainAPI methods are thin: they call an underlying client_http.Client and
// translate HTTP responses into DTOs.
type KeychainAPI struct {
	c *client_http.Client
}

// NewKeychainAPI creates a new KeychainAPI which uses the provided HTTP client.
func NewKeychainAPI(client *client_http.Client) *KeychainAPI {
	return &KeychainAPI{
		c: client,
	}
}

// AddCredential sends a credential to the server and returns the add result.
//
// ctx: request context.
// req: DTO with credential fields (title, login, password, etc).
// Returns AddSuccessResponse or an error. On non-2xx response the method
// returns an error (preferably *client_http.HTTPError).
func (a *KeychainAPI) AddCredential(ctx context.Context, req *dto.AddCredentialsDTO) (dto.AddSuccessResponse, error) {
	var out dto.AddSuccessResponse
	if err := a.c.Do(ctx, http.MethodPost, "/api/keychain/credential", req, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.AddSuccessResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.AddSuccessResponse{}, err
	}
	return out, nil
}

// AddCard sends a bank card DTO to the server the same AddCredential.
func (a *KeychainAPI) AddCard(ctx context.Context, req *dto.AddCardDTO) (dto.AddSuccessResponse, error) {
	var out dto.AddSuccessResponse
	if err := a.c.Do(ctx, http.MethodPost, "/api/keychain/card", req, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.AddSuccessResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.AddSuccessResponse{}, err
	}
	return out, nil
}

// AddText sends a text DTO to the server the same AddCredential.
func (a *KeychainAPI) AddText(ctx context.Context, req *dto.AddTextDTO) (dto.AddSuccessResponse, error) {
	var out dto.AddSuccessResponse
	if err := a.c.Do(ctx, http.MethodPost, "/api/keychain/text", req, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.AddSuccessResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.AddSuccessResponse{}, err
	}
	return out, nil
}

// AddFile uploads a file to the server together with optional metadata (title/note).
//
// The file is streamed using a pipe + multipart.Writer to avoid buffering the
// entire file in memory. filePath must point to a readable file.
func (a *KeychainAPI) AddFile(ctx context.Context, req *dto.AddFileDTO, filePath string) (dto.AddSuccessResponse, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return dto.AddSuccessResponse{}, err
	}

	defer f.Close()

	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)

	// start writer goroutine
	go func() {
		defer func() {
			_ = mw.Close()
			_ = pw.Close()
		}()

		part, err := mw.CreateFormFile("file", filepath.Base(filePath))
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, f); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		if req.Title != "" {
			if err := mw.WriteField("title", req.Title); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
		if req.Note != "" {
			if err := mw.WriteField("note", req.Note); err != nil {
				_ = pw.CloseWithError(err)
				return
			}
		}
	}()

	contentType := mw.FormDataContentType()

	var out dto.AddSuccessResponse
	if err := a.c.DoMultiPart(ctx, http.MethodPost, "/api/keychain/file", pr, contentType, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.AddSuccessResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.AddSuccessResponse{}, err
	}
	return out, nil
}

// GetKeysList fetches listing of keys. If keyType is non-empty it is used as a
// query parameter (server-side filtering).
func (a *KeychainAPI) GetKeysList(ctx context.Context, keyType string) (dto.GetKeysResponse, error) {
	var out dto.GetKeysResponse

	url := "/api/keychain"
	if keyType != "" {
		url = fmt.Sprintf("/api/keychain/?type=%s", keyType)
	}

	if err := a.c.Do(ctx, http.MethodGet, url, nil, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.GetKeysResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.GetKeysResponse{}, err
	}
	return out, nil
}

// GetKey fetches a single key by uuid.
func (a *KeychainAPI) GetKey(ctx context.Context, keyUUID string) (dto.GetKeyResponse, error) {
	var out dto.GetKeyResponse

	url := fmt.Sprintf("/api/keychain/%s", keyUUID)

	if err := a.c.Do(ctx, http.MethodGet, url, nil, &out); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return dto.GetKeyResponse{}, fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return dto.GetKeyResponse{}, err
	}
	return out, nil
}

// DeleteKey deletes a record by uuid.
func (a *KeychainAPI) DeleteKey(ctx context.Context, keyUUID string) error {
	url := fmt.Sprintf("/api/keychain/%s", keyUUID)

	if err := a.c.Do(ctx, http.MethodDelete, url, nil, nil); err != nil {
		var he *client_http.HTTPError
		if errors.As(err, &he) {
			return fmt.Errorf("http code %d: %s", he.StatusCode, he.Body)
		}
		return err
	}
	return nil
}
