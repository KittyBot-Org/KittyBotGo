package db

import (
	"context"
	"fmt"

	"github.com/kittybot-org/kittybotgo/service/bot/db/dbsqlc"
)

func (d *DB) SetGuildColumn(ctx context.Context, guildID int64, column string, value any) error {
	if err := d.EnsureGuild(ctx, guildID); err != nil {
		return err
	}
	settings, err := d.GetGuild(ctx, guildID)
	if err != nil {
		return err
	}
	switch column {
	case "announcement_channel_id":
		settings.AnnouncementChannelID, err = int64Value(value)
	case "join_message":
		settings.JoinMessage, err = stringValue(value)
	case "join_messages_enabled":
		settings.JoinMessagesEnabled, err = boolValue(value)
	case "leave_message":
		settings.LeaveMessage, err = stringValue(value)
	case "leave_messages_enabled":
		settings.LeaveMessagesEnabled, err = boolValue(value)
	case "nsfw_enabled":
		settings.NsfwEnabled, err = boolValue(value)
	case "dj_role_id":
		settings.DjRoleID, err = int64Value(value)
	case "role_saver_enabled":
		settings.RoleSaverEnabled, err = boolValue(value)
	default:
		return fmt.Errorf("unknown guild column %s", column)
	}
	if err != nil {
		return err
	}
	return d.UpdateGuild(ctx, dbsqlc.UpdateGuildParams{
		ID:                    settings.ID,
		AnnouncementChannelID: settings.AnnouncementChannelID,
		JoinMessage:           settings.JoinMessage,
		JoinMessagesEnabled:   settings.JoinMessagesEnabled,
		LeaveMessage:          settings.LeaveMessage,
		LeaveMessagesEnabled:  settings.LeaveMessagesEnabled,
		NsfwEnabled:           settings.NsfwEnabled,
		DjRoleID:              settings.DjRoleID,
		RoleSaverEnabled:      settings.RoleSaverEnabled,
	})
}

func (d *DB) ReplaceMemberRoles(ctx context.Context, guildID int64, userID int64, roleIDs []int64) error {
	tx, err := d.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := d.WithTx(tx)
	if err = q.DeleteMemberRoles(ctx, dbsqlc.DeleteMemberRolesParams{GuildID: guildID, UserID: userID}); err != nil {
		return err
	}
	for _, roleID := range roleIDs {
		if err = q.InsertMemberRole(ctx, dbsqlc.InsertMemberRoleParams{
			GuildID: guildID,
			UserID:  userID,
			RoleID:  roleID,
		}); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func int64Value(value any) (int64, error) {
	number, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("expected int64, got %T", value)
	}
	return number, nil
}

func stringValue(value any) (string, error) {
	text, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("expected string, got %T", value)
	}
	return text, nil
}

func boolValue(value any) (bool, error) {
	enabled, ok := value.(bool)
	if !ok {
		return false, fmt.Errorf("expected bool, got %T", value)
	}
	return enabled, nil
}
