package services

import (
	"log/slog"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
)

// SettingsService handles user settings operations
type SettingsService interface {
	GetUserSettings(userID int) (*models.Settings, error)
	UpdateUserSettings(userID int, settings models.SettingsData) error
	SetActiveCardBank(userID, bankID int) error
}

type settingsService struct {
	repo   repository.SettingsRepository
	logger *slog.Logger
}

// NewSettingsService creates a new settings service
func NewSettingsService(repo repository.SettingsRepository, logger *slog.Logger) SettingsService {
	return &settingsService{
		repo:   repo,
		logger: logger,
	}
}

// GetUserSettings retrieves settings for a user
func (s *settingsService) GetUserSettings(userID int) (*models.Settings, error) {
	s.logger.Debug("Getting user settings", "user_id", userID)

	settings, err := s.repo.GetByUserID(userID)
	if err != nil {
		// Create default settings if they don't exist
		settings = models.NewSettings(userID)
		err = s.repo.Create(settings)
		if err != nil {
			s.logger.Error("Failed to create settings", "error", err)
			return nil, err
		}
	}

	return settings, nil
}

// UpdateUserSettings updates a user's settings
func (s *settingsService) UpdateUserSettings(userID int, settingsData models.SettingsData) error {
	s.logger.Debug("Updating user settings", "user_id", userID)

	settings, err := s.GetUserSettings(userID)
	if err != nil {
		return err
	}

	settings.Settings = settingsData

	return s.repo.Update(settings)
}

// SetActiveCardBank sets the active card bank for a user
func (s *settingsService) SetActiveCardBank(userID, bankID int) error {
	s.logger.Debug("Setting active card bank", "user_id", userID, "bank_id", bankID)

	settings, err := s.GetUserSettings(userID)
	if err != nil {
		return err
	}

	settings.Settings.ActiveCardBankID = bankID

	return s.repo.Update(settings)
}
