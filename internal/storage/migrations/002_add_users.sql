-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id BIGINT NOT NULL UNIQUE,
    username TEXT,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT,
    photo_url TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE participants ADD COLUMN user_id UUID REFERENCES users(id);
CREATE INDEX idx_participants_user_event ON participants(user_id, event_id) WHERE user_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_participants_user_event;
ALTER TABLE participants DROP COLUMN IF EXISTS user_id;
DROP TABLE IF EXISTS users;
