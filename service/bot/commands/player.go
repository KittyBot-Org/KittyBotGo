package commands

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v3/lavalink"
	"github.com/disgoorg/json"
	"github.com/disgoorg/lavaqueue-plugin"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var bassBoostLevels = map[string]*lavalink.Equalizer{
	"Off": nil,
	"Low": {
		0:  0.2,
		1:  0.15,
		2:  0.1,
		3:  0.05,
		4:  0.0,
		5:  -0.05,
		6:  -0.1,
		7:  -0.1,
		8:  -0.1,
		9:  -0.1,
		10: -0.1,
		11: -0.1,
		12: -0.1,
		13: -0.1,
		14: -0.1,
	},
	"High": {
		0:  0.4,
		1:  0.3,
		2:  0.2,
		3:  0.1,
		4:  0.0,
		5:  -0.1,
		6:  -0.2,
		7:  -0.2,
		8:  -0.2,
		9:  -0.2,
		10: -0.2,
		11: -0.2,
		12: -0.2,
		13: -0.2,
		14: -0.2,
	},
}

var playerCommand = discord.SlashCommandCreate{
	Name:        "player",
	Description: "Shows the player status.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "play",
			Description: "Play a song",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "query",
					Description:  "The song or search to play",
					Required:     true,
					Autocomplete: true,
				},
				discord.ApplicationCommandOptionString{
					Name:        "source",
					Description: "The source to search on",
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "YouTube",
							Value: string(lavalink.SearchTypeYouTube),
						},
						{
							Name:  "YouTube Music",
							Value: string(lavalink.SearchTypeYouTubeMusic),
						},
						{
							Name:  "SoundCloud",
							Value: string(lavalink.SearchTypeSoundCloud),
						},
						{
							Name:  "Deezer",
							Value: "dzsearch",
						},
						{
							Name:  "Deezer ISRC",
							Value: "dzisrc",
						},
					},
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "status",
			Description: "Shows the player status.",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "pause",
			Description: "Pauses the player.",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "resume",
			Description: "Resumes the player.",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "next",
			Description: "Skips to the next song in the queue",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "previous",
			Description: "Skips to the previous song in the history",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "stop",
			Description: "Stops the player.",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "volume",
			Description: "Sets the player volume.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "volume",
					Description: "The volume to set.",
					Required:    true,
					MinValue:    json.Ptr(0),
					MaxValue:    json.Ptr(200),
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "bass-boost",
			Description: "Enables or disables bass boost of the music player.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "level",
					Description: "The bass boost level to use.",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{
							Name:  "Off",
							Value: "Off",
						},
						{
							Name:  "Low",
							Value: "Low",
						},
						{
							Name:  "High",
							Value: "High",
						},
					},
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "seek",
			Description: "Seeks to a position in the current song.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:        "position",
					Description: "The position to seek to.",
				},
				discord.ApplicationCommandOptionInt{
					Name:        "time-unit",
					Description: "The time unit to use.",
					Choices: []discord.ApplicationCommandOptionChoiceInt{
						{
							Name:  "Hours",
							Value: int(lavalink.Hour),
						},
						{
							Name:  "Minutes",
							Value: int(lavalink.Minute),
						},
						{
							Name:  "Seconds",
							Value: int(lavalink.Second),
						},
						{
							Name:  "Milliseconds",
							Value: int(lavalink.Millisecond),
						},
					},
				},
			},
		},
	},
}

func (c *commands) OnPlayerStatus(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	queue, err := lavaqueue.GetQueue(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get queue", err))
	}

	track := player.Track()
	if track == nil {
		return e.CreateMessage(res.CreateError("There is no song playing right now."))
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Playing:").
		SetColor(res.PrimaryColor).
		SetDescription(res.FormatTrack(*track, player.Position())).
		AddField("Author:", track.Info.Author, true).
		AddField("Volume:", fmt.Sprintf("%d%%", player.Volume()), true).
		SetFooterText(fmt.Sprintf("Songs in queue: %d", len(queue.Tracks)))

	if track.Info.ArtworkURL != nil {
		embed.SetThumbnail(*track.Info.ArtworkURL)
	}

	if !track.Info.IsStream {
		bar := [10]string{"▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬", "▬"}
		t1 := player.Position()
		t2 := track.Info.Length
		p := int(float64(t1) / float64(t2) * 10)
		bar[p] = "🔘"
		loopString := ""
		if queue.Type == lavaqueue.QueueTypeRepeatTrack {
			loopString = "🔂"
		} else if queue.Type == lavaqueue.QueueTypeRepeatQueue {
			loopString = "🔁"
		}
		embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", res.FormatDuration(t1), res.FormatDuration(t2), loopString, bar)
	}

	create := res.CreatePlayer("", true)
	create.Embeds = []discord.Embed{embed.Build()}
	return e.CreateMessage(create)
}

func (c *commands) OnPlayerPause(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := player.Update(e.Ctx, lavalink.WithPaused(true)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to pause the player", err))
	}
	return e.CreateMessage(res.CreatePlayer("⏸ Paused the player.", false))
}

func (c *commands) OnPlayerResume(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := player.Update(e.Ctx, lavalink.WithPaused(false)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to resume the player", err))
	}
	return e.CreateMessage(res.CreatePlayer("▶ Resumed the player.", false))
}

func (c *commands) OnPlayerStop(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := player.Destroy(e.Ctx); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to stop the player", err))
	}

	if err := c.Discord.UpdateVoiceState(e.Ctx, *e.GuildID(), nil, false, false); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to disconnect from the voice channel", err))
	}

	return e.CreateMessage(res.Create("⏹ Stopped the player."))
}

func (c *commands) OnPlayerNext(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	track, err := lavaqueue.QueueNextTrack(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		var eErr *lavalink.Error
		if errors.As(err, &eErr) && eErr.Status == http.StatusNotFound {
			return e.CreateMessage(res.CreateError("No more songs in queue"))
		}
		return e.CreateMessage(res.CreateErr("Failed to skip to the next song", err))
	}

	return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(*track, 0)))
}

func (c *commands) OnPlayerPrevious(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	track, err := lavaqueue.QueuePreviousTrack(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		var eErr *lavalink.Error
		if errors.As(err, &eErr) && eErr.Status == http.StatusNotFound {
			return e.CreateMessage(res.CreateError("No songs in history"))
		}
		return e.CreateMessage(res.CreateErr("Failed to skip to the next song", err))
	}

	return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(*track, 0)))
}

func (c *commands) OnPlayerVolume(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	volume := data.Int("volume")

	if err := player.Update(e.Ctx, lavalink.WithVolume(volume)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to set the volume", err))
	}
	return e.CreateMessage(res.CreatePlayerf("🔊 Set the volume to %d%%.", false, volume))
}

func (c *commands) OnPlayerBassBoost(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	level := data.String("level")

	filters := player.Filters()
	filters.Equalizer = bassBoostLevels[level]

	if err := player.Update(e.Ctx, lavalink.WithFilters(filters)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to set bass boost: %s", err))
	}
	return e.CreateMessage(res.CreatePlayerf("🔊 Set bass boost to %s.", false, level))
}

func (c *commands) OnPlayerSeek(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	position := data.Int("position")
	duration, ok := data.OptInt("time-unit")
	if !ok {
		duration = int(time.Second)
	}

	newPos := lavalink.Duration(position * duration)
	if err := player.Update(e.Ctx, lavalink.WithPosition(newPos)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to seek to %d", err))
	}
	return e.CreateMessage(res.CreatePlayerf("⏩ Seeked to %s.", false, res.FormatDuration(newPos)))
}

func (c *commands) OnPlayerNextButton(e *handler.ComponentEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	track, err := lavaqueue.QueueNextTrack(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		var eErr *lavalink.Error
		if errors.As(err, &eErr) && eErr.Status == http.StatusNotFound {
			return e.CreateMessage(res.CreateError("No more songs in queue"))
		}
		return e.CreateMessage(res.CreateErr("Failed to skip to the next song", err))
	}

	return e.UpdateMessage(res.UpdatePlayerf("▶ Playing: %s", true, res.FormatTrack(*track, 0)))
}

func (c *commands) OnPlayerPreviousButton(e *handler.ComponentEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	track, err := lavaqueue.QueuePreviousTrack(e.Ctx, player.Node(), *e.GuildID())
	if err != nil {
		var eErr *lavalink.Error
		if errors.As(err, &eErr) && eErr.Status == http.StatusNotFound {
			return e.CreateMessage(res.CreateError("No songs in history"))
		}
		return e.CreateMessage(res.CreateErr("Failed to skip to the next song", err))
	}

	return e.UpdateMessage(res.UpdatePlayerf("▶ Playing: %s", true, res.FormatTrack(*track, 0)))
}

func (c *commands) OnPlayerPlayPauseButton(e *handler.ComponentEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	paused := !player.Paused()
	if err := player.Update(e.Ctx, lavalink.WithPaused(paused)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to pause the player", err))
	}
	if paused {
		return e.UpdateMessage(res.UpdatePlayerf("⏸ Paused the player.", false))
	}
	return e.UpdateMessage(res.UpdatePlayerf("▶ Resumed the player.", false))
}

func (c *commands) OnPlayerStopButton(e *handler.ComponentEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := player.Destroy(e.Ctx); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to stop the player", err))
	}

	if err := c.Discord.UpdateVoiceState(e.Ctx, *e.GuildID(), nil, false, false); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to disconnect from the voice channel", err))
	}

	return e.UpdateMessage(res.Update("⏹ Stopped the player."))
}
