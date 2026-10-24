-- PostgreSQL cannot remove a value from an enum; demote moderators so the value is unused.
UPDATE room_members SET role = 'member' WHERE role::text = 'moderator';
