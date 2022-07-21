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
	Get(userID snowflake.ID) ([]PlayHistory, error)
	Add(userID snowflake.ID, query string, title string) error
}

type playHistoriesDBImpl struct {
	db *sql.DB
}

func (h *playHistoriesDBImpl) Get(userID snowflake.ID) ([]PlayHistory, error) {
	var playHistories []PlayHistory
	err := table.PlayHistory.SELECT(table.PlayHistory.AllColumns).
		WHERE(table.PlayHistory.UserID.EQ(String(userID.String()))).
		ORDER_BY(table.PlayHistory.LastUsedAt.DESC()).
		Query(h.db, &playHistories)
	return playHistories, err
}

func (h *playHistoriesDBImpl) Add(userID snowflake.ID, query string, title string) error {
	_, err := table.PlayHistory.INSERT(table.PlayHistory.AllColumns).
		VALUES(String(userID.String()), String(query), String(title), time.Now()).
		ON_CONFLICT(table.PlayHistory.UserID, table.PlayHistory.Title).
		DO_UPDATE(SET(table.PlayHistory.LastUsedAt.SET(TimestampT(time.Now())))).
		Exec(h.db)
	return err
}
