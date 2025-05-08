# Domain Services

This document describes the domain services for the Flash Cards Language Telegram Bot. These services implement the business logic of the application.

## User Service (internal/domain/services/user_service.go)

The UserService manages user-related operations.

```go
package services

import (
	"log/slog"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
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
```

## Flash Card Service (internal/domain/services/flashcard_service.go)

The FlashCardService manages flash card operations and interacts with the dictionary API.

```go
package services

import (
	"log/slog"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/username/flash-cards-language-tg-bot/internal/infrastructure/dictionary"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
)

// Definition represents a word definition from the dictionary
type Definition struct {
	ID          string
	Text        string
	PartOfSpeech string
	Examples    []Example
}

// Example represents a usage example for a word
type Example struct {
	ID   string
	Text string
}

// FlashCardService handles flash card operations
type FlashCardService interface {
	GetDefinitions(word string) ([]Definition, error)
	GetDefinition(definitionID string) (*Definition, error)
	GetExamples(word, definitionID string) ([]Example, error)
	GetExample(exampleID string) (*Example, error)
	CreateFlashCard(card *models.FlashCard) error
	GetFlashCard(cardID int) (*models.FlashCard, error)
	GetFlashCardsByBank(bankID int) ([]models.FlashCard, error)
	UpdateFlashCard(card *models.FlashCard) error
	DeleteFlashCard(cardID int) error
}

type flashCardService struct {
	repo           repository.FlashCardRepository
	dictService    dictionary.DictionaryService
	logger         *slog.Logger
	definitionCache map[string]Definition
	exampleCache    map[string]Example
}

// NewFlashCardService creates a new flash card service
func NewFlashCardService(repo repository.FlashCardRepository, dictService dictionary.DictionaryService, logger *slog.Logger) FlashCardService {
	return &flashCardService{
		repo:           repo,
		dictService:    dictService,
		logger:         logger,
		definitionCache: make(map[string]Definition),
		exampleCache:    make(map[string]Example),
	}
}

// GetDefinitions retrieves definitions for a word from the dictionary API
func (s *flashCardService) GetDefinitions(word string) ([]Definition, error) {
	s.logger.Debug("Getting definitions for word", "word", word)
	
	dictDefinitions, err := s.dictService.GetDefinitions(word)
	if err != nil {
		s.logger.Error("Failed to get definitions from dictionary API", "error", err)
		return nil, err
	}
	
	var definitions []Definition
	for _, d := range dictDefinitions {
		def := Definition{
			ID:          d.ID,
			Text:        d.Text,
			PartOfSpeech: d.PartOfSpeech,
		}
		
		var examples []Example
		for _, e := range d.Examples {
			ex := Example{
				ID:   e.ID,
				Text: e.Text,
			}
			examples = append(examples, ex)
			s.exampleCache[ex.ID] = ex
		}
		
		def.Examples = examples
		definitions = append(definitions, def)
		s.definitionCache[def.ID] = def
	}
	
	return definitions, nil
}

// GetDefinition retrieves a cached definition by ID
func (s *flashCardService) GetDefinition(definitionID string) (*Definition, error) {
	def, ok := s.definitionCache[definitionID]
	if !ok {
		s.logger.Error("Definition not found in cache", "definition_id", definitionID)
		return nil, ErrNotFound
	}
	return &def, nil
}

// GetExamples retrieves examples for a word and definition
func (s *flashCardService) GetExamples(word, definitionID string) ([]Example, error) {
	def, err := s.GetDefinition(definitionID)
	if err != nil {
		return nil, err
	}
	return def.Examples, nil
}

// GetExample retrieves a cached example by ID
func (s *flashCardService) GetExample(exampleID string) (*Example, error) {
	ex, ok := s.exampleCache[exampleID]
	if !ok {
		s.logger.Error("Example not found in cache", "example_id", exampleID)
		return nil, ErrNotFound
	}
	return &ex, nil
}

// CreateFlashCard creates a new flash card
func (s *flashCardService) CreateFlashCard(card *models.FlashCard) error {
	s.logger.Info("Creating flash card", "word", card.Word, "bank_id", card.CardBankID)
	return s.repo.Create(card)
}

// GetFlashCard retrieves a flash card by ID
func (s *flashCardService) GetFlashCard(cardID int) (*models.FlashCard, error) {
	s.logger.Debug("Getting flash card", "card_id", cardID)
	return s.repo.GetByID(cardID)
}

// GetFlashCardsByBank retrieves all flash cards in a bank
func (s *flashCardService) GetFlashCardsByBank(bankID int) ([]models.FlashCard, error) {
	s.logger.Debug("Getting flash cards for bank", "bank_id", bankID)
	return s.repo.GetCardsForBank(bankID)
}

// UpdateFlashCard updates an existing flash card
func (s *flashCardService) UpdateFlashCard(card *models.FlashCard) error {
	s.logger.Debug("Updating flash card", "card_id", card.ID)
	return s.repo.Update(card)
}

// DeleteFlashCard deletes a flash card
func (s *flashCardService) DeleteFlashCard(cardID int) error {
	s.logger.Info("Deleting flash card", "card_id", cardID)
	return s.repo.Delete(cardID)
}
```

## Card Bank Service (internal/domain/services/cardbank_service.go)

The CardBankService manages card banks, memberships, and group chat links.

```go
package services

import (
	"log/slog"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
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
```

## Spaced Repetition Service (internal/domain/services/spaced_repetition_service.go)

The SpacedRepetitionService implements the spaced repetition algorithm for reviewing flash cards.

```go
package services

import (
	"log/slog"
	"time"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
	"github.com/username/flash-cards-language-tg-bot/pkg/spaced_repetition"
)

// SpacedRepetitionService handles spaced repetition operations
type SpacedRepetitionService interface {
	GetDueCards(userID, bankID int, limit int) ([]models.FlashCard, error)
	ProcessReview(userID, cardID int, quality int) error
	GetReviewStats(userID int) (int, int, error) // total cards, due cards
}

type spacedRepetitionService struct {
	reviewRepo    repository.ReviewRepository
	flashcardRepo repository.FlashCardRepository
	algorithm     spaced_repetition.Algorithm
	logger        *slog.Logger
}

// NewSpacedRepetitionService creates a new spaced repetition service
func NewSpacedRepetitionService(
	reviewRepo repository.ReviewRepository,
	flashcardRepo repository.FlashCardRepository,
	logger *slog.Logger,
) SpacedRepetitionService {
	// Default to SM-2 algorithm
	algorithm := spaced_repetition.NewSM2Algorithm()
	
	return &spacedRepetitionService{
		reviewRepo:    reviewRepo,
		flashcardRepo: flashcardRepo,
		algorithm:     algorithm,
		logger:        logger,
	}
}

// GetDueCards retrieves cards that are due for review
func (s *spacedRepetitionService) GetDueCards(userID, bankID int, limit int) ([]models.FlashCard, error) {
	s.logger.Debug("Getting due cards", "user_id", userID, "bank_id", bankID, "limit", limit)
	
	// Get due reviews
	reviews, err := s.reviewRepo.GetDueReviews(userID, bankID, time.Now(), limit)
	if err != nil {
		s.logger.Error("Failed to get due reviews", "error", err)
		return nil, err
	}
	
	// If there are not enough due reviews, get new cards
	if len(reviews) < limit {
		newCardsLimit := limit - len(reviews)
		newCards, err := s.getNewCards(userID, bankID, newCardsLimit)
		if err != nil {
			s.logger.Error("Failed to get new cards", "error", err)
			return nil, err
		}
		
		// Create initial reviews for new cards
		for _, card := range newCards {
			review := models.NewReview(userID, card.ID)
			err := s.reviewRepo.Create(review)
			if err != nil {
				s.logger.Error("Failed to create review for new card", "error", err)
				continue
			}
		}
		
		// Get cards for due reviews
		var dueCards []models.FlashCard
		for _, review := range reviews {
			card, err := s.flashcardRepo.GetByID(review.FlashCardID)
			if err != nil {
				s.logger.Error("Failed to get card for review", "error", err)
				continue
			}
			dueCards = append(dueCards, *card)
		}
		
		// Combine due cards and new cards
		return append(dueCards, newCards...), nil
	}
	
	// Get cards for due reviews
	var dueCards []models.FlashCard
	for _, review := range reviews {
		card, err := s.flashcardRepo.GetByID(review.FlashCardID)
		if err != nil {
			s.logger.Error("Failed to get card for review", "error", err)
			continue
		}
		dueCards = append(dueCards, *card)
	}
	
	return dueCards, nil
}

// getNewCards retrieves cards that the user hasn't reviewed yet
func (s *spacedRepetitionService) getNewCards(userID, bankID, limit int) ([]models.FlashCard, error) {
	return s.flashcardRepo.GetNewCards(userID, bankID, limit)
}

// ProcessReview processes a card review and updates the review schedule
func (s *spacedRepetitionService) ProcessReview(userID, cardID int, quality int) error {
	s.logger.Debug("Processing review", "user_id", userID, "card_id", cardID, "quality", quality)
	
	// Get existing review or create a new one
	review, err := s.reviewRepo.GetByUserAndCard(userID, cardID)
	if err != nil {
		// Create new review if it doesn't exist
		review = models.NewReview(userID, cardID)
		err = s.reviewRepo.Create(review)
		if err != nil {
			s.logger.Error("Failed to create review", "error", err)
			return err
		}
	}
	
	// Calculate next review date using the algorithm
	dueDate, interval, easeFactor := s.algorithm.CalculateNextReview(review, quality)
	
	// Update review
	review.DueDate = dueDate
	review.Interval = interval
	review.EaseFactor = easeFactor
	review.Repetitions++
	review.LastReviewed = time.Now()
	
	// Save updated review
	err = s.reviewRepo.Update(review)
	if err != nil {
		s.logger.Error("Failed to update review", "error", err)
		return err
	}
	
	return nil
}

// GetReviewStats retrieves review statistics for a user
func (s *spacedRepetitionService) GetReviewStats(userID int) (int, int, error) {
	s.logger.Debug("Getting review stats", "user_id", userID)
	
	totalCards, err := s.reviewRepo.CountTotalReviews(userID)
	if err != nil {
		s.logger.Error("Failed to count total reviews", "error", err)
		return 0, 0, err
	}
	
	dueCards, err := s.reviewRepo.CountDueReviews(userID, time.Now())
	if err != nil {
		s.logger.Error("Failed to count due reviews", "error", err)
		return 0, 0, err
	}
	
	return totalCards, dueCards, nil
}
```

## Statistics Service (internal/domain/services/statistics_service.go)

The StatisticsService tracks and reports user learning statistics.

```go
package services

import (
	"log/slog"
	"time"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
)

// StatisticsService handles statistics operations
type StatisticsService interface {
	GetUserStatistics(userID int) ([]models.Statistics, error)
	GetBankStatistics(userID, bankID int) (*models.Statistics, error)
	IncrementReviewed(userID, bankID int) error
	IncrementLearned(userID, bankID int) error
	UpdateStreak(userID int) error
}

type statisticsService struct {
	repo   repository.StatisticsRepository
	logger *slog.Logger
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(repo repository.StatisticsRepository, logger *slog.Logger) StatisticsService {
	return &statisticsService{
		repo:   repo,
		logger: logger,
	}
}

// GetUserStatistics retrieves all statistics for a user
func (s *statisticsService) GetUserStatistics(userID int) ([]models.Statistics, error) {
	s.logger.Debug("Getting user statistics", "user_id", userID)
	return s.repo.GetByUserID(userID)
}

// GetBankStatistics retrieves statistics for a specific bank
func (s *statisticsService) GetBankStatistics(userID, bankID int) (*models.Statistics, error) {
	s.logger.Debug("Getting bank statistics", "user_id", userID, "bank_id", bankID)
	
	stats, err := s.repo.GetByUserAndBank(userID, bankID)
	if err != nil {
		// Create new statistics if they don't exist
		stats = models.NewStatistics(userID, bankID)
		err = s.repo.Create(stats)
		if err != nil {
			s.logger.Error("Failed to create statistics", "error", err)
			return nil, err
		}
	}
	
	return stats, nil
}

// IncrementReviewed increments the cards reviewed count
func (s *statisticsService) IncrementReviewed(userID, bankID int) error {
	s.logger.Debug("Incrementing cards reviewed", "user_id", userID, "bank_id", bankID)
	
	stats, err := s.GetBankStatistics(userID, bankID)
	if err != nil {
		return err
	}
	
	stats.CardsReviewed++
	stats.UpdatedAt = time.Now()
	
	return s.repo.Update(stats)
}

// IncrementLearned increments the cards learned count
func (s *statisticsService) IncrementLearned(userID, bankID int) error {
	s.logger.Debug("Incrementing cards learned", "user_id", userID, "bank_id", bankID)
	
	stats, err := s.GetBankStatistics(userID, bankID)
	if err != nil {
		return err
	}
	
	stats.CardsLearned++
	stats.UpdatedAt = time.Now()
	
	return s.repo.Update(stats)
}

// UpdateStreak updates the user's streak days
func (s *statisticsService) UpdateStreak(userID int) error {
	s.logger.Debug("Updating streak", "user_id", userID)
	
	// Get all user statistics
	allStats, err := s.GetUserStatistics(userID)
	if err != nil {
		return err
	}
	
	// Check if user has reviewed today
	today := time.Now().Format("2006-01-02")
	lastReviewDate, err := s.repo.GetLastReviewDate(userID)
	if err != nil {
		s.logger.Error("Failed to get last review date", "error", err)
		return err
	}
	
	// Update streak for all banks
	for _, stats := range allStats {
		if lastReviewDate == today {
			// User has reviewed today, increment streak if not already done
			if stats.UpdatedAt.Format("2006-01-02") != today {
				stats.StreakDays++
				stats.UpdatedAt = time.Now()
				err = s.repo.Update(&stats)
				if err != nil {
					s.logger.Error("Failed to update streak", "error", err)
					return err
				}
			}
		} else {
			// User hasn't reviewed today, reset streak if last review was not yesterday
			yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			if lastReviewDate != yesterday {
				stats.StreakDays = 0
				stats.UpdatedAt = time.Now()
				err = s.repo.Update(&stats)
				if err != nil {
					s.logger.Error("Failed to reset streak", "error", err)
					return err
				}
			}
		}
	}
	
	return nil
}
```

## Settings Service (internal/domain/services/settings_service.go)

The SettingsService manages user settings.

```go
package services

import (
	"log/slog"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
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
```

## Admin Service (internal/domain/services/admin_service.go)

The AdminService handles admin-specific functionality.

```go
package services

import (
	"log/slog"

	"github.com/username/flash-cards-language-tg-bot/internal/repository"
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
```

These domain services implement the core business logic of the application, encapsulating the rules and operations for each domain entity.