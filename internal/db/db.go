package db

import (
	"database/sql"
	"fmt"
)

type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"db_name"`
	SSLMode  string `json:"ssl_mode"`
}

func SetupDatabase(config DatabaseConfig) (DB, error) {
	db, err := sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode))
	if err != nil {
		return nil, err
	}

	return &dbImpl{
		likedSongs: &likedSongsDBImpl{db: db},
	}, nil
}

type DB interface {
	GuildSettings() GuildSettingsDB
	LikedSongs() LikedSongsDB
	MusicPlayers() MusicPlayersDB
	PlayHistory() PlayHistoriesDB
	Tags() TagsDB
	Voters() VotersDB
}

type dbImpl struct {
	guildSettings GuildSettingsDB
	likedSongs    LikedSongsDB
	musicPlayers  MusicPlayersDB
	playHistory   PlayHistoriesDB
	tags          TagsDB
	voters        VotersDB
}

func (d *dbImpl) GuildSettings() GuildSettingsDB {
	return d.guildSettings
}

func (d *dbImpl) LikedSongs() LikedSongsDB {
	return d.likedSongs
}

func (d *dbImpl) MusicPlayers() MusicPlayersDB {
	return d.musicPlayers
}

func (d *dbImpl) PlayHistory() PlayHistoriesDB {
	return d.playHistory
}

func (d *dbImpl) Tags() TagsDB {
	return d.tags
}

func (d *dbImpl) Voters() VotersDB {
	return d.voters
}
