# Domain Models

This document describes the domain models for the Flash Cards Language Telegram Bot. These models represent the core entities in the system.

## User Model (internal/domain/models/user.go)

The User model represents a Telegram user who interacts with the bot.

```go
package models

import (
	"time"
)

// User represents a Telegram user
type User struct {
	ID         int       `db:"id"`
	TelegramID int64     `db:"telegram_id"`
	Username   string    `db:"username"`
	FirstName  string    `db:"first_name"`
	LastName   string    `db:"last_name"`
	CreatedAt  time.Time `db:"created_at"`
	UpdatedAt  time.Time `db:"updated_at"`
	IsAdmin    bool      `db:"is_admin"`
}

// NewUser creates a new user from Telegram user data
func NewUser(telegramID int64, username, firstName, lastName string) *User {
	now := time.Now()
	return &User{
		TelegramID: telegramID,
		Username:   username,
		FirstName:  firstName,
		LastName:   lastName,
		CreatedAt:  now,
		UpdatedAt:  now,
		IsAdmin:    false,
	}
}
```

## Flash Card Model (internal/domain/models/flashcard.go)

The FlashCard model represents a vocabulary flash card with a word, definition, examples, and optional image.

```go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringArray is a custom type for storing string arrays in JSON format
type StringArray []string

// Value implements the driver.Valuer interface for StringArray
func (a StringArray) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Scan implements the sql.Scanner interface for StringArray
func (a *StringArray) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, a)
}

// FlashCard represents a vocabulary flash card
type FlashCard struct {
	ID         int         `db:"id"`
	CardBankID int         `db:"card_bank_id"`
	Word       string      `db:"word"`
	Definition string      `db:"definition"`
	Examples   StringArray `db:"examples"`
	ImageURL   string      `db:"image_url"`
	CreatedAt  time.Time   `db:"created_at"`
	UpdatedAt  time.Time   `db:"updated_at"`
}

// NewFlashCard creates a new flash card
func NewFlashCard(cardBankID int, word, definition string, examples []string, imageURL string) *FlashCard {
	now := time.Now()
	return &FlashCard{
		CardBankID: cardBankID,
		Word:       word,
		Definition: definition,
		Examples:   examples,
		ImageURL:   imageURL,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
```

## Card Bank Model (internal/domain/models/cardbank.go)

The CardBank model represents a collection of flash cards that can be shared among users.

```go
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
	ID            int       `db:"id"`
	TelegramChatID int64    `db:"telegram_chat_id"`
	Title         string    `db:"title"`
	CardBankID    int       `db:"card_bank_id"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
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
```

## Review Model (internal/domain/models/review.go)

The Review model represents a user's review of a flash card, used for the spaced repetition system.

```go
package models

import (
	"time"
)

// Review represents a user's review of a flash card
type Review struct {
	ID           int       `db:"id"`
	UserID       int       `db:"user_id"`
	FlashCardID  int       `db:"flash_card_id"`
	EaseFactor   float64   `db:"ease_factor"`
	DueDate      time.Time `db:"due_date"`
	Interval     int       `db:"interval"`      // in days
	Repetitions  int       `db:"repetitions"`   // number of times reviewed
	LastReviewed time.Time `db:"last_reviewed"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// NewReview creates a new review
func NewReview(userID, flashCardID int) *Review {
	now := time.Now()
	return &Review{
		UserID:       userID,
		FlashCardID:  flashCardID,
		EaseFactor:   2.5, // Default ease factor
		DueDate:      now, // Due immediately
		Interval:     0,
		Repetitions:  0,
		LastReviewed: time.Time{}, // Zero time
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}
```

## Statistics Model (internal/domain/models/statistics.go)

The Statistics model tracks a user's learning progress.

```go
package models

import (
	"time"
)

// Statistics represents a user's learning statistics
type Statistics struct {
	ID            int       `db:"id"`
	UserID        int       `db:"user_id"`
	CardBankID    int       `db:"card_bank_id"`
	CardsReviewed int       `db:"cards_reviewed"`
	CardsLearned  int       `db:"cards_learned"`
	StreakDays    int       `db:"streak_days"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}

// NewStatistics creates new statistics for a user and card bank
func NewStatistics(userID, cardBankID int) *Statistics {
	now := time.Now()
	return &Statistics{
		UserID:        userID,
		CardBankID:    cardBankID,
		CardsReviewed: 0,
		CardsLearned:  0,
		StreakDays:    0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}
```

## Settings Model (internal/domain/models/settings.go)

The Settings model stores user preferences.

```go
package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// SettingsData represents user settings data stored as JSON
type SettingsData struct {
	ActiveCardBankID int    `json:"active_card_bank_id"`
	ReviewLimit      int    `json:"review_limit"`
	Language         string `json:"language"`
	NotificationsOn  bool   `json:"notifications_on"`
	DarkMode         bool   `json:"dark_mode"`
	// Add more settings as needed
}

// Value implements the driver.Valuer interface for SettingsData
func (s SettingsData) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements the sql.Scanner interface for SettingsData
func (s *SettingsData) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, s)
}

// Settings represents user settings
type Settings struct {
	ID        int          `db:"id"`
	UserID    int          `db:"user_id"`
	Settings  SettingsData `db:"settings"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
}

// NewSettings creates new settings for a user with default values
func NewSettings(userID int) *Settings {
	now := time.Now()
	return &Settings{
		UserID: userID,
		Settings: SettingsData{
			ReviewLimit:     20,
			Language:        "en",
			NotificationsOn: true,
			DarkMode:        false,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
```

These domain models form the core entities of the system and are used throughout the application for data storage, business logic, and presentation.