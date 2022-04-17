package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type PlayHistoriesDB interface {
	Get(userID snowflake.Snowflake) ([]PlayHistories, error)
	Add(userID snowflake.Snowflake, query string, title string) error
}

type playHistoriesDBImpl struct {
	db *sql.DB
}

func (h *playHistoriesDBImpl) Get(userID snowflake.Snowflake) ([]PlayHistories, error) {
	return nil, nil
}

func (h *playHistoriesDBImpl) Add(userID snowflake.Snowflake, query string, title string) error {
	return nil
}
