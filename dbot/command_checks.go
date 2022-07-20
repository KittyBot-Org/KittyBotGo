package dbot

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var (
	IsDev CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) bool {
		for _, v := range b.Config.DevUserIDs {
			if v == e.User().ID {
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

	HasMusicPlayer CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) bool {
		if !b.MusicPlayers.Has(*e.GuildID()) {
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

	HasQueueItems CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) bool {
		if b.MusicPlayers.Get(*e.GuildID()).Queue.Len() == 0 {
			if err := e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("checks.has.queue.items")}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	HasHistoryItems CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) bool {
		if b.MusicPlayers.Get(*e.GuildID()).History.Len() == 0 {
			if err := e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("checks.has.history.items")}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

	IsMemberConnectedToVoiceChannel CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) bool {
		if voiceState, ok := e.Client().Caches().VoiceStates().Get(*e.GuildID(), e.User().ID); !ok || voiceState.ChannelID == nil {
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

	IsPlaying CommandCheck = func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) bool {
		if b.MusicPlayers.Get(*e.GuildID()).PlayingTrack() == nil {
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
