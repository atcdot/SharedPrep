-- +goose Up
ALTER TABLE events ADD COLUMN link TEXT;

-- +goose Down
ALTER TABLE events DROP COLUMN IF EXISTS link;
