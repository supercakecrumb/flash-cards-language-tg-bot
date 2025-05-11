package models

import (
	"time"
)

// CardBank represents a collection of flash cards
type CardBank struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	OwnerID     int       `db:"owner_id"`
	IsPublic    bool      `db:"is_public"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// BankMembership represents a user's membership in a card bank
type BankMembership struct {
	ID         int       `db:"id"`
	UserID     int       `db:"user_id"`
	CardBankID int       `db:"card_bank_id"`
	Role       string    `db:"role"` // "owner", "editor", "viewer"
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
}

// GroupChat represents a Telegram group chat linked to a card bank
type GroupChat struct {
	ID             int       `db:"id"`
	TelegramChatID int64     `db:"telegram_chat_id"`
	Title          string    `db:"title"`
	CardBankID     int       `db:"card_bank_id"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

// NewCardBank creates a new card bank
func NewCardBank(name, description string, ownerID int, isPublic bool) *CardBank {
	now := time.Now()
	return &CardBank{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
		IsPublic:    isPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// NewBankMembership creates a new bank membership
func NewBankMembership(userID, cardBankID int, role string) *BankMembership {
	now := time.Now()
	return &BankMembership{
		UserID:     userID,
		CardBankID: cardBankID,
		Role:       role,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// NewGroupChat creates a new group chat
func NewGroupChat(telegramChatID int64, title string, cardBankID int) *GroupChat {
	now := time.Now()
	return &GroupChat{
		TelegramChatID: telegramChatID,
		Title:          title,
		CardBankID:     cardBankID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}
