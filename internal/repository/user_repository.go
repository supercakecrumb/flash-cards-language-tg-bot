package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(user *models.User) error
	GetByID(userID int) (*models.User, error)
	GetByTelegramID(telegramID int64) (*models.User, error)
	Update(user *models.User) error
	Delete(userID int) error
	IsAdmin(telegramID int64) bool
}

// userRepository implements the UserRepository interface
type userRepository struct {
	db *sqlx.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user
func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (telegram_id, username, first_name, last_name, created_at, updated_at, is_admin)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		user.TelegramID,
		user.Username,
		user.FirstName,
		user.LastName,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsAdmin,
	).Scan(&user.ID)

	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(userID int) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, created_at, updated_at, is_admin
		FROM users
		WHERE id = $1
	`

	var user models.User
	err := r.db.Get(&user, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetByTelegramID retrieves a user by Telegram ID
func (r *userRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	query := `
		SELECT id, telegram_id, username, first_name, last_name, created_at, updated_at, is_admin
		FROM users
		WHERE telegram_id = $1
	`

	var user models.User
	err := r.db.Get(&user, query, telegramID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &user, nil
}

// Update updates an existing user
func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users
		SET username = $1, first_name = $2, last_name = $3, updated_at = $4, is_admin = $5
		WHERE id = $6
	`

	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		user.Username,
		user.FirstName,
		user.LastName,
		user.UpdatedAt,
		user.IsAdmin,
		user.ID,
	)

	return err
}

// Delete deletes a user
func (r *userRepository) Delete(userID int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.Exec(query, userID)
	return err
}

// IsAdmin checks if a user is an admin
func (r *userRepository) IsAdmin(telegramID int64) bool {
	query := `SELECT is_admin FROM users WHERE telegram_id = $1`

	var isAdmin bool
	err := r.db.Get(&isAdmin, query, telegramID)
	if err != nil {
		return false
	}

	return isAdmin
}
