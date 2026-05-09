-- +goose Up
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    share_code VARCHAR(8) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    event_date DATE,
    author_name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    cookie_token UUID UNIQUE NOT NULL DEFAULT gen_random_uuid(),
    is_author BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    creator_id UUID NOT NULL REFERENCES participants(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    assigned_participant_id UUID REFERENCES participants(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_participants_event ON participants(event_id);
CREATE INDEX idx_participants_token ON participants(cookie_token);
CREATE INDEX idx_items_event ON items(event_id);
CREATE INDEX idx_items_assigned ON items(assigned_participant_id);
