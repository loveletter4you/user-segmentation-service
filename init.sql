CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS segments (
    id   SERIAL PRIMARY KEY,
    slug VARCHAR(256) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_segment (
    id          SERIAL PRIMARY KEY,
    user_id     INTEGER NOT NULL REFERENCES users,
    segment_id  INTEGER NOT NULL REFERENCES segments,
    active_from TIMESTAMP NOT NULL DEFAULT now(),
    active_to   TIMESTAMP NOT NULL DEFAULT '9999-12-31 23:59:59'
);

CREATE TABLE IF NOT EXISTS segments_auto_insert (
    id          SERIAL PRIMARY KEY,
    segment_id  INTEGER NOT NULl REFERENCES segments,
    chance      FLOAT NOT NULL,
    active_from TIMESTAMP NOT NULL DEFAULT now(),
    active_to   TIMESTAMP NOT NULL DEFAULT '9999-12-31 23:59:59'
);