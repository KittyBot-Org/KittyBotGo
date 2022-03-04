package types

import (
	"os"

	"github.com/DisgoOrg/disgo/json"
	"github.com/DisgoOrg/disgolink/lavalink"
	"github.com/DisgoOrg/log"
	"github.com/DisgoOrg/snowflake"
	"github.com/pkg/errors"
)

func (b *Bot) LoadConfig() error {
	b.Logger.Info("Loading config...")
	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		if file, err = os.Create("config.json"); err != nil {
			return err
		}
		var data []byte
		if data, err = json.MarshalIndent(Config{}, "", "\t"); err != nil {
			return err
		}
		if _, err = file.Write(data); err != nil {
			return err
		}
		return errors.New("config.json not found, created new one")
	} else if err != nil {
		return err
	}

	var cfg Config
	if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		return err
	}
	b.Config = cfg
	return nil
}

type Config struct {
	DevMode     bool                  `json:"dev_mode"`
	DevGuildIDs []snowflake.Snowflake `json:"dev_guild_ids"`
	DevUserIDs  []snowflake.Snowflake `json:"dev_user_ids"`
	LogLevel    log.Level             `json:"log_level"`
	Bot         BotConfig             `json:"bot"`
	Database    DatabaseConfig        `json:"database"`
	Lavalink    LavalinkConfig        `json:"lavalink"`
}

type BotConfig struct {
	Token      string `json:"token"`
	ShardIDs   []int  `json:"shard_ids"`
	ShardCount int    `json:"shard_count"`
}

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
}

type LavalinkConfig struct {
	Nodes           []lavalink.NodeConfig `json:"nodes"`
	ResumingTimeout int                   `json:"resuming_timeout"`
}
