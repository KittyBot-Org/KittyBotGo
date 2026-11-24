-- name: CreateTag :exec
INSERT INTO guild_tags (name, guild_id, user_id, content)
VALUES ($1, $2, $3, $4);

-- name: GetTag :one
SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id
FROM guild_tags
WHERE guild_id = $1
  AND name = $2;

-- name: GetTagByID :one
SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id
FROM guild_tags
WHERE guild_id = $1
  AND id = $2;

-- name: ListTags :many
SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id
FROM guild_tags
WHERE guild_id = $1
ORDER BY name;

-- name: SearchTags :many
SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id
FROM guild_tags
WHERE guild_id = $1
  AND name ILIKE '%' || sqlc.arg(query) || '%'
ORDER BY name
LIMIT 25;

-- name: UpdateTag :exec
UPDATE guild_tags
SET content    = $3,
    updated_at = NOW()
WHERE guild_id = $1
  AND name = $2;

-- name: DeleteTag :one
DELETE
FROM guild_tags
WHERE guild_id = $1
  AND name = $2
RETURNING id, name, guild_id, user_id, content, created_at, updated_at, command_id;

-- name: SetTagCommandID :exec
UPDATE guild_tags
SET command_id = $2
WHERE id = $1;

-- name: GetTagByCommandID :one
SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id
FROM guild_tags
WHERE command_id = $1;
