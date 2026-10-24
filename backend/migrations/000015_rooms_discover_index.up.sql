-- Keyset pagination of room discovery orders by (name, id) over non-direct rooms; this partial
-- index lets each page read only its own rows instead of sorting every room.
CREATE INDEX IF NOT EXISTS idx_rooms_discover_name ON rooms (name, id) WHERE NOT is_direct;
