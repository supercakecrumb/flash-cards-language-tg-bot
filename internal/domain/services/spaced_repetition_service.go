package services

import (
	"log/slog"
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/pkg/spaced_repetition"
)

// SpacedRepetitionService handles spaced repetition operations
type SpacedRepetitionService interface {
	GetDueCards(userID, bankID int, limit int) ([]models.FlashCard, error)
	ProcessReview(userID, cardID int, quality int) error
	GetReviewStats(userID int) (int, int, error) // total cards, due cards
}

type spacedRepetitionService struct {
	reviewRepo    repository.ReviewRepository
	flashcardRepo repository.FlashCardRepository
	algorithm     spaced_repetition.Algorithm
	logger        *slog.Logger
}

// NewSpacedRepetitionService creates a new spaced repetition service
func NewSpacedRepetitionService(
	reviewRepo repository.ReviewRepository,
	flashcardRepo repository.FlashCardRepository,
	algorithm spaced_repetition.Algorithm,
	logger *slog.Logger,
) SpacedRepetitionService {
	return &spacedRepetitionService{
		reviewRepo:    reviewRepo,
		flashcardRepo: flashcardRepo,
		algorithm:     algorithm,
		logger:        logger,
	}
}

// GetDueCards retrieves cards that are due for review
func (s *spacedRepetitionService) GetDueCards(userID, bankID int, limit int) ([]models.FlashCard, error) {
	s.logger.Debug("Getting due cards", "user_id", userID, "bank_id", bankID, "limit", limit)

	// Get due reviews
	reviews, err := s.reviewRepo.GetDueReviews(userID, bankID, time.Now(), limit)
	if err != nil {
		s.logger.Error("Failed to get due reviews", "error", err)
		return nil, err
	}

	// If there are not enough due reviews, get new cards
	if len(reviews) < limit {
		newCardsLimit := limit - len(reviews)
		newCards, err := s.getNewCards(userID, bankID, newCardsLimit)
		if err != nil {
			s.logger.Error("Failed to get new cards", "error", err)
			return nil, err
		}

		// Create initial reviews for new cards
		for _, card := range newCards {
			review := models.NewReview(userID, card.ID)
			err := s.reviewRepo.Create(review)
			if err != nil {
				s.logger.Error("Failed to create review for new card", "error", err)
				continue
			}
		}

		// Get cards for due reviews
		var dueCards []models.FlashCard
		for _, review := range reviews {
			card, err := s.flashcardRepo.GetByID(review.FlashCardID)
			if err != nil {
				s.logger.Error("Failed to get card for review", "error", err)
				continue
			}
			dueCards = append(dueCards, *card)
		}

		// Combine due cards and new cards
		return append(dueCards, newCards...), nil
	}

	// Get cards for due reviews
	var dueCards []models.FlashCard
	for _, review := range reviews {
		card, err := s.flashcardRepo.GetByID(review.FlashCardID)
		if err != nil {
			s.logger.Error("Failed to get card for review", "error", err)
			continue
		}
		dueCards = append(dueCards, *card)
	}

	return dueCards, nil
}

// getNewCards retrieves cards that the user hasn't reviewed yet
func (s *spacedRepetitionService) getNewCards(userID, bankID, limit int) ([]models.FlashCard, error) {
	return s.flashcardRepo.GetNewCards(userID, bankID, limit)
}

// ProcessReview processes a card review and updates the review schedule
func (s *spacedRepetitionService) ProcessReview(userID, cardID int, quality int) error {
	s.logger.Debug("Processing review", "user_id", userID, "card_id", cardID, "quality", quality)

	// Get existing review or create a new one
	review, err := s.reviewRepo.GetByUserAndCard(userID, cardID)
	if err != nil {
		// Create new review if it doesn't exist
		review = models.NewReview(userID, cardID)
		err = s.reviewRepo.Create(review)
		if err != nil {
			s.logger.Error("Failed to create review", "error", err)
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
	err = s.reviewRepo.Update(review)
	if err != nil {
		s.logger.Error("Failed to update review", "error", err)
		return err
	}

	return nil
}

// GetReviewStats retrieves review statistics for a user
func (s *spacedRepetitionService) GetReviewStats(userID int) (int, int, error) {
	s.logger.Debug("Getting review stats", "user_id", userID)

	totalCards, err := s.reviewRepo.CountTotalReviews(userID)
	if err != nil {
		s.logger.Error("Failed to count total reviews", "error", err)
		return 0, 0, err
	}

	dueCards, err := s.reviewRepo.CountDueReviews(userID, time.Now())
	if err != nil {
		s.logger.Error("Failed to count due reviews", "error", err)
		return 0, 0, err
	}

	return totalCards, dueCards, nil
}
