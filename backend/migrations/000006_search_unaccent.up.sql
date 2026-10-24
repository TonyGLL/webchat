-- Accent-insensitive search ("cafe" finds "café"). unaccent() is only STABLE, which indexes
-- reject, so it is wrapped in an IMMUTABLE function with the dictionary fixed.
CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE OR REPLACE FUNCTION search_unaccent(text) RETURNS text
    LANGUAGE sql IMMUTABLE PARALLEL SAFE STRICT
    AS $$ SELECT public.unaccent('public.unaccent'::regdictionary, $1) $$;

DROP INDEX IF EXISTS idx_messages_fts;
CREATE INDEX idx_messages_fts ON messages USING GIN (to_tsvector('simple', search_unaccent(content)));
