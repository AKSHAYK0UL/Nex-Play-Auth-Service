package passwordstrengthchecker

import (
	"errors"
	"unicode"
)

// CheckPasswordStrength validates password rules:
//
//	length >= 8
//	at least 1 uppercase letter
//	at least 1 lowercase letter
//	at least 1 special character
func CheckPasswordStrength(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case !unicode.IsLetter(ch) && !unicode.IsNumber(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}

	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}
