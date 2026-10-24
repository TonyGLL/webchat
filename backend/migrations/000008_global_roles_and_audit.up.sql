-- Application-wide role. Granted only from the server side (see cmd/admin), never through the API.
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'moderator', 'admin'));

-- Append-only record of sensitive actions: who did what to whom, where and when.
CREATE TABLE audit_log (
    id BIGSERIAL PRIMARY KEY,
    actor_id INT REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    room_id UUID REFERENCES rooms(id) ON DELETE SET NULL,
    target_user_id INT REFERENCES users(id) ON DELETE SET NULL,
    target_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    details JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_audit_room_created ON audit_log (room_id, created_at DESC);
CREATE INDEX idx_audit_created ON audit_log (created_at DESC);
