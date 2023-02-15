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