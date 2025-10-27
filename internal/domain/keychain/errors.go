package keychain

import (
	"github.com/thxhix/passKeeper/internal/apperr"
)

var (
	ErrTitleEmpty = apperr.NewValidationError("title cannot be empty")
	ErrTitleLong  = apperr.NewValidationError("title cannot be greater than 128 characters")

	ErrCredentialEmptyLogin = apperr.NewValidationError("login field is required")

	ErrCardNumberInvalid = apperr.NewValidationError("invalid card number")
	ErrCardCVVInvalid    = apperr.NewValidationError("invalid CVV, should be 3 chars")

	ErrEmptyTextProvided = apperr.NewValidationError("text field is required")
)
