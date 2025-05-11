package spaced_repetition

import (
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// Algorithm types
const (
	AlgorithmTypeSM2    = "sm2"
	AlgorithmTypeCustom = "custom"
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
		return nil, models.ErrInvalidParameter
	}
}
