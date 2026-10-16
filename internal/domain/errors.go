package domain

import "errors"

var (
	ErrEmailTaken         = errors.New("email already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked      = errors.New("account is locked")
	ErrUserDeleted        = errors.New("user is deleted")
)
