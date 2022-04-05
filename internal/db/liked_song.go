package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake"
	. "github.com/go-jet/jet/v2/postgres"
)

type LikedSongsDB interface {
	Get(userID snowflake.Snowflake) ([]LikedSongs, error)
	Add(model LikedSongs) error
	Delete(userID snowflake.Snowflake, title string) error
}

type likedSongsDBImpl struct {
	db *sql.DB
}

func (s *likedSongsDBImpl) Get(userID snowflake.Snowflake) ([]LikedSongs, error) {
	var models []LikedSongs
	err := SELECT(table.LikedSongs.AllColumns).
		FROM(table.LikedSongs).
		WHERE(table.LikedSongs.UserID.EQ(String(userID.String()))).
		Query(s.db, &models)

	return models, err
}
func (s *likedSongsDBImpl) Add(likedSong LikedSongs) error {
	_, err := table.LikedSongs.INSERT(table.LikedSongs.AllColumns).
		MODEL(likedSong).
		ON_CONFLICT(table.LikedSongs.UserID, table.LikedSongs.Query).
		DO_UPDATE(SET(
			table.LikedSongs.Title.SET(String(likedSong.Title)),
			table.LikedSongs.CreatedAt.SET(likedSong.CreatedAt),
		)).
		Exec(s.db)
	return err
}

func (s *likedSongsDBImpl) Delete(userID snowflake.Snowflake, title string) error {
	_, err := table.LikedSongs.DELETE().WHERE(table.LikedSongs.UserID.EQ(String(userID.String())).AND(table.LikedSongs.Title.EQ(String(title)))).Exec(s.db)
	return err
}
