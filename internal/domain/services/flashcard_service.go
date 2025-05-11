package services

import (
	"log/slog"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/infrastructure/dictionary"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
)

// Definition represents a word definition from the dictionary
type Definition struct {
	ID           string
	Text         string
	PartOfSpeech string
	Examples     []Example
}

// Example represents a usage example for a word
type Example struct {
	ID   string
	Text string
}

// FlashCardService handles flash card operations
type FlashCardService interface {
	GetDefinitions(word string) ([]Definition, error)
	GetDefinition(definitionID string) (*Definition, error)
	GetExamples(word, definitionID string) ([]Example, error)
	GetExample(exampleID string) (*Example, error)
	CreateFlashCard(card *models.FlashCard) error
	GetFlashCard(cardID int) (*models.FlashCard, error)
	GetFlashCardsByBank(bankID int) ([]models.FlashCard, error)
	UpdateFlashCard(card *models.FlashCard) error
	DeleteFlashCard(cardID int) error
}

type flashCardService struct {
	repo            repository.FlashCardRepository
	dictService     dictionary.DictionaryService
	logger          *slog.Logger
	definitionCache map[string]Definition
	exampleCache    map[string]Example
}

// NewFlashCardService creates a new flash card service
func NewFlashCardService(repo repository.FlashCardRepository, dictService dictionary.DictionaryService, logger *slog.Logger) FlashCardService {
	return &flashCardService{
		repo:            repo,
		dictService:     dictService,
		logger:          logger,
		definitionCache: make(map[string]Definition),
		exampleCache:    make(map[string]Example),
	}
}

// GetDefinitions retrieves definitions for a word from the dictionary API
func (s *flashCardService) GetDefinitions(word string) ([]Definition, error) {
	s.logger.Debug("Getting definitions for word", "word", word)

	dictDefinitions, err := s.dictService.GetDefinitions(word)
	if err != nil {
		s.logger.Error("Failed to get definitions from dictionary API", "error", err)
		return nil, err
	}

	var definitions []Definition
	for _, d := range dictDefinitions {
		def := Definition{
			ID:           d.ID,
			Text:         d.Text,
			PartOfSpeech: d.PartOfSpeech,
		}

		var examples []Example
		for _, e := range d.Examples {
			ex := Example{
				ID:   e.ID,
				Text: e.Text,
			}
			examples = append(examples, ex)
			s.exampleCache[ex.ID] = ex
		}

		def.Examples = examples
		definitions = append(definitions, def)
		s.definitionCache[def.ID] = def
	}

	return definitions, nil
}

// GetDefinition retrieves a cached definition by ID
func (s *flashCardService) GetDefinition(definitionID string) (*Definition, error) {
	def, ok := s.definitionCache[definitionID]
	if !ok {
		s.logger.Error("Definition not found in cache", "definition_id", definitionID)
		return nil, ErrNotFound
	}
	return &def, nil
}

// GetExamples retrieves examples for a word and definition
func (s *flashCardService) GetExamples(word, definitionID string) ([]Example, error) {
	def, err := s.GetDefinition(definitionID)
	if err != nil {
		return nil, err
	}
	return def.Examples, nil
}

// GetExample retrieves a cached example by ID
func (s *flashCardService) GetExample(exampleID string) (*Example, error) {
	ex, ok := s.exampleCache[exampleID]
	if !ok {
		s.logger.Error("Example not found in cache", "example_id", exampleID)
		return nil, ErrNotFound
	}
	return &ex, nil
}

// CreateFlashCard creates a new flash card
func (s *flashCardService) CreateFlashCard(card *models.FlashCard) error {
	s.logger.Info("Creating flash card", "word", card.Word, "bank_id", card.CardBankID)
	return s.repo.Create(card)
}

// GetFlashCard retrieves a flash card by ID
func (s *flashCardService) GetFlashCard(cardID int) (*models.FlashCard, error) {
	s.logger.Debug("Getting flash card", "card_id", cardID)
	return s.repo.GetByID(cardID)
}

// GetFlashCardsByBank retrieves all flash cards in a bank
func (s *flashCardService) GetFlashCardsByBank(bankID int) ([]models.FlashCard, error) {
	s.logger.Debug("Getting flash cards for bank", "bank_id", bankID)
	return s.repo.GetCardsForBank(bankID)
}

// UpdateFlashCard updates an existing flash card
func (s *flashCardService) UpdateFlashCard(card *models.FlashCard) error {
	s.logger.Debug("Updating flash card", "card_id", card.ID)
	return s.repo.Update(card)
}

// DeleteFlashCard deletes a flash card
func (s *flashCardService) DeleteFlashCard(cardID int) error {
	s.logger.Info("Deleting flash card", "card_id", cardID)
	return s.repo.Delete(cardID)
}
