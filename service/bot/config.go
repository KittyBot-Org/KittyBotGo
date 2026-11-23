package bot

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/kittybot-org/kittybotgo/internal/log"
	"github.com/kittybot-org/kittybotgo/service/bot/db"
)

type Config struct {
	Log       log.Config      `toml:"log"`
	Bot       BotConfig       `toml:"bot"`
	Database  db.Config       `toml:"database"`
	Dashboard DashboardConfig `toml:"dashboard"`
	Nodes     Nodes           `toml:"nodes"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n Log: %v\n Bot: %s\n Database: %s\n Dashboard: %s\n Nodes: %s\n",
		c.Log,
		c.Bot,
		c.Database,
		c.Dashboard,
		c.Nodes,
	)
}

type BotConfig struct {
	SyncCommands bool           `toml:"sync_commands"`
	GuildIDs     []snowflake.ID `toml:"guild_ids"`
	GatewayURL   string         `toml:"gateway_url"`
	RestURL      string         `toml:"rest_url"`
	Token        string         `toml:"token"`
	Secret       string         `toml:"secret"`
	DevIDs       []snowflake.ID `toml:"dev_ids"`
}

func (c BotConfig) String() string {
	return fmt.Sprintf("\n  SyncCommands: %t\n  GuildIDs: %v\n  GatewayURL: %s\n  RestURL: %s\n  Token: %s\n  Secret: %s\n  DevIDs: %v\n",
		c.SyncCommands,
		c.GuildIDs,
		c.GatewayURL,
		c.RestURL,
		strings.Repeat("*", len(c.Token)),
		strings.Repeat("*", len(c.Secret)),
		c.DevIDs,
	)
}

type DashboardConfig struct {
	Host          string `toml:"host"`
	Port          int    `toml:"port"`
	OriginURL     string `toml:"origin_url"`
	RedirectURL   string `toml:"redirect_url"`
	SigningKey    string `toml:"signing_key"`
	HastebinURL   string `toml:"hastebin_url"`
	SupportInvite string `toml:"support_invite"`
	BotInviteURL  string `toml:"bot_invite_url"`
	PrivacyURL    string `toml:"privacy_url"`
	VoteURL       string `toml:"vote_url"`
}

func (c DashboardConfig) String() string {
	return fmt.Sprintf("\n  Host: %s\n  Port: %d\n  OriginURL: %s\n  RedirectURL: %s\n  SigningKey: %s\n",
		c.Host,
		c.Port,
		c.OriginURL,
		c.RedirectURL,
		strings.Repeat("*", len(c.SigningKey)),
	)
}

type NodeConfig struct {
	Name     string `toml:"name"`
	Address  string `toml:"address"`
	Password string `toml:"password"`
	Secure   bool   `toml:"secure"`
}

func (n NodeConfig) String() string {
	return fmt.Sprintf("\n  Name: %s\n  Address: %s\n  Password: %s\n  Secure: %t\n",
		n.Name,
		n.Address,
		strings.Repeat("*", len(n.Password)),
		n.Secure,
	)
}

func (n NodeConfig) ToLavalink(sessionID string) disgolink.NodeConfig {
	return disgolink.NodeConfig{
		Name:      n.Name,
		Address:   n.Address,
		Password:  n.Password,
		Secure:    n.Secure,
		SessionID: sessionID,
	}
}

type Nodes []NodeConfig

func (n Nodes) String() string {
	var s string
	for _, node := range n {
		s += node.String()
	}
	return s
}
