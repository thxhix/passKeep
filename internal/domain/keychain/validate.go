package keychain

import (
	"strings"
	"unicode"
)

// ValidateTitle checks if the provided title is valid.
//
// It trims whitespace and ensures the title is not empty and does not exceed 128 characters.
// Returns ErrTitleEmpty if empty, ErrTitleLong if too long, or nil if valid.
func ValidateTitle(title string) error {
	t := strings.TrimSpace(title)
	if t == "" {
		return ErrTitleEmpty
	}
	if len(t) > 128 {
		return ErrTitleLong
	}
	return nil
}

// ValidateCredential checks if the provided login for a credential key is valid.
//
// It trims whitespace and ensures the login is not empty.
// Returns ErrCredentialEmptyLogin if empty, or nil if valid.
func ValidateCredential(login string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return ErrCredentialEmptyLogin
	}

	return nil
}

// ValidateCard checks if the provided card number and CVV are valid.
//
// It validates the card number using the Luhn algorithm and checks that CVV is exactly 3 digits.
// Returns ErrCardNumberInvalid or ErrCardCVVInvalid if invalid, or nil if valid.
func ValidateCard(num string, cvv string) error {
	num = strings.TrimSpace(num)
	if !isValidLuhn(num) {
		return ErrCardNumberInvalid
	}

	cvv = strings.TrimSpace(cvv)
	if len(cvv) != 3 {
		return ErrCardCVVInvalid
	}

	return nil
}

// ValidateText checks if the provided text is valid.
//
// It trims whitespace and ensures the text is not empty.
// Returns ErrEmptyTextProvided if empty, or nil if valid.
func ValidateText(text string) error {
	text = strings.TrimSpace(text)
	if text == "" {
		return ErrEmptyTextProvided
	}

	return nil
}

// isValidLuhn validates a card number using the Luhn algorithm.
//
// Returns true if the number passes the Luhn checksum, false otherwise.
func isValidLuhn(number string) bool {
	number = strings.ReplaceAll(number, " ", "")

	var sum int
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		r := rune(number[i])
		if !unicode.IsDigit(r) {
			return false
		}
		digit := int(r - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
