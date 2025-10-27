package apperr

type AuthError struct {
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}

func NewAuthError(m string) *AuthError {
	return &AuthError{
		Message: m,
	}
}
