# Infrastructure Components

This document describes the infrastructure components for the Flash Cards Language Telegram Bot, including database connections, external API clients, and the Telegram bot implementation.

## Database (internal/infrastructure/database/postgres.go)

The PostgresDB component handles database connections and migrations.

```go
package database

import (
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Config represents database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config Config, logger *slog.Logger) (*PostgresDB, error) {
	logger.Info("Connecting to PostgreSQL database", 
		"host", config.Host, 
		"port", config.Port, 
		"user", config.User, 
		"database", config.Name,
	)
	
	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name,
	)
	
	// Connect to database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return nil, err
	}
	
	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	
	// Run migrations
	if err := runMigrations(db.DB, config.Name, logger); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		return nil, err
	}
	
	return &PostgresDB{
		db:     db,
		logger: logger,
	}, nil
}

// DB returns the underlying sqlx.DB instance
func (p *PostgresDB) DB() *sqlx.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	p.logger.Info("Closing database connection")
	return p.db.Close()
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB, dbName string, logger *slog.Logger) error {
	logger.Info("Running database migrations")
	
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}
	
	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/infrastructure/database/migrations",
		dbName,
		driver,
	)
	if err != nil {
		return err
	}
	
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	
	logger.Info("Database migrations completed successfully")
	return nil
}
```

## Dictionary Service (internal/infrastructure/dictionary/dictionary.go)

The dictionary service interface and implementations for retrieving word definitions and examples.

```go
package dictionary

// Definition represents a word definition from the dictionary API
type Definition struct {
	ID          string
	Text        string
	PartOfSpeech string
	Examples    []Example
}

// Example represents a usage example for a word
type Example struct {
	ID   string
	Text string
}

// DictionaryService defines the interface for dictionary services
type DictionaryService interface {
	GetDefinitions(word string) ([]Definition, error)
}
```

### Free Dictionary API Implementation (internal/infrastructure/dictionary/free_dictionary.go)

```go
package dictionary

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// FreeDictionaryService implements the DictionaryService interface using the Free Dictionary API
type FreeDictionaryService struct {
	apiKey     string
	httpClient *http.Client
}

// FreeDictionaryResponse represents the response from the Free Dictionary API
type FreeDictionaryResponse []struct {
	Word      string `json:"word"`
	Phonetics []struct {
		Text  string `json:"text"`
		Audio string `json:"audio"`
	} `json:"phonetics"`
	Meanings []struct {
		PartOfSpeech string `json:"partOfSpeech"`
		Definitions  []struct {
			Definition string   `json:"definition"`
			Example    string   `json:"example"`
			Synonyms   []string `json:"synonyms"`
			Antonyms   []string `json:"antonyms"`
		} `json:"definitions"`
	} `json:"meanings"`
}

// NewFreeDictionaryService creates a new Free Dictionary service
func NewFreeDictionaryService(apiKey string) DictionaryService {
	return &FreeDictionaryService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetDefinitions retrieves definitions for a word from the Free Dictionary API
func (s *FreeDictionaryService) GetDefinitions(word string) ([]Definition, error) {
	// Clean up the word
	word = strings.TrimSpace(strings.ToLower(word))
	
	// Build the API URL
	url := fmt.Sprintf("https://api.dictionaryapi.dev/api/v2/entries/en/%s", word)
	
	// Make the request
	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// Check response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}
	
	// Parse the response
	var apiResp FreeDictionaryResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}
	
	// Convert to our Definition type
	var definitions []Definition
	
	for _, entry := range apiResp {
		for i, meaning := range entry.Meanings {
			for j, def := range meaning.Definitions {
				// Create a unique ID for the definition
				defID := fmt.Sprintf("%s_%d_%d", word, i, j)
				
				// Create definition
				definition := Definition{
					ID:          defID,
					Text:        def.Definition,
					PartOfSpeech: meaning.PartOfSpeech,
				}
				
				// Add example if available
				if def.Example != "" {
					exID := fmt.Sprintf("%s_%d_%d_ex", word, i, j)
					example := Example{
						ID:   exID,
						Text: def.Example,
					}
					definition.Examples = append(definition.Examples, example)
				}
				
				definitions = append(definitions, definition)
			}
		}
	}
	
	return definitions, nil
}
```

### Mock Dictionary Service for Testing (internal/infrastructure/dictionary/mock_dictionary.go)

```go
package dictionary

// MockDictionaryService implements the DictionaryService interface for testing
type MockDictionaryService struct {
	Definitions map[string][]Definition
}

// NewMockDictionaryService creates a new mock dictionary service
func NewMockDictionaryService() *MockDictionaryService {
	return &MockDictionaryService{
		Definitions: make(map[string][]Definition),
	}
}

// GetDefinitions returns mock definitions for a word
func (s *MockDictionaryService) GetDefinitions(word string) ([]Definition, error) {
	if defs, ok := s.Definitions[word]; ok {
		return defs, nil
	}
	
	// Return some default mock data if word not found
	return []Definition{
		{
			ID:          "mock_1",
			Text:        "A mock definition for testing purposes",
			PartOfSpeech: "noun",
			Examples: []Example{
				{
					ID:   "mock_ex_1",
					Text: "This is a mock example",
				},
			},
		},
	}, nil
}

// AddMockDefinition adds a mock definition for a word
func (s *MockDictionaryService) AddMockDefinition(word string, definition Definition) {
	if _, ok := s.Definitions[word]; !ok {
		s.Definitions[word] = []Definition{}
	}
	s.Definitions[word] = append(s.Definitions[word], definition)
}
```

## Telegram Bot (internal/infrastructure/telegram/bot.go)

The Telegram bot implementation that handles user interactions.

```go
package telegram

import (
	"context"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/services"
	"github.com/username/flash-cards-language-tg-bot/internal/infrastructure/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log/slog"
)

// Bot represents the Telegram bot
type Bot struct {
	api              *tgbotapi.BotAPI
	logger           *slog.Logger
	userService      services.UserService
	flashcardService services.FlashCardService
	cardbankService  services.CardBankService
	spacedRepService services.SpacedRepetitionService
	statsService     services.StatisticsService
	settingsService  services.SettingsService
	adminService     services.AdminService
	
	// State management for multi-step operations
	userStates       map[int64]UserState
}

// UserState represents the current state of a user's interaction with the bot
type UserState struct {
	State       string
	CurrentWord string
	CurrentBank int
	SelectedDef string
	Examples    []string
	PhotoURL    string
	// Other state fields as needed
}

// NewBot creates a new Telegram bot
func NewBot(
	token string,
	logger *slog.Logger,
	userService services.UserService,
	flashcardService services.FlashCardService,
	cardbankService services.CardBankService,
	spacedRepService services.SpacedRepetitionService,
	statsService services.StatisticsService,
	settingsService services.SettingsService,
	adminService services.AdminService,
) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	
	return &Bot{
		api:              api,
		logger:           logger,
		userService:      userService,
		flashcardService: flashcardService,
		cardbankService:  cardbankService,
		spacedRepService: spacedRepService,
		statsService:     statsService,
		settingsService:  settingsService,
		adminService:     adminService,
		userStates:       make(map[int64]UserState),
	}, nil
}

// Start starts the Telegram bot
func (b *Bot) Start(ctx context.Context) error {
	b.logger.Info("Starting Telegram bot")
	
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	
	updates := b.api.GetUpdatesChan(u)
	
	for {
		select {
		case update := <-updates:
			go b.handleUpdate(update)
		case <-ctx.Done():
			b.logger.Info("Stopping Telegram bot")
			return nil
		}
	}
}

// handleUpdate handles a Telegram update
func (b *Bot) handleUpdate(update tgbotapi.Update) {
	// Log incoming update
	b.logger.Debug("Received update", 
		"update_id", update.UpdateID,
		"chat_id", getChatID(update),
		"user_id", getUserID(update),
	)
	
	// Route update to appropriate handler
	switch {
	case update.Message != nil && update.Message.IsCommand():
		b.handleCommand(update)
	case update.CallbackQuery != nil:
		b.handleCallback(update)
	case update.Message != nil && update.Message.Photo != nil:
		b.handlePhoto(update)
	case update.Message != nil:
		b.handleMessage(update)
	}
}

// Helper functions to extract IDs
func getChatID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func getUserID(update tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.From.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.From.ID
	}
	return 0
}

// ensureUser ensures that a user exists in the database
func (b *Bot) ensureUser(from *tgbotapi.User) (*models.User, error) {
	user, err := b.userService.GetByTelegramID(from.ID)
	if err != nil {
		// Create user if not found
		return b.userService.CreateUser(
			from.ID,
			from.UserName,
			from.FirstName,
			from.LastName,
		)
	}
	return user, nil
}

// sendMessage sends a text message to a chat
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Error("Failed to send message", "error", err)
	}
}

// sendErrorMessage sends an error message to a chat
func (b *Bot) sendErrorMessage(chatID int64, text string) {
	b.sendMessage(chatID, "❌ "+text)
}
```

## Telegram Bot Handlers (internal/infrastructure/telegram/handlers.go)

The handlers for different types of Telegram updates.

```go
package telegram

import (
	"fmt"
	"strconv"
	"strings"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// Command handler
func (b *Bot) handleCommand(update tgbotapi.Update) {
	cmd := update.Message.Command()
	args := update.Message.CommandArguments()
	
	// Ensure user exists in our system
	user, err := b.ensureUser(update.Message.From)
	if err != nil {
		b.logger.Error("Failed to ensure user exists", 
			"error", err,
			"user_id", update.Message.From.ID,
		)
		b.sendErrorMessage(update.Message.Chat.ID, "Internal error. Please try again later.")
		return
	}
	
	// Check if command is admin-only
	if isAdminCommand(cmd) && !b.adminService.IsAdmin(user.TelegramID) {
		b.logger.Warn("Unauthorized admin command attempt", 
			"user_id", user.TelegramID,
			"command", cmd,
		)
		b.sendMessage(update.Message.Chat.ID, "You don't have permission to use this command.")
		return
	}
	
	// Route to specific command handler
	switch cmd {
	case "start":
		b.handleStartCommand(update, user)
	case "help":
		b.handleHelpCommand(update, user)
	case "add":
		b.handleAddWordCommand(update, user, args)
	case "review":
		b.handleReviewCommand(update, user, args)
	case "stats":
		b.handleStatsCommand(update, user, args)
	case "banks":
		b.handleBanksCommand(update, user)
	case "create_bank":
		b.handleCreateBankCommand(update, user, args)
	case "share_bank":
		b.handleShareBankCommand(update, user, args)
	case "join_bank":
		b.handleJoinBankCommand(update, user, args)
	case "settings":
		b.handleSettingsCommand(update, user)
	case "admin":
		b.handleAdminCommand(update, user, args)
	default:
		b.sendMessage(update.Message.Chat.ID, "Unknown command. Type /help to see available commands.")
	}
}

// isAdminCommand checks if a command requires admin privileges
func isAdminCommand(cmd string) bool {
	adminCommands := []string{"admin", "broadcast", "stats_all"}
	for _, c := range adminCommands {
		if c == cmd {
			return true
		}
	}
	return false
}

// Callback query handler
func (b *Bot) handleCallback(update tgbotapi.Update) {
	data := update.CallbackQuery.Data
	chatID := update.CallbackQuery.Message.Chat.ID
	
	// Acknowledge the callback query to remove loading indicator
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	b.api.Request(callback)
	
	// Ensure user exists in our system
	user, err := b.ensureUser(update.CallbackQuery.From)
	if err != nil {
		b.logger.Error("Failed to ensure user exists in callback", 
			"error", err,
			"user_id", update.CallbackQuery.From.ID,
		)
		b.sendErrorMessage(chatID, "Internal error. Please try again later.")
		return
	}
	
	// Parse callback data
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		b.logger.Error("Invalid callback data format", "data", data)
		return
	}
	
	callbackType := parts[0]
	
	// Route to specific callback handler
	switch callbackType {
	case "def":
		b.handleDefinitionCallback(update, user, parts[1:])
	case "ex":
		b.handleExampleCallback(update, user, parts[1:])
	case "rev":
		b.handleReviewCallback(update, user, parts[1:])
	case "bank":
		b.handleBankCallback(update, user, parts[1:])
	case "set":
		b.handleSettingsCallback(update, user, parts[1:])
	case "page":
		b.handlePaginationCallback(update, user, parts[1:])
	default:
		b.logger.Warn("Unknown callback type", "type", callbackType)
	}
}

// Message handler (non-command text messages)
func (b *Bot) handleMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text
	
	// Ensure user exists in our system
	user, err := b.ensureUser(update.Message.From)
	if err != nil {
		b.logger.Error("Failed to ensure user exists in message handler", 
			"error", err,
			"user_id", update.Message.From.ID,
		)
		b.sendErrorMessage(chatID, "Internal error. Please try again later.")
		return
	}
	
	// Check if we're in a group chat
	isGroup := update.Message.Chat.IsGroup() || update.Message.Chat.IsSuperGroup()
	
	// Get user's current state
	state, exists := b.userStates[user.TelegramID]
	
	// Handle based on state or default to word addition in private chats
	if exists {
		switch state.State {
		case "awaiting_word":
			b.handleWordInput(update, user, text)
		case "awaiting_bank_name":
			b.handleBankNameInput(update, user, text)
		case "awaiting_settings":
			b.handleSettingsInput(update, user, text)
		case "awaiting_admin_input":
			b.handleAdminInput(update, user, text)
		default:
			// If in private chat and no specific state, treat as word to add
			if !isGroup {
				b.handleWordInput(update, user, text)
			}
		}
	} else if !isGroup {
		// Default behavior in private chat: treat as word to add
		b.handleWordInput(update, user, text)
	} else if isGroup {
		// In group chat, check if this group is linked to a card bank
		groupChat, err := b.cardbankService.GetGroupChat(chatID)
		if err == nil && groupChat != nil {
			// This group is linked to a card bank, process word
			b.handleGroupWordInput(update, user, text, groupChat.CardBankID)
		}
	}
}

// Photo handler
func (b *Bot) handlePhoto(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	
	// Ensure user exists in our system
	user, err := b.ensureUser(update.Message.From)
	if err != nil {
		b.logger.Error("Failed to ensure user exists in photo handler", 
			"error", err,
			"user_id", update.Message.From.ID,
		)
		b.sendErrorMessage(chatID, "Internal error. Please try again later.")
		return
	}
	
	// Get user's current state
	state, exists := b.userStates[user.TelegramID]
	
	// Only process photos if we're expecting one for a flash card
	if exists && state.State == "awaiting_photo" {
		b.handleContextPhoto(update, user)
	} else {
		b.sendMessage(chatID, "I can only process photos when you're creating a flash card. Please send a word first.")
	}
}
```

## Telegram Bot Keyboards (internal/infrastructure/telegram/keyboards.go)

Utility functions for creating Telegram inline keyboards.

```go
package telegram

import (
	"fmt"
	
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/username/flash-cards-language-tg-bot/internal/domain/models"
)

// createDefinitionsKeyboard creates an inline keyboard with definitions
func (b *Bot) createDefinitionsKeyboard(word string, definitions []services.Definition) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	for i, def := range definitions {
		if i >= 5 {
			// Limit to 5 definitions to avoid huge messages
			break
		}
		
		// Truncate definition if too long
		defText := def.Text
		if len(defText) > 30 {
			defText = defText[:27] + "..."
		}
		
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d. %s (%s)", i+1, defText, def.PartOfSpeech),
			fmt.Sprintf("def:%s", def.ID),
		)
		
		rows = append(rows, []tgbotapi.InlineKeyboardButton{button})
	}
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createExamplesKeyboard creates an inline keyboard with examples
func (b *Bot) createExamplesKeyboard(examples []services.Example, selectedExamples []string) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	for i, ex := range examples {
		if i >= 5 {
			// Limit to 5 examples to avoid huge messages
			break
		}
		
		// Check if this example is already selected
		isSelected := false
		for _, id := range selectedExamples {
			if id == ex.ID {
				isSelected = true
				break
			}
		}
		
		// Truncate example if too long
		exText := ex.Text
		if len(exText) > 30 {
			exText = exText[:27] + "..."
		}
		
		// Add checkmark if selected
		if isSelected {
			exText = "✅ " + exText
		}
		
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%d. %s", i+1, exText),
			fmt.Sprintf("ex:select:%s", ex.ID),
		)
		
		rows = append(rows, []tgbotapi.InlineKeyboardButton{button})
	}
	
	// Add buttons to continue
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Add context photo", "ex:photo"),
		tgbotapi.NewInlineKeyboardButtonData("Create card", "ex:create"),
	})
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createReviewKeyboard creates an inline keyboard for card review
func (b *Bot) createReviewKeyboard(isFlipped bool) tgbotapi.InlineKeyboardMarkup {
	if !isFlipped {
		// Show flip button if card is not flipped
		return tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Flip card", "rev:flip"),
			),
		)
	}
	
	// Show rating buttons if card is flipped
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Again", "rev:rate:0"),
			tgbotapi.NewInlineKeyboardButtonData("Hard", "rev:rate:1"),
			tgbotapi.NewInlineKeyboardButtonData("Good", "rev:rate:2"),
			tgbotapi.NewInlineKeyboardButtonData("Easy", "rev:rate:3"),
		),
	)
}

// createBanksKeyboard creates an inline keyboard with card banks
func (b *Bot) createBanksKeyboard(banks []models.CardBank, currentPage, totalPages int) tgbotapi.InlineKeyboardMarkup {
	var rows [][]tgbotapi.InlineKeyboardButton
	
	for _, bank := range banks {
		button := tgbotapi.NewInlineKeyboardButtonData(
			bank.Name,
			fmt.Sprintf("bank:select:%d", bank.ID),
		)
		
		rows = append(rows, []tgbotapi.InlineKeyboardButton{button})
	}
	
	// Add pagination buttons if needed
	if totalPages > 1 {
		var paginationRow []tgbotapi.InlineKeyboardButton
		
		if currentPage > 1 {
			paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Previous",
				fmt.Sprintf("page:banks:%d", currentPage-1),
			))
		}
		
		if currentPage < totalPages {
			paginationRow = append(paginationRow, tgbotapi.NewInlineKeyboardButtonData(
				"Next ▶️",
				fmt.Sprintf("page:banks:%d", currentPage+1),
			))
		}
		
		rows = append(rows, paginationRow)
	}
	
	// Add create bank button
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("Create new bank", "bank:create"),
	})
	
	return tgbotapi.NewInlineKeyboardMarkup(rows...)
}

// createSettingsKeyboard creates an inline keyboard for settings
func (b *Bot) createSettingsKeyboard(settings *models.Settings) tgbotapi.InlineKeyboardMarkup {
	notificationsText := "Notifications: OFF"
	if settings.Settings.NotificationsOn {
		notificationsText = "Notifications: ON"
	}
	
	darkModeText := "Dark Mode: OFF"
	if settings.Settings.DarkMode {
		darkModeText = "Dark Mode: ON"
	}
	
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Change Active Bank", "set:bank"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("Review Limit: %d", settings.Settings.ReviewLimit),
				"set:limit",
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(notificationsText, "set:notifications"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(darkModeText, "set:darkmode"),
		),
	)
}
```

## Logging (internal/infrastructure/logging/logger.go)

The logging setup using Go's slog package.

```go
package logging

import (
	"os"
	"strings"

	"log/slog"
)

// SetupLogger creates and configures a structured logger
func SetupLogger() *slog.Logger {
	logLevel := getLogLevel()
	
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}
	
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	
	return logger
}

// getLogLevel determines the log level from environment variables
func getLogLevel() slog.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))
	
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error", "crit":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
```

These infrastructure components provide the technical foundation for the application, handling external integrations, database access, and user interactions.