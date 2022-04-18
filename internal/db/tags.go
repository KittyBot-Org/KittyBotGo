package db

import (
	"database/sql"
	"github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/table"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type TagsDB interface {
	Get(guildID snowflake.Snowflake, name string) (Tag, error)
	GetAll(guildID snowflake.Snowflake) ([]Tag, error)
	Create(guildID snowflake.Snowflake, ownerID snowflake.Snowflake, name string, content string) error
	Edit(guildID snowflake.Snowflake, name string, content string) error
	IncrementUses(guildID snowflake.Snowflake, name string) error
	Delete(guildID snowflake.Snowflake, name string) error
}

type tagsDBImpl struct {
	db *sql.DB
}

func (t *tagsDBImpl) Get(guildID snowflake.Snowflake, name string) (Tag, error) {
	return Tag{}, nil
}

func (t *tagsDBImpl) GetAll(guildID snowflake.Snowflake) ([]Tag, error) {
	return nil, nil
}

func (t *tagsDBImpl) Create(guildID snowflake.Snowflake, ownerID snowflake.Snowflake, name string, content string) error {
	_, err := table.Tag.INSERT(table.Tag.AllColumns).VALUES(guildID, ownerID, name, content, 0, time.Now()).Exec(t.db)
	return err
}

func (t *tagsDBImpl) Edit(guildID snowflake.Snowflake, name string, content string) error {
	return nil
}

func (t *tagsDBImpl) IncrementUses(guildID snowflake.Snowflake, name string) error {
	return nil
}

func (t *tagsDBImpl) Delete(guildID snowflake.Snowflake, name string) error {
	return nil
}
