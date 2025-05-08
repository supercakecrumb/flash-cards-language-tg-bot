# File Structure

This document outlines the file structure for the Flash Cards Language Telegram Bot project. Each file's purpose and responsibilities are described to guide implementation.

```
flash-cards-language-tg-bot/
├── cmd/
│   └── bot/
│       └── main.go                  # Application entry point
├── internal/
│   ├── app/
│   │   ├── app.go                   # Application initialization and wiring
│   │   └── config.go                # Configuration loading and validation
│   ├── domain/
│   │   ├── models/                  # Domain entities
│   │   │   ├── user.go              # User model
│   │   │   ├── flashcard.go         # Flash card model
│   │   │   ├── cardbank.go          # Card bank model
│   │   │   ├── review.go            # Review model for spaced repetition
│   │   │   ├── statistics.go        # Statistics model
│   │   │   └── settings.go          # User settings model
│   │   └── services/                # Business logic services
│   │       ├── user_service.go      # User management service
│   │       ├── flashcard_service.go # Flash card management service
│   │       ├── cardbank_service.go  # Card bank management service
│   │       ├── spaced_repetition_service.go # Spaced repetition algorithm service
│   │       ├── statistics_service.go # Statistics tracking service
│   │       ├── settings_service.go  # User settings service
│   │       └── admin_service.go     # Admin functionality service
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── postgres.go          # PostgreSQL connection and management
│   │   │   └── migrations/          # Database migration files
│   │   ├── dictionary/
│   │   │   ├── dictionary.go        # Dictionary service interface
│   │   │   ├── free_dictionary.go   # Free Dictionary API implementation
│   │   │   └── mock_dictionary.go   # Mock implementation for testing
│   │   ├── telegram/
│   │   │   ├── bot.go               # Telegram bot initialization and update handling
│   │   │   ├── handlers.go          # Command, callback, and message handlers
│   │   │   └── keyboards.go         # Inline keyboard builders
│   │   └── logging/
│   │       └── logger.go            # Structured logging setup
│   └── repository/                  # Data access layer
│       ├── user_repository.go       # User data access
│       ├── flashcard_repository.go  # Flash card data access
│       ├── cardbank_repository.go   # Card bank data access
│       ├── review_repository.go     # Review data access
│       ├── statistics_repository.go # Statistics data access
│       └── settings_repository.go   # Settings data access
├── pkg/
│   ├── utils/
│   │   └── helpers.go               # Utility functions
│   └── spaced_repetition/
│       ├── algorithm.go             # Spaced repetition algorithm interface
│       └── sm2.go                   # SM-2 algorithm implementation
├── .env.example                     # Example environment variables
├── .gitignore                       # Git ignore file
├── docker-compose.yml               # Docker Compose configuration
├── Dockerfile                       # Docker build configuration
├── go.mod                           # Go module definition
├── go.sum                           # Go module checksums
└── README.md                        # Project documentation