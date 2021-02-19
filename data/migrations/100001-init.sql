CREATE TABLE strava_activities (
    strava_id   BIGINT   PRIMARY KEY
                         NOT NULL
                         UNIQUE,
    external_id STRING,
    name        STRING,
    started_on  DATE     NOT NULL,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL
)
WITHOUT ROWID;
