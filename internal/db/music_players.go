package db

import (
	"database/sql"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake"
	. "github.com/go-jet/jet/v2/postgres"
)

type MusicPlayersDB interface {
	GetAndDelete(guildID snowflake.Snowflake) (MusicPlayers, error)
	Add(model MusicPlayers) error
}

type musicPlayersDBImpl struct {
	db *sql.DB
}

func (s *musicPlayersDBImpl) GetAndDelete(guildID snowflake.Snowflake) (MusicPlayers, error) {
	var models []MusicPlayers
	err := SELECT(table.LikedSongs.AllColumns).
		FROM(table.LikedSongs).
		WHERE(table.LikedSongs.UserID.EQ(String(userID.String()))).
		Query(s.db, &models)

	return models, err
}
func (s *musicPlayersDBImpl) Add(model MusicPlayers) error {
	_, err := table.MusicPlayers.INSERT(table.MusicPlayers.AllColumns).MODEL(model).ON_CONFLICT(table.MusicPlayers.GuildID).DO_UPDATE(nil).Exec(s.db)
	return err
}
