-- Invitations addressed to a specific user: they see them in the app and accept or decline.
-- (The anonymous `invites` codes stay as they are.)
CREATE TABLE room_invitations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    inviter_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    invitee_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'declined')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    responded_at TIMESTAMPTZ
);

-- One open invitation per person and room.
CREATE UNIQUE INDEX idx_room_invitations_pending ON room_invitations(room_id, invitee_id) WHERE status = 'pending';
CREATE INDEX idx_room_invitations_invitee ON room_invitations(invitee_id) WHERE status = 'pending';
