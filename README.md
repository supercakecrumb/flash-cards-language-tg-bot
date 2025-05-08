# Flash Cards Language Telegram Bot

A Telegram bot for automatic creation and review of language flash cards, built with Go.

![License](https://img.shields.io/github/license/username/flash-cards-language-tg-bot)
![Go Version](https://img.shields.io/github/go-mod/go-version/username/flash-cards-language-tg-bot)
![Build Status](https://img.shields.io/github/workflow/status/username/flash-cards-language-tg-bot/Build%20and%20Deploy)

## Features

- **Automatic Flash Card Creation**: Send a word to the bot, and it will fetch definitions and examples from a dictionary API
- **Spaced Repetition System**: Review cards using an Anki-like spaced repetition algorithm (SM-2)
- **Card Banks**: Organize your flash cards into different collections
- **Sharing**: Share your card banks with other users
- **Group Chat Support**: Add the bot to group chats for collaborative card creation
- **Statistics**: Track your learning progress
- **Customizable Settings**: Adjust the bot's behavior to your preferences
- **Admin Controls**: Restrict access to the bot

## How It Works

### Adding a Word

1. Send a word to the bot
2. The bot fetches definitions from a dictionary API
3. Choose a definition by clicking a button
4. (Optional) Send a photo to add context
5. The bot shows examples of usage for the chosen definition
6. Choose which examples to include
7. The flash card is created!

### Reviewing Words

1. Use the `/review` command to start a review session
2. The bot shows the definition and examples
3. Flip the card to see the word
4. Rate your recall (Again/Hard/Good/Easy)
5. The spaced repetition algorithm schedules the next review

## Installation

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL (or use the Docker Compose setup)
- A Telegram Bot Token (get one from [@BotFather](https://t.me/BotFather))

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/username/flash-cards-language-tg-bot.git
   cd flash-cards-language-tg-bot
   ```

2. Create a `.env` file based on `.env.example`:
   ```bash
   cp .env.example .env
   ```

3. Edit the `.env` file with your configuration.

4. Start the services using Docker Compose:
   ```bash
   docker-compose up -d
   ```

5. Check the logs:
   ```bash
   docker-compose logs -f bot
   ```

### Production Deployment

See the [Deployment Documentation](docs/deployment.md) for detailed instructions.

## Usage

### Basic Commands

- `/start` - Start the bot and get a welcome message
- `/help` - Show help information
- `/add [word]` - Add a word as a flash card
- `/review` - Start a review session
- `/stats` - View your learning statistics
- `/banks` - Manage your card banks
- `/settings` - Configure your preferences

### Card Banks

- `/create_bank [name]` - Create a new card bank
- `/share_bank [username]` - Share a bank with another user
- `/join_bank [code]` - Join a shared card bank

### Admin Commands

- `/admin` - Access admin features (restricted to admin users)

## Architecture

The bot is built with a clean architecture approach, separating concerns into different layers:

- **Domain Layer**: Core business logic and entities
- **Application Layer**: Orchestration and use cases
- **Infrastructure Layer**: External integrations (Telegram, database, dictionary API)
- **Repository Layer**: Data access and persistence

For more details, see the [Architecture Documentation](docs/architecture.md).

## Project Structure

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
│   │   └── services/                # Business logic services
│   ├── infrastructure/
│   │   ├── database/                # Database connection and migrations
│   │   ├── dictionary/              # Dictionary API client
│   │   ├── telegram/                # Telegram bot implementation
│   │   └── logging/                 # Structured logging
│   └── repository/                  # Data access layer
├── pkg/
│   ├── utils/                       # Utility functions
│   └── spaced_repetition/           # Spaced repetition algorithm
├── docs/                            # Documentation
├── .env.example                     # Example environment variables
├── docker-compose.yml               # Docker Compose configuration
├── Dockerfile                       # Docker build configuration
└── README.md                        # Project documentation
```

For more details, see the [File Structure Documentation](docs/file_structure.md).

## Documentation

- [Architecture](docs/architecture.md) - System architecture and design
- [File Structure](docs/file_structure.md) - Project file organization
- [Main Application](docs/main_application.md) - Entry point and application setup
- [Domain Models](docs/domain_models.md) - Core entities and data structures
- [Domain Services](docs/domain_services.md) - Business logic services
- [Infrastructure](docs/infrastructure.md) - External integrations
- [Repositories](docs/repositories.md) - Data access layer
- [Spaced Repetition](docs/spaced_repetition.md) - Spaced repetition algorithm
- [Deployment](docs/deployment.md) - Deployment instructions

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Free Dictionary API](https://dictionaryapi.dev/)
- [SuperMemo 2 Algorithm](https://www.supermemo.com/en/archives1990-2015/english/ol/sm2)
- [Go Telegram Bot API](https://github.com/go-telegram-bot-api/telegram-bot-api)
