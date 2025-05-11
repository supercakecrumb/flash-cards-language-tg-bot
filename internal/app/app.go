package app

import (
	"context"
	"log/slog"

	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/services"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/infrastructure/database"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/infrastructure/dictionary"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/infrastructure/telegram"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/repository"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/pkg/spaced_repetition"
)

// App represents the application
type App struct {
	config *Config
	logger *slog.Logger
	db     *database.PostgresDB
	bot    *telegram.Bot
}

// NewApp creates a new application instance
func NewApp(config *Config, logger *slog.Logger) (*App, error) {
	// Initialize database
	dbConfig := database.Config{
		Host:     config.Database.Host,
		Port:     config.Database.Port,
		User:     config.Database.User,
		Password: config.Database.Password,
		Name:     config.Database.Name,
	}

	db, err := database.NewPostgresDB(dbConfig, logger)
	if err != nil {
		return nil, err
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db.DB())
	flashcardRepo := repository.NewFlashCardRepository(db.DB())
	cardbankRepo := repository.NewCardBankRepository(db.DB())
	reviewRepo := repository.NewReviewRepository(db.DB())
	statisticsRepo := repository.NewStatisticsRepository(db.DB())
	settingsRepo := repository.NewSettingsRepository(db.DB())

	// Initialize dictionary service
	var dictService dictionary.DictionaryService
	switch config.Dictionary.Provider {
	case "freedictionary":
		dictService = dictionary.NewFreeDictionaryService(config.Dictionary.APIKey)
	default:
		dictService = dictionary.NewFreeDictionaryService(config.Dictionary.APIKey)
	}

	// Initialize spaced repetition algorithm
	algorithm := spaced_repetition.NewSM2Algorithm()

	// Initialize services
	userService := services.NewUserService(userRepo, logger)
	flashcardService := services.NewFlashCardService(flashcardRepo, dictService, logger)
	cardbankService := services.NewCardBankService(cardbankRepo, logger)
	spacedRepService := services.NewSpacedRepetitionService(reviewRepo, flashcardRepo, algorithm, logger)
	statsService := services.NewStatisticsService(statisticsRepo, logger)
	settingsService := services.NewSettingsService(settingsRepo, logger)
	adminService := services.NewAdminService(config.AdminIDs, userRepo, logger)

	// Initialize Telegram bot
	bot, err := telegram.NewBot(
		config.Telegram.Token,
		logger,
		userService,
		flashcardService,
		cardbankService,
		spacedRepService,
		statsService,
		settingsService,
		adminService,
	)
	if err != nil {
		return nil, err
	}

	return &App{
		config: config,
		logger: logger,
		db:     db,
		bot:    bot,
	}, nil
}

// Start starts the application
func (a *App) Start(ctx context.Context) error {
	// Start the bot
	return a.bot.Start(ctx)
}

// Shutdown gracefully shuts down the application
func (a *App) Shutdown(ctx context.Context) error {
	// Close database connection
	if err := a.db.Close(); err != nil {
		return err
	}

	return nil
}
