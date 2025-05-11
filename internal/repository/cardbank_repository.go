package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
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
