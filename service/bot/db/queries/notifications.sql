-- name: CreateNotification :exec
INSERT INTO notifications (guild_id, user_id, channel_id, message_id, content, notification_time)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: ListNotifications :many
SELECT id, guild_id, user_id, channel_id, message_id, content, created_at, notification_time
FROM notifications
WHERE user_id = $1
ORDER BY notification_time;

-- name: DeleteNotification :exec
DELETE
FROM notifications
WHERE id = $1
  AND user_id = $2;

-- name: DueNotifications :many
SELECT id, guild_id, user_id, channel_id, message_id, content, created_at, notification_time
FROM notifications
WHERE notification_time <= $1;
