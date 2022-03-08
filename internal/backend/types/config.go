package types

import (
	"github.com/DisgoOrg/snowflake"
	"github.com/KittyBot-Org/KittyBotGo/internal/config"
	"github.com/KittyBot-Org/KittyBotGo/internal/database"
)

type Config struct {
	config.Config
	Backend  BackendConfig   `json:"backend"`
	Database database.Config `json:"database"`

	BotLists BotListsConfig `json:"bot_lists"`

	BotInvite   string `json:"bot_invite"`
	GuildInvite string `json:"guild_invite"`
}

type BackendConfig struct {
	Token string `json:"token"`
	Port  string `json:"port"`
}

type BotListsConfig struct {
	VoterRoleID snowflake.Snowflake `json:"voter_role_id"`
	Tokens      map[BotList]string  `json:"tokens"`
}

type BotList string

const (
	TopGG              BotList = "top_gg"
	BotListSpace       BotList = "botlist_space"
	BotsForDiscordCom  BotList = "bots_for_discord_com"
	DiscordBotListCom  BotList = "discord_bot_list_com"
	DiscordservicesNet BotList = "discordservices_net"
)
