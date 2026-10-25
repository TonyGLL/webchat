-- Personal invitations expire and can be cancelled by room admins.
ALTER TABLE room_invitations ADD COLUMN expires_at TIMESTAMPTZ NOT NULL DEFAULT NOW() + INTERVAL '30 days';
ALTER TABLE room_invitations DROP CONSTRAINT room_invitations_status_check;
ALTER TABLE room_invitations ADD CONSTRAINT room_invitations_status_check
    CHECK (status IN ('pending', 'accepted', 'declined', 'cancelled'));
CREATE INDEX idx_room_invitations_room ON room_invitations(room_id) WHERE status = 'pending';

-- A ban remembers the rank of whoever issued it: lifting or overwriting it needs a higher rank
-- (or being the issuer). Existing rows are backfilled from the issuer's current power; rows whose
-- issuer is gone or no longer a member keep the issuer's global power only (0 if none), so any
-- moderator can still lift them.
ALTER TABLE bans ADD COLUMN banned_by_rank INT NOT NULL DEFAULT 0;
UPDATE bans b SET banned_by_rank = CASE
    WHEN u.role = 'admin' THEN 5
    ELSE GREATEST(
        CASE WHEN u.role = 'moderator' THEN 2 ELSE 0 END,
        COALESCE((SELECT CASE m.role::text WHEN 'owner' THEN 4 WHEN 'admin' THEN 3 WHEN 'moderator' THEN 2 WHEN 'member' THEN 1 END
                  FROM room_members m WHERE m.room_id = b.room_id AND m.user_id = u.id), 0))
    END
FROM users u
WHERE u.id = b.banned_by_id;
