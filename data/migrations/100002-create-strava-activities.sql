CREATE TABLE strava_activities (
    remote_id   BIGINT   PRIMARY KEY
                         NOT NULL
                         UNIQUE,
    external_id TEXT,
    name        TEXT,
    data        TEXT,
    started_on  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME DEFAULT CURRENT_TIMESTAMP
)
WITHOUT ROWID;
