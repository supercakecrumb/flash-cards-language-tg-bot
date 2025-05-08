# Spaced Repetition Algorithm

This document describes the spaced repetition algorithm implementation for the Flash Cards Language Telegram Bot.

## Algorithm Interface (pkg/spaced_repetition/algorithm.go)

The Algorithm interface defines the contract for spaced repetition algorithms.

```go
package spaced_repetition

import (
	"time"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// Algorithm defines the interface for spaced repetition algorithms
type Algorithm interface {
	// CalculateNextReview calculates the next review date based on the current review and difficulty rating
	// Returns: next review date, new interval (days), new ease factor
	CalculateNextReview(review *models.Review, difficulty int) (time.Time, int, float64)
}
```

## SM-2 Algorithm Implementation (pkg/spaced_repetition/sm2.go)

The SM2 algorithm is an implementation of the SuperMemo 2 spaced repetition algorithm.

```go
package spaced_repetition

import (
	"math"
	"time"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// Quality ratings for SM-2 algorithm
const (
	QualityAgain = 0 // Complete blackout, wrong response
	QualityHard  = 1 // Correct response but with difficulty
	QualityGood  = 2 // Correct response with some hesitation
	QualityEasy  = 3 // Perfect response
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
		MinEaseFactor: 1.3,
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
```

## Customizable Algorithm (pkg/spaced_repetition/custom.go)

A customizable spaced repetition algorithm that allows users to adjust parameters.

```go
package spaced_repetition

import (
	"math"
	"time"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
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
		MinEaseFactor: 1.3,
		DefaultEaseFactor: 2.5,
		EaseFactorAdjustments: map[int]float64{
			QualityAgain: -0.3,
			QualityHard:  -0.15,
			QualityGood:  0.0,
			QualityEasy:  0.15,
		},
		InitialIntervals: []int{1, 3, 7},
		IntervalModifiers: map[int]float64{
			QualityAgain: 0.0,  // Reset
			QualityHard:  0.8,  // Reduce interval
			QualityGood:  1.0,  // Keep as calculated
			QualityEasy:  1.2,  // Increase interval
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
	
	return ErrInvalidParameter
}
```

## Algorithm Factory (pkg/spaced_repetition/factory.go)

A factory for creating spaced repetition algorithms based on configuration.

```go
package spaced_repetition

import (
	"errors"
)

// Algorithm types
const (
	AlgorithmTypeSM2    = "sm2"
	AlgorithmTypeCustom = "custom"
)

// Errors
var (
	ErrInvalidAlgorithmType = errors.New("invalid algorithm type")
	ErrInvalidParameter     = errors.New("invalid parameter")
)

// AlgorithmConfig represents configuration for a spaced repetition algorithm
type AlgorithmConfig struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// CreateAlgorithm creates a spaced repetition algorithm based on configuration
func CreateAlgorithm(config AlgorithmConfig) (Algorithm, error) {
	switch config.Type {
	case AlgorithmTypeSM2:
		return NewSM2Algorithm(), nil
		
	case AlgorithmTypeCustom:
		algorithm := NewCustomAlgorithm()
		
		// Apply custom parameters
		for name, value := range config.Parameters {
			if err := algorithm.SetParameter(name, value); err != nil {
				return nil, err
			}
		}
		
		return algorithm, nil
		
	default:
		return nil, ErrInvalidAlgorithmType
	}
}
```

## Usage in the Application

The spaced repetition algorithm is used in the SpacedRepetitionService to calculate when cards should be reviewed next. The service can be configured to use different algorithms based on user preferences.

```go
// Example usage in SpacedRepetitionService
func (s *spacedRepetitionService) ProcessReview(userID, cardID int, quality int) error {
	// Get existing review or create a new one
	review, err := s.reviewRepo.GetByUserAndCard(userID, cardID)
	if err != nil {
		// Create new review if it doesn't exist
		review = models.NewReview(userID, cardID)
		err = s.reviewRepo.Create(review)
		if err != nil {
			return err
		}
	}
	
	// Calculate next review date using the algorithm
	dueDate, interval, easeFactor := s.algorithm.CalculateNextReview(review, quality)
	
	// Update review
	review.DueDate = dueDate
	review.Interval = interval
	review.EaseFactor = easeFactor
	review.Repetitions++
	review.LastReviewed = time.Now()
	
	// Save updated review
	return s.reviewRepo.Update(review)
}
```

The spaced repetition system is a core feature of the Flash Cards Language Telegram Bot, enabling effective vocabulary learning through optimized review scheduling.