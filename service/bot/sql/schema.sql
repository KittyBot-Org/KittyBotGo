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

CREATE TABLE IF NOT EXISTS play_histories
(
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT    NOT NULL,
    played_at TIMESTAMP NOT NULL,
    track     JSONB     NOT NULL,
    UNIQUE (user_id, track)
);

-- The Java database is the base. These objects are unused by the Go bot.
DROP TABLE IF EXISTS self_assignable_roles;
DROP TABLE IF EXISTS self_assignable_role_messages;
DROP TABLE IF EXISTS reactive_messages;
DROP TABLE IF EXISTS self_assignable_role_groups;
DROP TABLE IF EXISTS guild_invite_roles;
DROP TABLE IF EXISTS guild_invites;
DROP TABLE IF EXISTS snipe_disabled_channels;
DROP TABLE IF EXISTS stream_user_events;
DROP TABLE IF EXISTS stream_users;
DROP TABLE IF EXISTS requests;
DROP TABLE IF EXISTS user_statistics;
DROP TABLE IF EXISTS bot_disabled_channels;
DROP TABLE IF EXISTS bot_ignored_members;

ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS snipes_enabled;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS log_messages_enabled;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS log_channel_id;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS stream_announcement_channel_id;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS stream_announcement_message;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS request_channel_id;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS requests_enabled;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS inactive_role_id;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS inactive_duration;
ALTER TABLE IF EXISTS guilds DROP COLUMN IF EXISTS inactive_role_enabled;

ALTER TABLE IF EXISTS sessions ALTER COLUMN access_token TYPE TEXT;
ALTER TABLE IF EXISTS sessions ALTER COLUMN refresh_token TYPE TEXT;