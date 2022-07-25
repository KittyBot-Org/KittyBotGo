package dbot

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/handler"
)

func IsDev(b *Bot) handler.Check[*handler.CommandContext] {
	return func(ctx *handler.CommandContext) bool {
		for _, v := range b.Config.DevUserIDs {
			if v == ctx.User().ID {
				return true
			}
		}
		if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(ctx.Printer.Sprintf("checks.is.dev")).
			SetEphemeral(true).
			Build(),
		); err != nil {
			b.Logger.Error(err)
		}
		return false
	}
}

func HasMusicPlayer(b *Bot) handler.Check[*handler.CommandContext] {
	return func(ctx *handler.CommandContext) bool {
		if !b.MusicPlayers.Has(*ctx.GuildID()) {
			if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(ctx.Printer.Sprintf("checks.has.music.player")).
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

func HasQueueItems(b *Bot) handler.Check[*handler.CommandContext] {
	return func(ctx *handler.CommandContext) bool {
		if b.MusicPlayers.Get(*ctx.GuildID()).Queue.Len() == 0 {
			if err := ctx.CreateMessage(discord.MessageCreate{Content: ctx.Printer.Sprintf("checks.has.queue.items")}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
}

func HasHistoryItems(b *Bot) handler.Check[*handler.CommandContext] {
	return func(ctx *handler.CommandContext) bool {
		if b.MusicPlayers.Get(*ctx.GuildID()).History.Len() == 0 {
			if err := ctx.CreateMessage(discord.MessageCreate{Content: ctx.Printer.Sprintf("checks.has.history.items")}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}
}

func IsMemberConnectedToVoiceChannel(b *Bot) handler.Check[*handler.CommandContext] {
	return func(ctx *handler.CommandContext) bool {
		if voiceState, ok := ctx.Client().Caches().VoiceStates().Get(*ctx.GuildID(), ctx.User().ID); !ok || voiceState.ChannelID == nil {
			if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(ctx.Printer.Sprintf("modules.music.not.in.voice")).
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

func IsPlaying(b *Bot) handler.Check[*handler.CommandContext] {
	return func(ctx *handler.CommandContext) bool {
		if b.MusicPlayers.Get(*ctx.GuildID()).PlayingTrack() == nil {
			if err := ctx.CreateMessage(discord.NewMessageCreateBuilder().
				SetContent(ctx.Printer.Sprintf("checks.is.playing")).
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
