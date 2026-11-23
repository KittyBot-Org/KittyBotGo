package db

import (
	"context"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Tag struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	GuildID   snowflake.ID   `db:"guild_id"`
	UserID    snowflake.ID   `db:"user_id"`
	Content   string         `db:"content"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt *time.Time     `db:"updated_at"`
	CommandID int64          `db:"command_id"`
}

type TagAttachment struct {
	ID    int    `db:"id"`
	TagID int    `db:"tag_id"`
	URL   string `db:"url"`
}

func (d *DB) CreateTag(ctx context.Context, guildID snowflake.ID, userID snowflake.ID, name string, content string) error {
	_, err := d.dbx.ExecContext(ctx, `INSERT INTO guild_tags (name, guild_id, user_id, content) VALUES ($1, $2, $3, $4)`, name, guildID, userID, content)
	return err
}

func (d *DB) GetTag(ctx context.Context, guildID snowflake.ID, name string) (Tag, error) {
	var tag Tag
	err := d.dbx.GetContext(ctx, &tag, `SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id FROM guild_tags WHERE guild_id = $1 AND name = $2`, guildID, name)
	return tag, err
}

func (d *DB) GetTagByID(ctx context.Context, guildID snowflake.ID, id int) (Tag, error) {
	var tag Tag
	err := d.dbx.GetContext(ctx, &tag, `SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id FROM guild_tags WHERE guild_id = $1 AND id = $2`, guildID, id)
	return tag, err
}

func (d *DB) ListTags(ctx context.Context, guildID snowflake.ID) ([]Tag, error) {
	var tags []Tag
	err := d.dbx.SelectContext(ctx, &tags, `SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id FROM guild_tags WHERE guild_id = $1 ORDER BY name`, guildID)
	return tags, err
}

func (d *DB) SearchTags(ctx context.Context, guildID snowflake.ID, query string) ([]Tag, error) {
	var tags []Tag
	err := d.dbx.SelectContext(ctx, &tags, `SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id FROM guild_tags WHERE guild_id = $1 AND name ILIKE '%' || $2 || '%' ORDER BY name LIMIT 25`, guildID, query)
	return tags, err
}

func (d *DB) UpdateTag(ctx context.Context, guildID snowflake.ID, name string, content string) error {
	_, err := d.dbx.ExecContext(ctx, `UPDATE guild_tags SET content = $3, updated_at = NOW() WHERE guild_id = $1 AND name = $2`, guildID, name, content)
	return err
}

func (d *DB) DeleteTag(ctx context.Context, guildID snowflake.ID, name string) (Tag, error) {
	var tag Tag
	err := d.dbx.GetContext(ctx, &tag, `DELETE FROM guild_tags WHERE guild_id = $1 AND name = $2 RETURNING id, name, guild_id, user_id, content, created_at, updated_at, command_id`, guildID, name)
	return tag, err
}

func (d *DB) SetTagCommandID(ctx context.Context, tagID int, commandID int64) error {
	_, err := d.dbx.ExecContext(ctx, `UPDATE guild_tags SET command_id = $2 WHERE id = $1`, tagID, commandID)
	return err
}

func (d *DB) GetTagByCommandID(ctx context.Context, commandID int64) (Tag, error) {
	var tag Tag
	err := d.dbx.GetContext(ctx, &tag, `SELECT id, name, guild_id, user_id, content, created_at, updated_at, command_id FROM guild_tags WHERE command_id = $1`, commandID)
	return tag, err
}
