//
// Code generated by go-jet DO NOT EDIT.
//
// WARNING: Changes to this file may cause incorrect behavior
// and will be lost if the code is regenerated
//

package table

import (
	"github.com/go-jet/jet/v2/postgres"
)

var GuildSetting = newGuildSettingTable("public", "guild_setting", "")

type guildSettingTable struct {
	postgres.Table

	//Columns
	ID                        postgres.ColumnString
	ModerationLogWebhookID    postgres.ColumnString
	ModerationLogWebhookToken postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type GuildSettingTable struct {
	guildSettingTable

	EXCLUDED guildSettingTable
}

// AS creates new GuildSettingTable with assigned alias
func (a GuildSettingTable) AS(alias string) *GuildSettingTable {
	return newGuildSettingTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new GuildSettingTable with assigned schema name
func (a GuildSettingTable) FromSchema(schemaName string) *GuildSettingTable {
	return newGuildSettingTable(schemaName, a.TableName(), a.Alias())
}

func newGuildSettingTable(schemaName, tableName, alias string) *GuildSettingTable {
	return &GuildSettingTable{
		guildSettingTable: newGuildSettingTableImpl(schemaName, tableName, alias),
		EXCLUDED:          newGuildSettingTableImpl("", "excluded", ""),
	}
}

func newGuildSettingTableImpl(schemaName, tableName, alias string) guildSettingTable {
	var (
		IDColumn                        = postgres.StringColumn("id")
		ModerationLogWebhookIDColumn    = postgres.StringColumn("moderation_log_webhook_id")
		ModerationLogWebhookTokenColumn = postgres.StringColumn("moderation_log_webhook_token")
		allColumns                      = postgres.ColumnList{IDColumn, ModerationLogWebhookIDColumn, ModerationLogWebhookTokenColumn}
		mutableColumns                  = postgres.ColumnList{ModerationLogWebhookIDColumn, ModerationLogWebhookTokenColumn}
	)

	return guildSettingTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:                        IDColumn,
		ModerationLogWebhookID:    ModerationLogWebhookIDColumn,
		ModerationLogWebhookToken: ModerationLogWebhookTokenColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}