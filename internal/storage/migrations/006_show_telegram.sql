-- +goose Up
ALTER TABLE users ADD COLUMN show_telegram BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE users DROP COLUMN show_telegram;
