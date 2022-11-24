package config

import (
	"errors"
	"os"

	"github.com/KittyBot-Org/KittyBotGo/db"
	"github.com/disgoorg/json"
	"github.com/disgoorg/log"
	"github.com/disgoorg/snowflake/v2"
)

func LoadConfig(v interface{}) error {
	file, err := os.Open("config.json")
	if os.IsNotExist(err) {
		if file, err = os.Create("config.json"); err != nil {
			return err
		}
		var data []byte
		if data, err = json.MarshalIndent(v, "", "\t"); err != nil {
			return err
		}
		if _, err = file.Write(data); err != nil {
			return err
		}
		return errors.New("config.json not found, created new one")
	} else if err != nil {
		return err
	}
	return json.NewDecoder(file).Decode(v)
}

type Config struct {
	DevMode         bool              `json:"dev_mode"`
	DevGuildIDs     []snowflake.ID    `json:"dev_guild_ids"`
	SupportGuildID  snowflake.ID      `json:"support_guild_id"`
	DevUserIDs      []snowflake.ID    `json:"dev_user_ids"`
	LogLevel        log.Level         `json:"log_level"`
	ErrorLogWebhook LogWebhookConfig  `json:"error_log_webhook"`
	InfoLogWebhook  LogWebhookConfig  `json:"info_log_webhook"`
	Token           string            `json:"token"`
	Database        db.DatabaseConfig `json:"database"`
}

type LogWebhookConfig struct {
	ID    snowflake.ID `json:"id"`
	Token string       `json:"token"`
}
