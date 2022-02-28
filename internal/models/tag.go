package models

import (
	"context"
	"time"

	"github.com/DisgoOrg/snowflake"
	"github.com/uptrace/bun"
)

type Tag struct {
	ID        int                 `bun:"id,type:bigserial,pk,notnull"`
	GuildID   snowflake.Snowflake `bun:"guild_id,notnull,unique:name-guild"`
	OwnerID   snowflake.Snowflake `bun:"owner_id,notnull"`
	Name      string              `bun:"name,notnull,unique:name-guild"`
	Content   string              `bun:"content,notnull"`
	Uses      int                 `bun:"uses,default:0"`
	CreatedAt time.Time           `bun:",nullzero,notnull,default:current_timestamp"`
}

var _ bun.BeforeAppendModelHook = (*Tag)(nil)

func (t *Tag) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery:
		t.CreatedAt = time.Now()
	}
	return nil
}
