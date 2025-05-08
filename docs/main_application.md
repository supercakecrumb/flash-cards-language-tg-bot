# Main Application

## Entry Point (cmd/bot/main.go)

The main.go file serves as the entry point for the application. It is responsible for:

1. Loading configuration from environment variables
2. Setting up the logger
3. Establishing database connection
4. Initializing the dictionary service
5. Creating repositories and services
6. Starting the Telegram bot
7. Handling graceful shutdown

```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/username/flash-cards-language-tg-bot/internal/app"
	"github.com/username/flash-cards-language-tg-bot/internal/infrastructure/logging"
)

func main() {
	// Setup logger
	logger := logging.SetupLogger()
	
	// Load configuration
	config, err := app.LoadConfig()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}
	
	// Create application
	application, err := app.NewApp(config, logger)
	if err != nil {
		logger.Error("Failed to create application", "error", err)
		os.Exit(1)
	}
	
	// Create context that listens for termination signals
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// Handle graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		
		logger.Info("Received termination signal, shutting down...")
		
		// Create a timeout context for shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := application.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during shutdown", "error", err)
		}
		
		cancel()
	}()
	
	// Start the application
	if err := application.Start(ctx); err != nil {
		logger.Error("Application error", "error", err)
		os.Exit(1)
	}
}
```

## Application (internal/app/app.go)

The app.go file is responsible for wiring together all components of the application:

1. Creating database connection
2. Initializing repositories
3. Creating services
4. Setting up the Telegram bot

```go
package app

import (
	"context"
	"log/slog"

	"github.com/username/flash-cards-language-tg-bot/internal/domain/services"
	"github.com/username/flash-cards-language-tg-bot/internal/infrastructure/database"
	"github.com/username/flash-cards-language-tg-bot/internal/infrastructure/dictionary"
	"github.com/username/flash-cards-language-tg-bot/internal/infrastructure/telegram"
	"github.com/username/flash-cards-language-tg-bot/internal/repository"
)

// App represents the application
type App struct {
	config     *Config
	logger     *slog.Logger
	db         *database.PostgresDB
	bot        *telegram.Bot
}

// NewApp creates a new application instance
func NewApp(config *Config, logger *slog.Logger) (*App, error) {
	// Initialize database
	db, err := database.NewPostgresDB(config.Database, logger)
	if err != nil {
		return nil, err
	}
	
	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	flashcardRepo := repository.NewFlashCardRepository(db)
	cardbankRepo := repository.NewCardBankRepository(db)
	reviewRepo := repository.NewReviewRepository(db)
	statisticsRepo := repository.NewStatisticsRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)
	
	// Initialize dictionary service
	var dictService dictionary.DictionaryService
	switch config.Dictionary.Provider {
	case "freedictionary":
		dictService = dictionary.NewFreeDictionaryService(config.Dictionary.APIKey)
	default:
		dictService = dictionary.NewFreeDictionaryService(config.Dictionary.APIKey)
	}
	
	// Initialize services
	userService := services.NewUserService(userRepo, logger)
	flashcardService := services.NewFlashCardService(flashcardRepo, dictService, logger)
	cardbankService := services.NewCardBankService(cardbankRepo, logger)
	spacedRepService := services.NewSpacedRepetitionService(reviewRepo, flashcardRepo, logger)
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
```

## Configuration (internal/app/config.go)

The config.go file is responsible for loading and validating configuration from environment variables:

```go
package app

import (
	"errors"
	"os"
	"strconv"
	"strings"
)

// Config represents the application configuration
type Config struct {
	Telegram struct {
		Token string
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Dictionary struct {
		Provider string
		APIKey   string
	}
	AdminIDs []int64
	LogLevel string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	config := &Config{}
	
	// Telegram configuration
	config.Telegram.Token = os.Getenv("TELEGRAM_TOKEN")
	if config.Telegram.Token == "" {
		return nil, errors.New("TELEGRAM_TOKEN environment variable is required")
	}
	
	// Database configuration
	config.Database.Host = os.Getenv("DB_HOST")
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}
	
	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		config.Database.Port = 5432 // Default PostgreSQL port
	} else {
		config.Database.Port = dbPort
	}
	
	config.Database.User = os.Getenv("DB_USER")
	if config.Database.User == "" {
		return nil, errors.New("DB_USER environment variable is required")
	}
	
	config.Database.Password = os.Getenv("DB_PASSWORD")
	if config.Database.Password == "" {
		return nil, errors.New("DB_PASSWORD environment variable is required")
	}
	
	config.Database.Name = os.Getenv("DB_NAME")
	if config.Database.Name == "" {
		return nil, errors.New("DB_NAME environment variable is required")
	}
	
	// Dictionary configuration
	config.Dictionary.Provider = os.Getenv("DICTIONARY_API")
	if config.Dictionary.Provider == "" {
		config.Dictionary.Provider = "freedictionary"
	}
	
	config.Dictionary.APIKey = os.Getenv("DICTIONARY_API_KEY")
	
	// Admin IDs
	adminIDsStr := os.Getenv("ADMIN_IDS")
	if adminIDsStr != "" {
		adminIDsSlice := strings.Split(adminIDsStr, ",")
		for _, idStr := range adminIDsSlice {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				return nil, errors.New("invalid admin ID format in ADMIN_IDS")
			}
			config.AdminIDs = append(config.AdminIDs, id)
		}
	}
	
	// Log level
	config.LogLevel = os.Getenv("LOG_LEVEL")
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	
	return config, nil
}