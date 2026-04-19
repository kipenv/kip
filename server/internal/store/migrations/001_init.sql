CREATE TABLE IF NOT EXISTS teams (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    invite_code TEXT NOT NULL UNIQUE,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS members (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES teams(id),
    username TEXT NOT NULL,
    token TEXT NOT NULL UNIQUE,
    joined_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, username)
);

CREATE TABLE IF NOT EXISTS pins (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES teams(id),
    filename TEXT NOT NULL DEFAULT '.env',
    ciphertext BLOB NOT NULL,
    nonce BLOB NOT NULL,
    pinned_by TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(team_id, filename)
);

CREATE TABLE IF NOT EXISTS share_log (
    id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL REFERENCES teams(id),
    from_user TEXT NOT NULL,
    to_user TEXT,
    filename TEXT NOT NULL,
    secret_id TEXT NOT NULL,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
