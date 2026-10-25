-- BP-4: the directory's default listing (empty query) is ORDER BY name, username over active
-- users; without an ordered index it scanned and sorted every user on each poll.
CREATE INDEX idx_users_directory_order ON users (name, username) WHERE deleted = FALSE;

-- BP-3: ListMembers pages ORDER BY joined_at, user_id inside one room; this index serves that
-- order directly instead of sorting every member of the room for each page.
CREATE INDEX idx_room_members_room_joined ON room_members (room_id, joined_at, user_id);
