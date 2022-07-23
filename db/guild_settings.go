package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"
)

type GuildSettingsDB interface {
	CreateIfNotExist(guildID snowflake.ID) error
	Get(guildID snowflake.ID) (GuildSetting, error)
	UpdateModeration(guildID snowflake.ID, webhookID snowflake.ID, webhookToken string) error
	Delete(guildID snowflake.ID) error
}

type guildSettingsDBImpl struct {
	db *sql.DB
}

func (s *guildSettingsDBImpl) CreateIfNotExist(guildID snowflake.ID) error {
	_, err := table.GuildSetting.INSERT(table.GuildSetting.AllColumns).
		VALUES(String(guildID.String()), String("0"), String("")).
		ON_CONFLICT(table.GuildSetting.ID).DO_NOTHING().
		Exec(s.db)
	return err
}

func (s *guildSettingsDBImpl) Get(guildID snowflake.ID) (GuildSetting, error) {
	var model GuildSetting
	return model, table.GuildSetting.SELECT(table.GuildSetting.AllColumns).WHERE(table.GuildSetting.ID.EQ(String(guildID.String()))).Query(s.db, &model)
}

func (s *guildSettingsDBImpl) UpdateModeration(guildID snowflake.ID, webhookID snowflake.ID, webhookToken string) error {
	_, err := table.GuildSetting.
		UPDATE(table.GuildSetting.ModerationLogWebhookID, table.GuildSetting.ModerationLogWebhookToken).
		SET(String(webhookID.String()), String(webhookToken)).
		WHERE(table.GuildSetting.ID.EQ(String(guildID.String()))).
		Exec(s.db)
	return err
}

func (s *guildSettingsDBImpl) Delete(guildID snowflake.ID) error {
	_, err := table.GuildSetting.DELETE().WHERE(table.GuildSetting.ID.EQ(String(guildID.String()))).Exec(s.db)
	return err
}
