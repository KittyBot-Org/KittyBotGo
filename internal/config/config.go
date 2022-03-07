package config

import (
	"os"

	"github.com/DisgoOrg/disgo/json"
	"github.com/DisgoOrg/log"
	"github.com/DisgoOrg/snowflake"
	"github.com/pkg/errors"
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
	DevMode     bool                  `json:"dev_mode"`
	DevGuildIDs []snowflake.Snowflake `json:"dev_guild_ids"`
	DevUserIDs  []snowflake.Snowflake `json:"dev_user_ids"`
	LogLevel    log.Level             `json:"log_level"`
}
