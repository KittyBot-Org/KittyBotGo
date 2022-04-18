package backend

import (
	"time"

	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/disgoorg/snowflake"
)

type Config struct {
	config.Config
	Address string `json:"address"`

	BotLists BotListsConfig `json:"bot_lists"`

	BotInvite          string `json:"bot_invite"`
	GuildInvite        string `json:"guild_invite"`
	PrometheusEndpoint string `json:"prometheus_endpoint"`
}

type BotListsConfig struct {
	VoterRoleID snowflake.Snowflake `json:"voter_role_id"`
	Tokens      map[string]string   `json:"tokens"`
}

type BotList struct {
	Name         string
	URL          string
	BotURL       string
	VoteCooldown time.Duration
}

var (
	TopGG = BotList{
		Name:         "top_gg",
		URL:          "https://top.gg",
		BotURL:       "/kbot/%s",
		VoteCooldown: 12 * time.Hour,
	}
	BotListSpace = BotList{
		Name:         "botlist_space",
		URL:          "https://botlist.space",
		BotURL:       "/kbot/%s",
		VoteCooldown: 24 * time.Hour,
	}
	DiscordBotsGG = BotList{
		Name:   "discord_bots_gg",
		URL:    "https://botlist.space",
		BotURL: "/bots/%s",
	}
	DiscordExtremeListXYZ = BotList{
		Name:   "discord_extreme_list_xyz",
		URL:    "https://discordextremelist.xyz",
		BotURL: "/bots/%s",
	}
	BotsForDiscordCom = BotList{
		Name:         "bots_for_discord_com",
		URL:          "https://botsfordiscord.com",
		BotURL:       "/kbot/%s",
		VoteCooldown: 24 * time.Hour,
	}
	DiscordBotListCom = BotList{
		Name:         "discord_bot_list_com",
		URL:          "https://discordbotlist.com",
		BotURL:       "/bots/%s",
		VoteCooldown: 12 * time.Hour,
	}
	DiscordservicesNet = BotList{
		Name:         "discordservices_net",
		URL:          "https://discordservices.net",
		BotURL:       "/kbot/%s",
		VoteCooldown: 12 * time.Hour,
	}
)