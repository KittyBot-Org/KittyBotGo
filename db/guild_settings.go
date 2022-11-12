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
	Get(guildID snowflake.ID) (GuildSettings, error)
	UpdateModeration(guildID snowflake.ID, webhookID snowflake.ID, webhookToken string) error
	Delete(guildID snowflake.ID) error
}

type guildSettingsDBImpl struct {
	db *sql.DB
}

func (s *guildSettingsDBImpl) CreateIfNotExist(guildID snowflake.ID) error {
	_, err := table.GuildSettings.INSERT(table.GuildSettings.AllColumns).
		VALUES(String(guildID.String()), String("0"), String("")).
		ON_CONFLICT(table.GuildSettings.ID).DO_NOTHING().
		Exec(s.db)
	return err
}

func (s *guildSettingsDBImpl) Get(guildID snowflake.ID) (GuildSettings, error) {
	var model GuildSettings
	return model, table.GuildSettings.SELECT(table.GuildSettings.AllColumns).WHERE(table.GuildSettings.ID.EQ(String(guildID.String()))).Query(s.db, &model)
}

func (s *guildSettingsDBImpl) UpdateModeration(guildID snowflake.ID, webhookID snowflake.ID, webhookToken string) error {
	_, err := table.GuildSettings.
		UPDATE(table.GuildSettings.ModerationLogWebhookID, table.GuildSettings.ModerationLogWebhookToken).
		SET(String(webhookID.String()), String(webhookToken)).
		WHERE(table.GuildSettings.ID.EQ(String(guildID.String()))).
		Exec(s.db)
	return err
}

func (s *guildSettingsDBImpl) Delete(guildID snowflake.ID) error {
	_, err := table.GuildSettings.DELETE().WHERE(table.GuildSettings.ID.EQ(String(guildID.String()))).Exec(s.db)
	return err
}
