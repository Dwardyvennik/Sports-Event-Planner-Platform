CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
<<<<<<< Updated upstream
    sport TEXT NOT NULL,
    venue TEXT NOT NULL,
    scheduled_at TIMESTAMPTZ NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
=======
    description TEXT,
    start_time TIMESTAMPTZ NOT NULL,
    end_time TIMESTAMPTZ,
    status TEXT NOT NULL DEFAULT 'scheduled',
    country TEXT,
    city TEXT,
>>>>>>> Stashed changes
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

<<<<<<< Updated upstream
CREATE TABLE registrations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    status TEXT NOT NULL DEFAULT 'registered',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(event_id, user_id)
);

CREATE INDEX idx_events_scheduled_at ON events(scheduled_at);
CREATE INDEX idx_registrations_user_id ON registrations(user_id);
=======
CREATE INDEX idx_events_start_time ON events(start_time);
CREATE INDEX idx_events_sport_start_time ON events(sport, start_time);
CREATE INDEX idx_events_competition_start_time ON events(competition, start_time);
CREATE INDEX idx_events_country_start_time ON events(country, start_time);
CREATE INDEX idx_events_filter_lookup ON events(sport, competition, country, start_time);
>>>>>>> Stashed changes
