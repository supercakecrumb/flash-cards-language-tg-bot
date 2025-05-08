# Flash Cards Language Telegram Bot - Architecture

## System Overview

The Flash Cards Language Telegram Bot is a Golang application that helps users learn English vocabulary through a spaced repetition system similar to Anki, but within Telegram. The bot allows users to create flash cards with definitions and examples, review them using spaced repetition, track statistics, share card banks, and collaborate in group chats.

```mermaid
graph TD
    User[User] -->|Interacts with| TelegramAPI[Telegram API]
    TelegramAPI -->|Sends messages to| Bot[Flash Cards Bot]
    Bot -->|Responds to| TelegramAPI
    Bot -->|Queries| DictionaryService[Dictionary Service]
    Bot -->|Stores/Retrieves data| Database[(PostgreSQL Database)]
    Bot -->|Logs to| Logging[Structured Logging]
    
    subgraph "Bot Components"
        CommandHandler[Command Handler]
        FlashCardManager[Flash Card Manager]
        SpacedRepetitionEngine[Spaced Repetition Engine]
        UserManager[User Manager]
        CardBankManager[Card Bank Manager]
        StatisticsManager[Statistics Manager]
        SettingsManager[Settings Manager]
        AdminManager[Admin Manager]
    end
```

## Core Components

### 1. Application Layer

```mermaid
graph TD
    Main[Main Application] --> Config[Configuration]
    Main --> TelegramBot[Telegram Bot Handler]
    Main --> DatabaseConn[Database Connection]
    Main --> DictionaryClient[Dictionary Client]
    Main --> Logger[Logger]
    
    TelegramBot --> CommandRouter[Command Router]
    CommandRouter --> CommandHandlers[Command Handlers]
    CommandRouter --> CallbackHandlers[Callback Handlers]
    CommandRouter --> MessageHandlers[Message Handlers]
```

### 2. Domain Layer

```mermaid
graph TD
    subgraph "Domain Models"
        User[User]
        FlashCard[FlashCard]
        CardBank[CardBank]
        Review[Review]
        Statistics[Statistics]
        Settings[Settings]
    end
    
    subgraph "Domain Services"
        FlashCardService[FlashCardService]
        UserService[UserService]
        CardBankService[CardBankService]
        SpacedRepetitionService[SpacedRepetitionService]
        StatisticsService[StatisticsService]
        SettingsService[SettingsService]
        AdminService[AdminService]
    end
```

### 3. Infrastructure Layer

```mermaid
graph TD
    subgraph "External Interfaces"
        TelegramAPI[Telegram API Client]
        DictionaryAPI[Dictionary API Interface]
    end
    
    subgraph "Repositories"
        UserRepo[UserRepository]
        FlashCardRepo[FlashCardRepository]
        CardBankRepo[CardBankRepository]
        ReviewRepo[ReviewRepository]
        StatisticsRepo[StatisticsRepository]
        SettingsRepo[SettingsRepository]
    end
    
    subgraph "Infrastructure Services"
        Logger[Structured Logger]
        Database[Database Connection]
        Cache[Cache]
    end
```

## Database Schema

```mermaid
erDiagram
    USERS {
        int id PK
        bigint telegram_id
        string username
        string first_name
        string last_name
        timestamp created_at
        timestamp updated_at
        bool is_admin
    }
    
    CARD_BANKS {
        int id PK
        string name
        string description
        int owner_id FK
        bool is_public
        timestamp created_at
        timestamp updated_at
    }
    
    FLASH_CARDS {
        int id PK
        int card_bank_id FK
        string word
        string definition
        json examples
        string image_url
        timestamp created_at
        timestamp updated_at
    }
    
    REVIEWS {
        int id PK
        int user_id FK
        int flash_card_id FK
        int ease_factor
        timestamp due_date
        int interval
        int repetitions
        timestamp last_reviewed
        timestamp created_at
        timestamp updated_at
    }
    
    STATISTICS {
        int id PK
        int user_id FK
        int card_bank_id FK
        int cards_reviewed
        int cards_learned
        int streak_days
        timestamp created_at
        timestamp updated_at
    }
    
    USER_SETTINGS {
        int id PK
        int user_id FK
        json settings
        timestamp created_at
        timestamp updated_at
    }
    
    BANK_MEMBERSHIPS {
        int id PK
        int user_id FK
        int card_bank_id FK
        string role
        timestamp created_at
        timestamp updated_at
    }
    
    GROUP_CHATS {
        int id PK
        bigint telegram_chat_id
        string title
        int card_bank_id FK
        timestamp created_at
        timestamp updated_at
    }
    
    USERS ||--o{ CARD_BANKS : "owns"
    USERS ||--o{ BANK_MEMBERSHIPS : "has"
    CARD_BANKS ||--o{ BANK_MEMBERSHIPS : "has"
    CARD_BANKS ||--o{ FLASH_CARDS : "contains"
    USERS ||--o{ REVIEWS : "has"
    FLASH_CARDS ||--o{ REVIEWS : "has"
    USERS ||--o{ STATISTICS : "has"
    CARD_BANKS ||--o{ STATISTICS : "has"
    USERS ||--o{ USER_SETTINGS : "has"
    CARD_BANKS ||--o{ GROUP_CHATS : "linked to"
```

## Key Features Implementation

### 1. Adding Words to Create Flash Cards

```mermaid
sequenceDiagram
    participant User
    participant Bot
    participant DictionaryAPI
    participant Database
    
    User->>Bot: Send word to learn
    Bot->>DictionaryAPI: Request definitions
    DictionaryAPI-->>Bot: Return definitions with examples
    Bot->>User: Show definitions with buttons
    User->>Bot: Select definition
    Bot->>User: Show examples with buttons
    User->>Bot: Select examples
    User->>Bot: (Optional) Send image with context
    Bot->>Database: Store flash card
    Bot->>User: Confirm card creation
```

### 2. Reviewing Words (Spaced Repetition)

```mermaid
sequenceDiagram
    participant User
    participant Bot
    participant SpacedRepetition
    participant Database
    
    User->>Bot: Request review
    Bot->>Database: Get due cards
    Database-->>Bot: Return due cards
    Bot->>SpacedRepetition: Get next card
    SpacedRepetition-->>Bot: Return next card
    Bot->>User: Show card definition/examples
    User->>Bot: Flip card
    Bot->>User: Show word
    User->>Bot: Rate difficulty (Again/Hard/Good/Easy)
    Bot->>SpacedRepetition: Update card scheduling
    SpacedRepetition->>Database: Save updated review data
    Bot->>User: Show next card or finish review
```

### 3. Card Banks and Sharing

```mermaid
sequenceDiagram
    participant User1
    participant User2
    participant Bot
    participant Database
    
    User1->>Bot: Create card bank
    Bot->>Database: Store card bank
    User1->>Bot: Share bank with User2
    Bot->>Database: Create bank membership
    Bot->>User2: Notify about shared bank
    User2->>Bot: Accept shared bank
    Bot->>Database: Update bank membership
    User2->>Bot: Access shared bank
```

### 4. Group Chat Functionality

```mermaid
sequenceDiagram
    participant Users
    participant GroupChat
    participant Bot
    participant Database
    
    Users->>GroupChat: Add bot to group
    GroupChat->>Bot: Bot added event
    Bot->>Users: Request card bank link
    Users->>Bot: Link to card bank
    Bot->>Database: Link group to card bank
    Users->>GroupChat: Send word to learn
    GroupChat->>Bot: Word message
    Bot->>Database: Process and add to linked bank
```

## Command Structure

```mermaid
graph TD
    TelegramUpdate[Telegram Update] --> UpdateRouter[Update Router]
    
    UpdateRouter --> CommandHandler[Command Handler]
    UpdateRouter --> CallbackHandler[Callback Handler]
    UpdateRouter --> MessageHandler[Message Handler]
    UpdateRouter --> PhotoHandler[Photo Handler]
    
    CommandHandler --> StartCommand[/start]
    CommandHandler --> HelpCommand[/help]
    CommandHandler --> AddWordCommand[/add]
    CommandHandler --> ReviewCommand[/review]
    CommandHandler --> StatsCommand[/stats]
    CommandHandler --> BanksCommand[/banks]
    CommandHandler --> CreateBankCommand[/create_bank]
    CommandHandler --> ShareBankCommand[/share_bank]
    CommandHandler --> JoinBankCommand[/join_bank]
    CommandHandler --> SettingsCommand[/settings]
    CommandHandler --> AdminCommand[/admin]
    
    CallbackHandler --> DefinitionCallback[Definition Selection]
    CallbackHandler --> ExampleCallback[Example Selection]
    CallbackHandler --> ReviewCallback[Review Rating]
    CallbackHandler --> BankCallback[Bank Selection]
    CallbackHandler --> SettingsCallback[Settings Selection]
    CallbackHandler --> PaginationCallback[Pagination]
    
    MessageHandler --> WordMessage[Word Input]
    MessageHandler --> SettingsMessage[Settings Input]
    MessageHandler --> AdminMessage[Admin Input]
    
    PhotoHandler --> ContextPhoto[Context Photo]
```

## Technical Architecture

### Project Structure

```
flash-cards-language-tg-bot/
├── cmd/
│   └── bot/
│       └── main.go
├── internal/
│   ├── app/
│   │   ├── app.go
│   │   └── config.go
│   ├── domain/
│   │   ├── models/
│   │   │   ├── user.go
│   │   │   ├── flashcard.go
│   │   │   ├── cardbank.go
│   │   │   ├── review.go
│   │   │   ├── statistics.go
│   │   │   └── settings.go
│   │   └── services/
│   │       ├── user_service.go
│   │       ├── flashcard_service.go
│   │       ├── cardbank_service.go
│   │       ├── spaced_repetition_service.go
│   │       ├── statistics_service.go
│   │       ├── settings_service.go
│   │       └── admin_service.go
│   ├── infrastructure/
│   │   ├── database/
│   │   │   ├── postgres.go
│   │   │   └── migrations/
│   │   ├── dictionary/
│   │   │   ├── dictionary.go
│   │   │   ├── free_dictionary.go
│   │   │   └── mock_dictionary.go
│   │   ├── telegram/
│   │   │   ├── bot.go
│   │   │   ├── handlers.go
│   │   │   └── keyboards.go
│   │   └── logging/
│   │       └── logger.go
│   └── repository/
│       ├── user_repository.go
│       ├── flashcard_repository.go
│       ├── cardbank_repository.go
│       ├── review_repository.go
│       ├── statistics_repository.go
│       └── settings_repository.go
├── pkg/
│   ├── utils/
│   │   └── helpers.go
│   └── spaced_repetition/
│       ├── algorithm.go
│       └── sm2.go
├── .env.example
├── .gitignore
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

## Deployment Architecture

```mermaid
graph TD
    GitHub[GitHub Repository] -->|CI/CD Pipeline| GitHubActions[GitHub Actions]
    GitHubActions -->|Build & Push| GHCR[GitHub Container Registry]
    GHCR -->|Pull Image| DockerHost[Docker Host]
    DockerHost -->|Run Container| BotContainer[Bot Container]
    DockerHost -->|Run Container| PostgresContainer[PostgreSQL Container]
    BotContainer -->|Connect to| PostgresContainer
    BotContainer -->|Connect to| TelegramAPI[Telegram API]
    BotContainer -->|Connect to| DictionaryAPI[Dictionary API]
```

## Configuration Management

The application will use environment variables for configuration, with an `.env.example` file provided as a template:

```
# Telegram Bot Configuration
TELEGRAM_TOKEN=your_telegram_bot_token
ADMIN_IDS=123456789,987654321

# Database Configuration
DB_HOST=postgres
DB_PORT=5432
DB_USER=flashcards
DB_PASSWORD=your_secure_password
DB_NAME=flashcards_db

# Dictionary API Configuration
DICTIONARY_API=freedictionary
DICTIONARY_API_KEY=your_api_key_if_needed

# Logging Configuration
LOG_LEVEL=info  # debug, info, warn, error
```

## Logging Strategy

Using Go's `slog` package for structured JSON logging with configurable log levels via environment variables.

## Testing Strategy

1. **Unit Tests**: For core business logic, services, and algorithms
2. **Integration Tests**: For repository implementations and external API clients
3. **End-to-End Tests**: For complete user flows using a mock Telegram API

## Security Considerations

1. **Environment Variables**: Sensitive information stored in environment variables
2. **Admin Access Control**: Restricted by Telegram IDs in environment variables
3. **Database Security**: Proper connection pooling, prepared statements to prevent SQL injection
4. **Input Validation**: All user input validated before processing
5. **Error Handling**: Proper error handling without leaking sensitive information

## Scalability Considerations

1. **Database Indexing**: Proper indexes on frequently queried fields
2. **Connection Pooling**: Efficient database connection management
3. **Caching**: Potential for caching frequently accessed data
4. **Stateless Design**: Bot designed to be stateless for horizontal scaling
5. **Containerization**: Easy deployment and scaling with Docker