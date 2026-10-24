-- Directory search: trigram index over the exact expression the query filters on, so
-- substring matches (LIKE '%q%') stop scanning the whole users table on every keystroke.
-- Queries shorter than 3 characters cannot use a trigram index and still scan.
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX IF NOT EXISTS idx_users_directory_trgm ON users
    USING gin (lower(username || ' ' || name || ' ' || COALESCE(last_name, '')) gin_trgm_ops)
    WHERE deleted = FALSE;
