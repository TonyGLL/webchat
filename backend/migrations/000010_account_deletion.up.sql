-- Deleting an account must not be blocked by the bans that user issued: keep the ban, forget who set it.
ALTER TABLE bans ALTER COLUMN banned_by_id DROP NOT NULL;
ALTER TABLE bans DROP CONSTRAINT bans_banned_by_id_fkey;
ALTER TABLE bans ADD CONSTRAINT bans_banned_by_id_fkey FOREIGN KEY (banned_by_id) REFERENCES users(id) ON DELETE SET NULL;
