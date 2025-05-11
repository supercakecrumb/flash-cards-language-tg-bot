package repository

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/supercakecrumb/flash-cards-language-tg-bot/internal/domain/models"
)

// SettingsRepository defines the interface for settings data access
type SettingsRepository interface {
	Create(settings *models.Settings) error
	GetByID(settingsID int) (*models.Settings, error)
	GetByUserID(userID int) (*models.Settings, error)
	Update(settings *models.Settings) error
	Delete(settingsID int) error
}

// settingsRepository implements the SettingsRepository interface
type settingsRepository struct {
	db *sqlx.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *sqlx.DB) SettingsRepository {
	return &settingsRepository{
		db: db,
	}
}

// Create creates new settings
func (r *settingsRepository) Create(settings *models.Settings) error {
	query := `
		INSERT INTO user_settings (user_id, settings, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		settings.UserID,
		settings.Settings,
		settings.CreatedAt,
		settings.UpdatedAt,
	).Scan(&settings.ID)

	return err
}

// GetByID retrieves settings by ID
func (r *settingsRepository) GetByID(settingsID int) (*models.Settings, error) {
	query := `
		SELECT id, user_id, settings, created_at, updated_at
		FROM user_settings
		WHERE id = $1
	`

	var settings models.Settings
	err := r.db.Get(&settings, query, settingsID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &settings, nil
}

// GetByUserID retrieves settings by user ID
func (r *settingsRepository) GetByUserID(userID int) (*models.Settings, error) {
	query := `
		SELECT id, user_id, settings, created_at, updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	var settings models.Settings
	err := r.db.Get(&settings, query, userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return &settings, nil
}

// Update updates existing settings
func (r *settingsRepository) Update(settings *models.Settings) error {
	query := `
		UPDATE user_settings
		SET settings = $1, updated_at = $2
		WHERE id = $3
	`

	settings.UpdatedAt = time.Now()

	_, err := r.db.Exec(
		query,
		settings.Settings,
		settings.UpdatedAt,
		settings.ID,
	)

	return err
}

// Delete deletes settings
func (r *settingsRepository) Delete(settingsID int) error {
	query := `DELETE FROM user_settings WHERE id = $1`
	_, err := r.db.Exec(query, settingsID)
	return err
}
