package types

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

var (
	IsDevUser CommandCheck = func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool {
		for _, v := range b.Config.DevUserIDs {
			if v == e.User.ID {
				return true
			}
		}
		return false
	}

	HasMusicPlayer CommandCheck = func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool {
		if !b.MusicPlayers.Has(*e.GuildID) {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("No player found in this server.").
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	IsMemberConnectedToVoiceChannel CommandCheck = func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool {
		if voiceState := e.Member.VoiceState(); voiceState == nil || voiceState.ChannelID == nil {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("You must be in a voice channel to use this command.").
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	HasPlayingTrack CommandCheck = func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool {
		if b.MusicPlayers.Get(*e.GuildID).PlayingTrack() == nil {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("No track is currently playing.").
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
)
