-- Drop indexes
DROP INDEX IF EXISTS idx_group_chats_card_bank_id;
DROP INDEX IF EXISTS idx_group_chats_telegram_chat_id;
DROP INDEX IF EXISTS idx_user_settings_user_id;
DROP INDEX IF EXISTS idx_statistics_card_bank_id;
DROP INDEX IF EXISTS idx_statistics_user_id;
DROP INDEX IF EXISTS idx_reviews_due_date;
DROP INDEX IF EXISTS idx_reviews_flash_card_id;
DROP INDEX IF EXISTS idx_reviews_user_id;
DROP INDEX IF EXISTS idx_flash_cards_word;
DROP INDEX IF EXISTS idx_flash_cards_card_bank_id;
DROP INDEX IF EXISTS idx_bank_memberships_card_bank_id;
DROP INDEX IF EXISTS idx_bank_memberships_user_id;
DROP INDEX IF EXISTS idx_card_banks_owner_id;
DROP INDEX IF EXISTS idx_users_telegram_id;

-- Drop tables in reverse order to avoid foreign key constraints
DROP TABLE IF EXISTS group_chats;
DROP TABLE IF EXISTS user_settings;
DROP TABLE IF EXISTS statistics;
DROP TABLE IF EXISTS reviews;
DROP TABLE IF EXISTS flash_cards;
DROP TABLE IF EXISTS bank_memberships;
DROP TABLE IF EXISTS card_banks;
DROP TABLE IF EXISTS users;