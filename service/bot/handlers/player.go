package handlers

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/lavalink"
	"github.com/disgoorg/json"

	"github.com/KittyBot-Org/KittyBotGo/interal/database"
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

func (h *Handlers) OnPlayerStatus(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	dbPlayer, _ := h.Database.GetPlayer(*e.GuildID(), player.Node().Config().Name)
	tracks, _ := h.Database.GetQueue(*e.GuildID())
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
		SetFooterText(fmt.Sprintf("Songs in queue: %d", len(tracks)))

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
		if dbPlayer.QueueType == database.QueueTypeRepeatTrack {
			loopString = "🔂"
		} else if dbPlayer.QueueType == database.QueueTypeRepeatQueue {
			loopString = "🔁"
		}
		embed.Description += fmt.Sprintf("\n\n%s / %s %s\n%s", res.FormatDuration(t1), res.FormatDuration(t2), loopString, bar)
	}

	create := res.CreatePlayer("", true)
	create.Embeds = []discord.Embed{embed.Build()}
	return e.CreateMessage(create)
}

func (h *Handlers) OnPlayerPause(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	if err := player.Update(context.Background(), lavalink.WithPaused(true)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to pause the player", err))
	}
	return e.CreateMessage(res.CreatePlayer("⏸ Paused the player.", false))
}

func (h *Handlers) OnPlayerResume(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	if err := player.Update(context.Background(), lavalink.WithPaused(false)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to resume the player", err))
	}
	return e.CreateMessage(res.CreatePlayer("▶ Resumed the player.", false))
}

func (h *Handlers) OnPlayerStop(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	if err := player.Destroy(context.Background()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to stop the player", err))
	}

	if err := h.Database.DeletePlayer(*e.GuildID()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to delete the player from the database", err))
	}

	if err := h.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), nil, false, false); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to disconnect from the voice channel", err))
	}

	return e.CreateMessage(res.Create("⏹ Stopped the player."))
}

func (h *Handlers) OnPlayerNext(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	track, err := h.Database.NextQueueTrack(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get next song", err))
	}

	if err = player.Update(context.Background(), lavalink.WithTrack(track.Track)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to play next song", err))
	}
	return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(track.Track, 0)))
}

func (h *Handlers) OnPlayerPrevious(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	track, err := h.Database.PreviousHistoryTrack(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get previous song", err))
	}

	if err = player.Update(context.Background(), lavalink.WithTrack(track.Track)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to play previous song", err))
	}
	return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(track.Track, 0)))
}

func (h *Handlers) OnPlayerVolume(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	volume := e.SlashCommandInteractionData().Int("volume")

	if err := player.Update(context.Background(), lavalink.WithVolume(volume)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to set the volume", err))
	}
	return e.CreateMessage(res.CreatePlayerf("🔊 Set the volume to %d%%.", false, volume))
}

func (h *Handlers) OnPlayerBassBoost(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	level := e.SlashCommandInteractionData().String("level")

	filters := player.Filters()
	filters.Equalizer = bassBoostLevels[level]

	if err := player.Update(context.Background(), lavalink.WithFilters(filters)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to set bass boost: %s", err))
	}
	return e.CreateMessage(res.CreatePlayerf("🔊 Set bass boost to %s.", false, level))
}

func (h *Handlers) OnPlayerSeek(e *handler.CommandEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	data := e.SlashCommandInteractionData()
	position := data.Int("position")
	duration, ok := data.OptInt("time-unit")
	if !ok {
		duration = int(time.Second)
	}

	newPos := lavalink.Duration(position * duration)
	if err := player.Update(context.Background(), lavalink.WithPosition(newPos)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to seek to %d", err))
	}
	return e.CreateMessage(res.CreatePlayerf("⏩ Seeked to %d.", false, res.FormatDuration(newPos)))
}

func (h *Handlers) OnPlayerPreviousButton(e *handler.ComponentEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	track, err := h.Database.PreviousHistoryTrack(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No songs in history"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get previous song", err))
	}

	if err = player.Update(context.Background(), lavalink.WithTrack(track.Track)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to play previous song", err))
	}
	return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(track.Track, 0)))
}

func (h *Handlers) OnPlayerPlayPauseButton(e *handler.ComponentEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	paused := !player.Paused()
	if err := player.Update(context.Background(), lavalink.WithPaused(paused)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to pause the player", err))
	}
	if paused {
		return e.CreateMessage(res.CreatePlayer("⏸ Paused the player.", false))
	}
	return e.CreateMessage(res.CreatePlayer("▶ Resumed the player.", false))
}

func (h *Handlers) OnPlayerNextButton(e *handler.ComponentEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	track, err := h.Database.NextQueueTrack(*e.GuildID())
	if errors.Is(err, sql.ErrNoRows) {
		return e.CreateMessage(res.CreateError("No more songs in queue"))
	}
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get next song", err))
	}

	if err = player.Update(context.Background(), lavalink.WithTrack(track.Track)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to play next song", err))
	}
	return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(track.Track, 0)))
}

func (h *Handlers) OnPlayerStopButton(e *handler.ComponentEvent) error {
	player := h.Lavalink.Player(*e.GuildID())
	if err := player.Destroy(context.Background()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to stop the player", err))
	}

	if err := h.Database.DeletePlayer(*e.GuildID()); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to delete the player from the database", err))
	}

	if err := h.Discord.UpdateVoiceState(context.Background(), *e.GuildID(), nil, false, false); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to disconnect from the voice channel", err))
	}

	return e.CreateMessage(res.Create("⏹ Stopped the player."))
}
