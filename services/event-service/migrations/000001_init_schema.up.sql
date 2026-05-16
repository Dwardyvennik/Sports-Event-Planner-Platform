CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'scheduled',
    country TEXT,
    city TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_sport_start_time ON events(sport, start_time);
CREATE INDEX idx_events_competition_start_time ON events(competition, start_time);
CREATE INDEX idx_events_country_start_time ON events(country, start_time);
CREATE INDEX idx_events_filter_lookup ON events(sport, competition, country, start_time);
