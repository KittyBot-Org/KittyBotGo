package commands

import (
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var infoCommand = discord.SlashCommandCreate{
	Name:        "info",
	Description: "Information about the bot and this server",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{Name: "help", Description: "Lists the command groups"},
		discord.ApplicationCommandOptionSubCommand{Name: "info", Description: "Shows bot information"},
		discord.ApplicationCommandOptionSubCommand{Name: "avatar", Description: "Shows a user's avatar", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionUser{Name: "user", Description: "The user, defaults to you"},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "servericon", Description: "Shows this server's icon"},
		discord.ApplicationCommandOptionSubCommand{Name: "serverbanner", Description: "Shows this server's banner"},
		discord.ApplicationCommandOptionSubCommand{Name: "uptime", Description: "Shows how long the bot has been running"},
		discord.ApplicationCommandOptionSubCommand{Name: "vote", Description: "Shows your vote status"},
		discord.ApplicationCommandOptionSubCommand{Name: "privacy", Description: "Shows the privacy policy"},
		discord.ApplicationCommandOptionSubCommand{Name: "commands", Description: "Shows the command list"},
		discord.ApplicationCommandOptionSubCommand{Name: "dashboard", Description: "Shows the dashboard link"},
	},
}

func (c *commands) OnInfoHelp(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(res.Create(commandList()))
}

func (c *commands) OnInfoInfo(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	user, _ := c.Discord.Caches.SelfUser()
	guilds := 0
	for range c.Discord.Caches.Guilds() {
		guilds++
	}
	return e.CreateMessage(res.Createf("%s\nVersion: `%s`\nCommit: `%s`\nGuilds: `%d`\nSupport: %s", user.Tag(), c.Version, c.Commit, guilds, c.Config.Dashboard.SupportInvite))
}

func (c *commands) OnInfoAvatar(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	user := e.User()
	if option, ok := data.OptUser("user"); ok {
		user = option
	}
	return e.CreateMessage(res.Create(user.EffectiveAvatarURL()))
}

func (c *commands) OnInfoServerIcon(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	guild, ok := c.Discord.Caches.Guild(*e.GuildID())
	if !ok || guild.Icon == nil {
		return e.CreateMessage(res.CreateError("This server has no icon"))
	}
	return e.CreateMessage(res.Create(*guild.IconURL()))
}

func (c *commands) OnInfoServerBanner(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	guild, ok := c.Discord.Caches.Guild(*e.GuildID())
	if !ok || guild.Banner == nil {
		return e.CreateMessage(res.CreateError("This server has no banner"))
	}
	return e.CreateMessage(res.Create(*guild.BannerURL()))
}

func (c *commands) OnInfoUptime(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(res.Createf("Up for %s", time.Since(c.Started).Round(time.Second)))
}

func (c *commands) OnInfoVote(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	expiry, err := c.Database.VoteExpiry(e.Ctx, e.User().ID)
	if err != nil || time.Now().After(expiry) {
		return e.CreateMessage(res.Createf("You have no active vote. %s", c.Config.Dashboard.VoteURL))
	}
	return e.CreateMessage(res.Createf("Your vote lasts until %s", expiry.Format(time.RFC822)))
}

func (c *commands) OnInfoPrivacy(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(res.Create(c.Config.Dashboard.PrivacyURL))
}

func (c *commands) OnInfoCommands(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(res.Create(commandList()))
}

func (c *commands) OnInfoDashboard(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return e.CreateMessage(res.Create(c.Config.Dashboard.OriginURL))
}

func commandList() string {
	return strings.Join([]string{
		"`/music` `/queue` `/history` `/playlists` `/liked-songs`",
		"`/settings` `/admin` `/tags` `/notification` `/neko` `/utility` `/info` `/ping`",
	}, "\n")
}
