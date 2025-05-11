package models

import (
	"time"
)

// Statistics represents a user's learning statistics
type Statistics struct {
	ID            int       `db:"id"`
	UserID        int       `db:"user_id"`
	CardBankID    int       `db:"card_bank_id"`
	CardsReviewed int       `db:"cards_reviewed"`
	CardsLearned  int       `db:"cards_learned"`
	StreakDays    int       `db:"streak_days"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// NewStatistics creates new statistics for a user and card bank
func NewStatistics(userID, cardBankID int) *Statistics {
	now := time.Now()
	return &Statistics{
		UserID:        userID,
		CardBankID:    cardBankID,
		CardsReviewed: 0,
		CardsLearned:  0,
		StreakDays:    0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
