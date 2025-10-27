package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/mocks"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"github.com/thxhix/passKeeper/internal/transport/http/middleware"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// helper для создания хендлеров с моками
func makeKeychainHandlers(keySvc *mocks.KeychainServiceMock) *Handlers {
	return &Handlers{
		keychainService: keySvc,
		logger:          zap.NewNop(),
	}
}

// helper для контекста с userID
func contextWithUserID(uid int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, middleware.CtxUserID, strconv.Itoa(int(uid)))
	return ctx
}

func TestHandlers_GetKeys(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)

		userID := int64(1)
		keys := []*keychain.KeyRecord{
			{
				KeyUUID:   uuid.New(),
				KeyType:   keychain.KeyText,
				Title:     "title1",
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		keySvc.On("GetKeys", mock.Anything, userID, (*keychain.KeyType)(nil)).Return(keys, nil)

		req := httptest.NewRequest(http.MethodGet, "/keys", nil)
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.GetKeys(rec, req)

		res := rec.Result()

		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)

		b, _ := io.ReadAll(res.Body)
		resp := dto.GetKeysResponse{}
		_ = json.Unmarshal(b, &resp)
		assert.Len(t, resp.Keys, 1)

		keySvc.AssertExpectations(t)
	})

	t.Run("unauthorized", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)

		req := httptest.NewRequest(http.MethodGet, "/keys", nil)
		rec := httptest.NewRecorder()

		h.GetKeys(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})
}

func TestHandlers_GetKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)

		userID := int64(1)
		keyUUID := uuid.New()
		record := &keychain.KeyRecord{
			KeyUUID:   keyUUID,
			KeyType:   keychain.KeyText,
			Title:     "title",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		data := dto.AddCredentialsDTO{
			Title:    record.Title,
			Login:    "login",
			Password: "password",
			Site:     "site",
			Note:     "note",
		}

		marshalJSON, err := data.MarshalJSON()
		if err != nil {
			return
		}

		keySvc.On("GetKey", mock.Anything, userID, keyUUID.String()).Return(record, marshalJSON, nil)

		req := httptest.NewRequest(http.MethodGet, "/keys/"+keyUUID.String(), nil)
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		// chi.URLParam подменить не получится напрямую, поэтому используем chi route context
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", keyUUID.String())
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

		h.GetKey(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode)
	})
	t.Run("not found", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)

		userID := int64(1)
		keyUUID := uuid.New().String()
		keySvc.On("GetKey", mock.Anything, userID, keyUUID).Return(&keychain.KeyRecord{}, []byte{}, sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/keys/"+keyUUID, nil)
		req = req.WithContext(contextWithUserID(userID))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", keyUUID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()

		h.GetKey(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestHandlers_DeleteKey(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)

		userID := int64(1)
		keyUUID := uuid.New().String()

		keySvc.On("DeleteKey", mock.Anything, userID, keyUUID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/keys/"+keyUUID, nil)
		req = req.WithContext(contextWithUserID(userID))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", keyUUID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()

		h.DeleteKey(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNoContent, res.StatusCode)
	})
	t.Run("not found", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)

		userID := int64(1)
		keyUUID := uuid.New().String()

		keySvc.On("DeleteKey", mock.Anything, userID, keyUUID).Return(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodDelete, "/keys/"+keyUUID, nil)
		req = req.WithContext(contextWithUserID(userID))
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("uuid", keyUUID)
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()

		h.DeleteKey(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func TestHandlers_AddCredential(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		body := `{"title":"t1","login":"l1","password":"p1"}`
		keyUUID := uuid.New().String()
		keySvc.On("AddCredential", mock.Anything, userID, mock.Anything).Return(keyUUID, nil)

		req := httptest.NewRequest(http.MethodPost, "/keys/credential", strings.NewReader(body))
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddCredential(rec, req)

		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
	})
	t.Run("bad request json", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		body := `{"title":`
		req := httptest.NewRequest(http.MethodPost, "/keys/credential", strings.NewReader(body))
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddCredential(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestHandlers_AddCard(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		body := `{"title":"t1","number":"1111","exp":"12/30"}`
		keyUUID := uuid.New().String()
		keySvc.On("AddCard", mock.Anything, userID, mock.Anything).Return(keyUUID, nil)

		req := httptest.NewRequest(http.MethodPost, "/keys/card", strings.NewReader(body))
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddCard(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
	})

	t.Run("bad request json", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		req := httptest.NewRequest(http.MethodPost, "/keys/card", strings.NewReader(`{"title":`))
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddCard(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestHandlers_AddText(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		body := `{"title":"t1","text":"hello"}`
		keyUUID := uuid.New().String()
		keySvc.On("AddText", mock.Anything, userID, mock.Anything).Return(keyUUID, nil)

		req := httptest.NewRequest(http.MethodPost, "/keys/text", strings.NewReader(body))
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddText(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
	})

	t.Run("bad request json", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		req := httptest.NewRequest(http.MethodPost, "/keys/text", strings.NewReader(`{"title":`))
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddText(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}

func TestHandlers_AddFile(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		keyUUID := uuid.New().String()
		keySvc.On("AddFile", mock.Anything, userID, mock.Anything).Return(keyUUID, nil)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("file content"))
		_ = writer.WriteField("title", "t1")
		_ = writer.WriteField("note", "n1")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/keys/file", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddFile(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusCreated, res.StatusCode)
	})

	t.Run("file too large", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write(bytes.Repeat([]byte("a"), 11<<20)) // > 10MB
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/keys/file", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddFile(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusRequestEntityTooLarge, res.StatusCode)
	})

	t.Run("file not found", func(t *testing.T) {
		keySvc := new(mocks.KeychainServiceMock)
		h := makeKeychainHandlers(keySvc)
		userID := int64(1)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		// no file part
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/keys/file", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = req.WithContext(contextWithUserID(userID))
		rec := httptest.NewRecorder()

		h.AddFile(rec, req)
		res := rec.Result()
		defer res.Body.Close()
		assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	})
}
