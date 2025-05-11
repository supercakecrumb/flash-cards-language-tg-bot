package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
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
