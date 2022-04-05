package db

import (
	"time"

	"github.com/disgoorg/snowflake"
)

type VotersDB interface {
	Get(userID snowflake.Snowflake) (VoterModel, error)
	GetAll(expiresAt time.Time) ([]VoterModel, error)
	Set(model VoterModel) error
	Delete(userID snowflake.Snowflake) error
}

type VoterModel struct {
	UserID    snowflake.Snowflake `bun:"user_id,pk,notnull"`
	ExpiresAt time.Time           `bun:"expires_at,notnull"`
}
