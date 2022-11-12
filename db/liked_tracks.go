package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"
)

type LikedTracksDB interface {
	Get(userID snowflake.ID, title string) (LikedTracks, error)
	GetAll(userID snowflake.ID) ([]LikedTracks, error)
	Add(userID snowflake.ID, query string, title string) error
	Delete(userID snowflake.ID, title string) error
	DeleteAll(userID snowflake.ID) error
}

type likedTracksDBImpl struct {
	db *sql.DB
}

func (s *likedTracksDBImpl) Get(userID snowflake.ID, title string) (LikedTracks, error) {
	var model LikedTracks
	err := table.LikedTracks.SELECT(table.LikedTracks.AllColumns).
		WHERE(table.LikedTracks.UserID.EQ(String(userID.String())).AND(table.LikedTracks.Title.EQ(String(title)))).
		Query(s.db, &model)
	return model, err
}

func (s *likedTracksDBImpl) GetAll(userID snowflake.ID) ([]LikedTracks, error) {
	var models []LikedTracks
	err := table.LikedTracks.SELECT(table.LikedTracks.AllColumns).
		WHERE(table.LikedTracks.UserID.EQ(String(userID.String()))).
		Query(s.db, &models)
	return models, err
}

func (s *likedTracksDBImpl) Add(userID snowflake.ID, query string, title string) error {
	_, err := table.LikedTracks.INSERT(table.LikedTracks.AllColumns).
		VALUES(userID, query, title, time.Now()).
		ON_CONFLICT(table.LikedTracks.UserID, table.LikedTracks.Title).
		DO_UPDATE(SET(
			table.LikedTracks.Query.SET(String(query)),
			table.LikedTracks.LikedAt.SET(TimestampT(time.Now())),
		)).
		Exec(s.db)
	return err
}

func (s *likedTracksDBImpl) Delete(userID snowflake.ID, title string) error {
	_, err := table.LikedTracks.DELETE().WHERE(table.LikedTracks.UserID.EQ(String(userID.String())).AND(table.LikedTracks.Title.EQ(String(title)))).Exec(s.db)
	return err
}

func (s *likedTracksDBImpl) DeleteAll(userID snowflake.ID) error {
	_, err := table.LikedTracks.DELETE().WHERE(table.LikedTracks.UserID.EQ(String(userID.String()))).Exec(s.db)
	return err
}
