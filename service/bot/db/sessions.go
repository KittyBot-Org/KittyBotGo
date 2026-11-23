package db

import (
	"context"
	"time"

	"github.com/disgoorg/snowflake/v2"
)

type Session struct {
	ID           int          `db:"id"`
	UserID       snowflake.ID `db:"user_id"`
	AccessToken  string       `db:"access_token"`
	RefreshToken string       `db:"refresh_token"`
	Expiration   time.Time    `db:"expiration"`
}

func (d *DB) UpsertSession(ctx context.Context, session Session) error {
	_, err := d.dbx.ExecContext(ctx, `INSERT INTO sessions (user_id, access_token, refresh_token, expiration) VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token, expiration = EXCLUDED.expiration`,
		session.UserID, session.AccessToken, session.RefreshToken, session.Expiration)
	return err
}

func (d *DB) GetSession(ctx context.Context, userID snowflake.ID) (Session, error) {
	var session Session
	err := d.dbx.GetContext(ctx, &session, `SELECT id, user_id, access_token, refresh_token, expiration FROM sessions WHERE user_id = $1`, userID)
	return session, err
}

func (d *DB) DeleteSession(ctx context.Context, userID snowflake.ID) error {
	_, err := d.dbx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, userID)
	return err
}

func (d *DB) UpsertVote(ctx context.Context, userID snowflake.ID, expiry time.Time) error {
	_, err := d.dbx.ExecContext(ctx, `INSERT INTO voters (user_id, vote_expiry) VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET vote_expiry = EXCLUDED.vote_expiry`, userID, expiry)
	return err
}

func (d *DB) VoteExpiry(ctx context.Context, userID snowflake.ID) (time.Time, error) {
	var expiry time.Time
	err := d.dbx.GetContext(ctx, &expiry, `SELECT vote_expiry FROM voters WHERE user_id = $1`, userID)
	return expiry, err
}
