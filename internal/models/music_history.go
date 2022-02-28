package models

import (
	"context"
	"time"

	"github.com/DisgoOrg/snowflake"
	"github.com/uptrace/bun"
)

type PlayHistory struct {
	ID         int                 `bun:"id,type:bigserial,pk,notnull"`
	UserID     snowflake.Snowflake `bun:"user_id,notnull"`
	Query      string              `bun:"query,notnull"`
	Title      string              `bun:"title,notnull"`
	LastUsedAt time.Time           `bun:"last_used_at,nullzero,notnull,default:current_timestamp"`
}

var _ bun.BeforeAppendModelHook = (*Tag)(nil)

func (t *PlayHistory) BeforeAppendModel(_ context.Context, query bun.Query) error {
	switch query.(type) {
	case *bun.InsertQuery, *bun.UpdateQuery:
		t.LastUsedAt = time.Now()
	}
	return nil
}
