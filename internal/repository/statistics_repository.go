package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// StatisticsRepository defines the interface for statistics data access
type StatisticsRepository interface {
	Create(stats *models.Statistics) error
	GetByID(statsID int) (*models.Statistics, error)
	GetByUserID(userID int) ([]models.Statistics, error)
	GetByUserAndBank(userID, bankID int) (*models.Statistics, error)
	Update(stats *models.Statistics) error
	Delete(statsID int) error
	GetLastReviewDate(userID int) (string, error)
}

// statisticsRepository implements the StatisticsRepository interface
type statisticsRepository struct {
	db *sqlx.DB
}

// NewStatisticsRepository creates a new statistics repository
func NewStatisticsRepository(db *sqlx.DB) StatisticsRepository {
	return &statisticsRepository{
		db: db,
	}
}

// Create creates new statistics
func (r *statisticsRepository) Create(stats *models.Statistics) error {
	query := `
		INSERT INTO statistics (user_id, card_bank_id, cards_reviewed, cards_learned, streak_days, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		stats.UserID,
		stats.CardBankID,
		stats.CardsReviewed,
		stats.CardsLearned,
		stats.StreakDays,
		stats.CreatedAt,
		stats.UpdatedAt,
	).Scan(&stats.ID)

	return err
}

// GetByID retrieves statistics by ID
func (r *statisticsRepository) GetByID(statsID int) (*models.Statistics, error) {
	query := `
		SELECT id, user_id, card_bank_id, cards_reviewed, cards_learned, streak_days, created_at, updated_at
		FROM statistics
		WHERE id = $1
	`

	var stats models.Statistics
	err := r.db.Get(&stats, query, statsID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &stats, nil
}

// GetByUserID retrieves all statistics for a user
func (r *statisticsRepository) GetByUserID(userID int) ([]models.Statistics, error) {
	query := `
		SELECT id, user_id, card_bank_id, cards_reviewed, cards_learned, streak_days, created_at, updated_at
		FROM statistics
		WHERE user_id = $1
	`

	var stats []models.Statistics
	err := r.db.Select(&stats, query, userID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetByUserAndBank retrieves statistics for a specific user and bank
func (r *statisticsRepository) GetByUserAndBank(userID, bankID int) (*models.Statistics, error) {
	query := `
		SELECT id, user_id, card_bank_id, cards_reviewed, cards_learned, streak_days, created_at, updated_at
		FROM statistics
		WHERE user_id = $1 AND card_bank_id = $2
	`

	var stats models.Statistics
	err := r.db.Get(&stats, query, userID, bankID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &stats, nil
}

// Update updates existing statistics
func (r *statisticsRepository) Update(stats *models.Statistics) error {
	query := `
		UPDATE statistics
		SET cards_reviewed = $1, cards_learned = $2, streak_days = $3, updated_at = $4
		WHERE id = $5
	`

	stats.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		stats.CardsReviewed,
		stats.CardsLearned,
		stats.StreakDays,
		stats.UpdatedAt,
		stats.ID,
	)

	return err
}

// Delete deletes statistics
func (r *statisticsRepository) Delete(statsID int) error {
	query := `DELETE FROM statistics WHERE id = $1`
	_, err := r.db.Exec(query, statsID)
	return err
}

// GetLastReviewDate gets the date of the last review for a user
func (r *statisticsRepository) GetLastReviewDate(userID int) (string, error) {
	query := `
		SELECT TO_CHAR(r.last_reviewed, 'YYYY-MM-DD')
		FROM reviews r
		WHERE r.user_id = $1 AND r.last_reviewed IS NOT NULL
		ORDER BY r.last_reviewed DESC
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
