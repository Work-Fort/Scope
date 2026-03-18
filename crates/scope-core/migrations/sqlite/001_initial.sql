CREATE TABLE IF NOT EXISTS forts (
    name TEXT PRIMARY KEY,
    local INTEGER NOT NULL DEFAULT 1,
    gateway TEXT,
    active INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS fort_services (
    fort_name TEXT NOT NULL REFERENCES forts(name) ON DELETE CASCADE,
    url TEXT NOT NULL,
    PRIMARY KEY (fort_name, url)
);

CREATE TABLE IF NOT EXISTS notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service TEXT NOT NULL,
    title TEXT NOT NULL,
    body TEXT,
    urgency TEXT NOT NULL CHECK (urgency IN ('passive', 'active')),
    route TEXT,
    read INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%fZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_notifications_unread
    ON notifications (read, created_at DESC);

CREATE TABLE IF NOT EXISTS preferences (
    service TEXT PRIMARY KEY,
    level TEXT NOT NULL DEFAULT 'allow_urgent'
        CHECK (level IN ('mute', 'passive_only', 'allow_urgent'))
);
