package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type PlayHistoriesDB interface {
	Get(userID snowflake.Snowflake) ([]PlayHistories, error)
	Add(model PlayHistories) error
}

type PlayHistoriesDBImpl struct {
	db *sql.DB
}
