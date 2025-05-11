package spaced_repetition

import (
	"math"
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// SM2Algorithm implements the SuperMemo 2 spaced repetition algorithm
type SM2Algorithm struct {
	// Minimum ease factor to prevent cards from getting too difficult
	MinEaseFactor float64

	// Default ease factor for new cards
	DefaultEaseFactor float64

	// Ease factor adjustment values
	EaseFactorAdjustments map[int]float64
}

// NewSM2Algorithm creates a new SM-2 algorithm with default parameters
func NewSM2Algorithm() *SM2Algorithm {
	return &SM2Algorithm{
		MinEaseFactor:     1.3,
		DefaultEaseFactor: 2.5,
		EaseFactorAdjustments: map[int]float64{
			QualityAgain: -0.3,
			QualityHard:  -0.15,
			QualityGood:  0.0,
			QualityEasy:  0.15,
		},
	}
}

// CalculateNextReview implements the Algorithm interface
func (a *SM2Algorithm) CalculateNextReview(review *models.Review, quality int) (time.Time, int, float64) {
	// Get current values or set defaults for new cards
	repetitions := review.Repetitions
	easeFactor := review.EaseFactor
	if easeFactor == 0 {
		easeFactor = a.DefaultEaseFactor
	}

	// Calculate new ease factor
	easeFactor += a.EaseFactorAdjustments[quality]
	if easeFactor < a.MinEaseFactor {
		easeFactor = a.MinEaseFactor
	}

	// Calculate new interval
	var interval int

	if quality == QualityAgain {
		// Reset the card if the answer was wrong
		repetitions = 0
		interval = 1 // Review again tomorrow
	} else {
		// Increment repetitions
		repetitions++

		// Calculate interval based on repetitions
		switch repetitions {
		case 1:
			interval = 1 // 1 day
		case 2:
			interval = 3 // 3 days
		default:
			// For repetitions > 2, use the formula: interval = previous_interval * ease_factor
			interval = int(math.Round(float64(review.Interval) * easeFactor))
		}

		// Adjust interval based on difficulty
		switch quality {
		case QualityHard:
			interval = int(float64(interval) * 0.8) // Reduce interval for hard cards
		case QualityEasy:
			interval = int(float64(interval) * 1.2) // Increase interval for easy cards
		}

		// Ensure minimum interval of 1 day
		if interval < 1 {
			interval = 1
		}
	}

	// Calculate next review date
	nextReview := time.Now().AddDate(0, 0, interval)

	return nextReview, interval, easeFactor
}
