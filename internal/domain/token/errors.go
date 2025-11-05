package token

import (
	"errors"
	"github.com/thxhix/passKeeper/internal/apperr"
)

var (
	ErrTokenAlreadyRotatedOrExpired = errors.New(`token already rotated or expired`)
	ErrTokenDoesntExistsByJTI       = errors.New(`token doesn't exists by JTI`)

	ErrMissingBearerToken = errors.New(`missing bearer token`)
	ErrInvalidAuthToken   = errors.New(`invalid or expired access token`)

	ErrEmptyRefreshCredentials = apperr.NewValidationError("refresh token isnt provided")
)
