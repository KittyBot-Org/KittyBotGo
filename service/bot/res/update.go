package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/json"
)

func Update(content string) discord.MessageUpdate {
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Description: content,
				Color:       PrimaryColor,
			},
		},
	}
}

func Updatef(format string, a ...any) discord.MessageUpdate {
	return Update(fmt.Sprintf(format, a...))
}

func UpdateErr(message string, err error, a ...any) discord.MessageUpdate {
	return UpdateError(message + ": " + err.Error())
}

func UpdateError(message string, a ...any) discord.MessageUpdate {
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Description: fmt.Sprintf(message, a...),
				Color:       DangerColor,
			},
		},
		Flags: json.Ptr(discord.MessageFlagEphemeral),
	}
}
