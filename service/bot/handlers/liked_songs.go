package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v2/lavalink"

	"github.com/KittyBot-Org/KittyBotGo/service/bot/res"
)

var trackRegex = regexp.MustCompile(`\[\x60(.+)\x60]\(<(?P<url>.+)?>\)`)

var likedSongsCommand = discord.SlashCommandCreate{
	Name:        "liked-songs",
	Description: "Shows your liked songs.",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{
			Name:        "add",
			Description: "Adds a song to your liked songs.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "query",
					Description: "The song to add.",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "Removes a song from your liked songs.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionInt{
					Name:         "song",
					Description:  "The song to remove.",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "show",
			Description: "Shows your liked songs.",
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "clear",
			Description: "Clears your liked songs.",
		},
	},
}

func findTrackURL(content string) string {
	allMatches := trackRegex.FindAllStringSubmatch(content, -1)
	if allMatches == nil || len(allMatches) == 0 || len(allMatches[0]) == 0 {
		return ""
	}

	return allMatches[0][trackRegex.SubexpIndex("url")]
}

func (h *Handlers) OnLikedSongsAddButton(e *handler.ComponentEvent) error {
	url := findTrackURL(e.Message.Content)
	if url == "" {
		for _, embed := range e.Message.Embeds {
			url = findTrackURL(embed.Description)
			if url != "" {
				break
			}
		}
	}
	if url == "" {
		return e.CreateMessage(res.CreateError("Failed to find a song URL."))
	}

	likedTrack, err := h.Database.FindLikedTrack(e.User().ID, url)
	if err != nil && err != sql.ErrNoRows {
		return e.CreateMessage(res.CreateErr("Failed to like song", err))
	}

	if err == sql.ErrNoRows {
		result, err := h.Lavalink.BestNode().Rest().LoadTracks(context.Background(), url)
		if err != nil {
			return e.CreateMessage(res.CreateErr("Failed to like song", err))
		}
		if result.LoadType == lavalink.LoadTypeLoadFailed {
			return e.CreateMessage(res.CreateErr("Failed to like song", err))
		} else if result.LoadType == lavalink.LoadTypeNoMatches || len(result.Tracks) == 0 {
			return e.CreateMessage(res.CreateError("Failed to like song: No matches found."))
		}

		track := result.Tracks[0]
		if err = h.Database.AddLikedTrack(e.User().ID, track); err != nil {
			return e.CreateMessage(res.CreateError("Failed to add song to your liked songs. Please try again."))
		}
		create := res.Createf("❤ Added %s to your liked songs.", res.FormatTrack(track, 0))
		create.Flags = discord.MessageFlagEphemeral
		return e.CreateMessage(create)
	}

	if err = h.Database.RemoveLikedTrack(likedTrack.ID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove song from your liked songs", err))
	}

	create := res.Createf("💔 Removed %s from your liked songs.", res.FormatTrack(likedTrack.Track, 0))
	create.Flags = discord.MessageFlagEphemeral
	return e.CreateMessage(create)
}

func (h *Handlers) OnLikedSongsShow(e *handler.CommandEvent) error {
	likedTracks, err := h.Database.GetLikedTracks(e.User().ID)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get liked songs", err))
	}

	if len(likedTracks) == 0 {
		return e.CreateMessage(res.CreateError("You don't have any liked songs."))
	}

	content := fmt.Sprintf("Liked Songs(`%d`):\n", len(likedTracks))
	for i, track := range likedTracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track.Track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}
	return e.CreateMessage(res.Create(content))
}

func (h *Handlers) OnLikedSongsRemove(e *handler.CommandEvent) error {
	trackID := e.SlashCommandInteractionData().Int("song")

	if err := h.Database.RemoveLikedTrack(trackID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove song from your liked songs", err))
	}

	return e.CreateMessage(res.Create("Removed song from your liked songs."))
}

func (h *Handlers) OnLikedSongsAutocomplete(e *handler.AutocompleteEvent) error {
	query := e.Data.String("song")
	likedTracks, err := h.Database.SearchLikedTracks(e.User().ID, query, 25)
	if err != nil {
		return e.Result(nil)
	}

	choices := make([]discord.AutocompleteChoice, len(likedTracks))
	for i, track := range likedTracks {
		choices[i] = discord.AutocompleteChoiceInt{
			Name:  res.Trim(track.Track.Info.Title, 100),
			Value: track.ID,
		}
	}
	return e.Result(choices)
}

func (h *Handlers) OnLikedSongsClear(e *handler.CommandEvent) error {
	if err := h.Database.ClearLikedTracks(e.User().ID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to clear liked songs", err))
	}
	return e.CreateMessage(res.Create("Cleared liked songs."))
}
