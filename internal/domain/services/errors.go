package services

import (
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// Common errors
var (
	ErrNotFound         = models.ErrNotFound
	ErrAlreadyExists    = models.ErrAlreadyExists
	ErrInvalidInput     = models.ErrInvalidInput
	ErrUnauthorized     = models.ErrUnauthorized
	ErrInternalError    = models.ErrInternalError
	ErrDatabaseError    = models.ErrDatabaseError
	ErrExternalAPIError = models.ErrExternalAPIError
	ErrInvalidParameter = models.ErrInvalidParameter
)
