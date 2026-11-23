package commands

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/omit"
	"github.com/disgoorg/snowflake/v2"

	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var settingsCommand = discord.SlashCommandCreate{
	Name:                     "settings",
	Description:              "View or change guild settings",
	DefaultMemberPermissions: omit.NewPtr(discord.PermissionManageGuild),
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{Name: "list", Description: "Lists the current settings"},
		discord.ApplicationCommandOptionSubCommand{Name: "joinmessage", Description: "Sets or enables join messages", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{Name: "enabled", Description: "Whether join messages are enabled"},
			discord.ApplicationCommandOptionString{Name: "message", Description: "The join message template"},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "leavemessage", Description: "Sets or enables leave messages", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{Name: "enabled", Description: "Whether leave messages are enabled"},
			discord.ApplicationCommandOptionString{Name: "message", Description: "The leave message template"},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "announcementchannel", Description: "Sets the announcement channel", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionChannel{Name: "channel", Description: "The announcement channel", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "djrole", Description: "Sets the DJ role", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionRole{Name: "role", Description: "The DJ role. Omit to clear it"},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "nsfw", Description: "Enables or disables NSFW commands", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{Name: "enabled", Description: "Whether NSFW commands are enabled", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "rolesaver", Description: "Enables or disables restoring roles when a member rejoins", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionBool{Name: "enabled", Description: "Whether role saving is enabled", Required: true},
		}},
	},
}

func (c *commands) OnSettingsList(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	settings, err := c.Database.GetGuild(e.Ctx, *e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to load settings", err))
	}
	return e.CreateMessage(res.Create(fmt.Sprintf("Join messages: `%t`\nJoin message: %s\nLeave messages: `%t`\nLeave message: %s\nAnnouncement channel: %s\nDJ role: %s\nNSFW: `%t`\nRole saver: `%t`",
		settings.JoinMessagesEnabled, settings.JoinMessage, settings.LeaveMessagesEnabled, settings.LeaveMessage, mentionChannel(settings.AnnouncementChannel()), mentionRole(settings.DJRole()), settings.NSFWEnabled, settings.RoleSaverEnabled)))
}

func (c *commands) OnSettingsJoin(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return c.updateMessageSetting(e, data, "join_messages_enabled", "join_message", "Join")
}

func (c *commands) OnSettingsLeave(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return c.updateMessageSetting(e, data, "leave_messages_enabled", "leave_message", "Leave")
}

func (c *commands) updateMessageSetting(e *handler.CommandEvent, data discord.SlashCommandInteractionData, enabledColumn string, messageColumn string, label string) error {
	var parts []string
	if enabled, ok := data.OptBool("enabled"); ok {
		if err := c.Database.SetGuildColumn(e.Ctx, *e.GuildID(), enabledColumn, enabled); err != nil {
			return e.CreateMessage(res.CreateErr("Failed to update settings", err))
		}
		parts = append(parts, fmt.Sprintf("%s messages `%s`", label, enabledText(enabled)))
	}
	if message, ok := data.OptString("message"); ok {
		if err := c.Database.SetGuildColumn(e.Ctx, *e.GuildID(), messageColumn, message); err != nil {
			return e.CreateMessage(res.CreateErr("Failed to update settings", err))
		}
		parts = append(parts, fmt.Sprintf("%s message set to:\n%s", label, message))
	}
	if len(parts) == 0 {
		return c.OnSettingsList(data, e)
	}
	return e.CreateMessage(res.Create(strings.Join(parts, "\n")))
}

func (c *commands) OnSettingsAnnouncement(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	channel := data.Channel("channel")
	if err := c.Database.SetGuildColumn(e.Ctx, *e.GuildID(), "announcement_channel_id", int64(channel.ID)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update settings", err))
	}
	return e.CreateMessage(res.Createf("Announcement channel set to %s", discord.ChannelMention(channel.ID)))
}

func (c *commands) OnSettingsDJRole(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	roleID := int64(-1)
	message := "Cleared the DJ role."
	if role, ok := data.OptRole("role"); ok {
		roleID = int64(role.ID)
		message = fmt.Sprintf("DJ role set to %s", discord.RoleMention(role.ID))
	}
	if err := c.Database.SetGuildColumn(e.Ctx, *e.GuildID(), "dj_role_id", roleID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update settings", err))
	}
	return e.CreateMessage(res.Create(message))
}

func (c *commands) OnSettingsNSFW(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	enabled := data.Bool("enabled")
	if err := c.Database.SetGuildColumn(e.Ctx, *e.GuildID(), "nsfw_enabled", enabled); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update settings", err))
	}
	return e.CreateMessage(res.Createf("NSFW commands `%s`", enabledText(enabled)))
}

func (c *commands) OnSettingsRoleSaver(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	enabled := data.Bool("enabled")
	if err := c.Database.SetGuildColumn(e.Ctx, *e.GuildID(), "role_saver_enabled", enabled); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update settings", err))
	}
	return e.CreateMessage(res.Createf("Role saver `%s`", enabledText(enabled)))
}

func enabledText(enabled bool) string {
	if enabled {
		return "enabled"
	}
	return "disabled"
}

func mentionChannel(id snowflake.ID) string {
	if id == 0 {
		return "unset"
	}
	return discord.ChannelMention(id)
}

func mentionRole(id snowflake.ID) string {
	if id == 0 {
		return "unset"
	}
	return discord.RoleMention(id)
}
