package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type TagsDB interface {
	Get(guildID snowflake.Snowflake, name string) (Tags, error)
	GetAll(guildID snowflake.Snowflake) ([]Tags, error)
	Set(model Tags) error
	Delete(guildID snowflake.Snowflake, name string) error
}

type tagsDBImpl struct {
	db *sql.DB
}

func (t *tagsDBImpl) Get(guildID snowflake.Snowflake, name string) (Tags, error) {
	return Tags{}, nil
}

func (t *tagsDBImpl) GetAll(guildID snowflake.Snowflake) ([]Tags, error) {
	return nil, nil
}

func (t *tagsDBImpl) Set(model Tags) error {
	return nil
}

func (t *tagsDBImpl) Delete(guildID snowflake.Snowflake, name string) error {
	return nil
}
