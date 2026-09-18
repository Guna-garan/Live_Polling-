package auth

import (
	"errors"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

var (
	ErrEmailRequired    = errors.New("email is required")
	ErrEmailInvalid     = errors.New("email format is invalid")
	ErrPasswordRequired = errors.New("password is required")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordMismatch = errors.New("passwords do not match")
)

// normalizeEmail lowercases and trims an email so that
// "User@Example.com" and "user@example.com " are treated as the same
// account both for uniqueness and for lookups.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateSignup(req SignupRequest) (SignupRequest, error) {
	req.Email = normalizeEmail(req.Email)

	if req.Email == "" {
		return req, ErrEmailRequired
	}
	if !emailRegex.MatchString(req.Email) {
		return req, ErrEmailInvalid
	}
	if req.Password == "" {
		return req, ErrPasswordRequired
	}
	if len(req.Password) < 8 {
		return req, ErrPasswordTooShort
	}
	if req.Password != req.ConfirmPassword {
		return req, ErrPasswordMismatch
	}
	return req, nil
}

func validateLogin(req LoginRequest) (LoginRequest, error) {
	req.Email = normalizeEmail(req.Email)
	if req.Email == "" {
		return req, ErrEmailRequired
	}
	if req.Password == "" {
		return req, ErrPasswordRequired
	}
	return req, nil
}
