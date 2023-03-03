package db

import (
	"database/sql"
	"time"

	"github.com/disgoorg/snowflake/v2"
	. "github.com/go-jet/jet/v2/postgres"

	. "github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/model"
	"github.com/KittyBot-Org/KittyBotGo/db/.gen/kittybot-go/public/table"
)

type ReportsDB interface {
	Get(id int32) (Reports, error)
	GetCount(userID snowflake.ID, guildID snowflake.ID) (int, error)
	GetAll(userID snowflake.ID, guildID snowflake.ID) ([]Reports, error)
	Create(userID snowflake.ID, guildID snowflake.ID, description string, createdAt time.Time, messageID snowflake.ID, channelID snowflake.ID) (int32, error)
	Confirm(id int32) error
	Delete(id int32) error
	DeleteAll(userID snowflake.ID, guildID snowflake.ID) error
}

type reportsDBImpl struct {
	db *sql.DB
}

func (s *reportsDBImpl) Get(id int32) (Reports, error) {
	var model Reports
	err := table.Reports.SELECT(table.Reports.AllColumns).
		WHERE(table.Reports.ID.EQ(Int32(id))).
		Query(s.db, &model)
	return model, err
}

func (s *reportsDBImpl) GetCount(userID snowflake.ID, guildID snowflake.ID) (int, error) {
	var count struct {
		Count int
	}
	err := table.Reports.SELECT(COUNT(table.Reports.ID)).
		WHERE(table.Reports.UserID.EQ(String(userID.String())).AND(table.Reports.GuildID.EQ(String(guildID.String()))).AND(table.Reports.Confirmed.EQ(Bool(true)))).
		Query(s.db, &count)
	return count.Count, err
}

func (s *reportsDBImpl) GetAll(userID snowflake.ID, guildID snowflake.ID) ([]Reports, error) {
	var model []Reports
	err := table.Reports.SELECT(table.Reports.AllColumns).
		WHERE(table.Reports.UserID.EQ(String(userID.String())).AND(table.Reports.GuildID.EQ(String(guildID.String())))).
		ORDER_BY(table.Reports.CreatedAt.ASC()).
		Query(s.db, &model)
	return model, err
}

func (s *reportsDBImpl) Create(userID snowflake.ID, guildID snowflake.ID, description string, createdAt time.Time, messageID snowflake.ID, channelID snowflake.ID) (int32, error) {
	var model Reports
	err := table.Reports.INSERT(table.Reports.UserID, table.Reports.GuildID, table.Reports.Description, table.Reports.CreatedAt, table.Reports.MessageID, table.Reports.ChannelID).
		VALUES(userID, guildID, description, createdAt, messageID, channelID).
		RETURNING(table.Reports.ID).
		Query(s.db, &model)
	return model.ID, err
}

func (s *reportsDBImpl) Confirm(id int32) error {
	_, err := table.Reports.UPDATE(table.Reports.Confirmed).SET(Bool(true)).WHERE(table.Reports.ID.EQ(Int32(id))).Exec(s.db)
	return err
}

func (s *reportsDBImpl) Delete(id int32) error {
	_, err := table.Reports.DELETE().WHERE(table.Reports.ID.EQ(Int32(id))).Exec(s.db)
	return err
}

func (s *reportsDBImpl) DeleteAll(userID snowflake.ID, guildID snowflake.ID) error {
	_, err := table.Reports.DELETE().WHERE(table.Reports.UserID.EQ(String(userID.String())).AND(table.Reports.GuildID.EQ(String(guildID.String())))).Exec(s.db)
	return err
}
