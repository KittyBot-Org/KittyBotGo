package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"
)

type LikedSongsDB interface {
	Get(userID snowflake.ID, title string) (LikedSong, error)
	GetAll(userID snowflake.ID) ([]LikedSong, error)
	Add(userID snowflake.ID, query string, title string) error
	Delete(userID snowflake.ID, title string) error
	DeleteAll(userID snowflake.ID) error
}

type likedSongsDBImpl struct {
	db *sql.DB
}

func (s *likedSongsDBImpl) Get(userID snowflake.ID, title string) (LikedSong, error) {
	var model LikedSong
	err := table.LikedSong.SELECT(table.LikedSong.AllColumns).
		WHERE(table.LikedSong.UserID.EQ(String(userID.String())).AND(table.LikedSong.Title.EQ(String(title)))).
		Query(s.db, &model)
	return model, err
}

func (s *likedSongsDBImpl) GetAll(userID snowflake.ID) ([]LikedSong, error) {
	var models []LikedSong
	err := table.LikedSong.SELECT(table.LikedSong.AllColumns).
		WHERE(table.LikedSong.UserID.EQ(String(userID.String()))).
		Query(s.db, &models)
	return models, err
}

func (s *likedSongsDBImpl) Add(userID snowflake.ID, query string, title string) error {
	_, err := table.LikedSong.INSERT(table.LikedSong.AllColumns).
		VALUES(userID, query, title, time.Now()).
		ON_CONFLICT(table.LikedSong.UserID, table.LikedSong.Title).
		DO_UPDATE(SET(
			table.LikedSong.Query.SET(String(query)),
			table.LikedSong.CreatedAt.SET(TimestampT(time.Now())),
		)).
		Exec(s.db)
	return err
}

func (s *likedSongsDBImpl) Delete(userID snowflake.ID, title string) error {
	_, err := table.LikedSong.DELETE().WHERE(table.LikedSong.UserID.EQ(String(userID.String())).AND(table.LikedSong.Title.EQ(String(title)))).Exec(s.db)
	return err
}

func (s *likedSongsDBImpl) DeleteAll(userID snowflake.ID) error {
	_, err := table.LikedSong.DELETE().WHERE(table.LikedSong.UserID.EQ(String(userID.String()))).Exec(s.db)
	return err
}
