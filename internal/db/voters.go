package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/internal/db/.gen/kittybot-go/public/model"
	"github.com/disgoorg/snowflake"
)

type VotersDB interface {
	Get(userID snowflake.Snowflake) (Voters, error)
	GetAll(expiresAt time.Time) ([]Voters, error)
	Set(model Voters) error
	Delete(userID snowflake.Snowflake) error
}

type votersDBImpl struct {
	db *sql.DB
}

func (v *votersDBImpl) Get(userID snowflake.Snowflake) (Voters, error) {
	return Voters{}, nil
}

func (v *votersDBImpl) GetAll(expiresAt time.Time) ([]Voters, error) {
	return nil, nil
}

func (v *votersDBImpl) Set(model Voters) error {
	return nil
}

func (v *votersDBImpl) Delete(userID snowflake.Snowflake) error {
	return nil
}
