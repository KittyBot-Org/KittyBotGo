package bot

import (
	"context"
	"log/slog"
	"math/rand"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/topi314/tint"
)

var joinMessages = []string{
	"${user} just joined the server - glhf!",
	"${user} just joined. Everyone, look busy!",
	"${user} just joined. Can I get a heal?",
	"${user} joined your party.",
	"${user} joined. You must construct additional pylons.",
	"Welcome, ${user}. Stay awhile and listen.",
	"Welcome, ${user}. We hope you brought pizza.",
	"A wild ${user} appeared.",
	"${user} just joined. Hide your bananas.",
	"Never gonna give ${user} up. Never gonna let ${user} down.",
	"${user} has joined the server! It's super effective!",
	"${user} is here, as the prophecy foretold.",
}

var leaveMessages = []string{
	"${user}(${user_tag}) sprinted away!",
}

func (b *Bot) cacheGuildInvites(guildID snowflake.ID) {
	perms, err := b.Discord.Rest.GetGuildMemberPermissions(guildID, b.Discord.ApplicationID)
	if err != nil || !perms.Has(discord.PermissionManageGuild) {
		return
	}
	invites, err := b.Discord.Rest.GetGuildInvites(guildID)
	if err != nil {
		slog.Debug("failed to cache invites", slog.String("guild", guildID.String()), tint.Err(err))
		return
	}
	cached := make(map[string]int, len(invites))
	for _, invite := range invites {
		cached[invite.Code] = invite.Uses
	}
	b.inviteMu.Lock()
	b.invites[guildID] = cached
	b.inviteMu.Unlock()
}

func (b *Bot) rememberInvite(guildID snowflake.ID, code string, uses int) {
	b.inviteMu.Lock()
	defer b.inviteMu.Unlock()
	if b.invites[guildID] == nil {
		b.invites[guildID] = map[string]int{}
	}
	b.invites[guildID][code] = uses
}

func (b *Bot) forgetInvite(guildID snowflake.ID, code string) {
	b.inviteMu.Lock()
	defer b.inviteMu.Unlock()
	delete(b.invites[guildID], code)
}

func (b *Bot) usedInvite(guildID snowflake.ID) (string, bool) {
	perms, err := b.Discord.Rest.GetGuildMemberPermissions(guildID, b.Discord.ApplicationID)
	if err != nil || !perms.Has(discord.PermissionManageGuild) {
		return "", false
	}
	current, err := b.Discord.Rest.GetGuildInvites(guildID)
	if err != nil {
		return "", false
	}
	b.inviteMu.Lock()
	defer b.inviteMu.Unlock()
	cached := b.invites[guildID]
	if cached == nil {
		cached = map[string]int{}
		b.invites[guildID] = cached
	}
	var used string
	for _, invite := range current {
		if old, ok := cached[invite.Code]; ok && invite.Uses > old {
			used = invite.Code
		}
		cached[invite.Code] = invite.Uses
	}
	return used, used != ""
}

func (b *Bot) onMemberJoin(e *events.GuildMemberJoin) {
	code, ok := b.usedInvite(e.GuildID)
	if ok {
		b.usedMu.Lock()
		if b.used[e.GuildID] == nil {
			b.used[e.GuildID] = map[snowflake.ID]string{}
		}
		b.used[e.GuildID][e.Member.User.ID] = code
		b.usedMu.Unlock()
	}
	settings, err := b.Database.GetGuild(context.Background(), e.GuildID)
	if err != nil {
		return
	}
	if settings.RoleSaverEnabled {
		b.restoreRoles(e.GuildID, e.Member.User.ID)
	}
	if settings.JoinMessagesEnabled {
		inviteCode := "unknown"
		inviteLink := "unknown"
		if ok {
			inviteCode = code
			inviteLink = discord.InviteURL(code)
		}
		b.sendAnnouncement(e.GuildID, settings.AnnouncementChannel(), replacePlaceholders(settings.JoinMessage, map[string]string{
			"random_join_message": joinMessages[rand.Intn(len(joinMessages))],
			"user":                e.Member.User.Mention(),
			"user_tag":            e.Member.User.Tag(),
			"user_name":           e.Member.User.Username,
			"invite_code":         inviteCode,
			"invite_link":         inviteLink,
		}))
	}
}

func (b *Bot) onMemberLeave(e *events.GuildMemberLeave) {
	b.usedMu.Lock()
	delete(b.used[e.GuildID], e.User.ID)
	b.usedMu.Unlock()
	settings, err := b.Database.GetGuild(context.Background(), e.GuildID)
	if err != nil || !settings.LeaveMessagesEnabled {
		return
	}
	b.sendAnnouncement(e.GuildID, settings.AnnouncementChannel(), replacePlaceholders(settings.LeaveMessage, map[string]string{
		"random_leave_message": leaveMessages[rand.Intn(len(leaveMessages))],
		"user":                 e.User.Mention(),
		"user_tag":             e.User.Tag(),
		"user_name":            e.User.Username,
	}))
}

func (b *Bot) onMemberUpdate(e *events.GuildMemberUpdate) {
	var roles []snowflake.ID
	for _, roleID := range e.Member.RoleIDs {
		if roleID != e.GuildID {
			roles = append(roles, roleID)
		}
	}
	if err := b.Database.ReplaceMemberRoles(context.Background(), e.GuildID, e.Member.User.ID, roles); err != nil {
		slog.Debug("failed to store member roles", tint.Err(err))
	}
}

func (b *Bot) restoreRoles(guildID snowflake.ID, userID snowflake.ID) {
	perms, err := b.Discord.Rest.GetGuildMemberPermissions(guildID, b.Discord.ApplicationID)
	if err != nil || !perms.Has(discord.PermissionManageRoles) {
		return
	}
	roles, err := b.Database.GetMemberRoles(context.Background(), guildID, userID)
	if err != nil {
		return
	}
	for _, roleID := range roles {
		if err := b.Discord.Rest.AddMemberRole(guildID, userID, roleID); err != nil {
			slog.Debug("failed to restore role", slog.String("role", roleID.String()), tint.Err(err))
		}
	}
}

func (b *Bot) sendAnnouncement(guildID snowflake.ID, channelID snowflake.ID, message string) {
	if channelID == 0 || message == "" {
		return
	}
	if _, err := b.Discord.Rest.CreateMessage(channelID, discord.MessageCreate{Content: message}); err != nil {
		slog.Debug("failed to send announcement", slog.String("guild", guildID.String()), tint.Err(err))
	}
}

func replacePlaceholders(message string, values map[string]string) string {
	for key, value := range values {
		message = strings.ReplaceAll(message, "${"+key+"}", value)
	}
	return message
}
