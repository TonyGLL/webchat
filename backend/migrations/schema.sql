-- Enable the UUID extension if you plan to use it for new tables.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- =============================================================================
-- 1. USERS & AUTHENTICATION (Based on your User and Password structs)
-- =============================================================================

-- This table maps directly to your `User` struct in `user.go`.
-- We use SERIAL for the ID to match your `int` type.
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(30),
    avatar_url TEXT, -- Changed `avatar` to `avatar_url` for clarity.
    
    -- This field directly maps to `LastAccess` in your struct.
    last_access TIMESTAMPTZ,
    
    deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- === ADDITIONS FOR FULL USE CASES ===
    -- For Google OAuth2 (Use Case #1)
    google_sub TEXT UNIQUE, -- The unique subject ID from Google.
    email_verified_at TIMESTAMPTZ -- To confirm the user's email is valid.
);

-- This table maps directly to your `Password` struct in `user.go`.
-- It establishes a one-to-one relationship with the `users` table.
CREATE TABLE passwords (
    user_id INT PRIMARY KEY,
    hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    CONSTRAINT fk_user
        FOREIGN KEY(user_id) 
        REFERENCES users(id)
        ON DELETE CASCADE -- If a user is deleted, their password record is also deleted.
);

-- === ADDITION FOR AUTHENTICATION FLOW (Use Case #1) ===
-- Required for the JWT Refresh Token strategy.
CREATE TABLE refresh_tokens (
    token_hash VARCHAR(128) PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


-- =============================================================================
-- 2. ROOMS AND MEMBERSHIPS (New entities required by Use Case #3)
-- =============================================================================

-- ENUM type for roles within a room.
CREATE TYPE room_role AS ENUM (
    'owner',
    'admin',
    'member'
);

-- Rooms are the "channels" where messages are sent.
CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    name VARCHAR(150) NOT NULL,
    topic TEXT,
    is_private BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Join table to manage which users are in which rooms and their roles.
CREATE TABLE room_members (
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role room_role NOT NULL DEFAULT 'member',
    joined_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (room_id, user_id)
);


-- =============================================================================
-- 3. MESSAGING (Based on your Message struct and expanded)
-- =============================================================================

-- This table is based on your `Message` struct.
-- We use UUID for the ID to match your `string` type.
CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Maps to your `Text` field. Renamed to `content` for clarity.
    content TEXT NOT NULL,
    
    -- Maps to your `AuthorID` field. Links to the `users` table.
    -- ON DELETE SET NULL keeps the message even if the author's account is deleted.
    author_id INT REFERENCES users(id) ON DELETE SET NULL,
    
    -- Maps to your `ChannelID` field. Links to the new `rooms` table.
    room_id UUID NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    
    -- Maps to your `CreatedAt` field.
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- === ADDITIONS FOR FULL USE CASES ===
    -- For message replies (Use Case #6)
    reply_to_message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    
    -- For editing messages (Use Case #4)
    edited_at TIMESTAMPTZ,
    
    -- For soft-deleting messages (Use Case #4)
    deleted_at TIMESTAMPTZ
);

-- Index for fetching messages in a room, ordered by creation time (most common query).
CREATE INDEX idx_messages_room_id_created_at ON messages(room_id, created_at DESC);


-- =============================================================================
-- 4. FEATURES & MODERATION (New entities for Use Cases #5, #6, #7)
-- =============================================================================

-- For reactions to messages (Use Case #6)
CREATE TABLE reactions (
    message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    emoji VARCHAR(32) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (message_id, user_id, emoji)
);

-- For file attachments (Use Case #5)
CREATE TYPE attachment_status AS ENUM ('pending', 'uploaded', 'processing', 'complete', 'error');

CREATE TABLE attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    message_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    uploader_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    storage_key TEXT NOT NULL UNIQUE, -- The key/path in your S3-compatible storage
    filename VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    size_bytes BIGINT NOT NULL,
    status attachment_status NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- For banning users from rooms or globally (Use Case #7)
CREATE TABLE bans (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    room_id UUID REFERENCES rooms(id) ON DELETE CASCADE, -- NULL for a global ban
    banned_by_id INT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    reason TEXT,
    expires_at TIMESTAMPTZ, -- NULL for a permanent ban
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE (user_id, room_id) -- A user can only have one active ban per room.
);

-- For user-submitted reports (Use Case #7)
CREATE TYPE report_status AS ENUM ('pending', 'resolved', 'ignored');

CREATE TABLE reports (
    id SERIAL PRIMARY KEY,
    reported_message_id UUID NOT NULL REFERENCES messages(id) ON DELETE CASCADE,
    reported_by_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT NOT NULL,
    status report_status NOT NULL DEFAULT 'pending',
    resolved_by_id INT REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);