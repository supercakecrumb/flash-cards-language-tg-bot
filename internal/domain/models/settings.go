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
