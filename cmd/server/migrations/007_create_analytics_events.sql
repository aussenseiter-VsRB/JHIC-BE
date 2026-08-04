CREATE TABLE IF NOT EXISTS analytics_events (
    id BIGSERIAL PRIMARY KEY,
    event_name TEXT NOT NULL,
    session_hash TEXT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    properties JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_analytics_events_created_at ON analytics_events (created_at);
CREATE INDEX idx_analytics_events_name_created_at ON analytics_events (event_name, created_at);
