package spaced_repetition

import (
	"math"
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// CustomAlgorithm implements a customizable spaced repetition algorithm
type CustomAlgorithm struct {
	// Minimum ease factor
	MinEaseFactor float64

	// Default ease factor for new cards
	DefaultEaseFactor float64

	// Ease factor adjustments for each quality rating
	EaseFactorAdjustments map[int]float64

	// Initial intervals (in days) for the first few repetitions
	InitialIntervals []int

	// Interval modifiers for each quality rating
	IntervalModifiers map[int]float64

	// Maximum interval (in days)
	MaxInterval int
}

// NewCustomAlgorithm creates a new customizable algorithm with default parameters
func NewCustomAlgorithm() *CustomAlgorithm {
	return &CustomAlgorithm{
		MinEaseFactor:     1.3,
		DefaultEaseFactor: 2.5,
		EaseFactorAdjustments: map[int]float64{
			QualityAgain: -0.3,
			QualityHard:  -0.15,
			QualityGood:  0.0,
			QualityEasy:  0.15,
		},
		InitialIntervals: []int{1, 3, 7},
		IntervalModifiers: map[int]float64{
			QualityAgain: 0.0, // Reset
			QualityHard:  0.8, // Reduce interval
			QualityGood:  1.0, // Keep as calculated
			QualityEasy:  1.2, // Increase interval
		},
		MaxInterval: 365, // Maximum interval of 1 year
	}
}

// CalculateNextReview implements the Algorithm interface
func (a *CustomAlgorithm) CalculateNextReview(review *models.Review, quality int) (time.Time, int, float64) {
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
		if repetitions <= len(a.InitialIntervals) {
			// Use predefined intervals for the first few repetitions
			interval = a.InitialIntervals[repetitions-1]
		} else {
			// For later repetitions, use the formula: interval = previous_interval * ease_factor
			interval = int(math.Round(float64(review.Interval) * easeFactor))
		}

		// Apply interval modifier based on quality
		interval = int(float64(interval) * a.IntervalModifiers[quality])

		// Ensure minimum interval of 1 day
		if interval < 1 {
			interval = 1
		}

		// Apply maximum interval
		if interval > a.MaxInterval {
			interval = a.MaxInterval
		}
	}

	// Calculate next review date
	nextReview := time.Now().AddDate(0, 0, interval)

	return nextReview, interval, easeFactor
}

// SetParameter sets a parameter of the algorithm
func (a *CustomAlgorithm) SetParameter(name string, value interface{}) error {
	switch name {
	case "min_ease_factor":
		if val, ok := value.(float64); ok {
			a.MinEaseFactor = val
			return nil
		}
	case "default_ease_factor":
		if val, ok := value.(float64); ok {
			a.DefaultEaseFactor = val
			return nil
		}
	case "max_interval":
		if val, ok := value.(int); ok {
			a.MaxInterval = val
			return nil
		}
	case "initial_intervals":
		if val, ok := value.([]int); ok {
			a.InitialIntervals = val
			return nil
		}
	}

	return models.ErrInvalidParameter
}
