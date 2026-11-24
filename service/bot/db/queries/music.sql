-- name: ListLavalinkNodes :many
SELECT name, COALESCE(session_id, '') AS session_id
FROM lavalink_nodes;

-- name: UpsertLavalinkNode :exec
INSERT INTO lavalink_nodes (name, session_id)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE
    SET session_id = EXCLUDED.session_id;

-- name: GetPlaylists :many
SELECT id, name, user_id
FROM playlists
WHERE user_id = $1;

-- name: SearchPlaylists :many
SELECT id, name, user_id
FROM playlists
WHERE user_id = $1
ORDER BY strict_word_similarity_dist(name, $2)
LIMIT $3;

-- name: GetPlaylist :one
SELECT id, name, user_id
FROM playlists
WHERE id = $1;

-- name: CreatePlaylist :one
INSERT INTO playlists (name, user_id)
VALUES ($1, $2)
RETURNING id, name, user_id;

-- name: DeletePlaylist :exec
DELETE
FROM playlists
WHERE id = $1
  AND user_id = $2;

-- name: ListPlaylistTracks :many
SELECT id, playlist_id, position, track
FROM playlist_tracks
WHERE playlist_id = $1;

-- name: InsertPlaylistTrack :exec
INSERT INTO playlist_tracks (playlist_id, track)
VALUES ($1, $2);

-- name: SearchPlaylistTracks :many
SELECT id, playlist_id, position, track
FROM playlist_tracks
WHERE playlist_id = $1
ORDER BY strict_word_similarity_dist(track -> 'info' ->> 'title', $2)
LIMIT $3;

-- name: RemovePlaylistTrack :exec
DELETE
FROM playlist_tracks
WHERE id = $1;

-- name: AddLikedTrack :exec
INSERT INTO liked_tracks (user_id, track)
VALUES ($1, $2);

-- name: RemoveLikedTrack :exec
DELETE
FROM liked_tracks
WHERE id = $1;

-- name: GetLikedTracks :many
SELECT id, user_id, track
FROM liked_tracks
WHERE user_id = $1;

-- name: SearchLikedTracks :many
SELECT id, user_id, track
FROM liked_tracks
WHERE user_id = $1
ORDER BY strict_word_similarity_dist(track -> 'info' ->> 'title', $2)
LIMIT $3;

-- name: ClearLikedTracks :exec
DELETE
FROM liked_tracks
WHERE user_id = $1;

-- name: GetLikedTrack :one
SELECT id, user_id, track
FROM liked_tracks
WHERE id = $1;

-- name: FindLikedTrack :one
SELECT id, user_id, track
FROM liked_tracks
WHERE user_id = $1
  AND track -> 'info' ->> 'uri' = sqlc.arg(uri)::text;

-- name: UpsertPlayHistory :exec
INSERT INTO play_histories (user_id, played_at, track)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, track) DO UPDATE
    SET played_at = EXCLUDED.played_at;

-- name: TrimPlayHistory :exec
DELETE
FROM play_histories AS histories
WHERE histories.user_id = $1
  AND histories.id NOT IN (SELECT recent.id
                           FROM play_histories AS recent
                           WHERE recent.user_id = histories.user_id
                           ORDER BY recent.played_at DESC
                           LIMIT 10);

-- name: GetPlayHistoryTrack :one
SELECT id, user_id, played_at, track
FROM play_histories
WHERE id = $1;

-- name: SearchPlayHistory :many
SELECT id, user_id, played_at, track
FROM play_histories
WHERE user_id = $1
ORDER BY strict_word_similarity_dist(track -> 'info' ->> 'title', $2)
LIMIT $3;

-- name: SearchPlay :many
SELECT results.id, results.type, results.name
FROM (SELECT liked.id, 'liked_track' AS type, liked.track -> 'info' ->> 'title' AS name
      FROM liked_tracks AS liked
      WHERE liked.user_id = sqlc.arg(user_id)
      UNION ALL
      SELECT lists.id, 'playlist', lists.name
      FROM playlists AS lists
      WHERE lists.user_id = sqlc.arg(user_id)
      UNION ALL
      SELECT histories.id, 'play_history', histories.track -> 'info' ->> 'title'
      FROM play_histories AS histories
      WHERE histories.user_id = sqlc.arg(user_id)) AS results
ORDER BY strict_word_similarity_dist(results.name, sqlc.arg(query))
LIMIT sqlc.arg(result_limit);
