-- Indexes on foreign-key / filter columns that had none (the composite primary keys lead with
-- another column, so they cannot serve these lookups), plus removal of a duplicate index.
-- golang-migrate runs this file in one transaction, so plain CREATE INDEX (a short write lock
-- on each table) is used instead of CREATE INDEX CONCURRENTLY, which cannot run in a transaction.

-- Rooms of a user, membership export, ON DELETE CASCADE from users.
CREATE INDEX IF NOT EXISTS idx_room_members_user_id ON room_members (user_id);

-- Export of a user's messages, ON DELETE SET NULL from users.
CREATE INDEX IF NOT EXISTS idx_messages_author_id ON messages (author_id) WHERE author_id IS NOT NULL;

-- ON DELETE SET NULL from messages (replies).
CREATE INDEX IF NOT EXISTS idx_messages_reply_to ON messages (reply_to_message_id) WHERE reply_to_message_id IS NOT NULL;

-- Export of a user's reactions, ON DELETE CASCADE from users (PK leads with message_id).
CREATE INDEX IF NOT EXISTS idx_reactions_user_id ON reactions (user_id);

-- Rooms owned by a user (ON DELETE RESTRICT check on user deletion).
CREATE INDEX IF NOT EXISTS idx_rooms_owner_id ON rooms (owner_id);

-- Reports filed by a user: export and ON DELETE CASCADE from users.
CREATE INDEX IF NOT EXISTS idx_reports_reported_by ON reports (reported_by_id);

-- Bans of a room: ban list, is-banned check, ON DELETE CASCADE from rooms
-- (UNIQUE (user_id, room_id) leads with user_id).
CREATE INDEX IF NOT EXISTS idx_bans_room_id ON bans (room_id);

-- idx_messages_room_created (migration 2) is identical to idx_messages_room_id_created_at
-- (migration 1): (room_id, created_at DESC). Keep the original, drop the duplicate.
DROP INDEX IF EXISTS idx_messages_room_created;
