-- name: UpsertSession :exec
INSERT INTO sessions (user_id, access_token, refresh_token, expiration)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_id) DO UPDATE
    SET access_token  = EXCLUDED.access_token,
        refresh_token = EXCLUDED.refresh_token,
        expiration    = EXCLUDED.expiration;

-- name: GetSession :one
SELECT id, user_id, access_token, refresh_token, expiration
FROM sessions
WHERE user_id = $1;

-- name: DeleteSession :exec
DELETE
FROM sessions
WHERE user_id = $1;

-- name: UpsertVote :exec
INSERT INTO voters (user_id, vote_expiry)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
    SET vote_expiry = EXCLUDED.vote_expiry;

-- name: VoteExpiry :one
SELECT vote_expiry
FROM voters
WHERE user_id = $1;
