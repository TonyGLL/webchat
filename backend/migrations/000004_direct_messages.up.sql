-- A direct conversation is a private room with exactly two members. direct_rooms guarantees
-- there is at most one per pair of users (user_a < user_b makes the pair order-independent).
ALTER TABLE rooms ADD COLUMN is_direct BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE direct_rooms (
    user_a INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    user_b INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id UUID NOT NULL UNIQUE REFERENCES rooms(id) ON DELETE CASCADE,
    PRIMARY KEY (user_a, user_b),
    CHECK (user_a < user_b)
);
