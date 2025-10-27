package security

import (
	"errors"
	"github.com/thxhix/passKeeper/internal/apperr"
)

var (
	ErrAccessExpiredToken   = apperr.NewAuthError("expired access token")
	ErrAccessInvalidSubject = apperr.NewAuthError("invalid access subject")

	ErrRefreshExpiredToken  = apperr.NewAuthError("expired refresh token")
	ErrRefreshInvalidClaims = apperr.NewAuthError("invalid refresh claims")

	ErrUnexpectedSigningMethod = apperr.NewAuthError("unexpected signing method")

	ErrSecretTooShort  = errors.New("secret too short")
	ErrAEADWrongLength = errors.New("invalid AEAD length, expect 32 bytes")
)
