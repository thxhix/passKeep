package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mailru/easyjson"
	"github.com/thxhix/passKeeper/internal/apperr"
	"github.com/thxhix/passKeeper/internal/domain/keychain"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"github.com/thxhix/passKeeper/internal/transport/http/middleware"
	"go.uber.org/zap"
	"io"
	"net/http"
	"time"
)

// GetKeys returns a list of all user keys.
//
// Query parameters:
//
//	type (optional) – filters keys by type (credential, card, file, text).
//
// Status codes:
//
//	200 OK – the key list was returned successfully.
//	400 BadRequest – invalid 'type' query parameter.
//	401 Unauthorized – if the user is not authenticated.
//	500 InternalServerError – internal service error.
func (h *Handlers) GetKeys(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	var typePtr *keychain.KeyType
	if raw := r.URL.Query().Get("type"); raw != "" {
		if t, ok := keychain.ParseKeyType(raw); ok {
			typePtr = &t
		} else {
			h.PublicError(w, http.StatusBadRequest, ErrBadQuery)
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	list, err := h.keychainService.GetKeys(ctx, userId, typePtr)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	mappedList := dto.GetKeysResponse{}

	for _, record := range list {
		mappedRecord := &dto.GetKeysRecord{
			KeyUUID:   record.KeyUUID,
			KeyType:   record.KeyType,
			Title:     record.Title,
			CreatedAt: record.CreatedAt,
			UpdatedAt: record.UpdatedAt,
		}
		mappedList.Keys = append(mappedList.Keys, mappedRecord)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := easyjson.MarshalToWriter(&mappedList, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// GetKey returns a specific user key by UUID.
//
// URL parameters:
//
//	uuid – the key UUID.
//
// Status codes:
//
//	200 OK – the key was found and returned.
//	400 BadRequest – invalid UUID.
//	401 Unauthorized – user is not authenticated.
//	404 NotFound – key not found.
//	500 InternalServerError – internal service error.
func (h *Handlers) GetKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	keyUUID := chi.URLParam(r, "uuid")
	if _, err := uuid.Parse(keyUUID); err != nil {
		h.logger.Error(ErrBadRequest.Error(), zap.Error(err))
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	keyRecord, plainDecrypted, err := h.keychainService.GetKey(ctx, userId, keyUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.PublicError(w, http.StatusNotFound, ErrNotFound)
			return
		}
		h.InternalError(w, err)
		return
	}

	var data json.RawMessage

	switch keyRecord.KeyType {
	case keychain.KeyCredential:
		var d dto.CredentialsResponseDTO
		if err := json.Unmarshal(plainDecrypted, &d); err != nil {
			h.InternalError(w, err)
			return
		}
		b, _ := json.Marshal(d)
		data = b

	case keychain.KeyBankCard:
		var d dto.CardResponseDTO
		if err := json.Unmarshal(plainDecrypted, &d); err != nil {
			h.InternalError(w, err)
			return
		}
		b, _ := json.Marshal(d)
		data = b

	case keychain.KeyFile:
		var d dto.FileResponseDTO
		if err := json.Unmarshal(plainDecrypted, &d); err != nil {
			h.InternalError(w, err)
			return
		}
		b, _ := json.Marshal(d)
		data = b

	case keychain.KeyText:
		var d dto.TextResponseDTO
		if err := json.Unmarshal(plainDecrypted, &d); err != nil {
			h.InternalError(w, err)
			return
		}
		b, _ := json.Marshal(d)
		data = b

	default:
		data = json.RawMessage(plainDecrypted)
	}

	respObj := dto.GetKeyResponse{
		KeyUUID:   keyRecord.KeyUUID,
		KeyType:   keyRecord.KeyType,
		Title:     keyRecord.Title,
		Data:      data,
		CreatedAt: keyRecord.CreatedAt,
		UpdatedAt: keyRecord.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// DeleteKey deletes a user key by UUID.

// URL parameters:
//
//	uuid – the key UUID.
//
// Status codes:
//
//	204 NoContent – the key was successfully deleted.
//	400 BadRequest – invalid UUID.
//	401 Unauthorized – user is not authenticated.
//	404 NotFound – key not found.
//	500 InternalServerError – internal service error.
func (h *Handlers) DeleteKey(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	keyUUID := chi.URLParam(r, "uuid")
	if _, err := uuid.Parse(keyUUID); err != nil {
		h.logger.Error(ErrBadRequest.Error(), zap.Error(err))
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	err := h.keychainService.DeleteKey(ctx, userId, keyUUID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.PublicError(w, http.StatusNotFound, ErrNotFound)
			return
		}
		h.InternalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNoContent)
}

// AddCredential adds a new credential key for the user.

// Body (JSON):
//
//	{
//	  "title": "string",
//	  "login": "string",
//	  "password": "string"
//	}
//
// Status codes:
//
//	201 Created – the key was successfully added.
//	400 BadRequest – invalid JSON or validation error.
//	401 Unauthorized – user is not authenticated.
//	500 InternalServerError – internal service error.
func (h *Handlers) AddCredential(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	reqObj := dto.AddCredentialsDTO{}
	err = easyjson.Unmarshal(body, &reqObj)
	if err != nil {
		h.logger.Error(ErrBadRequest.Error(), zap.Error(err))
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	keyUUID, err := h.keychainService.AddCredential(r.Context(), userId, reqObj)
	if err != nil {
		var ve *apperr.ValidationError
		if errors.As(err, &ve) {
			h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
			return
		}
		h.InternalError(w, err)
		return
	}

	respObj := dto.AddSuccessResponse{
		UUID: keyUUID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// AddCard adds a bank card.
//
// Body (JSON):
//
//	{
//	  "title": "string",
//	  "number": "string",
//	  "exp": "string"
//	}
//
// Status codes:
//
//	201 Created – the card was successfully added.
//	400 BadRequest – invalid JSON or validation error.
//	401 Unauthorized – user is not authenticated.
//	500 InternalServerError – internal service error.
func (h *Handlers) AddCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	reqObj := dto.AddCardDTO{}
	err = easyjson.Unmarshal(body, &reqObj)
	if err != nil {
		h.logger.Error(ErrBadRequest.Error(), zap.Error(err))
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	keyUUID, err := h.keychainService.AddCard(r.Context(), userId, reqObj)
	if err != nil {
		var ve *apperr.ValidationError
		if errors.As(err, &ve) {
			h.PublicError(w, http.StatusBadRequest, err)
			return
		}
		h.InternalError(w, err)
		return
	}

	respObj := dto.AddSuccessResponse{
		UUID: keyUUID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// AddText adds a text entry for the user.
//
// Body (JSON):
//
//	{
//	  "title": "string",
//	  "text": "string"
//	}
//
// Status codes:
//
//	201 Created – the text entry was successfully added.
//	400 BadRequest – invalid JSON or validation error.
//	401 Unauthorized – user is not authenticated.
//	500 InternalServerError – internal service error.
func (h *Handlers) AddText(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	reqObj := dto.AddTextDTO{}
	err = easyjson.Unmarshal(body, &reqObj)
	if err != nil {
		h.logger.Error(ErrBadRequest.Error(), zap.Error(err))
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	keyUUID, err := h.keychainService.AddText(r.Context(), userId, reqObj)
	if err != nil {
		var ve *apperr.ValidationError
		if errors.As(err, &ve) {
			h.PublicError(w, http.StatusBadRequest, err)
			return
		}
		h.InternalError(w, err)
		return
	}

	respObj := dto.AddSuccessResponse{
		UUID: keyUUID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// AddFile adds a file for the user.
//
// Body (multipart/form-data):
//
//	file – file content
//	title – file title
//	note – optional note
//
// Constraints:
//
//	Maximum file size – 10MB.
//
// Status codes:
//
//	201 Created – the file was successfully added.
//	400 BadRequest – file not found in the request.
//	401 Unauthorized – user is not authenticated.
//	413 RequestEntityTooLarge – file size exceeds the limit.
//	500 InternalServerError – internal service error.
func (h *Handlers) AddFile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId, ok := middleware.GetUserIDFromCtx(ctx)
	if !ok {
		h.PublicError(w, http.StatusUnauthorized, ErrUnauthorizedError)
		return
	}

	defer r.Body.Close()

	const maxUploadSize = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		h.PublicError(w, http.StatusRequestEntityTooLarge, ErrPayloadFileLimit)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		h.PublicError(w, http.StatusBadRequest, ErrPayloadFileNotFound)
		return
	}
	defer file.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, file); err != nil {
		h.InternalError(w, err)
		return
	}
	raw := buf.Bytes()

	title := r.FormValue("title")
	note := r.FormValue("note")

	reqObj := dto.AddFileDTO{
		Title: title,
		File:  raw,
		Note:  note,
	}

	keyUUID, err := h.keychainService.AddFile(r.Context(), userId, reqObj)
	if err != nil {
		var ve *apperr.ValidationError
		if errors.As(err, &ve) {
			h.PublicError(w, http.StatusBadRequest, err)
			return
		}
		h.InternalError(w, err)
		return
	}

	respObj := dto.AddSuccessResponse{
		UUID: keyUUID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}
