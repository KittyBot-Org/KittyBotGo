package dbot

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
)

func IsDev(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(ctx *events.ApplicationCommandInteractionCreate) bool {
		for _, v := range b.Config.DevUserIDs {
			if v == ctx.User().ID {
				return true
			}
		}
		if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("checks.is.dev").
			SetEphemeral(true).
			Build(),
		); err != nil {
			b.Logger.Error(err)
		}
		return false
	}
}

func HasMusicPlayer(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(e *events.ApplicationCommandInteractionCreate) bool {
		if !b.MusicPlayers.Has(*e.GuildID()) {
			if err := e.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("checks.has.music.player").
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
}

func HasQueueItems(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(e *events.ApplicationCommandInteractionCreate) bool {
		if b.MusicPlayers.Get(*e.GuildID()).Queue.Len() == 0 {
			if err := e.CreateMessage(discord.MessageCreate{Content: "checks.has.queue.items"}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
}

func HasHistoryItems(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(ctx *events.ApplicationCommandInteractionCreate) bool {
		if b.MusicPlayers.Get(*ctx.GuildID()).History.Len() == 0 {
			if err := ctx.CreateMessage(discord.MessageCreate{Content: "checks.has.history.items"}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
}

func IsMemberConnectedToVoiceChannel(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(ctx *events.ApplicationCommandInteractionCreate) bool {
		if voiceState, ok := ctx.Client().Caches().VoiceStates().Get(*ctx.GuildID(), ctx.User().ID); !ok || voiceState.ChannelID == nil {
			if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("modules.music.not.in.voice").
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
}

func IsPlaying(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(ctx *events.ApplicationCommandInteractionCreate) bool {
		if b.MusicPlayers.Get(*ctx.GuildID()).PlayingTrack() == nil {
			if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent("checks.is.playing").
				SetEphemeral(true).
				Build(),
			); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

}
