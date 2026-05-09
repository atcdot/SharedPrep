-- +goose Up
ALTER TABLE events DROP COLUMN author_name;
ALTER TABLE participants DROP COLUMN name;
ALTER TABLE participants DROP COLUMN cookie_token;
ALTER TABLE participants ALTER COLUMN user_id SET NOT NULL;

-- +goose Down
ALTER TABLE participants ALTER COLUMN user_id DROP NOT NULL;
ALTER TABLE participants ADD COLUMN cookie_token UUID UNIQUE NOT NULL DEFAULT gen_random_uuid();
ALTER TABLE participants ADD COLUMN name VARCHAR(100) NOT NULL DEFAULT '';
ALTER TABLE events ADD COLUMN author_name VARCHAR(100) NOT NULL DEFAULT '';
