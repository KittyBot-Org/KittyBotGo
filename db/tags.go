package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"
)

type TagsDB interface {
	Get(guildID snowflake.ID, name string) (Tag, error)
	GetAll(guildID snowflake.ID) ([]Tag, error)
	Create(guildID snowflake.ID, ownerID snowflake.ID, name string, content string) error
	Edit(guildID snowflake.ID, name string, content string) error
	IncrementUses(guildID snowflake.ID, name string) error
	Delete(guildID snowflake.ID, name string) error
}

type tagsDBImpl struct {
	db *sql.DB
}

func (t *tagsDBImpl) Get(guildID snowflake.ID, name string) (Tag, error) {
	var tag Tag
	err := table.Tag.SELECT(table.Tag.AllColumns).
		WHERE(table.Tag.GuildID.EQ(String(guildID.String())).AND(table.Tag.Name.EQ(String(name)))).
		Query(t.db, &tag)
	return tag, err
}

func (t *tagsDBImpl) GetAll(guildID snowflake.ID) ([]Tag, error) {
	var tags []Tag
	err := table.Tag.SELECT(table.Tag.AllColumns).
		WHERE(table.Tag.GuildID.EQ(String(guildID.String()))).
		Query(t.db, &tags)
	return tags, err
}

func (t *tagsDBImpl) Create(guildID snowflake.ID, ownerID snowflake.ID, name string, content string) error {
	_, err := table.Tag.INSERT(table.Tag.AllColumns).VALUES(guildID, ownerID, name, content, 0, time.Now()).Exec(t.db)
	return err
}

func (t *tagsDBImpl) Edit(guildID snowflake.ID, name string, content string) error {
	_, err := table.Tag.UPDATE(table.Tag.Content).
		SET(content).
		WHERE(table.Tag.GuildID.EQ(String(guildID.String())).AND(table.Tag.Name.EQ(String(name)))).
		Exec(t.db)
	return err
}

func (t *tagsDBImpl) IncrementUses(guildID snowflake.ID, name string) error {
	_, err := table.Tag.UPDATE(table.Tag.Uses).
		SET(table.Tag.Uses.ADD(Int(1))).
		WHERE(table.Tag.GuildID.EQ(String(guildID.String())).AND(table.Tag.Name.EQ(String(name)))).
		Exec(t.db)
	return err
}

func (t *tagsDBImpl) Delete(guildID snowflake.ID, name string) error {
	_, err := table.Tag.DELETE().
		WHERE(table.Tag.GuildID.EQ(String(guildID.String())).AND(table.Tag.Name.EQ(String(name)))).
		Exec(t.db)
	return err
}
