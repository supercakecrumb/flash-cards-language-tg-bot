package repository

import (
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// Common errors
var (
	ErrNotFound      = models.ErrNotFound
	ErrAlreadyExists = models.ErrAlreadyExists
	ErrInvalidInput  = models.ErrInvalidInput
	ErrDatabaseError = models.ErrDatabaseError
)
