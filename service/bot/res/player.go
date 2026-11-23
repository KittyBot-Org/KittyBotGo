package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

func playerComponents(likeButton bool) []discord.LayoutComponent {
	buttons := []discord.InteractiveComponent{
		discord.NewPrimaryButton("", "/player/previous").WithEmoji(discord.ComponentEmoji{Name: "⏮"}),
		discord.NewPrimaryButton("", "/player/pause_play").WithEmoji(discord.ComponentEmoji{Name: "⏯"}),
		discord.NewPrimaryButton("", "/player/next").WithEmoji(discord.ComponentEmoji{Name: "⏭"}),
		discord.NewPrimaryButton("", "/player/stop").WithEmoji(discord.ComponentEmoji{Name: "⏹"}),
	}
	if likeButton {
		buttons = append(buttons, discord.NewPrimaryButton("", "/liked-songs/add").WithEmoji(discord.ComponentEmoji{Name: "❤"}))
	}
	return []discord.LayoutComponent{discord.ActionRowComponent{Components: buttons}}
}

func CreatePlayer(content string, likeButton bool) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Description: content,
				Color:       PrimaryColor,
			},
		},
		Components: playerComponents(likeButton),
	}
}

func CreatePlayerf(format string, likeButton bool, a ...any) discord.MessageCreate {
	return CreatePlayer(fmt.Sprintf(format, a...), likeButton)
}

func UpdatePlayer(content string, likeButton bool) discord.MessageUpdate {
	components := playerComponents(likeButton)
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Description: content,
				Color:       PrimaryColor,
			},
		},
		Components: &components,
	}
}

func UpdatePlayerf(format string, likeButton bool, a ...any) discord.MessageUpdate {
	return UpdatePlayer(fmt.Sprintf(format, a...), likeButton)
}
