package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake"
	. "github.com/go-jet/jet/v2/postgres"
)

type LikedSongsDB interface {
	Get(userID snowflake.Snowflake, title string) (LikedSong, error)
	GetAll(userID snowflake.Snowflake) ([]LikedSong, error)
	Add(userID snowflake.Snowflake, query string, title string) error
	Delete(userID snowflake.Snowflake, title string) error
	DeleteAll(userID snowflake.Snowflake) error
}

type likedSongsDBImpl struct {
	db *sql.DB
}

func (s *likedSongsDBImpl) Get(userID snowflake.Snowflake, title string) (LikedSong, error) {
	var model LikedSong
	err := SELECT(table.LikedSong.AllColumns).
		FROM(table.LikedSong).
		WHERE(table.LikedSong.UserID.EQ(String(userID.String())).AND(table.LikedSong.Title.EQ(String(title)))).
		Query(s.db, &model)
	return model, err
}

func (s *likedSongsDBImpl) GetAll(userID snowflake.Snowflake) ([]LikedSong, error) {
	var models []LikedSong
	err := SELECT(table.LikedSong.AllColumns).
		FROM(table.LikedSong).
		WHERE(table.LikedSong.UserID.EQ(String(userID.String()))).
		Query(s.db, &models)
	return models, err
}

func (s *likedSongsDBImpl) Add(userID snowflake.Snowflake, query string, title string) error {
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

func (s *likedSongsDBImpl) Delete(userID snowflake.Snowflake, title string) error {
	_, err := table.LikedSong.DELETE().WHERE(table.LikedSong.UserID.EQ(String(userID.String())).AND(table.LikedSong.Title.EQ(String(title)))).Exec(s.db)
	return err
}

func (s *likedSongsDBImpl) DeleteAll(userID snowflake.Snowflake) error {
	_, err := table.LikedSong.DELETE().WHERE(table.LikedSong.UserID.EQ(String(userID.String()))).Exec(s.db)
	return err
}
