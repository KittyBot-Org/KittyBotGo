package db

import (
	"database/sql"
	"time"

	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
)

type VotersDB interface {
	GetAll(expiresAt time.Time) ([]Voters, error)
	Add(userID snowflake.ID, duration time.Duration) error
	Delete(userID snowflake.ID) error
}

type votersDBImpl struct {
	db *sql.DB
}

func (v *votersDBImpl) GetAll(expiresAt time.Time) ([]Voters, error) {
	var voters []Voters
	err := table.Voters.SELECT(table.Voters.AllColumns).
		WHERE(table.Voters.ExpiresAt.LT(TimestampT(expiresAt))).
		Query(v.db, &voters)
	return voters, err
}

func (v *votersDBImpl) Add(userID snowflake.ID, duration time.Duration) error {
	_, err := table.Voters.INSERT(table.Voters.AllColumns).
		VALUES(userID, time.Now().Add(duration)).
		ON_CONFLICT(table.Voters.UserID).
		DO_UPDATE(SET(table.Voters.ExpiresAt.SET(table.Voters.ExpiresAt.ADD(INTERVALd(duration))))).
		Exec(v.db)
	return err
}

func (v *votersDBImpl) Delete(userID snowflake.ID) error {
	_, err := table.Voters.DELETE().
		WHERE(table.Voters.UserID.EQ(String(userID.String()))).
		Exec(v.db)
	return err
}
