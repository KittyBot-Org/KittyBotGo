-- name: EnsureGuild :exec
INSERT INTO guilds (id)
VALUES ($1)
ON CONFLICT (id) DO NOTHING;

-- name: GetGuild :one
SELECT id,
       announcement_channel_id,
       join_message,
       join_messages_enabled,
       leave_message,
       leave_messages_enabled,
       nsfw_enabled,
       dj_role_id,
       role_saver_enabled
FROM guilds
WHERE id = $1;

-- name: UpdateGuild :exec
UPDATE guilds
SET announcement_channel_id = $2,
    join_message            = $3,
    join_messages_enabled   = $4,
    leave_message           = $5,
    leave_messages_enabled  = $6,
    nsfw_enabled            = $7,
    dj_role_id              = $8,
    role_saver_enabled      = $9
WHERE id = $1;

-- name: DeleteMemberRoles :exec
DELETE
FROM member_roles
WHERE guild_id = $1
  AND user_id = $2;

-- name: InsertMemberRole :exec
INSERT INTO member_roles (guild_id, user_id, role_id)
VALUES ($1, $2, $3)
ON CONFLICT (guild_id, user_id, role_id) DO NOTHING;

-- name: GetMemberRoles :many
SELECT role_id
FROM member_roles
WHERE guild_id = $1
  AND user_id = $2;
