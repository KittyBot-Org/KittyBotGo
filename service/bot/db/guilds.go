package db

import (
	"context"
	"database/sql"

	"github.com/disgoorg/snowflake/v2"
)

type GuildSettings struct {
	ID                    int64  `db:"id"`
	AnnouncementChannelID int64  `db:"announcement_channel_id"`
	JoinMessage           string `db:"join_message"`
	JoinMessagesEnabled   bool   `db:"join_messages_enabled"`
	LeaveMessage          string `db:"leave_message"`
	LeaveMessagesEnabled  bool   `db:"leave_messages_enabled"`
	NSFWEnabled           bool   `db:"nsfw_enabled"`
	DJRoleID              int64  `db:"dj_role_id"`
	RoleSaverEnabled      bool   `db:"role_saver_enabled"`
}

func (g GuildSettings) AnnouncementChannel() snowflake.ID {
	return positiveID(g.AnnouncementChannelID)
}

func (g GuildSettings) DJRole() snowflake.ID {
	return positiveID(g.DJRoleID)
}

func positiveID(id int64) snowflake.ID {
	if id <= 0 {
		return 0
	}
	return snowflake.ID(id)
}

func (d *DB) EnsureGuild(ctx context.Context, guildID snowflake.ID) error {
	_, err := d.dbx.ExecContext(ctx, `INSERT INTO guilds (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`, guildID)
	return err
}

func (d *DB) GetGuild(ctx context.Context, guildID snowflake.ID) (GuildSettings, error) {
	if err := d.EnsureGuild(ctx, guildID); err != nil {
		return GuildSettings{}, err
	}
	var settings GuildSettings
	err := d.dbx.GetContext(ctx, &settings, `SELECT id, announcement_channel_id, join_message, join_messages_enabled, leave_message, leave_messages_enabled, nsfw_enabled, dj_role_id, role_saver_enabled FROM guilds WHERE id = $1`, guildID)
	return settings, err
}

func (d *DB) SetGuildColumn(ctx context.Context, guildID snowflake.ID, column string, value any) error {
	if err := d.EnsureGuild(ctx, guildID); err != nil {
		return err
	}
	_, err := d.dbx.ExecContext(ctx, `UPDATE guilds SET `+column+` = $2 WHERE id = $1`, guildID, value)
	return err
}

func (d *DB) ReplaceMemberRoles(ctx context.Context, guildID snowflake.ID, userID snowflake.ID, roleIDs []snowflake.ID) error {
	tx, err := d.dbx.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `DELETE FROM member_roles WHERE guild_id = $1 AND user_id = $2`, guildID, userID); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if _, err = tx.ExecContext(ctx, `INSERT INTO member_roles (guild_id, user_id, role_id) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, guildID, userID, roleID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (d *DB) GetMemberRoles(ctx context.Context, guildID snowflake.ID, userID snowflake.ID) ([]snowflake.ID, error) {
	var roles []snowflake.ID
	err := d.dbx.SelectContext(ctx, &roles, `SELECT role_id FROM member_roles WHERE guild_id = $1 AND user_id = $2`, guildID, userID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return roles, err
}
