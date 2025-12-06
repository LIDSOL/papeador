package security

import (
	"net/mail"
	"regexp"
	"strings"
	"errors"
)

var (
	lower 	= regexp.MustCompile(`[a-z]`)
	upper 	= regexp.MustCompile(`[A-Z]`)
	digit 	= regexp.MustCompile(`\d`)
	special = regexp.MustCompile(`[!@#\$%\^&\*\(\)\-\_\=\+\[\]\{\};:'",.<>\/\?\\\|` + "`~]")
	space 	= regexp.MustCompile(`\s`)
	uname 	= regexp.MustCompile(`^[A-Za-z0-9_-]{3,20}$`)

	ErrInvalidUsername 	= errors.New("invalid username")
	ErrInvalidPassword 	= errors.New("invalid password")
	ErrInvalidEmail 	= errors.New("invalid email")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

func IsValidUsername(username string) error {
	
	if !uname.MatchString(username) {
		return ErrInvalidUsername
	}

	return nil
}

func IsValidPassword(password string) error {
	valid_length := len(password) >= 12 && len(password) <= 64

	if !valid_length {
		return ErrInvalidPassword
	}

	if !(lower.MatchString(password) && 
		 upper.MatchString(password) && 
		 digit.MatchString(password) && 
		 special.MatchString(password) && 
		 !space.MatchString(password)) {
			 return ErrInvalidPassword
	}

	return nil
}

func ValidateEmail(email string) (string, error) {
	email = strings.TrimSpace(email)
	_, err := mail.ParseAddress(email)

	if err != nil {
		return "", ErrInvalidEmail
	}

	return strings.ToLower(email), nil
}
