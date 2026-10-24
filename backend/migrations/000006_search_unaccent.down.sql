DROP INDEX IF EXISTS idx_messages_fts;
CREATE INDEX idx_messages_fts ON messages USING GIN (to_tsvector('simple', content));
DROP FUNCTION IF EXISTS search_unaccent(text);
