package music

import (
	"regexp"

	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/go-jet/jet/v2/qrm"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	"golang.org/x/text/message"
)

var trackRegex = regexp.MustCompile(`\[\x60(?P<title>.+)\x60]\(<(?P<url>.+)?>\)`)

func checkPlayer(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) (*kbot.MusicPlayer, error) {
	player := b.MusicPlayers.Get(*e.GuildID())
	if player == nil {
		return nil, e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.no.player"), Flags: discord.MessageFlagEphemeral})
	}
	return player, nil
}

func previousComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
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
	components := []discord.ContainerComponent{getMusicControllerComponents(nextTrack)}
	return e.UpdateMessage(discord.MessageUpdate{Content: &msg, Components: &components})
}

func playPauseComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
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
	track := player.PlayingTrack()
	info := track.Info()
	if paused {
		msg = p.Sprintf("modules.music.components.play.pause.pause.success", info.Title, *info.URI, info.Length, player.Position())
	} else {
		msg = p.Sprintf("modules.music.components.play.pause.play.success", info.Title, *info.URI, info.Length)
	}
	components := []discord.ContainerComponent{getMusicControllerComponents(track)}
	return e.UpdateMessage(discord.MessageUpdate{Content: &msg, Components: &components})
}

func nextComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
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
	components := []discord.ContainerComponent{getMusicControllerComponents(nextTrack)}
	return e.UpdateMessage(discord.MessageUpdate{Content: &msg, Components: &components})
}

func likeComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
	allMatches := trackRegex.FindAllStringSubmatch(e.Message.Content, -1)
	if allMatches == nil {
		return e.CreateMessage(discord.MessageCreate{Content: p.Sprintf("modules.music.components.like.no.track"), Flags: discord.MessageFlagEphemeral})
	}
	matches := allMatches[0]
	var (
		title string
		url   *string
	)
	title = matches[trackRegex.SubexpIndex("title")]
	if len(matches) > 2 {
		url = &matches[trackRegex.SubexpIndex("url")]
	}

	_, err := b.DB.LikedSongs().Get(e.User().ID, title)
	if err != nil && err != qrm.ErrNoRows {
		return err
	}

	var msg string
	if err == qrm.ErrNoRows {
		if err = b.DB.LikedSongs().Add(e.User().ID, getTrackQuery(title, url), title); err != nil {
			b.Logger.Error("Error adding music history entry: ", err)
		}
		msg = p.Sprintf("modules.music.components.like.added", title, url)
	} else {
		if err = b.DB.LikedSongs().Delete(e.User().ID, title); err != nil {
			b.Logger.Error("Error adding music history entry: ", err)
		}
		msg = p.Sprintf("modules.music.components.like.removed", title, url)
	}
	return e.CreateMessage(discord.MessageCreate{Content: msg, Flags: discord.MessageFlagEphemeral})
}

func getMusicControllerComponents(track lavalink.AudioTrack) discord.ContainerComponent {
	buttons := discord.ActionRowComponent{
		discord.NewPrimaryButton("", "cmd:now-playing:previous").WithEmoji(discord.ComponentEmoji{Name: "⏮"}),
		discord.NewPrimaryButton("", "cmd:now-playing:play-pause").WithEmoji(discord.ComponentEmoji{Name: "⏯"}),
		discord.NewPrimaryButton("", "cmd:now-playing:next").WithEmoji(discord.ComponentEmoji{Name: "⏭"}),
	}
	if track != nil {
		buttons = buttons.AddComponents(discord.NewPrimaryButton("", "cmd:now-playing:like").WithEmoji(discord.ComponentEmoji{Name: "❤"}))
	}
	return buttons
}

func getTrackQuery(title string, url *string) string {
	if url != nil {
		return *url
	}
	return title
}
