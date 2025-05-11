package telegram

import (
	"fmt"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/pkg/spaced_repetition"
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

// Command handlers implementation

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
	// Clean up the word
	word = strings.TrimSpace(strings.ToLower(word))

	// Get user's active card bank
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	// Use the bank ID from user state if available, otherwise use the active bank from settings
	bankID := settings.Settings.ActiveCardBankID
	state, exists := b.userStates[user.TelegramID]
	if exists && state.CurrentBank != 0 {
		bankID = state.CurrentBank
	}

	// Check if user has access to this bank
	hasAccess, err := b.cardbankService.UserHasAccess(user.ID, bankID)
	if err != nil || !hasAccess {
		b.logger.Error("User doesn't have access to active bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", bankID,
		)
		b.sendErrorMessage(chatID, "You don't have access to your active card bank. Please select another bank using /banks.")
		return
	}

	// Get definitions from dictionary service
	b.sendMessage(chatID, fmt.Sprintf("Looking up definitions for \"%s\"...", word))

	definitions, err := b.flashcardService.GetDefinitions(word)
	if err != nil {
		b.logger.Error("Failed to get definitions",
			"error", err,
			"word", word,
		)
		b.sendErrorMessage(chatID, "Failed to get definitions for this word. Please try another word.")
		return
	}

	if len(definitions) == 0 {
		b.sendMessage(chatID, "No definitions found for this word. Please check the spelling or try another word.")
		return
	}

	// Store word in user state
	state = UserState{
		State:       "selecting_definition",
		CurrentWord: word,
		CurrentBank: bankID,
	}
	b.userStates[user.TelegramID] = state

	// Send definitions with inline keyboard
	text := fmt.Sprintf("📝 *Definitions for \"%s\"*\n\nPlease select the definition you want to use:", word)

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = b.createDefinitionsKeyboard(word, definitions)

	b.api.Send(msg)
}

func (b *Bot) handleDefinitionCallback(update tgbotapi.Update, user *models.User, args []string) {
	chatID := update.CallbackQuery.Message.Chat.ID

	if len(args) < 1 {
		b.logger.Error("Invalid definition callback data", "args", args)
		return
	}

	definitionID := args[0]

	// Get user state
	state, exists := b.userStates[user.TelegramID]
	if !exists || state.CurrentWord == "" {
		b.sendErrorMessage(chatID, "Session expired. Please start again by sending a word.")
		return
	}

	// Update user state
	state.State = "selecting_examples"
	state.SelectedDef = definitionID
	b.userStates[user.TelegramID] = state

	// Get examples for this definition
	examples, err := b.flashcardService.GetExamples(state.CurrentWord, definitionID)
	if err != nil {
		b.logger.Error("Failed to get examples",
			"error", err,
			"word", state.CurrentWord,
			"def_id", definitionID,
		)
		b.sendErrorMessage(chatID, "Failed to get examples for this definition. Please try again.")
		return
	}

	// Get the definition text
	definition, err := b.flashcardService.GetDefinition(definitionID)
	if err != nil {
		b.logger.Error("Failed to get definition",
			"error", err,
			"def_id", definitionID,
		)
		b.sendErrorMessage(chatID, "Failed to get the definition. Please try again.")
		return
	}

	// Send message about selected definition
	selectionText := fmt.Sprintf("You selected: *%s*\n\nNow, please select examples you want to include:", definition.Text)

	msg := tgbotapi.NewMessage(chatID, selectionText)
	msg.ParseMode = "Markdown"
	b.api.Send(msg)

	// Send examples with inline keyboard
	if len(examples) == 0 {
		// No examples available
		text := "No examples available for this definition. You can add your own context later."

		// Add buttons to continue or add photo
		keyboard := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Add context photo", "ex:photo"),
				tgbotapi.NewInlineKeyboardButtonData("Create card", "ex:create"),
			),
		)

		msg := tgbotapi.NewMessage(chatID, text)
		msg.ReplyMarkup = keyboard

		b.api.Send(msg)
		return
	}

	// Examples available
	examplesText := fmt.Sprintf("📚 *Examples for \"%s\"*\n\nSelect the examples you want to include (you can select multiple):", state.CurrentWord)

	examplesMsg := tgbotapi.NewMessage(chatID, examplesText)
	examplesMsg.ParseMode = "Markdown"
	examplesMsg.ReplyMarkup = b.createExamplesKeyboard(examples, state.Examples)

	b.api.Send(examplesMsg)
}

func (b *Bot) handleExampleCallback(update tgbotapi.Update, user *models.User, args []string) {
	chatID := update.CallbackQuery.Message.Chat.ID

	if len(args) < 1 {
		b.logger.Error("Invalid example callback data", "args", args)
		return
	}

	action := args[0]

	// Get user state
	state, exists := b.userStates[user.TelegramID]
	if !exists || state.CurrentWord == "" || state.SelectedDef == "" {
		b.sendErrorMessage(chatID, "Session expired. Please start again by sending a word.")
		return
	}

	switch action {
	case "select":
		// User selected an example
		if len(args) < 2 {
			b.logger.Error("Invalid example selection data", "args", args)
			return
		}

		exampleID := args[1]

		// Add to selected examples if not already there
		found := false
		for _, id := range state.Examples {
			if id == exampleID {
				found = true
				break
			}
		}

		if !found {
			state.Examples = append(state.Examples, exampleID)
			b.userStates[user.TelegramID] = state

			// Get the example
			example, err := b.flashcardService.GetExample(exampleID)
			if err != nil {
				b.logger.Error("Failed to get example",
					"error", err,
					"example_id", exampleID,
				)
				return
			}

			// Update the message to show selection
			b.sendMessage(chatID, fmt.Sprintf("Example selected: \"%s\"\n\nYou have selected %d examples.", example.Text, len(state.Examples)))

			// Send updated examples keyboard
			examples, err := b.flashcardService.GetExamples(state.CurrentWord, state.SelectedDef)
			if err != nil {
				b.logger.Error("Failed to get examples for keyboard update",
					"error", err,
					"word", state.CurrentWord,
					"def_id", state.SelectedDef,
				)
				return
			}

			examplesText := fmt.Sprintf("📚 *Examples for \"%s\"*\n\nSelect the examples you want to include (you can select multiple):", state.CurrentWord)

			examplesMsg := tgbotapi.NewMessage(chatID, examplesText)
			examplesMsg.ParseMode = "Markdown"
			examplesMsg.ReplyMarkup = b.createExamplesKeyboard(examples, state.Examples)

			b.api.Send(examplesMsg)
		}

	case "photo":
		// User wants to add a context photo
		state.State = "awaiting_photo"
		b.userStates[user.TelegramID] = state

		b.sendMessage(chatID, "Please send a photo that provides context for this word.")

	case "create":
		// User wants to create the card with selected examples
		b.createFlashCard(chatID, user, state)
	}
}

func (b *Bot) handleContextPhoto(update tgbotapi.Update, user *models.User) {
	chatID := update.Message.Chat.ID

	// Get the largest photo
	photos := update.Message.Photo
	if len(photos) == 0 {
		b.sendErrorMessage(chatID, "Failed to process the photo. Please try again.")
		return
	}

	largestPhoto := photos[len(photos)-1]

	// Get file URL
	fileConfig := tgbotapi.FileConfig{
		FileID: largestPhoto.FileID,
	}

	file, err := b.api.GetFile(fileConfig)
	if err != nil {
		b.logger.Error("Failed to get photo file",
			"error", err,
			"file_id", largestPhoto.FileID,
		)
		b.sendErrorMessage(chatID, "Failed to process the photo. Please try again.")
		return
	}

	photoURL := file.Link(b.api.Token)

	// Update user state
	state := b.userStates[user.TelegramID]
	state.PhotoURL = photoURL
	b.userStates[user.TelegramID] = state

	// Create flash card with photo
	b.createFlashCard(chatID, user, state)
}

func (b *Bot) createFlashCard(chatID int64, user *models.User, state UserState) {
	// Get definition
	definition, err := b.flashcardService.GetDefinition(state.SelectedDef)
	if err != nil {
		b.logger.Error("Failed to get definition for card creation",
			"error", err,
			"def_id", state.SelectedDef,
		)
		b.sendErrorMessage(chatID, "Failed to create flash card. Please try again.")
		return
	}

	// Get examples
	var exampleTexts []string
	for _, exID := range state.Examples {
		ex, err := b.flashcardService.GetExample(exID)
		if err != nil {
			b.logger.Warn("Failed to get example",
				"error", err,
				"ex_id", exID,
			)
			continue
		}
		exampleTexts = append(exampleTexts, ex.Text)
	}

	// Create flash card
	card := models.NewFlashCard(
		state.CurrentBank,
		state.CurrentWord,
		definition.Text,
		exampleTexts,
		state.PhotoURL,
	)

	// Save to database
	err = b.flashcardService.CreateFlashCard(card)
	if err != nil {
		b.logger.Error("Failed to save flash card",
			"error", err,
			"word", state.CurrentWord,
		)
		b.sendErrorMessage(chatID, "Failed to save flash card. Please try again.")
		return
	}

	// Update statistics
	err = b.statsService.IncrementLearned(user.ID, state.CurrentBank)
	if err != nil {
		b.logger.Warn("Failed to update statistics",
			"error", err,
			"user_id", user.ID,
			"bank_id", state.CurrentBank,
		)
	}

	// Clear user state
	delete(b.userStates, user.TelegramID)

	// Send confirmation
	text := fmt.Sprintf("✅ Flash card created for *%s*!\n\n*Definition:*\n%s",
		state.CurrentWord, definition.Text)

	if len(exampleTexts) > 0 {
		text += "\n\n*Examples:*"
		for i, ex := range exampleTexts {
			text += fmt.Sprintf("\n%d. %s", i+1, ex)
		}
	}

	if state.PhotoURL != "" {
		text += "\n\n*Context photo added!*"
	}

	text += "\n\nUse /review to practice your flash cards."

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"

	b.api.Send(msg)
}

// Other handlers (placeholder implementations)

// ReviewState represents the state of a review session
type ReviewState struct {
	Cards       []models.FlashCard
	CurrentCard int
	IsFlipped   bool
	BankID      int
}

func (b *Bot) handleReviewCommand(update tgbotapi.Update, user *models.User, args string) {
	chatID := update.Message.Chat.ID

	// Parse limit argument if provided
	limit := 10 // Default limit
	if args != "" {
		customLimit, err := strconv.Atoi(args)
		if err == nil && customLimit > 0 && customLimit <= 50 {
			limit = customLimit
		}
	}

	// Get user's active card bank
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	activeBankID := settings.Settings.ActiveCardBankID

	// Check if user has access to this bank
	hasAccess, err := b.cardbankService.UserHasAccess(user.ID, activeBankID)
	if err != nil || !hasAccess {
		b.logger.Error("User doesn't have access to active bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", activeBankID,
		)
		b.sendErrorMessage(chatID, "You don't have access to your active card bank. Please select another bank using /banks.")
		return
	}

	// Get due cards
	dueCards, err := b.spacedRepService.GetDueCards(user.ID, activeBankID, limit)
	if err != nil {
		b.logger.Error("Failed to get due cards",
			"error", err,
			"user_id", user.ID,
			"bank_id", activeBankID,
		)
		b.sendErrorMessage(chatID, "Failed to get cards for review. Please try again.")
		return
	}

	if len(dueCards) == 0 {
		b.sendMessage(chatID, "You don't have any cards due for review. Great job! 🎉")
		return
	}

	// Create review state
	reviewState := ReviewState{
		Cards:       dueCards,
		CurrentCard: 0,
		IsFlipped:   false,
		BankID:      activeBankID,
	}

	// Store review state in user state
	b.userStates[user.TelegramID] = UserState{
		State:       "reviewing",
		CurrentBank: activeBankID,
		ReviewState: &reviewState,
	}

	// Show first card
	b.showReviewCard(chatID, user, reviewState.Cards[0], false)
}

func (b *Bot) handleReviewCallback(update tgbotapi.Update, user *models.User, args []string) {
	chatID := update.CallbackQuery.Message.Chat.ID

	if len(args) < 1 {
		b.logger.Error("Invalid review callback data", "args", args)
		return
	}

	action := args[0]

	// Get user state
	state, exists := b.userStates[user.TelegramID]
	if !exists || state.State != "reviewing" || state.ReviewState == nil {
		b.sendErrorMessage(chatID, "Review session expired. Please start a new review with /review.")
		return
	}

	reviewState := state.ReviewState

	switch action {
	case "flip":
		// Flip the card
		reviewState.IsFlipped = true
		b.userStates[user.TelegramID] = state

		// Show the flipped card
		b.showReviewCard(chatID, user, reviewState.Cards[reviewState.CurrentCard], true)

	case "rate":
		if len(args) < 2 {
			b.logger.Error("Invalid rating in review callback", "args", args)
			return
		}

		// Parse rating
		rating, err := strconv.Atoi(args[1])
		if err != nil || rating < 0 || rating > 3 {
			b.logger.Error("Invalid rating value", "rating", args[1])
			return
		}

		// Process the review
		currentCard := reviewState.Cards[reviewState.CurrentCard]
		err = b.spacedRepService.ProcessReview(user.ID, currentCard.ID, rating)
		if err != nil {
			b.logger.Error("Failed to process review",
				"error", err,
				"user_id", user.ID,
				"card_id", currentCard.ID,
			)
			b.sendErrorMessage(chatID, "Failed to save your review. Please try again.")
			return
		}

		// Update statistics
		err = b.statsService.IncrementReviewed(user.ID, reviewState.BankID)
		if err != nil {
			b.logger.Warn("Failed to update statistics",
				"error", err,
				"user_id", user.ID,
				"bank_id", reviewState.BankID,
			)
		}

		// Update streak
		err = b.statsService.UpdateStreak(user.ID)
		if err != nil {
			b.logger.Warn("Failed to update streak",
				"error", err,
				"user_id", user.ID,
			)
		}

		// Show feedback based on rating
		var feedbackText string
		switch rating {
		case spaced_repetition.QualityAgain:
			feedbackText = "You'll see this card again soon."
		case spaced_repetition.QualityHard:
			feedbackText = "You'll see this card again in a short while."
		case spaced_repetition.QualityGood:
			feedbackText = "Good job! You'll see this card again later."
		case spaced_repetition.QualityEasy:
			feedbackText = "Excellent! You'll see this card again much later."
		}

		b.sendMessage(chatID, fmt.Sprintf("✅ Card reviewed: *%s*\n\n%s", currentCard.Word, feedbackText))

		// Move to the next card or finish the review
		reviewState.CurrentCard++
		reviewState.IsFlipped = false

		if reviewState.CurrentCard >= len(reviewState.Cards) {
			// Review session completed
			delete(b.userStates, user.TelegramID)

			// Get review stats
			totalCards, dueCards, err := b.spacedRepService.GetReviewStats(user.ID)
			if err != nil {
				b.logger.Error("Failed to get review stats",
					"error", err,
					"user_id", user.ID,
				)
				b.sendMessage(chatID, "🎉 Review session completed!")
				return
			}

			b.sendMessage(chatID, fmt.Sprintf("🎉 Review session completed!\n\nYou've reviewed %d cards.\nYou have %d cards in total, with %d cards due for review.",
				len(reviewState.Cards), totalCards, dueCards))
			return
		}

		// Show the next card
		b.userStates[user.TelegramID] = state
		b.showReviewCard(chatID, user, reviewState.Cards[reviewState.CurrentCard], false)
	}
}

func (b *Bot) handleStatsCommand(update tgbotapi.Update, user *models.User, args string) {
	chatID := update.Message.Chat.ID

	// Get user's statistics
	stats, err := b.statsService.GetUserStatistics(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user statistics",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your statistics. Please try again.")
		return
	}

	// Get user's active bank
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	activeBankID := settings.Settings.ActiveCardBankID

	// Get review stats
	totalCards, dueCards, err := b.spacedRepService.GetReviewStats(user.ID)
	if err != nil {
		b.logger.Error("Failed to get review stats",
			"error", err,
			"user_id", user.ID,
		)
		totalCards = 0
		dueCards = 0
	}

	// Build statistics message
	statsText := "📊 *Your Learning Statistics*\n\n"

	// Overall stats
	statsText += "*Overall:*\n"
	statsText += fmt.Sprintf("Total cards: %d\n", totalCards)
	statsText += fmt.Sprintf("Cards due for review: %d\n", dueCards)

	// Calculate totals
	var totalLearned, totalReviewed, longestStreak int
	for _, s := range stats {
		totalLearned += s.CardsLearned
		totalReviewed += s.CardsReviewed
		if s.StreakDays > longestStreak {
			longestStreak = s.StreakDays
		}
	}

	statsText += fmt.Sprintf("Total cards learned: %d\n", totalLearned)
	statsText += fmt.Sprintf("Total reviews: %d\n", totalReviewed)
	statsText += fmt.Sprintf("Current streak: %d days\n", longestStreak)

	// Per-bank stats
	if len(stats) > 0 {
		statsText += "\n*Card Banks:*\n"
		for _, s := range stats {
			// Get bank name
			bank, err := b.cardbankService.GetCardBank(s.CardBankID)
			if err != nil {
				continue
			}

			activeMarker := ""
			if s.CardBankID == activeBankID {
				activeMarker = " ✅"
			}

			statsText += fmt.Sprintf("*%s*%s\n", bank.Name, activeMarker)
			statsText += fmt.Sprintf("Cards learned: %d\n", s.CardsLearned)
			statsText += fmt.Sprintf("Reviews: %d\n\n", s.CardsReviewed)
		}
	}

	// Send statistics
	msg := tgbotapi.NewMessage(chatID, statsText)
	msg.ParseMode = "Markdown"

	b.api.Send(msg)
}

func (b *Bot) handleBanksCommand(update tgbotapi.Update, user *models.User) {
	chatID := update.Message.Chat.ID

	// Get user's card banks
	banks, err := b.cardbankService.GetUserCardBanks(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user's card banks",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your card banks. Please try again.")
		return
	}

	// Get user's active bank
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	activeBankID := settings.Settings.ActiveCardBankID

	// If user has no banks, create a default one
	if len(banks) == 0 {
		defaultBank, err := b.createDefaultBank(user)
		if err != nil {
			b.logger.Error("Failed to create default bank",
				"error", err,
				"user_id", user.ID,
			)
			b.sendErrorMessage(chatID, "Failed to create a default card bank. Please try again.")
			return
		}

		banks = append(banks, *defaultBank)

		// Set as active bank
		err = b.settingsService.SetActiveCardBank(user.ID, defaultBank.ID)
		if err != nil {
			b.logger.Error("Failed to set active bank",
				"error", err,
				"user_id", user.ID,
				"bank_id", defaultBank.ID,
			)
		}

		activeBankID = defaultBank.ID

		b.sendMessage(chatID, "Created a default card bank for you: \"My Cards\"")
	}

	// Show banks with keyboard
	var banksText string
	if len(banks) == 0 {
		banksText = "You don't have any card banks yet. Create one using the button below."
	} else {
		banksText = "📚 *Your Card Banks*\n\n"
		for i, bank := range banks {
			activeMarker := ""
			if bank.ID == activeBankID {
				activeMarker = " ✅ (Active)"
			}

			banksText += fmt.Sprintf("%d. *%s*%s\n", i+1, bank.Name, activeMarker)
			if bank.Description != "" {
				banksText += fmt.Sprintf("   %s\n", bank.Description)
			}
			banksText += fmt.Sprintf("   Cards: %d\n\n", b.countCardsInBank(bank.ID))
		}
	}

	msg := tgbotapi.NewMessage(chatID, banksText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = b.createBanksKeyboard(banks, 1, 1)

	b.api.Send(msg)
}

// createDefaultBank creates a default card bank for a user
func (b *Bot) createDefaultBank(user *models.User) (*models.CardBank, error) {
	return b.cardbankService.CreateCardBank("My Cards", "Default card bank", user.ID, false)
}

// countCardsInBank counts the number of cards in a bank
func (b *Bot) countCardsInBank(bankID int) int {
	cards, err := b.flashcardService.GetFlashCardsByBank(bankID)
	if err != nil {
		return 0
	}
	return len(cards)
}

func (b *Bot) handleBankCallback(update tgbotapi.Update, user *models.User, args []string) {
	chatID := update.CallbackQuery.Message.Chat.ID

	if len(args) < 1 {
		b.logger.Error("Invalid bank callback data", "args", args)
		return
	}

	action := args[0]

	switch action {
	case "select":
		// User selected a bank
		if len(args) < 2 {
			b.logger.Error("Invalid bank selection data", "args", args)
			return
		}

		bankID, err := strconv.Atoi(args[1])
		if err != nil {
			b.logger.Error("Invalid bank ID",
				"error", err,
				"bank_id", args[1],
			)
			return
		}

		// Check if user has access to this bank
		hasAccess, err := b.cardbankService.UserHasAccess(user.ID, bankID)
		if err != nil || !hasAccess {
			b.logger.Error("User doesn't have access to bank",
				"error", err,
				"user_id", user.ID,
				"bank_id", bankID,
			)
			b.sendErrorMessage(chatID, "You don't have access to this card bank.")
			return
		}

		// Set as active bank
		err = b.settingsService.SetActiveCardBank(user.ID, bankID)
		if err != nil {
			b.logger.Error("Failed to set active bank",
				"error", err,
				"user_id", user.ID,
				"bank_id", bankID,
			)
			b.sendErrorMessage(chatID, "Failed to set active card bank. Please try again.")
			return
		}

		// Get bank name
		bank, err := b.cardbankService.GetCardBank(bankID)
		if err != nil {
			b.logger.Error("Failed to get bank",
				"error", err,
				"bank_id", bankID,
			)
			b.sendMessage(chatID, "Card bank set as active.")
			return
		}

		b.sendMessage(chatID, fmt.Sprintf("Card bank \"%s\" is now active. You can add words to this bank.", bank.Name))

	case "create":
		// User wants to create a new bank
		b.userStates[user.TelegramID] = UserState{
			State: "awaiting_bank_name",
		}

		b.sendMessage(chatID, "Please send the name for your new card bank.")
	}
}

func (b *Bot) handleCreateBankCommand(update tgbotapi.Update, user *models.User, args string) {
	chatID := update.Message.Chat.ID

	// If no name provided in command, prompt user and set state
	if args == "" {
		b.userStates[user.TelegramID] = UserState{
			State: "awaiting_bank_name",
		}
		b.sendMessage(chatID, "Please send the name for your new card bank.")
		return
	}

	// Create bank with provided name
	bank, err := b.cardbankService.CreateCardBank(args, "", user.ID, false)
	if err != nil {
		b.logger.Error("Failed to create card bank",
			"error", err,
			"user_id", user.ID,
			"name", args,
		)
		b.sendErrorMessage(chatID, "Failed to create card bank. Please try again.")
		return
	}

	// Set as active bank
	err = b.settingsService.SetActiveCardBank(user.ID, bank.ID)
	if err != nil {
		b.logger.Error("Failed to set active bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", bank.ID,
		)
	}

	b.sendMessage(chatID, fmt.Sprintf("Card bank \"%s\" created and set as active. You can now add words to this bank.", bank.Name))
}

func (b *Bot) handleBankNameInput(update tgbotapi.Update, user *models.User, text string) {
	chatID := update.Message.Chat.ID

	// Create bank with provided name
	bank, err := b.cardbankService.CreateCardBank(text, "", user.ID, false)
	if err != nil {
		b.logger.Error("Failed to create card bank",
			"error", err,
			"user_id", user.ID,
			"name", text,
		)
		b.sendErrorMessage(chatID, "Failed to create card bank. Please try again.")
		return
	}

	// Clear user state
	delete(b.userStates, user.TelegramID)

	// Set as active bank
	err = b.settingsService.SetActiveCardBank(user.ID, bank.ID)
	if err != nil {
		b.logger.Error("Failed to set active bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", bank.ID,
		)
	}

	b.sendMessage(chatID, fmt.Sprintf("Card bank \"%s\" created and set as active. You can now add words to this bank.", bank.Name))
}

func (b *Bot) handleShareBankCommand(update tgbotapi.Update, user *models.User, args string) {
	chatID := update.Message.Chat.ID

	// Check if username is provided
	if args == "" {
		b.sendMessage(chatID, "Please provide a username to share with: /share_bank username")
		return
	}

	// Get target username
	targetUsername := strings.TrimSpace(args)
	if targetUsername[0] == '@' {
		targetUsername = targetUsername[1:] // Remove @ if present
	}

	// Get user's active bank
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	activeBankID := settings.Settings.ActiveCardBankID

	// Check if user has access to this bank
	hasAccess, err := b.cardbankService.UserHasAccess(user.ID, activeBankID)
	if err != nil || !hasAccess {
		b.logger.Error("User doesn't have access to active bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", activeBankID,
		)
		b.sendErrorMessage(chatID, "You don't have access to your active card bank. Please select another bank using /banks.")
		return
	}

	// Get bank details
	bank, err := b.cardbankService.GetCardBank(activeBankID)
	if err != nil {
		b.logger.Error("Failed to get bank",
			"error", err,
			"bank_id", activeBankID,
		)
		b.sendErrorMessage(chatID, "Failed to get bank details. Please try again.")
		return
	}

	// Generate a share code (we'll use the bank ID for simplicity)
	shareCode := fmt.Sprintf("%d", bank.ID)

	// Send share message
	shareText := fmt.Sprintf("You've shared your card bank \"%s\" with @%s.\n\nThey can join using the command:\n/join_bank %s",
		bank.Name, targetUsername, shareCode)

	b.sendMessage(chatID, shareText)

	// Note: In a real implementation, we would store the share information in the database
	// and validate that the target user exists before sharing.
}

func (b *Bot) handleJoinBankCommand(update tgbotapi.Update, user *models.User, args string) {
	chatID := update.Message.Chat.ID

	// Check if code is provided
	if args == "" {
		b.sendMessage(chatID, "Please provide a share code to join a bank: /join_bank code")
		return
	}

	// Parse the share code (bank ID)
	bankID, err := strconv.Atoi(strings.TrimSpace(args))
	if err != nil {
		b.logger.Error("Invalid bank ID",
			"error", err,
			"code", args,
		)
		b.sendErrorMessage(chatID, "Invalid share code. Please check and try again.")
		return
	}

	// Check if bank exists
	bank, err := b.cardbankService.GetCardBank(bankID)
	if err != nil {
		b.logger.Error("Failed to get bank",
			"error", err,
			"bank_id", bankID,
		)
		b.sendErrorMessage(chatID, "Card bank not found. Please check the share code and try again.")
		return
	}

	// Check if user already has access
	hasAccess, err := b.cardbankService.UserHasAccess(user.ID, bankID)
	if err == nil && hasAccess {
		b.sendMessage(chatID, fmt.Sprintf("You already have access to the card bank \"%s\".", bank.Name))
		return
	}

	// Add user to bank with viewer role
	err = b.cardbankService.AddUserToBank(user.ID, bankID, "viewer")
	if err != nil {
		b.logger.Error("Failed to add user to bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", bankID,
		)
		b.sendErrorMessage(chatID, "Failed to join card bank. Please try again.")
		return
	}

	// Set as active bank
	err = b.settingsService.SetActiveCardBank(user.ID, bankID)
	if err != nil {
		b.logger.Error("Failed to set active bank",
			"error", err,
			"user_id", user.ID,
			"bank_id", bankID,
		)
	}

	b.sendMessage(chatID, fmt.Sprintf("You've joined the card bank \"%s\" and it's now your active bank.", bank.Name))
}

func (b *Bot) handleSettingsCommand(update tgbotapi.Update, user *models.User) {
	chatID := update.Message.Chat.ID

	// Get user's settings
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	// Get active bank name
	activeBankName := "None"
	if settings.Settings.ActiveCardBankID != 0 {
		bank, err := b.cardbankService.GetCardBank(settings.Settings.ActiveCardBankID)
		if err == nil {
			activeBankName = bank.Name
		}
	}

	// Build settings message
	settingsText := "⚙️ *Your Settings*\n\n"
	settingsText += fmt.Sprintf("*Active Card Bank:* %s\n", activeBankName)
	settingsText += fmt.Sprintf("*Review Limit:* %d cards per session\n", settings.Settings.ReviewLimit)

	notificationsStatus := "Off"
	if settings.Settings.NotificationsOn {
		notificationsStatus = "On"
	}
	settingsText += fmt.Sprintf("*Notifications:* %s\n", notificationsStatus)

	darkModeStatus := "Off"
	if settings.Settings.DarkMode {
		darkModeStatus = "On"
	}
	settingsText += fmt.Sprintf("*Dark Mode:* %s\n", darkModeStatus)

	// Send settings with keyboard
	msg := tgbotapi.NewMessage(chatID, settingsText)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = b.createSettingsKeyboard(settings)

	b.api.Send(msg)
}

func (b *Bot) handleSettingsCallback(update tgbotapi.Update, user *models.User, args []string) {
	chatID := update.CallbackQuery.Message.Chat.ID

	if len(args) < 1 {
		b.logger.Error("Invalid settings callback data", "args", args)
		return
	}

	action := args[0]

	// Get user's settings
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	switch action {
	case "bank":
		// Show banks to select active bank
		b.handleBanksCommand(update, user)

	case "limit":
		// Set review limit
		b.userStates[user.TelegramID] = UserState{
			State:         "awaiting_settings",
			SettingsField: "review_limit",
		}

		b.sendMessage(chatID, "Please enter the number of cards you want to review per session (1-50):")

	case "notifications":
		// Toggle notifications
		settings.Settings.NotificationsOn = !settings.Settings.NotificationsOn

		err = b.settingsService.UpdateUserSettings(user.ID, settings.Settings)
		if err != nil {
			b.logger.Error("Failed to update settings",
				"error", err,
				"user_id", user.ID,
			)
			b.sendErrorMessage(chatID, "Failed to update settings. Please try again.")
			return
		}

		status := "disabled"
		if settings.Settings.NotificationsOn {
			status = "enabled"
		}

		b.sendMessage(chatID, fmt.Sprintf("Notifications %s.", status))

		// Show updated settings
		b.handleSettingsCommand(update, user)

	case "darkmode":
		// Toggle dark mode
		settings.Settings.DarkMode = !settings.Settings.DarkMode

		err = b.settingsService.UpdateUserSettings(user.ID, settings.Settings)
		if err != nil {
			b.logger.Error("Failed to update settings",
				"error", err,
				"user_id", user.ID,
			)
			b.sendErrorMessage(chatID, "Failed to update settings. Please try again.")
			return
		}

		status := "disabled"
		if settings.Settings.DarkMode {
			status = "enabled"
		}

		b.sendMessage(chatID, fmt.Sprintf("Dark mode %s.", status))

		// Show updated settings
		b.handleSettingsCommand(update, user)
	}
}

func (b *Bot) handleSettingsInput(update tgbotapi.Update, user *models.User, text string) {
	chatID := update.Message.Chat.ID

	// Get user state
	state, exists := b.userStates[user.TelegramID]
	if !exists || state.State != "awaiting_settings" {
		b.sendErrorMessage(chatID, "Session expired. Please start again with /settings.")
		return
	}

	// Get user's settings
	settings, err := b.settingsService.GetUserSettings(user.ID)
	if err != nil {
		b.logger.Error("Failed to get user settings",
			"error", err,
			"user_id", user.ID,
		)
		b.sendErrorMessage(chatID, "Failed to get your settings. Please try again.")
		return
	}

	// Process input based on settings field
	switch state.SettingsField {
	case "review_limit":
		// Parse review limit
		limit, err := strconv.Atoi(text)
		if err != nil || limit < 1 || limit > 50 {
			b.sendErrorMessage(chatID, "Please enter a valid number between 1 and 50.")
			return
		}

		// Update settings
		settings.Settings.ReviewLimit = limit

		err = b.settingsService.UpdateUserSettings(user.ID, settings.Settings)
		if err != nil {
			b.logger.Error("Failed to update settings",
				"error", err,
				"user_id", user.ID,
			)
			b.sendErrorMessage(chatID, "Failed to update settings. Please try again.")
			return
		}

		// Clear user state
		delete(b.userStates, user.TelegramID)

		b.sendMessage(chatID, fmt.Sprintf("Review limit set to %d cards per session.", limit))

		// Show updated settings
		msg := update.Message
		update.Message = nil
		update.CallbackQuery = &tgbotapi.CallbackQuery{
			Message: msg,
		}
		b.handleSettingsCommand(update, user)
	}
}

func (b *Bot) handlePaginationCallback(update tgbotapi.Update, user *models.User, args []string) {
	// Placeholder implementation
}

func (b *Bot) handleAdminCommand(update tgbotapi.Update, user *models.User, args string) {
	// Placeholder implementation
	chatID := update.Message.Chat.ID
	b.sendMessage(chatID, "Admin functionality is coming soon!")
}

func (b *Bot) handleAdminInput(update tgbotapi.Update, user *models.User, text string) {
	// Placeholder implementation
}

// showReviewCard shows a flash card for review
func (b *Bot) showReviewCard(chatID int64, user *models.User, card models.FlashCard, isFlipped bool) {
	var text string

	if isFlipped {
		// Show the word (answer)
		text = fmt.Sprintf("📝 *%s*\n\n*Definition:*\n%s", card.Word, card.Definition)

		if len(card.Examples) > 0 {
			text += "\n\n*Examples:*"
			for i, ex := range card.Examples {
				text += fmt.Sprintf("\n%d. %s", i+1, ex)
			}
		}

		text += "\n\nHow well did you remember this word?"
	} else {
		// Show the definition and examples (question)
		text = fmt.Sprintf("*Definition:*\n%s", card.Definition)

		if len(card.Examples) > 0 {
			text += "\n\n*Examples:*"
			for i, ex := range card.Examples {
				text += fmt.Sprintf("\n%d. %s", i+1, ex)
			}
		}
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = b.createReviewKeyboard(isFlipped)

	b.api.Send(msg)

	// If there's an image and the card is flipped, send it
	if card.ImageURL != "" && isFlipped {
		photo := tgbotapi.NewPhoto(chatID, tgbotapi.FileURL(card.ImageURL))
		photo.Caption = "Context image for: " + card.Word
		b.api.Send(photo)
	}
}
