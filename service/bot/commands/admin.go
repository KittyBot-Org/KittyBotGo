package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgo/rest"

	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var adminCommand = discord.SlashCommandCreate{
	Name:                     "admin",
	Description:              "Moderation commands",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommandGroup{
			Name:        "ban",
			Description: "Manage bans",
			Options: []discord.ApplicationCommandOptionSubCommand{
				{Name: "add", Description: "Bans a member", Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionUser{Name: "user", Description: "The user to ban", Required: true},
					discord.ApplicationCommandOptionString{Name: "reason", Description: "The ban reason"},
				}},
				{Name: "remove", Description: "Removes a ban", Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionUser{Name: "user", Description: "The user to unban", Required: true},
				}},
				{Name: "list", Description: "Lists bans"},
			},
		},
		discord.ApplicationCommandOptionSubCommand{Name: "kick", Description: "Kicks a member", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{Name: "user", Description: "The user to kick", Required: true},
			discord.ApplicationCommandOptionString{Name: "reason", Description: "The kick reason"},
		}},
	},
}

func (c *commands) OnAdminBanAdd(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionBanMembers) {
		return e.CreateMessage(res.CreateError("You need the Ban Members permission"))
	}
	user := data.User("user")
	if err := c.Discord.Rest.AddBan(*e.GuildID(), user.ID, time.Hour*24, reasonOpt(data)...); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to ban that user", err))
	}
	return e.CreateMessage(res.Createf("Banned %s", user.Mention()))
}

func (c *commands) OnAdminBanRemove(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionBanMembers) {
		return e.CreateMessage(res.CreateError("You need the Ban Members permission"))
	}
	user := data.User("user")
	if err := c.Discord.Rest.DeleteBan(*e.GuildID(), user.ID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to unban that user", err))
	}
	return e.CreateMessage(res.Createf("Unbanned %s", user.Mention()))
}

func (c *commands) OnAdminBanList(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionBanMembers) {
		return e.CreateMessage(res.CreateError("You need the Ban Members permission"))
	}
	bans, err := c.Discord.Rest.GetBans(*e.GuildID(), 0, 0, 100)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to list bans", err))
	}
	if len(bans) == 0 {
		return e.CreateMessage(res.Create("No bans"))
	}
	var lines []string
	for _, ban := range bans {
		reason := ""
		if ban.Reason != nil {
			reason = *ban.Reason
		}
		lines = append(lines, fmt.Sprintf("%s - %s", ban.User.Tag(), reason))
	}
	return e.CreateMessage(res.Create(strings.Join(lines, "\n")))
}

func (c *commands) OnAdminKick(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionKickMembers) {
		return e.CreateMessage(res.CreateError("You need the Kick Members permission"))
	}
	user := data.User("user")
	if err := c.Discord.Rest.RemoveMember(*e.GuildID(), user.ID, reasonOpt(data)...); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to kick that user", err))
	}
	return e.CreateMessage(res.Createf("Kicked %s", user.Mention()))
}

func hasPerm(e *handler.CommandEvent, perm discord.Permissions) bool {
	member := e.Member()
	return member != nil && member.Permissions.Has(perm)
}

func reasonOpt(data discord.SlashCommandInteractionData) []rest.RequestOpt {
	reason, ok := data.OptString("reason")
	if !ok || reason == "" {
		return nil
	}
	return []rest.RequestOpt{rest.WithReason(reason)}
}
