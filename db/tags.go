package db

import (
	"database/sql"
	"time"

	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
)

type TagsDB interface {
	Get(guildID snowflake.ID, name string) (Tags, error)
	GetAll(guildID snowflake.ID) ([]Tags, error)
	Create(guildID snowflake.ID, ownerID snowflake.ID, name string, content string) error
	Edit(guildID snowflake.ID, name string, content string) error
	IncrementUses(guildID snowflake.ID, name string) error
	Delete(guildID snowflake.ID, name string) error
}

type tagsDBImpl struct {
	db *sql.DB
}

func (t *tagsDBImpl) Get(guildID snowflake.ID, name string) (Tags, error) {
	var Tags Tags
	err := table.Tags.SELECT(table.Tags.AllColumns).
		WHERE(table.Tags.GuildID.EQ(String(guildID.String())).AND(table.Tags.Name.EQ(String(name)))).
		Query(t.db, &Tags)
	return Tags, err
}

func (t *tagsDBImpl) GetAll(guildID snowflake.ID) ([]Tags, error) {
	var tags []Tags
	err := table.Tags.SELECT(table.Tags.AllColumns).
		WHERE(table.Tags.GuildID.EQ(String(guildID.String()))).
		Query(t.db, &tags)
	return tags, err
}

func (t *tagsDBImpl) Create(guildID snowflake.ID, ownerID snowflake.ID, name string, content string) error {
	_, err := table.Tags.INSERT(table.Tags.AllColumns).VALUES(guildID, ownerID, name, content, 0, time.Now()).Exec(t.db)
	return err
}

func (t *tagsDBImpl) Edit(guildID snowflake.ID, name string, content string) error {
	_, err := table.Tags.UPDATE(table.Tags.Content).
		SET(content).
		WHERE(table.Tags.GuildID.EQ(String(guildID.String())).AND(table.Tags.Name.EQ(String(name)))).
		Exec(t.db)
	return err
}

func (t *tagsDBImpl) IncrementUses(guildID snowflake.ID, name string) error {
	_, err := table.Tags.UPDATE(table.Tags.Uses).
		SET(table.Tags.Uses.ADD(Int(1))).
		WHERE(table.Tags.GuildID.EQ(String(guildID.String())).AND(table.Tags.Name.EQ(String(name)))).
		Exec(t.db)
	return err
}

func (t *tagsDBImpl) Delete(guildID snowflake.ID, name string) error {
	_, err := table.Tags.DELETE().
		WHERE(table.Tags.GuildID.EQ(String(guildID.String())).AND(table.Tags.Name.EQ(String(name)))).
		Exec(t.db)
	return err
}
