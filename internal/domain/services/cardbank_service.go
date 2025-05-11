package services

import (
	"log/slog"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
)

// CardBankService handles card bank operations
type CardBankService interface {
	CreateCardBank(name, description string, ownerID int, isPublic bool) (*models.CardBank, error)
	GetCardBank(bankID int) (*models.CardBank, error)
	GetUserCardBanks(userID int) ([]models.CardBank, error)
	UpdateCardBank(bank *models.CardBank) error
	DeleteCardBank(bankID int) error

	// Membership operations
	AddUserToBank(userID, bankID int, role string) error
	RemoveUserFromBank(userID, bankID int) error
	UserHasAccess(userID, bankID int) (bool, error)

	// Group chat operations
	LinkGroupChat(telegramChatID int64, title string, bankID int) error
	GetGroupChat(telegramChatID int64) (*models.GroupChat, error)
	UnlinkGroupChat(telegramChatID int64) error
}

type cardBankService struct {
	repo   repository.CardBankRepository
	logger *slog.Logger
}

// NewCardBankService creates a new card bank service
func NewCardBankService(repo repository.CardBankRepository, logger *slog.Logger) CardBankService {
	return &cardBankService{
		repo:   repo,
		logger: logger,
	}
}

// CreateCardBank creates a new card bank
func (s *cardBankService) CreateCardBank(name, description string, ownerID int, isPublic bool) (*models.CardBank, error) {
	s.logger.Info("Creating card bank", "name", name, "owner_id", ownerID)

	bank := models.NewCardBank(name, description, ownerID, isPublic)
	err := s.repo.Create(bank)
	if err != nil {
		s.logger.Error("Failed to create card bank", "error", err)
		return nil, err
	}

	// Add owner as a member with "owner" role
	membership := models.NewBankMembership(ownerID, bank.ID, "owner")
	err = s.repo.CreateMembership(membership)
	if err != nil {
		s.logger.Error("Failed to create bank membership for owner", "error", err)
		return nil, err
	}

	return bank, nil
}

// GetCardBank retrieves a card bank by ID
func (s *cardBankService) GetCardBank(bankID int) (*models.CardBank, error) {
	s.logger.Debug("Getting card bank", "bank_id", bankID)
	return s.repo.GetByID(bankID)
}

// GetUserCardBanks retrieves all card banks a user has access to
func (s *cardBankService) GetUserCardBanks(userID int) ([]models.CardBank, error) {
	s.logger.Debug("Getting card banks for user", "user_id", userID)
	return s.repo.GetBanksForUser(userID)
}

// UpdateCardBank updates an existing card bank
func (s *cardBankService) UpdateCardBank(bank *models.CardBank) error {
	s.logger.Debug("Updating card bank", "bank_id", bank.ID)
	return s.repo.Update(bank)
}

// DeleteCardBank deletes a card bank
func (s *cardBankService) DeleteCardBank(bankID int) error {
	s.logger.Info("Deleting card bank", "bank_id", bankID)
	return s.repo.Delete(bankID)
}

// AddUserToBank adds a user to a card bank with the specified role
func (s *cardBankService) AddUserToBank(userID, bankID int, role string) error {
	s.logger.Info("Adding user to bank", "user_id", userID, "bank_id", bankID, "role", role)

	membership := models.NewBankMembership(userID, bankID, role)
	return s.repo.CreateMembership(membership)
}

// RemoveUserFromBank removes a user from a card bank
func (s *cardBankService) RemoveUserFromBank(userID, bankID int) error {
	s.logger.Info("Removing user from bank", "user_id", userID, "bank_id", bankID)
	return s.repo.DeleteMembership(userID, bankID)
}

// UserHasAccess checks if a user has access to a card bank
func (s *cardBankService) UserHasAccess(userID, bankID int) (bool, error) {
	s.logger.Debug("Checking if user has access to bank", "user_id", userID, "bank_id", bankID)
	return s.repo.UserHasAccess(userID, bankID)
}

// LinkGroupChat links a Telegram group chat to a card bank
func (s *cardBankService) LinkGroupChat(telegramChatID int64, title string, bankID int) error {
	s.logger.Info("Linking group chat to bank", "chat_id", telegramChatID, "bank_id", bankID)

	groupChat := models.NewGroupChat(telegramChatID, title, bankID)
	return s.repo.CreateGroupChat(groupChat)
}

// GetGroupChat retrieves a group chat by Telegram chat ID
func (s *cardBankService) GetGroupChat(telegramChatID int64) (*models.GroupChat, error) {
	s.logger.Debug("Getting group chat", "chat_id", telegramChatID)
	return s.repo.GetGroupChatByTelegramID(telegramChatID)
}

// UnlinkGroupChat unlinks a Telegram group chat from its card bank
func (s *cardBankService) UnlinkGroupChat(telegramChatID int64) error {
	s.logger.Info("Unlinking group chat", "chat_id", telegramChatID)
	return s.repo.DeleteGroupChat(telegramChatID)
}
