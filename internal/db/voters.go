package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type VotersDB interface {
	Get(userID snowflake.Snowflake) (Voter, error)
	GetAll(expiresAt time.Time) ([]Voter, error)
	Add(userID snowflake.Snowflake, duration time.Duration) error
	Delete(userID snowflake.Snowflake) error
}

type votersDBImpl struct {
	db *sql.DB
}

func (v *votersDBImpl) Get(userID snowflake.Snowflake) (Voter, error) {
	return Voter{}, nil
}

func (v *votersDBImpl) GetAll(expiresAt time.Time) ([]Voter, error) {
	return nil, nil
}

func (v *votersDBImpl) Add(userID snowflake.Snowflake, duration time.Duration) error {
	return nil
}

func (v *votersDBImpl) Delete(userID snowflake.Snowflake) error {
	return nil
}
