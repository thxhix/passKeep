package handlers

import (
	"errors"
	"github.com/mailru/easyjson"
	"github.com/thxhix/passKeeper/internal/apperr"
	"github.com/thxhix/passKeeper/internal/domain/token"
	"github.com/thxhix/passKeeper/internal/domain/user"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// Register handles user registration.
//
// Body (JSON):
//
//	{
//	  "login": "string",
//	  "password": "string"
//	}
//
// Status codes:
//
//	201 Created – the user was registered successfully, tokens returned.
//	400 BadRequest – invalid JSON or validation error.
//	409 Conflict – login already exists.
//	500 InternalServerError – internal service error.
func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	reqObj := dto.RegisterRequest{}
	err = easyjson.Unmarshal(body, &reqObj)
	if err != nil {
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	userId, accessToken, refreshToken, err := h.authService.Register(r.Context(), reqObj.Login, reqObj.Password)

	if err != nil {
		if errors.Is(err, user.ErrDuplicateLogin) {
			h.PublicError(w, http.StatusConflict, err)
			return
		}
		var ae *apperr.ValidationError
		if errors.As(err, &ae) {
			h.PublicError(w, http.StatusBadRequest, err)
			return
		}

		h.InternalError(w, err)
		return
	}

	respObj := dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusCreated)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// Login handles user authentication.
//
// Body (JSON):
//
//	{
//	  "login": "string",
//	  "password": "string"
//	}
//
// Status codes:
//
//	200 OK – login successful, tokens returned.
//	400 BadRequest – invalid JSON or validation error.
//	401 Unauthorized – authentication failed.
//	500 InternalServerError – internal service error.
func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	reqObj := dto.LoginRequest{}
	err = easyjson.Unmarshal(body, &reqObj)
	if err != nil {
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	userId, accessToken, refreshToken, err := h.authService.Login(r.Context(), reqObj.Login, reqObj.Password)

	if err != nil {
		var ae *apperr.AuthError
		if errors.As(err, &ae) {
			h.PublicError(w, http.StatusUnauthorized, err)
			return
		}

		var ve *apperr.ValidationError
		if errors.As(err, &ve) {
			h.PublicError(w, http.StatusBadRequest, err)
			return
		}

		h.InternalError(w, err)
		return
	}

	respObj := dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UserID:       userId,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}

// Refresh handles access token renewal using a refresh token.
//
// Body (JSON):
//
//	{
//	  "refresh_token": "string"
//	}
//
// Status codes:
//
//	200 OK – token refreshed successfully, new tokens returned.
//	400 BadRequest – invalid JSON or invalid refresh token.
//	401 Unauthorized – refresh token is invalid or expired.
//	500 InternalServerError – internal service error.
func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.InternalError(w, err)
		return
	}

	reqObj := dto.RefreshRequest{}
	err = easyjson.Unmarshal(body, &reqObj)
	if err != nil {
		h.PublicError(w, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if err := token.ValidateRefreshToken(reqObj.RefreshToken); err != nil {
		h.PublicError(w, http.StatusBadRequest, err)
		return
	}

	accessToken, refreshToken, err := h.authService.Refresh(r.Context(), reqObj.RefreshToken)
	if err != nil {
		var ae *apperr.AuthError
		if errors.As(err, &ae) {
			h.PublicError(w, http.StatusUnauthorized, err)
			return
		}
		h.InternalError(w, err)
		return
	}

	respObj := dto.RefreshedTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)

	if _, err := easyjson.MarshalToWriter(&respObj, w); err != nil {
		h.logger.Error(ErrCantWriteResponseBody.Error(), zap.Error(err))
		return
	}
}
