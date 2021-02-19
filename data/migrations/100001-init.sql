CREATE TABLE credentials (
    name TEXT NOT NULL UNIQUE,
    data  TEXT NOT NULL,
    expires_at  DATETIME NOT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
);
