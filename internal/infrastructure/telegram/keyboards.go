package telegram

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/services"
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
