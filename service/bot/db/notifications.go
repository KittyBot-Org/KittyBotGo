package db

import (
	"context"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Notification struct {
	ID               int            `db:"id"`
	GuildID          snowflake.ID   `db:"guild_id"`
	UserID           snowflake.ID   `db:"user_id"`
	ChannelID        snowflake.ID   `db:"channel_id"`
	MessageID        snowflake.ID   `db:"message_id"`
	Content          string         `db:"content"`
	CreatedAt        time.Time      `db:"created_at"`
	NotificationTime time.Time      `db:"notification_time"`
}

func (d *DB) CreateNotification(ctx context.Context, n Notification) error {
	_, err := d.dbx.ExecContext(ctx, `INSERT INTO notifications (guild_id, user_id, channel_id, message_id, content, notification_time) VALUES ($1, $2, $3, $4, $5, $6)`, n.GuildID, n.UserID, n.ChannelID, n.MessageID, n.Content, n.NotificationTime)
	return err
}

func (d *DB) ListNotifications(ctx context.Context, userID snowflake.ID) ([]Notification, error) {
	var notifications []Notification
	err := d.dbx.SelectContext(ctx, &notifications, `SELECT id, guild_id, user_id, channel_id, message_id, content, created_at, notification_time FROM notifications WHERE user_id = $1 ORDER BY notification_time`, userID)
	return notifications, err
}

func (d *DB) DeleteNotification(ctx context.Context, id int, userID snowflake.ID) error {
	_, err := d.dbx.ExecContext(ctx, `DELETE FROM notifications WHERE id = $1 AND user_id = $2`, id, userID)
	return err
}

func (d *DB) DueNotifications(ctx context.Context, until time.Time) ([]Notification, error) {
	var notifications []Notification
	err := d.dbx.SelectContext(ctx, &notifications, `SELECT id, guild_id, user_id, channel_id, message_id, content, created_at, notification_time FROM notifications WHERE notification_time <= $1`, until)
	return notifications, err
}
