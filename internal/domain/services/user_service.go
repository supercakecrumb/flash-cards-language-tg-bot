package services

import (
	"log/slog"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
)

// UserService handles user-related operations
type UserService interface {
	GetByTelegramID(telegramID int64) (*models.User, error)
	CreateUser(telegramID int64, username, firstName, lastName string) (*models.User, error)
	UpdateUser(user *models.User) error
}

type userService struct {
	repo   repository.UserRepository
	logger *slog.Logger
}

// NewUserService creates a new user service
func NewUserService(repo repository.UserRepository, logger *slog.Logger) UserService {
	return &userService{
		repo:   repo,
		logger: logger,
	}
}

// GetByTelegramID retrieves a user by their Telegram ID
func (s *userService) GetByTelegramID(telegramID int64) (*models.User, error) {
	s.logger.Debug("Getting user by Telegram ID", "telegram_id", telegramID)
	return s.repo.GetByTelegramID(telegramID)
}

// CreateUser creates a new user
func (s *userService) CreateUser(telegramID int64, username, firstName, lastName string) (*models.User, error) {
	s.logger.Info("Creating new user", "telegram_id", telegramID, "username", username)

	user := models.NewUser(telegramID, username, firstName, lastName)
	err := s.repo.Create(user)
	if err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, err
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *userService) UpdateUser(user *models.User) error {
	s.logger.Debug("Updating user", "user_id", user.ID, "telegram_id", user.TelegramID)
	return s.repo.Update(user)
}
