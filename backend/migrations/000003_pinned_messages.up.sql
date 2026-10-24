-- A pinned message has the time it was pinned; NULL means not pinned.
ALTER TABLE messages ADD COLUMN pinned_at TIMESTAMPTZ;

CREATE INDEX idx_messages_pinned ON messages (room_id, pinned_at DESC) WHERE pinned_at IS NOT NULL;
