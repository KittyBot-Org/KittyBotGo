package models

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type Voter struct {
	ID        snowflake.Snowflake `bun:"id,pk,notnull"`
	ExpiresAt time.Time           `bun:"expires_at,notnull"`
}
