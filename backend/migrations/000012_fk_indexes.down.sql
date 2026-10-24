CREATE INDEX IF NOT EXISTS idx_messages_room_created ON messages (room_id, created_at DESC);

DROP INDEX IF EXISTS idx_bans_room_id;
DROP INDEX IF EXISTS idx_reports_reported_by;
DROP INDEX IF EXISTS idx_rooms_owner_id;
DROP INDEX IF EXISTS idx_reactions_user_id;
DROP INDEX IF EXISTS idx_messages_reply_to;
DROP INDEX IF EXISTS idx_messages_author_id;
DROP INDEX IF EXISTS idx_room_members_user_id;
