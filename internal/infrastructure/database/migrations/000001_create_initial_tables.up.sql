-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL UNIQUE,
    username VARCHAR(255),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

-- Create card_banks table
CREATE TABLE IF NOT EXISTS card_banks (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    owner_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    is_public BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create bank_memberships table
CREATE TABLE IF NOT EXISTS bank_memberships (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_bank_id INTEGER NOT NULL REFERENCES card_banks(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL, -- "owner", "editor", "viewer"
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(user_id, card_bank_id)
);

-- Create flash_cards table
CREATE TABLE IF NOT EXISTS flash_cards (
    id SERIAL PRIMARY KEY,
    card_bank_id INTEGER NOT NULL REFERENCES card_banks(id) ON DELETE CASCADE,
    word VARCHAR(255) NOT NULL,
    definition TEXT NOT NULL,
    examples JSONB NOT NULL DEFAULT '[]'::JSONB,
    image_url TEXT,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create reviews table
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    flash_card_id INTEGER NOT NULL REFERENCES flash_cards(id) ON DELETE CASCADE,
    ease_factor FLOAT NOT NULL DEFAULT 2.5,
    due_date TIMESTAMP NOT NULL,
    interval INTEGER NOT NULL DEFAULT 0,
    repetitions INTEGER NOT NULL DEFAULT 0,
    last_reviewed TIMESTAMP,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(user_id, flash_card_id)
);

-- Create statistics table
CREATE TABLE IF NOT EXISTS statistics (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_bank_id INTEGER NOT NULL REFERENCES card_banks(id) ON DELETE CASCADE,
    cards_reviewed INTEGER NOT NULL DEFAULT 0,
    cards_learned INTEGER NOT NULL DEFAULT 0,
    streak_days INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    UNIQUE(user_id, card_bank_id)
);

-- Create user_settings table
CREATE TABLE IF NOT EXISTS user_settings (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE UNIQUE,
    settings JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create group_chats table
CREATE TABLE IF NOT EXISTS group_chats (
    id SERIAL PRIMARY KEY,
    telegram_chat_id BIGINT NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    card_bank_id INTEGER NOT NULL REFERENCES card_banks(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_users_telegram_id ON users(telegram_id);
CREATE INDEX IF NOT EXISTS idx_card_banks_owner_id ON card_banks(owner_id);
CREATE INDEX IF NOT EXISTS idx_bank_memberships_user_id ON bank_memberships(user_id);
CREATE INDEX IF NOT EXISTS idx_bank_memberships_card_bank_id ON bank_memberships(card_bank_id);
CREATE INDEX IF NOT EXISTS idx_flash_cards_card_bank_id ON flash_cards(card_bank_id);
CREATE INDEX IF NOT EXISTS idx_flash_cards_word ON flash_cards(word);
CREATE INDEX IF NOT EXISTS idx_reviews_user_id ON reviews(user_id);
CREATE INDEX IF NOT EXISTS idx_reviews_flash_card_id ON reviews(flash_card_id);
CREATE INDEX IF NOT EXISTS idx_reviews_due_date ON reviews(due_date);
CREATE INDEX IF NOT EXISTS idx_statistics_user_id ON statistics(user_id);
CREATE INDEX IF NOT EXISTS idx_statistics_card_bank_id ON statistics(card_bank_id);
CREATE INDEX IF NOT EXISTS idx_user_settings_user_id ON user_settings(user_id);
CREATE INDEX IF NOT EXISTS idx_group_chats_telegram_chat_id ON group_chats(telegram_chat_id);
CREATE INDEX IF NOT EXISTS idx_group_chats_card_bank_id ON group_chats(card_bank_id);