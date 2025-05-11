package models

import (
	"errors"
)

// Common errors
var (
	ErrNotFound         = errors.New("not found")
	ErrAlreadyExists    = errors.New("already exists")
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInternalError    = errors.New("internal error")
	ErrDatabaseError    = errors.New("database error")
	ErrExternalAPIError = errors.New("external API error")
	ErrInvalidParameter = errors.New("invalid parameter")
)
