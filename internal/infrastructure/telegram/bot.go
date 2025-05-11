package telegram

import (
	"context"

	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/services"
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
	userStates map[int64]UserState
}

// UserState represents the current state of a user's interaction with the bot
type UserState struct {
	State         string
	CurrentWord   string
	CurrentBank   int
	SelectedDef   string
	Examples      []string
	PhotoURL      string
	ReviewState   *ReviewState
	SettingsField string
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
	msg.ParseMode = "HTML"
	_, err := b.api.Send(msg)
	if err != nil {
		b.logger.Error("Failed to send message", "error", err)
	}
}

// sendErrorMessage sends an error message to a chat
func (b *Bot) sendErrorMessage(chatID int64, text string) {
	b.sendMessage(chatID, "❌ "+text)
}
