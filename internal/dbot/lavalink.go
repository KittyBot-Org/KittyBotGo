package dbot

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/disgolink/disgolink"
	"github.com/disgoorg/disgolink/lavalink"
	"github.com/disgoorg/snowflake"
)

func (b *Bot) SetupLavalink() {
	b.MusicPlayers = NewMusicPlayerMap(b)
	b.Lavalink = disgolink.New(b.Client /*lavalink.WithPlugins(source_extensions.NewSpotifyPlugin(), source_extensions.NewAppleMusicPlugin())*/)
	b.RegisterNodes()
	b.Client.EventManager().AddEventListeners(b.Lavalink)
	/*b.Bot.EventManager.AddEventListeners(events.ListenerAdapter{
		OnGuildReady: func(event *events.GuildReadyEvent) {
			b.LoadPlayer(event.GuildID)
		},
	})*/
}

func (b *Bot) RegisterNodes() {
	var wg sync.WaitGroup
	for i := range b.Config.Lavalink.Nodes {
		wg.Add(1)
		config := b.Config.Lavalink.Nodes[i]
		go func() {
			defer wg.Done()
			node, err := b.Lavalink.AddNode(context.TODO(), config)
			if err != nil {
				b.Logger.Error("Failed to add node: ", err)
				return
			}
			if config.ResumingKey != "" {
				if err = node.ConfigureResuming(config.ResumingKey, b.Config.Lavalink.ResumingTimeout); err != nil {
					b.Logger.Error("Failed to configure resuming: ", err)
				}
			}
		}()
	}
	wg.Wait()
}

func (b *Bot) LoadPlayer(guildID snowflake.Snowflake) {
	voiceState, ok := b.Client.Caches().VoiceStates().Get(guildID, b.Client.SelfUser().ID)
	if !ok {
		return
	}
	var player db.MusicPlayer
	if _, err := b.DB.NewDelete().Model(&player).Where("guild_id = ?", voiceState.GuildID).Returning("*").Exec(context.TODO()); err != nil {
		b.Logger.Error("Failed to delete & return player: ", err)
		return
	}

	var restoreState lavalink.PlayerRestoreState
	if err := json.Unmarshal(player.State, &restoreState); err != nil {
		b.Logger.Error("Failed to unmarshal player state: ", err)
		return
	}

	queue := NewMusicQueue(LoopingType(player.LoopingType))
	for _, track := range player.Queue {
		decodedTrack, err := b.Lavalink.DecodeTrack(track.Track)
		if err != nil {
			b.Logger.Error("Failed to decode track from queue: ", err)
			continue
		}
		decodedTrack.SetUserData(track.UserData)
		queue.Push(decodedTrack)
	}

	history := NewHistory()
	for _, track := range player.History {
		decodedTrack, err := b.Lavalink.DecodeTrack(track.Track)
		if err != nil {
			b.Logger.Error("Failed to decode track from history: ", err)
			continue
		}
		decodedTrack.SetUserData(track.UserData)
		history.Push(decodedTrack)
	}

	skipVotes := make(map[snowflake.Snowflake]struct{}, len(player.SkipVotes))
	for _, id := range player.SkipVotes {
		skipVotes[id] = struct{}{}
	}

	lavalinkPlayer, err := b.Lavalink.RestorePlayer(restoreState)
	if err != nil {
		b.Logger.Error("Failed to restore player: ", err)
		return
	}

	if track := lavalinkPlayer.PlayingTrack(); track != nil && player.PlayingTrackUserData != nil {
		track.SetUserData(player.PlayingTrackUserData)
	}

	b.MusicPlayers.Add(&MusicPlayer{
		Player:    lavalinkPlayer,
		Bot:       b,
		Type:      PlayerType(player.Type),
		Queue:     queue,
		History:   history,
		SkipVotes: skipVotes,
	})

}

func (b *Bot) SavePlayers() {
	for _, player := range b.MusicPlayers.All() {
		resumeData, err := json.Marshal(player.Export())
		if err != nil {
			b.Logger.Error("Failed to marshal player: ", err)
			continue
		}
		var trackData *db.AudioTrackData
		if player.PlayingTrack() != nil {
			data := player.PlayingTrack().UserData().(db.AudioTrackData)
			trackData = &data
		}
		queue := make([]db.AudioTrack, player.Queue.Len())
		for i, track := range player.Queue.Tracks() {
			encodedTrack, err := b.Lavalink.EncodeTrack(track)
			if err != nil {
				b.Logger.Error("Failed to encode queue track: ", err)
				continue
			}
			queue[i] = db.AudioTrack{Track: encodedTrack, UserData: track.UserData().(db.AudioTrackData)}
		}

		history := make([]db.AudioTrack, player.History.Len())
		for i, track := range player.History.Tracks() {
			encodedTrack, err := b.Lavalink.EncodeTrack(track)
			if err != nil {
				b.Logger.Error("Failed to encode history track: ", err)
				continue
			}
			history[i] = db.AudioTrack{Track: encodedTrack, UserData: track.UserData().(db.AudioTrackData)}
		}

		skipVotes := make([]snowflake.Snowflake, len(player.SkipVotes))
		i := 0
		for id := range player.SkipVotes {
			skipVotes[i] = id
			i++
		}

		if _, err = b.DB.NewInsert().Model(&db.MusicPlayer{
			GuildID:              player.GuildID(),
			State:                resumeData,
			PlayingTrackUserData: trackData,
			Type:                 int(player.Type),
			Queue:                queue,
			LoopingType:          int(player.Queue.LoopingType()),
			History:              history,
			SkipVotes:            skipVotes,
		}).On("CONFLICT (guild_id) DO UPDATE").Exec(context.TODO()); err != nil {
			b.Logger.Error("Failed to save player: ", err)
		}
	}
}
