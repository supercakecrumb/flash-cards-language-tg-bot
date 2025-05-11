package services

import (
	"log/slog"
	"time"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
)

// StatisticsService handles statistics operations
type StatisticsService interface {
	GetUserStatistics(userID int) ([]models.Statistics, error)
	GetBankStatistics(userID, bankID int) (*models.Statistics, error)
	IncrementReviewed(userID, bankID int) error
	IncrementLearned(userID, bankID int) error
	UpdateStreak(userID int) error
}

type statisticsService struct {
	repo   repository.StatisticsRepository
	logger *slog.Logger
}

// NewStatisticsService creates a new statistics service
func NewStatisticsService(repo repository.StatisticsRepository, logger *slog.Logger) StatisticsService {
	return &statisticsService{
		repo:   repo,
		logger: logger,
	}
}

// GetUserStatistics retrieves all statistics for a user
func (s *statisticsService) GetUserStatistics(userID int) ([]models.Statistics, error) {
	s.logger.Debug("Getting user statistics", "user_id", userID)
	return s.repo.GetByUserID(userID)
}

// GetBankStatistics retrieves statistics for a specific bank
func (s *statisticsService) GetBankStatistics(userID, bankID int) (*models.Statistics, error) {
	s.logger.Debug("Getting bank statistics", "user_id", userID, "bank_id", bankID)

	stats, err := s.repo.GetByUserAndBank(userID, bankID)
	if err != nil {
		// Create new statistics if they don't exist
		stats = models.NewStatistics(userID, bankID)
		err = s.repo.Create(stats)
		if err != nil {
			s.logger.Error("Failed to create statistics", "error", err)
			return nil, err
		}
	}

	return stats, nil
}

// IncrementReviewed increments the cards reviewed count
func (s *statisticsService) IncrementReviewed(userID, bankID int) error {
	s.logger.Debug("Incrementing cards reviewed", "user_id", userID, "bank_id", bankID)

	stats, err := s.GetBankStatistics(userID, bankID)
	if err != nil {
		return err
	}

	stats.CardsReviewed++
	stats.UpdatedAt = time.Now()

	return s.repo.Update(stats)
}

// IncrementLearned increments the cards learned count
func (s *statisticsService) IncrementLearned(userID, bankID int) error {
	s.logger.Debug("Incrementing cards learned", "user_id", userID, "bank_id", bankID)

	stats, err := s.GetBankStatistics(userID, bankID)
	if err != nil {
		return err
	}

	stats.CardsLearned++
	stats.UpdatedAt = time.Now()

	return s.repo.Update(stats)
}

// UpdateStreak updates the user's streak days
func (s *statisticsService) UpdateStreak(userID int) error {
	s.logger.Debug("Updating streak", "user_id", userID)

	// Get all user statistics
	allStats, err := s.GetUserStatistics(userID)
	if err != nil {
		return err
	}

	// Check if user has reviewed today
	today := time.Now().Format("2006-01-02")
	lastReviewDate, err := s.repo.GetLastReviewDate(userID)
	if err != nil {
		s.logger.Error("Failed to get last review date", "error", err)
		return err
	}

	// Update streak for all banks
	for _, stats := range allStats {
		if lastReviewDate == today {
			// User has reviewed today, increment streak if not already done
			if stats.UpdatedAt.Format("2006-01-02") != today {
				stats.StreakDays++
				stats.UpdatedAt = time.Now()
				err = s.repo.Update(&stats)
				if err != nil {
					s.logger.Error("Failed to update streak", "error", err)
					return err
				}
			}
		} else {
			// User hasn't reviewed today, reset streak if last review was not yesterday
			yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			if lastReviewDate != yesterday {
				stats.StreakDays = 0
				stats.UpdatedAt = time.Now()
				err = s.repo.Update(&stats)
				if err != nil {
					s.logger.Error("Failed to reset streak", "error", err)
					return err
				}
			}
		}
	}

	return nil
}
