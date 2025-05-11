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
