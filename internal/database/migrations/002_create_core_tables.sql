-- +goose Up

CREATE TYPE user_role AS ENUM ('photographer', 'customer');
CREATE TYPE event_status AS ENUM ('draft', 'published', 'archived');

CREATE TABLE users (
    id                       TEXT        PRIMARY KEY,
    email                    TEXT UNIQUE NOT NULL,
    password_hash            TEXT        NOT NULL,
    display_name             TEXT        NOT NULL,
    role                     user_role   NOT NULL DEFAULT 'customer',
    banned_at                TIMESTAMPTZ,
    email_verified_at        TIMESTAMPTZ,
    password_reset_token     TEXT,
    password_reset_expires_at TIMESTAMPTZ,
    created_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE events (
    id              TEXT         PRIMARY KEY,
    photographer_id TEXT         NOT NULL REFERENCES users(id),
    slug            TEXT UNIQUE  NOT NULL,
    name            TEXT         NOT NULL,
    sport           TEXT         NOT NULL,
    location        TEXT,
    description     TEXT,
    tags            TEXT[]       NOT NULL DEFAULT '{}',
    status          event_status NOT NULL DEFAULT 'draft',
    date            DATE,
    cover_photo_id  TEXT,
    sort_order      INTEGER      NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE TABLE galleries (
    id             TEXT        PRIMARY KEY,
    event_id       TEXT        NOT NULL REFERENCES events(id),
    slug           TEXT        NOT NULL,
    name           TEXT        NOT NULL,
    description    TEXT,
    sort_order     INTEGER     NOT NULL DEFAULT 0,
    cover_photo_id TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, slug)
);

CREATE TABLE photos (
    id           TEXT        PRIMARY KEY,
    gallery_id   TEXT        NOT NULL REFERENCES galleries(id),
    storage_key  TEXT        NOT NULL,
    preview_key  TEXT        NOT NULL,
    filename     TEXT        NOT NULL,
    content_type TEXT        NOT NULL,
    size_bytes   BIGINT      NOT NULL,
    width        INTEGER,
    height       INTEGER,
    sort_order   INTEGER     NOT NULL DEFAULT 0,
    exif_data    JSONB,
    taken_at     TIMESTAMPTZ,
    deleted_at   TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE events
    ADD CONSTRAINT fk_events_cover_photo
    FOREIGN KEY (cover_photo_id) REFERENCES photos(id) ON DELETE SET NULL;

ALTER TABLE galleries
    ADD CONSTRAINT fk_galleries_cover_photo
    FOREIGN KEY (cover_photo_id) REFERENCES photos(id) ON DELETE SET NULL;

CREATE INDEX idx_events_photographer_id ON events (photographer_id);
CREATE INDEX idx_events_slug            ON events (slug);
CREATE INDEX idx_events_status          ON events (status);
CREATE INDEX idx_galleries_event_id     ON galleries (event_id);
CREATE INDEX idx_photos_gallery_id      ON photos (gallery_id);
CREATE INDEX idx_photos_deleted_at      ON photos (deleted_at);

-- +goose Down

ALTER TABLE events    DROP CONSTRAINT IF EXISTS fk_events_cover_photo;
ALTER TABLE galleries DROP CONSTRAINT IF EXISTS fk_galleries_cover_photo;

DROP INDEX IF EXISTS idx_photos_deleted_at;
DROP INDEX IF EXISTS idx_photos_gallery_id;
DROP INDEX IF EXISTS idx_galleries_event_id;
DROP INDEX IF EXISTS idx_events_status;
DROP INDEX IF EXISTS idx_events_slug;
DROP INDEX IF EXISTS idx_events_photographer_id;

DROP TABLE IF EXISTS photos;
DROP TABLE IF EXISTS galleries;
DROP TABLE IF EXISTS events;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS event_status;
DROP TYPE IF EXISTS user_role;
