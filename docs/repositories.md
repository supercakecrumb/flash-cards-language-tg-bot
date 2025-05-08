# Repositories

This document describes the repository implementations for the Flash Cards Language Telegram Bot. Repositories handle data access to the database.

## User Repository (internal/repository/user_repository.go)

The UserRepository handles user data access.

```go
package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(userID int) (*models.User, error)
	GetByTelegramID(telegramID int64) (*models.User, error)
	Update(user *models.User) error
	Delete(userID int) error
	IsAdmin(telegramID int64) bool
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, created_at, updated_at, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsAdmin,
	).Scan(&user.ID)
	
	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(userID int) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, created_at, updated_at, is_admin
		FROM users
		WHERE id = $1
	`
	
	var user models.User
	err := r.db.Get(&user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// GetByTelegramID retrieves a user by Telegram ID
func (r *userRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, created_at, updated_at, is_admin
		FROM users
		WHERE telegram_id = $1
	`
	
	var user models.User
	err := r.db.Get(&user, query, telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, first_name = $2, last_name = $3, updated_at = $4, is_admin = $5
		WHERE id = $6
	`
	
	user.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.UpdatedAt,
		user.IsAdmin,
		user.ID,
	)
	
	return err
}

// Delete deletes a user
func (r *userRepository) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

// IsAdmin checks if a user is an admin
func (r *userRepository) IsAdmin(telegramID int64) bool {
	query := `SELECT is_admin FROM users WHERE telegram_id = $1`
	
	var isAdmin bool
	err := r.db.Get(&isAdmin, query, telegramID)
	if err != nil {
		return false
	}
	
	return isAdmin
}
```

## Flash Card Repository (internal/repository/flashcard_repository.go)

The FlashCardRepository handles flash card data access.

```go
package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// FlashCardRepository defines the interface for flash card data access
type FlashCardRepository interface {
	Create(card *models.FlashCard) error
	GetByID(cardID int) (*models.FlashCard, error)
	GetByWord(word string, bankID int) (*models.FlashCard, error)
	GetCardsForBank(bankID int) ([]models.FlashCard, error)
	GetNewCards(userID, bankID, limit int) ([]models.FlashCard, error)
	Update(card *models.FlashCard) error
	Delete(cardID int) error
}

// flashCardRepository implements the FlashCardRepository interface
type flashCardRepository struct {
	db *sqlx.DB
}

// NewFlashCardRepository creates a new flash card repository
func NewFlashCardRepository(db *sqlx.DB) FlashCardRepository {
	return &flashCardRepository{
		db: db,
	}
}

// Create creates a new flash card
func (r *flashCardRepository) Create(card *models.FlashCard) error {
	query := `
		INSERT INTO flash_cards (card_bank_id, word, definition, examples, image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		card.CardBankID,
		card.Word,
		card.Definition,
		card.Examples,
		card.ImageURL,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&card.ID)
	
	return err
}

// GetByID retrieves a flash card by ID
func (r *flashCardRepository) GetByID(cardID int) (*models.FlashCard, error) {
	query := `
		SELECT id, card_bank_id, word, definition, examples, image_url, created_at, updated_at
		FROM flash_cards
		WHERE id = $1
	`
	
	var card models.FlashCard
	err := r.db.Get(&card, query, cardID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &card, nil
}

// GetByWord retrieves a flash card by word and bank ID
func (r *flashCardRepository) GetByWord(word string, bankID int) (*models.FlashCard, error) {
	query := `
		SELECT id, card_bank_id, word, definition, examples, image_url, created_at, updated_at
		FROM flash_cards
		WHERE word = $1 AND card_bank_id = $2
	`
	
	var card models.FlashCard
	err := r.db.Get(&card, query, word, bankID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &card, nil
}

// GetCardsForBank retrieves all flash cards in a bank
func (r *flashCardRepository) GetCardsForBank(bankID int) ([]models.FlashCard, error) {
	query := `
		SELECT id, card_bank_id, word, definition, examples, image_url, created_at, updated_at
		FROM flash_cards
		WHERE card_bank_id = $1
		ORDER BY created_at DESC
	`
	
	var cards []models.FlashCard
	err := r.db.Select(&cards, query, bankID)
	if err != nil {
		return nil, err
	}
	
	return cards, nil
}

// GetNewCards retrieves cards that the user hasn't reviewed yet
func (r *flashCardRepository) GetNewCards(userID, bankID, limit int) ([]models.FlashCard, error) {
	query := `
		SELECT fc.id, fc.card_bank_id, fc.word, fc.definition, fc.examples, fc.image_url, fc.created_at, fc.updated_at
		FROM flash_cards fc
		LEFT JOIN reviews r ON fc.id = r.flash_card_id AND r.user_id = $1
		WHERE fc.card_bank_id = $2 AND r.id IS NULL
		ORDER BY fc.created_at DESC
		LIMIT $3
	`
	
	var cards []models.FlashCard
	err := r.db.Select(&cards, query, userID, bankID, limit)
	if err != nil {
		return nil, err
	}
	
	return cards, nil
}

// Update updates an existing flash card
func (r *flashCardRepository) Update(card *models.FlashCard) error {
	query := `
		UPDATE flash_cards
		SET word = $1, definition = $2, examples = $3, image_url = $4, updated_at = $5
		WHERE id = $6
	`
	
	card.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		card.Word,
		card.Definition,
## Card Bank Repository (internal/repository/cardbank_repository.go)

The CardBankRepository handles card bank data access.

```go
package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// CardBankRepository defines the interface for card bank data access
type CardBankRepository interface {
	Create(bank *models.CardBank) error
	GetByID(bankID int) (*models.CardBank, error)
	GetBanksForUser(userID int) ([]models.CardBank, error)
	Update(bank *models.CardBank) error
	Delete(bankID int) error
	
	// Membership operations
	CreateMembership(membership *models.BankMembership) error
	GetMembership(userID, bankID int) (*models.BankMembership, error)
	UpdateMembership(membership *models.BankMembership) error
	DeleteMembership(userID, bankID int) error
	UserHasAccess(userID, bankID int) (bool, error)
	
	// Group chat operations
	CreateGroupChat(groupChat *models.GroupChat) error
	GetGroupChatByTelegramID(telegramChatID int64) (*models.GroupChat, error)
	UpdateGroupChat(groupChat *models.GroupChat) error
	DeleteGroupChat(telegramChatID int64) error
}

// cardBankRepository implements the CardBankRepository interface
type cardBankRepository struct {
	db *sqlx.DB
}

// NewCardBankRepository creates a new card bank repository
func NewCardBankRepository(db *sqlx.DB) CardBankRepository {
	return &cardBankRepository{
		db: db,
	}
}

// Create creates a new card bank
func (r *cardBankRepository) Create(bank *models.CardBank) error {
	query := `
		INSERT INTO card_banks (name, description, owner_id, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		bank.Name,
		bank.Description,
		bank.OwnerID,
		bank.IsPublic,
		bank.CreatedAt,
		bank.UpdatedAt,
	).Scan(&bank.ID)
	
	return err
}

// GetByID retrieves a card bank by ID
func (r *cardBankRepository) GetByID(bankID int) (*models.CardBank, error) {
	query := `
		SELECT id, name, description, owner_id, is_public, created_at, updated_at
		FROM card_banks
		WHERE id = $1
	`
	
	var bank models.CardBank
	err := r.db.Get(&bank, query, bankID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &bank, nil
}

// GetBanksForUser retrieves all card banks a user has access to
func (r *cardBankRepository) GetBanksForUser(userID int) ([]models.CardBank, error) {
	query := `
		SELECT cb.id, cb.name, cb.description, cb.owner_id, cb.is_public, cb.created_at, cb.updated_at
		FROM card_banks cb
		JOIN bank_memberships bm ON cb.id = bm.card_bank_id
		WHERE bm.user_id = $1
		ORDER BY cb.created_at DESC
	`
	
	var banks []models.CardBank
	err := r.db.Select(&banks, query, userID)
	if err != nil {
		return nil, err
	}
	
	return banks, nil
}

// Update updates an existing card bank
func (r *cardBankRepository) Update(bank *models.CardBank) error {
	query := `
		UPDATE card_banks
		SET name = $1, description = $2, is_public = $3, updated_at = $4
		WHERE id = $5
	`
	
	bank.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		bank.Name,
		bank.Description,
		bank.IsPublic,
		bank.UpdatedAt,
		bank.ID,
	)
	
	return err
}

// Delete deletes a card bank
func (r *cardBankRepository) Delete(bankID int) error {
	query := `DELETE FROM card_banks WHERE id = $1`
	_, err := r.db.Exec(query, bankID)
	return err
}

// CreateMembership creates a new bank membership
func (r *cardBankRepository) CreateMembership(membership *models.BankMembership) error {
	query := `
		INSERT INTO bank_memberships (user_id, card_bank_id, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		membership.UserID,
		membership.CardBankID,
		membership.Role,
		membership.CreatedAt,
		membership.UpdatedAt,
	).Scan(&membership.ID)
	
	return err
}

// GetMembership retrieves a bank membership
func (r *cardBankRepository) GetMembership(userID, bankID int) (*models.BankMembership, error) {
	query := `
		SELECT id, user_id, card_bank_id, role, created_at, updated_at
		FROM bank_memberships
		WHERE user_id = $1 AND card_bank_id = $2
	`
	
	var membership models.BankMembership
	err := r.db.Get(&membership, query, userID, bankID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &membership, nil
}

// UpdateMembership updates a bank membership
func (r *cardBankRepository) UpdateMembership(membership *models.BankMembership) error {
	query := `
		UPDATE bank_memberships
		SET role = $1, updated_at = $2
		WHERE id = $3
	`
	
	membership.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		membership.Role,
		membership.UpdatedAt,
		membership.ID,
	)
	
	return err
}

// DeleteMembership deletes a bank membership
func (r *cardBankRepository) DeleteMembership(userID, bankID int) error {
	query := `DELETE FROM bank_memberships WHERE user_id = $1 AND card_bank_id = $2`
	_, err := r.db.Exec(query, userID, bankID)
	return err
}

// UserHasAccess checks if a user has access to a card bank
func (r *cardBankRepository) UserHasAccess(userID, bankID int) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM bank_memberships
		WHERE user_id = $1 AND card_bank_id = $2
	`
	
	var count int
	err := r.db.Get(&count, query, userID, bankID)
	if err != nil {
		return false, err
	}
	
	return count > 0, nil
}

// CreateGroupChat creates a new group chat
func (r *cardBankRepository) CreateGroupChat(groupChat *models.GroupChat) error {
	query := `
		INSERT INTO group_chats (telegram_chat_id, title, card_bank_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		groupChat.TelegramChatID,
		groupChat.Title,
		groupChat.CardBankID,
		groupChat.CreatedAt,
		groupChat.UpdatedAt,
	).Scan(&groupChat.ID)
	
	return err
}

// GetGroupChatByTelegramID retrieves a group chat by Telegram chat ID
func (r *cardBankRepository) GetGroupChatByTelegramID(telegramChatID int64) (*models.GroupChat, error) {
	query := `
		SELECT id, telegram_chat_id, title, card_bank_id, created_at, updated_at
		FROM group_chats
		WHERE telegram_chat_id = $1
	`
	
	var groupChat models.GroupChat
	err := r.db.Get(&groupChat, query, telegramChatID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &groupChat, nil
}

// UpdateGroupChat updates a group chat
func (r *cardBankRepository) UpdateGroupChat(groupChat *models.GroupChat) error {
	query := `
		UPDATE group_chats
		SET title = $1, card_bank_id = $2, updated_at = $3
		WHERE id = $4
	`
	
	groupChat.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		groupChat.Title,
		groupChat.CardBankID,
		groupChat.UpdatedAt,
		groupChat.ID,
	)
	
	return err
}

// DeleteGroupChat deletes a group chat
func (r *cardBankRepository) DeleteGroupChat(telegramChatID int64) error {
	query := `DELETE FROM group_chats WHERE telegram_chat_id = $1`
	_, err := r.db.Exec(query, telegramChatID)
	return err
}
```
		card.Examples,
		card.ImageURL,
		card.UpdatedAt,
		card.ID,
	)
	
	return err
}

// Delete deletes a flash card
func (r *flashCardRepository) Delete(cardID int) error {
	query := `DELETE FROM flash_cards WHERE id = $1`
	_, err := r.db.Exec(query, cardID)
	return err
}
## Review Repository (internal/repository/review_repository.go)

The ReviewRepository handles review data access for the spaced repetition system.

```go
package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// ReviewRepository defines the interface for review data access
type ReviewRepository interface {
	Create(review *models.Review) error
	GetByID(reviewID int) (*models.Review, error)
	GetByUserAndCard(userID, cardID int) (*models.Review, error)
	GetDueReviews(userID, bankID int, dueDate time.Time, limit int) ([]models.Review, error)
	Update(review *models.Review) error
	Delete(reviewID int) error
	CountTotalReviews(userID int) (int, error)
	CountDueReviews(userID int, dueDate time.Time) (int, error)
	GetLastReviewDate(userID int) (string, error)
}

// reviewRepository implements the ReviewRepository interface
type reviewRepository struct {
	db *sqlx.DB
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db *sqlx.DB) ReviewRepository {
	return &reviewRepository{
		db: db,
	}
}

// Create creates a new review
func (r *reviewRepository) Create(review *models.Review) error {
	query := `
		INSERT INTO reviews (user_id, flash_card_id, ease_factor, due_date, interval, repetitions, last_reviewed, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		review.UserID,
		review.FlashCardID,
		review.EaseFactor,
		review.DueDate,
		review.Interval,
		review.Repetitions,
		review.LastReviewed,
		review.CreatedAt,
		review.UpdatedAt,
	).Scan(&review.ID)
	
	return err
}

// GetByID retrieves a review by ID
func (r *reviewRepository) GetByID(reviewID int) (*models.Review, error) {
	query := `
		SELECT id, user_id, flash_card_id, ease_factor, due_date, interval, repetitions, last_reviewed, created_at, updated_at
		FROM reviews
		WHERE id = $1
	`
	
	var review models.Review
	err := r.db.Get(&review, query, reviewID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &review, nil
}

// GetByUserAndCard retrieves a review by user ID and card ID
func (r *reviewRepository) GetByUserAndCard(userID, cardID int) (*models.Review, error) {
	query := `
		SELECT id, user_id, flash_card_id, ease_factor, due_date, interval, repetitions, last_reviewed, created_at, updated_at
		FROM reviews
		WHERE user_id = $1 AND flash_card_id = $2
	`
	
	var review models.Review
	err := r.db.Get(&review, query, userID, cardID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &review, nil
}

// GetDueReviews retrieves reviews that are due for a user
func (r *reviewRepository) GetDueReviews(userID, bankID int, dueDate time.Time, limit int) ([]models.Review, error) {
	query := `
		SELECT r.id, r.user_id, r.flash_card_id, r.ease_factor, r.due_date, r.interval, r.repetitions, r.last_reviewed, r.created_at, r.updated_at
		FROM reviews r
		JOIN flash_cards fc ON r.flash_card_id = fc.id
		WHERE r.user_id = $1 AND fc.card_bank_id = $2 AND r.due_date <= $3
		ORDER BY r.due_date ASC
		LIMIT $4
	`
	
	var reviews []models.Review
	err := r.db.Select(&reviews, query, userID, bankID, dueDate, limit)
	if err != nil {
		return nil, err
	}
	
	return reviews, nil
}

// Update updates an existing review
func (r *reviewRepository) Update(review *models.Review) error {
	query := `
		UPDATE reviews
		SET ease_factor = $1, due_date = $2, interval = $3, repetitions = $4, last_reviewed = $5, updated_at = $6
		WHERE id = $7
	`
	
	review.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		review.EaseFactor,
		review.DueDate,
		review.Interval,
		review.Repetitions,
		review.LastReviewed,
		review.UpdatedAt,
		review.ID,
	)
	
	return err
}

// Delete deletes a review
func (r *reviewRepository) Delete(reviewID int) error {
	query := `DELETE FROM reviews WHERE id = $1`
	_, err := r.db.Exec(query, reviewID)
	return err
}

// CountTotalReviews counts the total number of reviews for a user
func (r *reviewRepository) CountTotalReviews(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE user_id = $1`
	
	var count int
	err := r.db.Get(&count, query, userID)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

// CountDueReviews counts the number of due reviews for a user
func (r *reviewRepository) CountDueReviews(userID int, dueDate time.Time) (int, error) {
	query := `SELECT COUNT(*) FROM reviews WHERE user_id = $1 AND due_date <= $2`
	
	var count int
	err := r.db.Get(&count, query, userID, dueDate)
	if err != nil {
		return 0, err
	}
	
	return count, nil
}

// GetLastReviewDate gets the date of the last review for a user
func (r *reviewRepository) GetLastReviewDate(userID int) (string, error) {
	query := `
		SELECT TO_CHAR(last_reviewed, 'YYYY-MM-DD')
		FROM reviews
		WHERE user_id = $1 AND last_reviewed IS NOT NULL
		ORDER BY last_reviewed DESC
		LIMIT 1
	`
	
	var date string
	err := r.db.Get(&date, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	
	return date, nil
}
```

## Settings Repository (internal/repository/settings_repository.go)

The SettingsRepository handles user settings data access.

```go
package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// SettingsRepository defines the interface for settings data access
type SettingsRepository interface {
	Create(settings *models.Settings) error
	GetByID(settingsID int) (*models.Settings, error)
	GetByUserID(userID int) (*models.Settings, error)
	Update(settings *models.Settings) error
	Delete(settingsID int) error
}

// settingsRepository implements the SettingsRepository interface
type settingsRepository struct {
	db *sqlx.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *sqlx.DB) SettingsRepository {
	return &settingsRepository{
		db: db,
	}
}

// Create creates new settings
func (r *settingsRepository) Create(settings *models.Settings) error {
	query := `
		INSERT INTO user_settings (user_id, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	
	err := r.db.QueryRow(
		query,
		settings.UserID,
		settings.Settings,
		settings.CreatedAt,
		settings.UpdatedAt,
	).Scan(&settings.ID)
	
	return err
}

// GetByID retrieves settings by ID
func (r *settingsRepository) GetByID(settingsID int) (*models.Settings, error) {
	query := `
		SELECT id, user_id, settings, created_at, updated_at
		FROM user_settings
		WHERE id = $1
	`
	
	var settings models.Settings
	err := r.db.Get(&settings, query, settingsID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &settings, nil
}

// GetByUserID retrieves settings by user ID
func (r *settingsRepository) GetByUserID(userID int) (*models.Settings, error) {
	query := `
		SELECT id, user_id, settings, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`
	
	var settings models.Settings
	err := r.db.Get(&settings, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}
	
	return &settings, nil
}

// Update updates existing settings
func (r *settingsRepository) Update(settings *models.Settings) error {
	query := `
		UPDATE user_settings
		SET settings = $1, updated_at = $2
		WHERE id = $3
	`
	
	settings.UpdatedAt = time.Now()
	
	_, err := r.db.Exec(
		query,
		settings.Settings,
		settings.UpdatedAt,
		settings.ID,
	)
	
	return err
}

// Delete deletes settings
func (r *settingsRepository) Delete(settingsID int) error {
	query := `DELETE FROM user_settings WHERE id = $1`
	_, err := r.db.Exec(query, settingsID)
	return err
}
```

# Repositories

This document describes the repository implementations for the Flash Cards Language Telegram Bot. Repositories handle data access to the database.