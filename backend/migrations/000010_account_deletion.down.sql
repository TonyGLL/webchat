-- Bans whose issuer was deleted cannot satisfy NOT NULL again; drop them before restoring it.
DELETE FROM bans WHERE banned_by_id IS NULL;
ALTER TABLE bans DROP CONSTRAINT bans_banned_by_id_fkey;
ALTER TABLE bans ADD CONSTRAINT bans_banned_by_id_fkey FOREIGN KEY (banned_by_id) REFERENCES users(id) ON DELETE RESTRICT;
ALTER TABLE bans ALTER COLUMN banned_by_id SET NOT NULL;
