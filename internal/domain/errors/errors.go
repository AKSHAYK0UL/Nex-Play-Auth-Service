package errors

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidOTP         = errors.New("invalid or expired code")
	ErrOTPExpired         = errors.New("code has expired")
	ErrOTPAlreadyUsed     = errors.New("code was already used")
	ErrInvalidToken       = errors.New("invalid token")
	ErrWeakPassword       = errors.New("password does not meet requirements")
	ErrAccountNotVerified = errors.New("account not verified")
)
