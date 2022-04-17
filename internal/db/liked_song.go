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
	Get(userID snowflake.Snowflake, title string) (LikedSongs, error)
	GetAll(userID snowflake.Snowflake) ([]LikedSongs, error)
	Add(userID snowflake.Snowflake, query string, title string) error
	Delete(userID snowflake.Snowflake, title string) error
	DeleteAll(userID snowflake.Snowflake) error
}

type likedSongsDBImpl struct {
	db *sql.DB
}

func (s *likedSongsDBImpl) Get(userID snowflake.Snowflake, title string) (LikedSongs, error) {
	var model LikedSongs
	err := SELECT(table.LikedSongs.AllColumns).
		FROM(table.LikedSongs).
		WHERE(table.LikedSongs.UserID.EQ(String(userID.String())).AND(table.LikedSongs.Title.EQ(String(title)))).
		Query(s.db, &model)
	return model, err
}

func (s *likedSongsDBImpl) GetAll(userID snowflake.Snowflake) ([]LikedSongs, error) {
	var models []LikedSongs
	err := SELECT(table.LikedSongs.AllColumns).
		FROM(table.LikedSongs).
		WHERE(table.LikedSongs.UserID.EQ(String(userID.String()))).
		Query(s.db, &models)
	return models, err
}

func (s *likedSongsDBImpl) Add(userID snowflake.Snowflake, query string, title string) error {
	model := LikedSongs{
		UserID:    userID.String(),
		Query:     query,
		Title:     title,
		CreatedAt: time.Now(),
	}
	_, err := table.LikedSongs.INSERT(table.LikedSongs.AllColumns).
		MODEL(model).
		ON_CONFLICT(table.LikedSongs.UserID, table.LikedSongs.Title).
		DO_UPDATE(SET(
			table.LikedSongs.Query.SET(String(model.Query)),
			table.LikedSongs.CreatedAt.SET(TimestampzT(model.CreatedAt)),
		)).
		Exec(s.db)
	return err
}

func (s *likedSongsDBImpl) Delete(userID snowflake.Snowflake, title string) error {
	_, err := table.LikedSongs.DELETE().WHERE(table.LikedSongs.UserID.EQ(String(userID.String())).AND(table.LikedSongs.Title.EQ(String(title)))).Exec(s.db)
	return err
}

func (s *likedSongsDBImpl) DeleteAll(userID snowflake.Snowflake) error {
	_, err := table.LikedSongs.DELETE().WHERE(table.LikedSongs.UserID.EQ(String(userID.String()))).Exec(s.db)
	return err
}
