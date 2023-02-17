CREATE TABLE IF NOT EXISTS players
(
    guild_id   bigint  NOT NULL,
    node       varchar NOT NULL,
    queue_type int     NOT NULL,
    CONSTRAINT players_pkey PRIMARY KEY (guild_id)
);

CREATE TABLE IF NOT EXISTS queues
(
    id       bigserial NOT NULL,
    guild_id bigint    NOT NULL REFERENCES players (guild_id) ON DELETE CASCADE,
    position bigserial NOT NULL,
    track    json      NOT NULL,
    CONSTRAINT queues_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS histories
(
    id       bigserial NOT NULL,
    guild_id bigint    NOT NULL REFERENCES players (guild_id) ON DELETE CASCADE,
    position bigserial NOT NULL,
    track    json      NOT NULL,
    CONSTRAINT histories_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS playlists
(
    id      bigserial NOT NULL,
    user_id bigint    NOT NULL,
    name    varchar   NOT NULL,
    CONSTRAINT playlists_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS playlist_tracks
(
    id          bigserial NOT NULL,
    playlist_id bigint    NOT NULL REFERENCES playlists (id) ON DELETE CASCADE,
    position    bigserial NOT NULL,
    track       json      NOT NULL,
    CONSTRAINT playlist_tracks_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS liked_tracks
(
    id      bigserial NOT NULL,
    user_id bigint    NOT NULL,
    track   json      NOT NULL,
    CONSTRAINT liked_tracks_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS play_histories
(
    id        bigserial NOT NULL,
    user_id   bigint    NOT NULL,
    played_at timestamp NOT NULL,
    track     jsonb     NOT NULL,
    CONSTRAINT play_histories_pkey PRIMARY KEY (id),
    CONSTRAINT play_histories_user_id_track_key UNIQUE (user_id, track)
);

CREATE EXTENSION IF NOT EXISTS pg_trgm;
