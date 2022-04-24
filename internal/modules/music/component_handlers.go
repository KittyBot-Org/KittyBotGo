package music

import (
	"regexp"

	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/responses"
	"github.com/go-jet/jet/v2/qrm"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgolink/lavalink"
	"golang.org/x/text/message"
)

var trackRegex = regexp.MustCompile(`\[\x60(?P<title>.+)\x60]\((?P<url>.+)?\)`)

func checkPlayer(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) (*kbot.MusicPlayer, error) {
	player := b.MusicPlayers.Get(*e.GuildID())
	if player == nil {
		return nil, e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.no.player"))
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
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.previous.empty"))
	}

	if err = player.Play(nextTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.previous.error"))
	}
	return e.UpdateMessage(responses.UpdateSuccessComponentsf(p, "modules.music.commands.previous.success", []any{nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}

func playPauseComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	if player.PlayingTrack() == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.play.pause.not.playing"))
	}
	paused := !player.Paused()
	if err = player.Pause(paused); err != nil {
		if paused {
			return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.play.pause.pause.error"))
		}
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.play.pause.play.error"))
	}
	track := player.PlayingTrack()
	info := track.Info()
	if paused {
		return e.UpdateMessage(responses.UpdateSuccessComponentsf(p, "modules.music.components.play.pause.pause.success", []any{info.Title, *info.URI, info.Length, player.Position()}, getMusicControllerComponents(track)))
	}
	return e.UpdateMessage(responses.UpdateSuccessComponentsf(p, "modules.music.components.play.pause.play.success", []any{info.Title, *info.URI, info.Length}, getMusicControllerComponents(track)))
}

func nextComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
	player, err := checkPlayer(b, p, e)
	if player == nil {
		return err
	}
	nextTrack := player.Queue.Pop()
	if nextTrack == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.next.empty"))
	}

	if err = player.Play(nextTrack); err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.next.error"))
	}
	return e.UpdateMessage(responses.UpdateSuccessComponentsf(p, "modules.music.commands.next.success", []any{nextTrack.Info().Title, *nextTrack.Info().URI, nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}

func likeComponentHandler(b *kbot.Bot, p *message.Printer, e *events.ComponentInteractionEvent) error {
	allMatches := trackRegex.FindAllStringSubmatch(e.Message.Content, -1)
	if allMatches == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.like.no.track"))
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

	if err == qrm.ErrNoRows {
		if err = b.DB.LikedSongs().Add(e.User().ID, getTrackQuery(title, url), title); err != nil {
			b.Logger.Error("Error adding music history entry: ", err)
			return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.like.add.error"))
		}
		res := responses.CreateSuccessf(p, "modules.music.components.like.add.success", title, url)
		res.Flags = discord.MessageFlagEphemeral
		return e.CreateMessage(res)

	}
	if err = b.DB.LikedSongs().Delete(e.User().ID, title); err != nil {
		b.Logger.Error("Error removing music history entry: ", err)
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.components.like.remove.error"))
	}
	res := responses.CreateSuccessf(p, "modules.music.components.like.remove.success", title, url)
	res.Flags = discord.MessageFlagEphemeral
	return e.CreateMessage(res)
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
