-- One report per person per message; the active-ban lookup also filters by user and room.
CREATE UNIQUE INDEX uq_reports_message_reporter ON reports (reported_message_id, reported_by_id);
CREATE INDEX idx_reports_status_created ON reports (status, created_at DESC);
