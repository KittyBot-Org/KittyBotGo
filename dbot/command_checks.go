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
		if err := ctx.CreateMessage(discord.MessageCreate{
			Content: "This action is only available for developer.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			b.Logger.Error(err)
		}
		return false
	}
}

func HasMusicPlayer(b *Bot) handler.Check[*events.ApplicationCommandInteractionCreate] {
	return func(e *events.ApplicationCommandInteractionCreate) bool {
		if !b.MusicPlayers.Has(*e.GuildID()) {
			if err := e.CreateMessage(discord.MessageCreate{
				Content: "No music player found in this server.",
				Flags:   discord.MessageFlagEphemeral,
			}); err != nil {
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
			if err := e.CreateMessage(discord.MessageCreate{
				Content: "The song queue is empty.",
				Flags:   discord.MessageFlagEphemeral,
			}); err != nil {
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
			if err := ctx.CreateMessage(discord.MessageCreate{
				Content: "The song history is empty.",
				Flags:   discord.MessageFlagEphemeral,
			}); err != nil {
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
			if err := ctx.CreateMessage(discord.MessageCreate{
				Content: "You need to be connected to a voice channel to play music.",
				Flags:   discord.MessageFlagEphemeral,
			}); err != nil {
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
			if err := ctx.CreateMessage(
				discord.MessageCreate{
					Content: "No song is currently playing.",
					Flags:   discord.MessageFlagEphemeral,
				}); err != nil {
				b.Logger.Error(err)
			}
			return false
		}
		return true
	}

}
