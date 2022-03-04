package music

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"golang.org/x/text/message"
)

func checkPlayer(b *types.Bot, p *message.Printer, e *events.ComponentInteractionEvent) (*types.MusicPlayer, error) {
	player := b.MusicPlayers.Get(*e.GuildID)
	if player == nil {
		return nil, e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.no.player"), Flags: discord.MessageFlagEphemeral})
	}
	return player, nil
}

func previousComponentHandler(b *types.Bot, p *message.Printer, e *events.ComponentInteractionEvent, action string) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	nextTrack := player.History.Last()
	if nextTrack == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.previous.empty"), Flags: discord.MessageFlagEphemeral})
	}

	if err = player.Play(nextTrack); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.previous.error"), Flags: discord.MessageFlagEphemeral})
	}
	msg := p.Sprintf("modules.music.commands.previous.success", nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length)
	return e.UpdateMessage(discord.MessageUpdate{Content: &msg})
}

func playPauseComponentHandler(b *types.Bot, p *message.Printer, e *events.ComponentInteractionEvent, action string) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	if player.PlayingTrack() == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.play.pause.not.playing"), Flags: discord.MessageFlagEphemeral})
	}
	paused := !player.Paused()
	if err := player.Pause(paused); err != nil {
		var msg string
		if paused {
			msg = p.Sprintf("modules.music.components.play.pause.pause.error")
		} else {
			msg = p.Sprintf("modules.music.components.play.pause.play.error")
		}
		return e.CreateMessage(discord.MessageCreate{Content: msg, Flags: discord.MessageFlagEphemeral})
	}
	var msg string
	info := player.PlayingTrack().Info()
	if paused {
		msg = p.Sprintf("modules.music.components.play.pause.pause.success", info.Title, *info.URI, info.Length, player.Position())
	} else {
		msg = p.Sprintf("modules.music.components.play.pause.play.success", info.Title, *info.URI, info.Length)
	}
	return e.UpdateMessage(discord.MessageUpdate{Content: &msg})
}

func nextComponentHandler(b *types.Bot, p *message.Printer, e *events.ComponentInteractionEvent, action string) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	nextTrack := player.Queue.Pop()
	if nextTrack == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.next.empty"), Flags: discord.MessageFlagEphemeral})
	}

	if err = player.Play(nextTrack); err != nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.next.error"), Flags: discord.MessageFlagEphemeral})
	}
	msg := p.Sprintf("modules.music.commands.next.success", nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length)
	return e.UpdateMessage(discord.MessageUpdate{Content: &msg})
}

func likeComponentHandler(b *types.Bot, p *message.Printer, e *events.ComponentInteractionEvent, action string) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	track := player.PlayingTrack()
	if track == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.like.not.playing"), Flags: discord.MessageFlagEphemeral})
	}
	// TODO: Add like/unlike
	return e.CreateMessage(discord.MessageCreate{Content: "not implemented yet", Flags: discord.MessageFlagEphemeral})
}

func getMusicControllerComponents() discord.ContainerComponent {
	return discord.ActionRowComponent{
		discord.NewPrimaryButton("", "cmd:now-playing:previous").WithEmoji(discord.ComponentEmoji{Name: "⏮"}),
		discord.NewPrimaryButton("", "cmd:now-playing:play-pause").WithEmoji(discord.ComponentEmoji{Name: "⏯"}),
		discord.NewPrimaryButton("", "cmd:now-playing:next").WithEmoji(discord.ComponentEmoji{Name: "⏭"}),
		discord.NewPrimaryButton("", "cmd:now-playing:like").WithEmoji(discord.ComponentEmoji{Name: "❤️"}),
	}
}
