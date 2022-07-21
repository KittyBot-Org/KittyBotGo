package commands

import (
	"fmt"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

var NowPlaying = dbot.Command{
	Create: discord.SlashCommandCreate{
		CommandName: "now-playing",
		Description: "Tells you about the currently playing song.",
	},
	Checks: dbot.HasMusicPlayer.And(dbot.IsPlaying),
	CommandHandler: map[string]dbot.CommandHandler{
		"": nowPlayingHandler,
	},
	ComponentHandler: map[string]dbot.ComponentHandler{
		"previous":   previousComponentHandler,
		"play-pause": playPauseComponentHandler,
		"next":       nextComponentHandler,
		"like":       likeComponentHandler,
	},
}

func nowPlayingHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	player := b.MusicPlayers.Get(*e.GuildID())
	track := player.PlayingTrack()

	if track == nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.music.commands.nowplaying.no.track"))
	}
	i := track.Info()
	embed := discord.NewEmbedBuilder().
		SetAuthorName(p.Sprintf("modules.music.commands.nowplaying.title")).
		SetTitle(i.Title).
		SetURL(*i.URI).
		AddField(p.Sprintf("modules.music.commands.nowplaying.author"), i.Author, true).
		AddField(p.Sprintf("modules.music.commands.nowplaying.requested.by"), discord.UserMention(track.UserData().(dbot.AudioTrackData).Requester), true).
		AddField(p.Sprintf("modules.music.commands.nowplaying.volume"), fmt.Sprintf("%d%%", player.Volume()), true).
		SetThumbnail(getArtworkURL(player.PlayingTrack())).
		SetFooterText(p.Sprintf("modules.music.commands.nowplaying.footer", player.Queue.Len()))
	if !i.IsStream {
		bar := [10]string{"▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬"}
		t1 := player.Position()
		t2 := i.Length
		p := int(float64(t1) / float64(t2) * 10)
		bar[p] = "🔘"
		loopString := ""
		if player.Queue.LoopingType() == dbot.LoopingTypeRepeatSong {
			loopString = "🔂"
		}
		if player.Queue.LoopingType() == dbot.LoopingTypeRepeatQueue {
			loopString = "🔁"
		}
		embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", formatPosition(t1), formatPosition(t2), loopString, bar)
	}
	return e.CreateMessage(responses.CreateSuccessEmbedComponents(embed.Build(), getMusicControllerComponents(track)))
}

func previousComponentHandler(b *dbot.Bot, _ []string, p *message.Printer, e *events.ComponentInteractionCreate) error {
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
	return e.UpdateMessage(responses.UpdateSuccessComponentsf(p, "modules.music.commands.previous.success", []any{formatTrack(nextTrack), nextTrack.Info().Length}, getMusicControllerComponents(nextTrack)))
}
