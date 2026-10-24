-- Emails and usernames are unique regardless of case. Existing data is never rewritten: if two
-- accounts already collide under lower(), the migration stops with the offending values so an
-- operator can resolve them by hand first.
DO $$
DECLARE
    dup RECORD;
    found TEXT := '';
BEGIN
    FOR dup IN
        SELECT 'email' AS kind, lower(email) AS value, count(*) AS n FROM users GROUP BY lower(email) HAVING count(*) > 1
        UNION ALL
        SELECT 'username', lower(username), count(*) FROM users GROUP BY lower(username) HAVING count(*) > 1
    LOOP
        found := found || format(E'\n  %s "%s" used by %s accounts', dup.kind, dup.value, dup.n);
    END LOOP;
    IF found <> '' THEN
        RAISE EXCEPTION 'migration 000013 aborted: accounts collide when compared case-insensitively; resolve them by hand (rename or delete) and restart:%', found;
    END IF;
END $$;

CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));
CREATE UNIQUE INDEX users_username_lower_key ON users (lower(username));
