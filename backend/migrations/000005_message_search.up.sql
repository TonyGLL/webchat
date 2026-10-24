-- Full-text search over message content. The 'simple' configuration does no stemming, which
-- suits a chat that mixes languages. The query must use the exact same expression to use it.
CREATE INDEX idx_messages_fts ON messages USING GIN (to_tsvector('simple', content));
