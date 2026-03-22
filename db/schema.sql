CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE users (
    email           TEXT PRIMARY KEY,
    agent_id        TEXT UNIQUE NOT NULL
);

CREATE TABLE magic_links (
    token           TEXT PRIMARY KEY,
    email           TEXT REFERENCES users(email),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE images (
    image_id        TEXT PRIMARY KEY,
    agent_id        TEXT REFERENCES users(agent_id),
    file_name       TEXT NOT NULL,
    file_type       TEXT NOT NULL,
    file_timestamp  TIMESTAMPTZ NOT NULL,
    embedding       vector(512),
    tags            TEXT[],
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE albums (
    album_id        TEXT PRIMARY KEY,
    album_name      TEXT NOT NULL,
    agent_id        TEXT REFERENCES users(agent_id),
    created_at      TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE albums_images(
    album_id        TEXT REFERENCES albums(album_id),
    image_id        TEXT REFERENCES images(image_id),
    PRIMARY KEY (album_id, image_id)
);

CREATE TABLE share_tokens (
    token           TEXT PRIMARY KEY,
    album_id        TEXT REFERENCES albums(album_id),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);
