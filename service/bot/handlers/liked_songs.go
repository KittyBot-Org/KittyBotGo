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
			Description: "Adds a track to your liked songs.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "query",
					Description: "The track to add.",
					Required:    true,
				},
			},
		},
		discord.ApplicationCommandOptionSubCommand{
			Name:        "remove",
			Description: "Removes a track from your liked songs.",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:         "track",
					Description:  "The track to remove.",
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
		return e.CreateMessage(res.CreateError("Failed to find track URL."))
	}

	likedTrack, err := h.Database.FindLikedTrack(e.User().ID, url)
	if err != nil && err != sql.ErrNoRows {
		return e.CreateMessage(res.CreateErr("Failed to like track", err))
	}

	if err == sql.ErrNoRows {
		result, err := h.Lavalink.BestNode().Rest().LoadTracks(context.Background(), url)
		if err != nil {
			return e.CreateMessage(res.CreateErr("Failed to like track", err))
		}
		if result.LoadType == lavalink.LoadTypeLoadFailed {
			return e.CreateMessage(res.CreateErr("Failed to like track", err))
		}
		if result.LoadType == lavalink.LoadTypeNoMatches || len(result.Tracks) == 0 {
			return e.CreateMessage(res.CreateError("Failed to like track: No matches found."))
		}

		track := result.Tracks[0]
		if err = h.Database.AddLikedTrack(e.User().ID, track); err != nil {
			return e.CreateMessage(res.CreateError("Failed to add track to your liked tracks. Please try again."))
		}
		create := res.Createf("❤ Added %s to your liked tracks.", res.FormatTrack(track, 0))
		create.Flags = discord.MessageFlagEphemeral
		return e.CreateMessage(create)
	}

	if err = h.Database.RemoveLikedTrack(likedTrack.ID); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to remove track from your liked tracks", err))
	}

	create := res.Createf("💔 Removed %s from your liked tracks.", res.FormatTrack(likedTrack.Track, 0))
	create.Flags = discord.MessageFlagEphemeral
	return e.CreateMessage(create)
}

func (h *Handlers) OnLikedSongsShow(e *handler.CommandEvent) error {
	likedTracks, err := h.Database.GetLikedTracks(e.User().ID)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to get liked tracks", err))
	}

	if len(likedTracks) == 0 {
		return e.CreateMessage(res.CreateError("You don't have any liked tracks."))
	}

	content := fmt.Sprintf("Liked tracks(`%d`):\n", len(likedTracks))
	for i, track := range likedTracks {
		line := fmt.Sprintf("%d. %s\n", i+1, res.FormatTrack(track.Track, 0))
		if len([]rune(content))+len([]rune(line)) > 2000 {
			break
		}
		content += line
	}
	return e.CreateMessage(res.Create(content))
}
