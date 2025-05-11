package database

import (
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// Config represents database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
}

// PostgresDB represents a PostgreSQL database connection
type PostgresDB struct {
	db     *sqlx.DB
	logger *slog.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection
func NewPostgresDB(config Config, logger *slog.Logger) (*PostgresDB, error) {
	logger.Info("Connecting to PostgreSQL database",
		"host", config.Host,
		"port", config.Port,
		"user", config.User,
		"database", config.Name,
	)

	// Create connection string
	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Name,
	)

	// Connect to database
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		logger.Error("Failed to connect to database", "error", err)
		return nil, err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	// Run migrations
	if err := runMigrations(db.DB, config.Name, logger); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		return nil, err
	}

	return &PostgresDB{
		db:     db,
		logger: logger,
	}, nil
}

// DB returns the underlying sqlx.DB instance
func (p *PostgresDB) DB() *sqlx.DB {
	return p.db
}

// Close closes the database connection
func (p *PostgresDB) Close() error {
	p.logger.Info("Closing database connection")
	return p.db.Close()
}

// runMigrations runs database migrations
func runMigrations(db *sql.DB, dbName string, logger *slog.Logger) error {
	logger.Info("Running database migrations")

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://internal/infrastructure/database/migrations",
		dbName,
		driver,
	)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Info("Database migrations completed successfully")
	return nil
}
