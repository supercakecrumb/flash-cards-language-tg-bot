package models

import (
	"time"
)

// Review represents a user's review of a flash card
type Review struct {
	ID           int       `db:"id"`
	UserID       int       `db:"user_id"`
	FlashCardID  int       `db:"flash_card_id"`
	EaseFactor   float64   `db:"ease_factor"`
	DueDate      time.Time `db:"due_date"`
	Interval     int       `db:"interval"`    // in days
	Repetitions  int       `db:"repetitions"` // number of times reviewed
	LastReviewed time.Time `db:"last_reviewed"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// NewReview creates a new review
func NewReview(userID, flashCardID int) *Review {
	now := time.Now()
	return &Review{
		UserID:       userID,
		FlashCardID:  flashCardID,
		EaseFactor:   2.5, // Default ease factor
		DueDate:      now, // Due immediately
		Interval:     0,
		Repetitions:  0,
		LastReviewed: time.Time{}, // Zero time
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
