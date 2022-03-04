package types

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"golang.org/x/text/message"
)

var (
	IsDev CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		for _, v := range b.Config.DevUserIDs {
			if v == e.User.ID {
				return true
			}
		}
		if err := e.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(p.Sprintf("checks.is.dev")).
			SetEphemeral(true).
			Build(),
		); err != nil {
			b.Logger.Error(err)
		}
		return false
	}

	HasMusicPlayer CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		if !b.MusicPlayers.Has(*e.GuildID) {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(p.Sprintf("checks.has.music.player")).
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	HasQueueItems CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		player := b.MusicPlayers.Get(*e.GuildID)
		if player.Queue.Len() == 0 {
			if err := e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("checks.has.queue.items")}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	HasHistoryItems CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		player := b.MusicPlayers.Get(*e.GuildID)
		if player.History.Len() == 0 {
			if err := e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("checks.has.history.items")}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	IsMemberConnectedToVoiceChannel CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		if voiceState := e.Member.VoiceState(); voiceState == nil || voiceState.ChannelID == nil {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(p.Sprintf("modules.music.not.in.voice")).
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	IsPlaying CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		if b.MusicPlayers.Get(*e.GuildID).PlayingTrack() == nil {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(p.Sprintf("checks.is.playing")).
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
