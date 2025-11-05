package user

// ValidateLogin checks whether the login meets length requirements.
//
// Login must be more than 3 characters and less than or equal to 64 characters.
// Returns ErrLoginTooShort or ErrLoginTooLong if invalid.
func ValidateLogin(l string) error {
	if len(l) <= 3 {
		return ErrLoginTooShort
	}
	if len(l) > 64 {
		return ErrLoginTooLong
	}
	return nil
}

// ValidatePassword checks whether the password meets length requirements.
//
// Password must be at least 8 characters long.
// Returns ErrPasswordTooShort if invalid.
func ValidatePassword(p string) error {
	if len(p) < 8 {
		return ErrPasswordTooShort
	}
	return nil
}
