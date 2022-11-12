package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"
)

type PlayHistoriesDB interface {
	Get(userID snowflake.ID) ([]PlayHistories, error)
	Add(userID snowflake.ID, query string, title string) error
}

type playHistoriesDBImpl struct {
	db *sql.DB
}

func (h *playHistoriesDBImpl) Get(userID snowflake.ID) ([]PlayHistories, error) {
	var playHistories []PlayHistories
	err := table.PlayHistories.SELECT(table.PlayHistories.AllColumns).
		WHERE(table.PlayHistories.UserID.EQ(String(userID.String()))).
		ORDER_BY(table.PlayHistories.LastUsedAt.DESC()).
		Query(h.db, &playHistories)
	return playHistories, err
}

func (h *playHistoriesDBImpl) Add(userID snowflake.ID, query string, title string) error {
	_, err := table.PlayHistories.INSERT(table.PlayHistories.AllColumns).
		VALUES(String(userID.String()), String(query), String(title), time.Now()).
		ON_CONFLICT(table.PlayHistories.UserID, table.PlayHistories.Title).
		DO_UPDATE(SET(table.PlayHistories.LastUsedAt.SET(TimestampT(time.Now())))).
		Exec(h.db)
	return err
}
