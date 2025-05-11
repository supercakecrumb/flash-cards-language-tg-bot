package spaced_repetition

import (
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// Quality ratings for SM-2 algorithm
const (
	QualityAgain = 0 // Complete blackout, wrong response
	QualityHard  = 1 // Correct response but with difficulty
	QualityGood  = 2 // Correct response with some hesitation
	QualityEasy  = 3 // Perfect response
)

// Algorithm defines the interface for spaced repetition algorithms
type Algorithm interface {
	// CalculateNextReview calculates the next review date based on the current review and difficulty rating
	// Returns: next review date, new interval (days), new ease factor
	CalculateNextReview(review *models.Review, difficulty int) (time.Time, int, float64)
}
