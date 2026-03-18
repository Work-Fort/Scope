CREATE TABLE IF NOT EXISTS forts (
    name TEXT PRIMARY KEY,
    local BOOLEAN NOT NULL DEFAULT true,
    gateway TEXT,
    active BOOLEAN NOT NULL DEFAULT false
);

CREATE TABLE IF NOT EXISTS fort_services (
    fort_name TEXT NOT NULL REFERENCES forts(name) ON DELETE CASCADE,
    url TEXT NOT NULL,
    PRIMARY KEY (fort_name, url)
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    service TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    urgency TEXT NOT NULL CHECK (urgency IN ('passive', 'active')),
    route TEXT,
    read BOOLEAN NOT NULL DEFAULT false,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_notifications_unread
    ON notifications (read, created_at DESC);

CREATE TABLE IF NOT EXISTS preferences (
    service TEXT PRIMARY KEY,
    level TEXT NOT NULL DEFAULT 'allow_urgent'
        CHECK (level IN ('mute', 'passive_only', 'allow_urgent'))
);
