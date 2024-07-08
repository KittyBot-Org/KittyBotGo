package bot

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgolink/v3/disgolink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/KittyBot-Org/KittyBotGo/internal/log"
	"github.com/KittyBot-Org/KittyBotGo/service/bot/db"
)

type Config struct {
	Log      log.Config `toml:"log"`
	Bot      BotConfig  `toml:"bot"`
	Database db.Config  `json:"database"`
	Nodes    Nodes      `json:"nodes"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n Log: %v\n Bot: %s\n Database: %s\n Nodes: %s\n",
		c.Log,
		c.Bot,
		c.Database,
		c.Nodes,
	)
}

type BotConfig struct {
	SyncCommands bool           `toml:"sync_commands"`
	GuildIDs     []snowflake.ID `toml:"guild_ids"`
	GatewayURL   string         `toml:"gateway_url"`
	RestURL      string         `toml:"rest_url"`
	Token        string         `toml:"token"`
}

func (c BotConfig) String() string {
	return fmt.Sprintf("\n  SyncCommands: %t\n  GuildIDs: %v\n  GatewayURL: %s\n  RestURL: %s\n  Token: %s\n",
		c.SyncCommands,
		c.GuildIDs,
		c.GatewayURL,
		c.RestURL,
		strings.Repeat("*", len(c.Token)),
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
