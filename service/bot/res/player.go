package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

func playerComponents(likeButton bool) []discord.ContainerComponent {
	actionRow := discord.ActionRowComponent{
		discord.NewPrimaryButton("", "/player/previous").WithEmoji(discord.ComponentEmoji{Name: "⏮"}),
		discord.NewPrimaryButton("", "/player/pause_play").WithEmoji(discord.ComponentEmoji{Name: "⏯"}),
		discord.NewPrimaryButton("", "/player/next").WithEmoji(discord.ComponentEmoji{Name: "⏭"}),
		discord.NewPrimaryButton("", "/player/stop").WithEmoji(discord.ComponentEmoji{Name: "⏹"}),
	}

	if likeButton {
		actionRow = append(actionRow, discord.NewPrimaryButton("", "/liked-songs/add").WithEmoji(discord.ComponentEmoji{Name: "❤"}))
	}

	return []discord.ContainerComponent{actionRow}
}

func CreatePlayer(content string, likeButton bool) discord.MessageCreate {
	return discord.MessageCreate{
		Content:    content,
		Components: playerComponents(likeButton),
	}
}

func CreatePlayerf(format string, likeButton bool, a ...any) discord.MessageCreate {
	return CreatePlayer(fmt.Sprintf(format, a...), likeButton)
}

func UpdatePlayer(content string, likeButton bool) discord.MessageUpdate {
	components := playerComponents(likeButton)
	return discord.MessageUpdate{
		Content:    &content,
		Components: &components,
	}
}

func UpdatePlayerf(format string, likeButton bool, a ...any) discord.MessageUpdate {
	return UpdatePlayer(fmt.Sprintf(format, a...), likeButton)
}
