package db

import (
	"database/sql"
	"time"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"
)

type VotersDB interface {
	GetAll(expiresAt time.Time) ([]Voter, error)
	Add(userID snowflake.ID, duration time.Duration) error
	Delete(userID snowflake.ID) error
}

type votersDBImpl struct {
	db *sql.DB
}

func (v *votersDBImpl) GetAll(expiresAt time.Time) ([]Voter, error) {
	var voters []Voter
	err := table.Voter.SELECT(table.Voter.AllColumns).
		WHERE(table.Voter.ExpiresAt.LT(TimestampT(expiresAt))).
		Query(v.db, &voters)
	return voters, err
}

func (v *votersDBImpl) Add(userID snowflake.ID, duration time.Duration) error {
	_, err := table.Voter.INSERT(table.Voter.AllColumns).
		VALUES(userID, time.Now().Add(duration)).
		ON_CONFLICT(table.Voter.UserID).
		DO_UPDATE(SET(table.Voter.ExpiresAt.SET(table.Voter.ExpiresAt.ADD(INTERVALd(duration))))).
		Exec(v.db)
	return err
}

func (v *votersDBImpl) Delete(userID snowflake.ID) error {
	_, err := table.Voter.DELETE().
		WHERE(table.Voter.UserID.EQ(String(userID.String()))).
		Exec(v.db)
	return err
}
