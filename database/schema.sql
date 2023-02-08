
CREATE TABLE IF NOT EXISTS reports(
    id serial NOT NULL,
    user_id varchar NOT NULL,
    guild_id varchar NOT NULL,
    description varchar NOT NULL,
    created_at timestamp without time zone NOT NULL,
    confirmed boolean NOT NULL,
    message_id varchar NOT NULL,
    channel_id varchar NOT NULL,
    CONSTRAINT reports_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS public.guild_settings(
    id varchar NOT NULL,
    moderation_log_webhook_id varchar NOT NULL,
    moderation_log_webhook_token varchar NOT NULL,
    CONSTRAINT guild_settings_pkey PRIMARY KEY (id)
);

CREATE TABLE IF NOT EXISTS liked_tracks(
    user_id varchar NOT NULL,
    query varchar NOT NULL,
    title varchar NOT NULL,
    liked_at timestamp NOT NULL,
    CONSTRAINT liked_tracks_pkey PRIMARY KEY (user_id, title)
);


CREATE TABLE IF NOT EXISTS public.play_histories(
    user_id varchar NOT NULL,
    query varchar NOT NULL,
    title varchar NOT NULL,
    last_used_at timestamp without time zone NOT NULL,
    CONSTRAINT play_histories_pkey PRIMARY KEY (user_id, title)
);

CREATE TABLE IF NOT EXISTS public.tags(
    guild_id varchar NOT NULL,
    owner_id varchar NOT NULL,
    name varchar NOT NULL,
    content varchar NOT NULL,
    uses integer NOT NULL,
    created_at timestamp NOT NULL,
    CONSTRAINT tags_pkey PRIMARY KEY (guild_id, name)
);

CREATE TABLE IF NOT EXISTS public.voters(
    user_id varchar NOT NULL,
    expires_at timestamp NOT NULL,
    CONSTRAINT voters_pkey PRIMARY KEY (user_id)
);