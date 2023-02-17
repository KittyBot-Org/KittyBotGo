package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

const (
	PrimaryColor = 0x5c5fea
	DangerColor  = 0xd43535
)

func Create(content string) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Description: content,
				Color:       PrimaryColor,
			},
		},
	}
}

func Createf(format string, a ...any) discord.MessageCreate {
	return Create(fmt.Sprintf(format, a...))
}

func CreateErr(message string, err error, a ...any) discord.MessageCreate {
	return CreateError(message + ": " + err.Error())
}

func CreateError(message string, a ...any) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Description: fmt.Sprintf(message, a...),
				Color:       DangerColor,
			},
		},
		Flags: discord.MessageFlagEphemeral,
	}
}
