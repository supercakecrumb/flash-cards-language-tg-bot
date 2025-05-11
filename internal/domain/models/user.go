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
