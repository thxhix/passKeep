package handlers

import (
	"github.com/mailru/easyjson"
	"github.com/thxhix/passKeeper/internal/services"
	"github.com/thxhix/passKeeper/internal/transport/http/dto"
	"go.uber.org/zap"
	"net/http"
)

// Handlers is a collection of HTTP handlers for authentication and key management.
//
// It holds references to the logger and underlying services for authentication
// and keychain operations.
type Handlers struct {
	logger          *zap.Logger
	authService     services.IAuthService
	keychainService services.IKeychainService
}

// NewHandlers creates a new instance of Handlers.
//
// Parameters:
//
//	logger *zap.Logger – the logger instance to be used.
//	authService services.IAuthService – the authentication service.
//	keychainService services.IKeychainService – the keychain service.
//
// Returns:
//
//	Handlers – a new Handlers instance.
func NewHandlers(logger *zap.Logger, authService services.IAuthService, keychainService services.IKeychainService) Handlers {
	return Handlers{
		logger:          logger,
		authService:     authService,
		keychainService: keychainService,
	}
}

// PublicError writes a public-facing error response with the given HTTP status code.
//
// Parameters:
//
//	w http.ResponseWriter – the HTTP response writer.
//	code int – the HTTP status code to return.
//	err error – the error message to include in the response.
//
// The method logs a warning with the error, sets the response headers to
// "application/json", and writes the error object to the response body.
func (h *Handlers) PublicError(w http.ResponseWriter, code int, err error) {
	h.logger.Warn(ErrInternalServerError.Error(), zap.Error(err))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	errObj := writeResponseError(code, err)

	if _, err := easyjson.MarshalToWriter(&errObj, w); err != nil {
		h.logger.Error("write error response failed", zap.Error(err))
	}
}

// InternalError writes a generic internal server error response.
//
// Parameters:
//
//	w http.ResponseWriter – the HTTP response writer.
//	err error – the internal error that occurred.
//
// The method logs the error, sets the response headers to "application/json",
// writes HTTP status 500, and returns a generic internal error message to the client.
func (h *Handlers) InternalError(w http.ResponseWriter, err error) {
	h.logger.Error(ErrInternalServerError.Error(), zap.Error(err))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)

	errObj := writeResponseError(http.StatusInternalServerError, ErrInternalPublicError)

	if _, err := easyjson.MarshalToWriter(&errObj, w); err != nil {
		h.logger.Error("write internal error response failed", zap.Error(err))
	}
}

// writeResponseError formats an error into a JSON-compatible ErrorResponse object.
//
// Parameters:
//
//	code int – the HTTP status code.
//	err error – the error to include in the response.
//
// Returns:
//
//	dto.ErrorResponse – the structured error object.
func writeResponseError(code int, err error) dto.ErrorResponse {
	return dto.ErrorResponse{
		Code:      code,
		ErrorText: err.Error(),
	}
}
