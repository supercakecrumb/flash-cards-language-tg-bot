package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
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
