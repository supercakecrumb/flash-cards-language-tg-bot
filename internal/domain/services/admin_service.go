package services

import (
	"log/slog"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
)

// AdminService handles admin operations
type AdminService interface {
	IsAdmin(telegramID int64) bool
	GetAdminIDs() []int64
}

type adminService struct {
	adminIDs []int64
	userRepo repository.UserRepository
	logger   *slog.Logger
}

// NewAdminService creates a new admin service
func NewAdminService(adminIDs []int64, userRepo repository.UserRepository, logger *slog.Logger) AdminService {
	return &adminService{
		adminIDs: adminIDs,
		userRepo: userRepo,
		logger:   logger,
	}
}

// IsAdmin checks if a user is an admin
func (s *adminService) IsAdmin(telegramID int64) bool {
	// Check if the Telegram ID is in the admin IDs list
	for _, id := range s.adminIDs {
		if id == telegramID {
			return true
		}
	}

	// Check if the user is marked as admin in the database
	return s.userRepo.IsAdmin(telegramID)
}

// GetAdminIDs returns the list of admin Telegram IDs
func (s *adminService) GetAdminIDs() []int64 {
	return s.adminIDs
}
