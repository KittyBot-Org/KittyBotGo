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

var GuildSettings = newGuildSettingsTable("public", "guild_settings", "")

type guildSettingsTable struct {
	postgres.Table

	//Columns
	ID                        postgres.ColumnString
	ModerationLogWebhookID    postgres.ColumnString
	ModerationLogWebhookToken postgres.ColumnString

	AllColumns     postgres.ColumnList
	MutableColumns postgres.ColumnList
}

type GuildSettingsTable struct {
	guildSettingsTable

	EXCLUDED guildSettingsTable
}

// AS creates new GuildSettingsTable with assigned alias
func (a GuildSettingsTable) AS(alias string) *GuildSettingsTable {
	return newGuildSettingsTable(a.SchemaName(), a.TableName(), alias)
}

// Schema creates new GuildSettingsTable with assigned schema name
func (a GuildSettingsTable) FromSchema(schemaName string) *GuildSettingsTable {
	return newGuildSettingsTable(schemaName, a.TableName(), a.Alias())
}

// WithPrefix creates new GuildSettingsTable with assigned table prefix
func (a GuildSettingsTable) WithPrefix(prefix string) *GuildSettingsTable {
	return newGuildSettingsTable(a.SchemaName(), prefix+a.TableName(), a.TableName())
}

// WithSuffix creates new GuildSettingsTable with assigned table suffix
func (a GuildSettingsTable) WithSuffix(suffix string) *GuildSettingsTable {
	return newGuildSettingsTable(a.SchemaName(), a.TableName()+suffix, a.TableName())
}

func newGuildSettingsTable(schemaName, tableName, alias string) *GuildSettingsTable {
	return &GuildSettingsTable{
		guildSettingsTable: newGuildSettingsTableImpl(schemaName, tableName, alias),
		EXCLUDED:           newGuildSettingsTableImpl("", "excluded", ""),
	}
}

func newGuildSettingsTableImpl(schemaName, tableName, alias string) guildSettingsTable {
	var (
		IDColumn                        = postgres.StringColumn("id")
		ModerationLogWebhookIDColumn    = postgres.StringColumn("moderation_log_webhook_id")
		ModerationLogWebhookTokenColumn = postgres.StringColumn("moderation_log_webhook_token")
		allColumns                      = postgres.ColumnList{IDColumn, ModerationLogWebhookIDColumn, ModerationLogWebhookTokenColumn}
		mutableColumns                  = postgres.ColumnList{ModerationLogWebhookIDColumn, ModerationLogWebhookTokenColumn}
	)

	return guildSettingsTable{
		Table: postgres.NewTable(schemaName, tableName, alias, allColumns...),

		//Columns
		ID:                        IDColumn,
		ModerationLogWebhookID:    ModerationLogWebhookIDColumn,
		ModerationLogWebhookToken: ModerationLogWebhookTokenColumn,

		AllColumns:     allColumns,
		MutableColumns: mutableColumns,
	}
}