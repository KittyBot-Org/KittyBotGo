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
		db:            db,
		guildSettings: &guildSettingsDBImpl{db: db},
		likedTracks:   &likedTracksDBImpl{db: db},
		playHistory:   &playHistoriesDBImpl{db: db},
		tags:          &tagsDBImpl{db: db},
		voters:        &votersDBImpl{db: db},
		reports:       &reportsDBImpl{db: db},
	}, nil
}

type DB interface {
	GuildSettings() GuildSettingsDB
	LikedTracks() LikedTracksDB
	PlayHistory() PlayHistoriesDB
	Tags() TagsDB
	Voters() VotersDB
	Reports() ReportsDB
	Close() error
}

type dbImpl struct {
	db            *sql.DB
	guildSettings GuildSettingsDB
	likedTracks   LikedTracksDB
	playHistory   PlayHistoriesDB
	tags          TagsDB
	voters        VotersDB
	reports       ReportsDB
}

func (d *dbImpl) GuildSettings() GuildSettingsDB {
	return d.guildSettings
}

func (d *dbImpl) LikedTracks() LikedTracksDB {
	return d.likedTracks
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

func (d *dbImpl) Reports() ReportsDB {
	return d.reports
}

func (d *dbImpl) Close() error {
	return d.db.Close()
}
