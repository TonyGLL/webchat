ALTER TABLE bans DROP COLUMN banned_by_rank;
DROP INDEX idx_room_invitations_room;
UPDATE room_invitations SET status = 'declined' WHERE status = 'cancelled';
ALTER TABLE room_invitations DROP CONSTRAINT room_invitations_status_check;
ALTER TABLE room_invitations ADD CONSTRAINT room_invitations_status_check
    CHECK (status IN ('pending', 'accepted', 'declined'));
ALTER TABLE room_invitations DROP COLUMN expires_at;
