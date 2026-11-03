package user

import (
	"errors"
	"net/mail"
)

var (
	ErrInvalidUsername = errors.New("invalid username")
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidEmail    = errors.New("invalid email")
)

func isAlphanumericASCII(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		isDigit := c >= '0' && c <= '9'
		isUpper := c >= 'A' && c <= 'Z'
		isLower := c >= 'a' && c <= 'z'
		if !isDigit && !isUpper && !isLower {
			return false
		}
	}
	return true
}

func ValidateUsername(username string) error {
	if username == "" ||
		len(username) < 3 ||
		len(username) > 64 {
		return ErrInvalidUsername
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" ||
		len(password) < 8 ||
		len(password) > 64 ||
		!isAlphanumericASCII(password) {
		return ErrInvalidPassword
	}
	return nil
}

func ValidateEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return ErrInvalidEmail
	}
	return nil
}
