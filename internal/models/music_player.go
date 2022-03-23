package models

import (
	"github.com/disgoorg/snowflake"
)

type MusicPlayer struct {
	GuildID              snowflake.Snowflake   `bun:"guild_id,pk,notnull"`
	State                []byte                `bun:"state,notnull"`
	PlayingTrackUserData *AudioTrackData       `bun:"playing_track_user_data"`
	Type                 int                   `bun:"type,notnull"`
	Queue                []AudioTrack          `bun:"queue,notnull"`
	LoopingType          int                   `bun:"looping_type,notnull"`
	History              []AudioTrack          `bun:"history,notnull"`
	SkipVotes            []snowflake.Snowflake `bun:"skip_votes,notnull"`
}

type AudioTrack struct {
	Track    string         `json:"track"`
	UserData AudioTrackData `json:"user_data"`
}

type AudioTrackData struct {
	Requester snowflake.Snowflake `json:"requester"`
}
