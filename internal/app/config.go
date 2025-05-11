package app

import (
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config represents the application configuration
type Config struct {
	Telegram struct {
		Token string
	}
	Database struct {
		Host     string
		Port     int
		User     string
		Password string
		Name     string
	}
	Dictionary struct {
		Provider string
		APIKey   string
	}
	AdminIDs []int64
	LogLevel string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{}

	// Telegram configuration
	config.Telegram.Token = os.Getenv("TELEGRAM_TOKEN")
	if config.Telegram.Token == "" {
		return nil, errors.New("TELEGRAM_TOKEN environment variable is required")
	}

	// Database configuration
	config.Database.Host = os.Getenv("DB_HOST")
	if config.Database.Host == "" {
		config.Database.Host = "localhost"
	}

	dbPort, err := strconv.Atoi(os.Getenv("DB_PORT"))
	if err != nil {
		config.Database.Port = 5432 // Default PostgreSQL port
	} else {
		config.Database.Port = dbPort
	}

	config.Database.User = os.Getenv("DB_USER")
	if config.Database.User == "" {
		return nil, errors.New("DB_USER environment variable is required")
	}

	config.Database.Password = os.Getenv("DB_PASSWORD")
	if config.Database.Password == "" {
		return nil, errors.New("DB_PASSWORD environment variable is required")
	}

	config.Database.Name = os.Getenv("DB_NAME")
	if config.Database.Name == "" {
		return nil, errors.New("DB_NAME environment variable is required")
	}

	// Dictionary configuration
	config.Dictionary.Provider = os.Getenv("DICTIONARY_API")
	if config.Dictionary.Provider == "" {
		config.Dictionary.Provider = "freedictionary"
	}

	config.Dictionary.APIKey = os.Getenv("DICTIONARY_API_KEY")

	// Admin IDs
	adminIDsStr := os.Getenv("ADMIN_IDS")
	if adminIDsStr != "" {
		adminIDsSlice := strings.Split(adminIDsStr, ",")
		for _, idStr := range adminIDsSlice {
			id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64)
			if err != nil {
				return nil, errors.New("invalid admin ID format in ADMIN_IDS")
			}
			config.AdminIDs = append(config.AdminIDs, id)
		}
	}

	// Log level
	config.LogLevel = os.Getenv("LOG_LEVEL")
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}

	return config, nil
}
