CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL,
    sport TEXT NOT NULL,
    category TEXT,
    competition TEXT,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'scheduled',
    country TEXT,
    city TEXT,
    max_participants INTEGER NOT NULL DEFAULT 0 CHECK (max_participants >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE event_participants (
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (event_id, user_id)
);

CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN NEW.updated_at = now(); RETURN NEW; END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER events_updated_at
BEFORE UPDATE ON events
FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE INDEX idx_events_creator_id ON events(creator_id);
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_sport_start_time ON events(sport, start_time);
CREATE INDEX idx_events_competition_start_time ON events(competition, start_time);
CREATE INDEX idx_events_country_start_time ON events(country, start_time);
CREATE INDEX idx_events_filter_lookup ON events(sport, competition, country, start_time);
CREATE INDEX idx_event_participants_user_id ON event_participants(user_id);
