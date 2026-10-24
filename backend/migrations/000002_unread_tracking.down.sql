DROP INDEX IF EXISTS idx_messages_room_created;
ALTER TABLE room_members DROP COLUMN IF EXISTS last_read_at;
