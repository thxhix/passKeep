package user

import (
	"errors"
	"github.com/thxhix/passKeeper/internal/apperr"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrDuplicateLogin = errors.New("provided login already exists")

	ErrLoginTooShort    = apperr.NewValidationError("login is too short, minimum length – 3")
	ErrLoginTooLong     = apperr.NewValidationError("login is too short, maximum length – 64")
	ErrPasswordTooShort = apperr.NewValidationError("password too short, minimum length – 8")

	ErrInvalidCredentials        = errors.New("wrong login or password")
	ErrInvalidRefreshCredentials = errors.New("unregistered refresh token provided")
)
