-- Preserve photo paths while removing embedded bot credentials.
UPDATE flash_cards
SET image_url = regexp_replace(image_url, '^https://api\.telegram\.org/file/bot[^/]+/', 'telegram-path:')
WHERE image_url LIKE 'https://api.telegram.org/file/bot%';
