package bot

import (
	"fmt"
	"strings"

	"github.com/disgoorg/disgolink/v2/disgolink"
	"github.com/disgoorg/snowflake/v2"

	"github.com/KittyBot-Org/KittyBotGo/interal/database"
)

type Config struct {
	DevMode      bool            `json:"dev_mode"`
	SyncCommands bool            `json:"sync_commands"`
	GuildIDs     []snowflake.ID  `json:"guild_ids"`
	GatewayURL   string          `json:"gateway_url"`
	RestURL      string          `json:"rest_url"`
	Token        string          `json:"token"`
	LogLevel     string          `json:"log_level"`
	Database     database.Config `json:"database"`
	Nodes        Nodes           `json:"nodes"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n DevMode: %t,\n Sync Commands: %t,\n Guild IDs: %v,\n Token: %s,\n Log Level: %s,\n Database: %s\n", c.DevMode, c.SyncCommands, c.GuildIDs, strings.Repeat("*", len(c.Token)), c.LogLevel, c.Database)
}

type Nodes []disgolink.NodeConfig

func (n Nodes) String() string {
	s := ""
	for _, node := range n {
		s += fmt.Sprintf("\n  Name: %s,\n  Address: %s,\n  Password: %s,\n  Secure: %t,\n  Session ID: %s\n", node.Name, node.Address, strings.Repeat("*", len(node.Password)), node.Secure, node.SessionID)
	}
	return s
}
