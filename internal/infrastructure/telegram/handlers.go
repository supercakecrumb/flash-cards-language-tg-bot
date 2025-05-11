package telegram

import (
	"fmt"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
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

// Placeholder implementations for command handlers
// These would be fully implemented in a real application

func (b *Bot) handleStartCommand(update tgbotapi.Update, user *models.User) {
	chatID := update.Message.Chat.ID

	welcomeText := `Welcome to the Flash Cards Language Bot! 🎉

This bot helps you learn English vocabulary using a spaced repetition system similar to Anki.

To get started:
• Send any English word to create a flash card
• Use /review to practice your vocabulary
• Use /banks to manage your card collections
• Use /stats to see your learning progress
• Use /help to see all available commands

Let's start learning! Send me an English word you want to learn.`

	b.sendMessage(chatID, welcomeText)
}

func (b *Bot) handleHelpCommand(update tgbotapi.Update, user *models.User) {
	chatID := update.Message.Chat.ID

	helpText := `📚 *Flash Cards Language Bot Help* 📚

*Basic Commands:*
• Send any word - Create a flash card for this word
• /add [word] - Explicitly add a word as a flash card
• /review - Start a review session with due cards
• /stats - View your learning statistics
• /help - Show this help message

*Card Banks:*
• /banks - List your card banks
• /create_bank [name] - Create a new card bank
• /share_bank [username] - Share a bank with another user
• /join_bank [code] - Join a shared card bank

*Settings:*
• /settings - Configure your preferences

*Group Chat:*
• Add this bot to a group chat and link it to a card bank to collaboratively create flash cards

*Tips:*
• You can add context photos to your flash cards
• Use the buttons to navigate through definitions and examples
• Regular review is key to effective learning!

For more detailed help on a specific command, type /help [command]`

	msg := tgbotapi.NewMessage(chatID, helpText)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)
}

func (b *Bot) handleAddWordCommand(update tgbotapi.Update, user *models.User, args string) {
	chatID := update.Message.Chat.ID

	// If no word provided in command, prompt user and set state
	if args == "" {
		b.userStates[user.TelegramID] = UserState{
			State: "awaiting_word",
		}
		b.sendMessage(chatID, "Please send the English word you want to add as a flash card.")
		return
	}

	// Process the word provided in command
	b.processWord(user, chatID, args)
}

func (b *Bot) handleWordInput(update tgbotapi.Update, user *models.User, text string) {
	chatID := update.Message.Chat.ID

	// Process the word
	b.processWord(user, chatID, text)
}

func (b *Bot) handleGroupWordInput(update tgbotapi.Update, user *models.User, text string, bankID int) {
	// Similar to handleWordInput but uses the specified bank ID
	chatID := update.Message.Chat.ID

	// Set the bank ID in the user state
	state := UserState{
		State:       "awaiting_word",
		CurrentBank: bankID,
	}
	b.userStates[user.TelegramID] = state

	// Process the word
	b.processWord(user, chatID, text)
}

func (b *Bot) processWord(user *models.User, chatID int64, word string) {
	// This is a placeholder implementation
	// In a real application, this would:
	// 1. Get definitions from the dictionary service
	// 2. Show definitions to the user
	// 3. Wait for user to select a definition
	// 4. Show examples to the user
	// 5. Wait for user to select examples
	// 6. Create the flash card

	b.sendMessage(chatID, fmt.Sprintf("Processing word: %s (placeholder implementation)", word))
}

// Placeholder implementations for other handlers
// These would be fully implemented in a real application

func (b *Bot) handleDefinitionCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder
}

func (b *Bot) handleExampleCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder
}

func (b *Bot) handleContextPhoto(update tgbotapi.Update, user *models.User) {
	// Placeholder
}

func (b *Bot) handleReviewCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder
}

func (b *Bot) handleReviewCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder
}

func (b *Bot) handleStatsCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder
}

func (b *Bot) handleBanksCommand(update tgbotapi.Update, user *models.User) {
	// Placeholder
}

func (b *Bot) handleBankCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder
}

func (b *Bot) handleCreateBankCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder
}

func (b *Bot) handleBankNameInput(update tgbotapi.Update, user *models.User, text string) {
	// Placeholder
}

func (b *Bot) handleShareBankCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder
}

func (b *Bot) handleJoinBankCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder
}

func (b *Bot) handleSettingsCommand(update tgbotapi.Update, user *models.User) {
	// Placeholder
}

func (b *Bot) handleSettingsCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder
}

func (b *Bot) handleSettingsInput(update tgbotapi.Update, user *models.User, text string) {
	// Placeholder
}

func (b *Bot) handlePaginationCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder
}

func (b *Bot) handleAdminCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder
}

func (b *Bot) handleAdminInput(update tgbotapi.Update, user *models.User, text string) {
	// Placeholder
}
