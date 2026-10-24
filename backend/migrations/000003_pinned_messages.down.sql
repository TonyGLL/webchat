DROP INDEX IF EXISTS idx_messages_pinned;
ALTER TABLE messages DROP COLUMN IF EXISTS pinned_at;
