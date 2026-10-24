-- When each member last read each room; unread = messages from others created after it.
-- Existing members start "caught up" so nobody opens the app to a pile of old unread counts.
ALTER TABLE room_members ADD COLUMN last_read_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX idx_messages_room_created ON messages (room_id, created_at DESC);
