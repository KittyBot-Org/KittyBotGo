CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE TABLE IF NOT EXISTS guilds
(
    id                             BIGINT PRIMARY KEY,
    announcement_channel_id        BIGINT NOT NULL DEFAULT -1,
    join_message                   TEXT NOT NULL DEFAULT 'Welcome ${user}!',
    join_messages_enabled          BOOLEAN NOT NULL DEFAULT false,
    leave_message                  TEXT NOT NULL DEFAULT 'Bye ${user}!',
    leave_messages_enabled         BOOLEAN NOT NULL DEFAULT false,
    log_channel_id                 BIGINT NOT NULL DEFAULT -1,
    log_messages_enabled           BOOLEAN NOT NULL DEFAULT false,
    request_channel_id             BIGINT NOT NULL DEFAULT -1,
    requests_enabled               BOOLEAN NOT NULL DEFAULT false,
    stream_announcement_channel_id BIGINT NOT NULL DEFAULT -1,
    stream_announcement_message    TEXT NOT NULL DEFAULT '${user} is now live!',
    nsfw_enabled                   BOOLEAN NOT NULL DEFAULT true,
    inactive_role_id               BIGINT NOT NULL DEFAULT -1,
    inactive_duration              INTERVAL NOT NULL DEFAULT '0',
    inactive_role_enabled          BOOLEAN NOT NULL DEFAULT false,
    dj_role_id                     BIGINT NOT NULL DEFAULT -1,
    snipes_enabled                 BOOLEAN NOT NULL DEFAULT true,
    role_saver_enabled             BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE IF NOT EXISTS sessions
(
    id            BIGSERIAL PRIMARY KEY,
    user_id       BIGINT NOT NULL UNIQUE,
    access_token  VARCHAR(32) NOT NULL,
    refresh_token VARCHAR(32) NOT NULL,
    expiration    TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS guild_tags
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(64) NOT NULL,
    guild_id   BIGINT NOT NULL REFERENCES guilds (id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL,
    content    TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    command_id BIGINT NOT NULL DEFAULT -1,
    UNIQUE (name, guild_id)
);

CREATE TABLE IF NOT EXISTS notifications
(
    id                BIGSERIAL PRIMARY KEY,
    guild_id          BIGINT NOT NULL REFERENCES guilds (id) ON DELETE CASCADE,
    user_id           BIGINT NOT NULL,
    channel_id        BIGINT NOT NULL,
    message_id        BIGINT NOT NULL,
    content           TEXT NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notification_time TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS member_roles
(
    id       BIGSERIAL PRIMARY KEY,
    guild_id BIGINT NOT NULL REFERENCES guilds (id) ON DELETE CASCADE,
    user_id  BIGINT NOT NULL,
    role_id  BIGINT NOT NULL,
    UNIQUE (guild_id, user_id, role_id)
);

CREATE TABLE IF NOT EXISTS voters
(
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL UNIQUE,
    vote_expiry TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS lavalink_nodes
(
    name       VARCHAR PRIMARY KEY,
    session_id VARCHAR
);

CREATE TABLE IF NOT EXISTS playlists
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    name    VARCHAR NOT NULL
);

CREATE TABLE IF NOT EXISTS playlist_tracks
(
    id          BIGSERIAL PRIMARY KEY,
    playlist_id BIGINT NOT NULL REFERENCES playlists (id) ON DELETE CASCADE,
    position    BIGSERIAL NOT NULL,
    track       JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS liked_tracks
(
    id      BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    track   JSONB NOT NULL
);

CREATE TABLE IF NOT EXISTS play_histories
(
    id        BIGSERIAL PRIMARY KEY,
    user_id   BIGINT NOT NULL,
    played_at TIMESTAMP NOT NULL,
    track     JSONB NOT NULL,
    UNIQUE (user_id, track)
);

DROP TABLE IF EXISTS self_assignable_roles, self_assignable_role_messages, self_assignable_role_groups, reactive_messages, guild_invite_roles, guild_invites, snipe_disabled_channels, stream_user_events, stream_users, requests, user_statistics, bot_disabled_channels, bot_ignored_members, guild_tag_attachments;

ALTER TABLE guilds
    DROP COLUMN IF EXISTS snipes_enabled,
    DROP COLUMN IF EXISTS log_messages_enabled,
    DROP COLUMN IF EXISTS log_channel_id,
    DROP COLUMN IF EXISTS stream_announcement_channel_id,
    DROP COLUMN IF EXISTS stream_announcement_message,
    DROP COLUMN IF EXISTS request_channel_id,
    DROP COLUMN IF EXISTS requests_enabled,
    DROP COLUMN IF EXISTS inactive_role_id,
    DROP COLUMN IF EXISTS inactive_duration,
    DROP COLUMN IF EXISTS inactive_role_enabled;

ALTER TABLE sessions
    ALTER COLUMN access_token TYPE TEXT,
    ALTER COLUMN refresh_token TYPE TEXT;
