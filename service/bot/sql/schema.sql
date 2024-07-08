CREATE TABLE IF NOT EXISTS lavalink_nodes
(
    name       VARCHAR PRIMARY KEY,
    session_id VARCHAR
);

CREATE TABLE IF NOT EXISTS playlists
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT    NOT NULL,
    name    VARCHAR   NOT NULL
);

CREATE TABLE IF NOT EXISTS playlist_tracks
(
    id          BIGSERIAL PRIMARY KEY,
    playlist_id BIGINT    NOT NULL REFERENCES playlists (id) ON DELETE CASCADE,
    position    BIGSERIAL NOT NULL,
    track       JSONB     NOT NULL
);

CREATE TABLE IF NOT EXISTS liked_tracks
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT    NOT NULL,
    track   JSONB     NOT NULL
);