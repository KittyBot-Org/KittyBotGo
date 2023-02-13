package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

func Create(content string) discord.MessageCreate {
	return discord.MessageCreate{
		Content: content,
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
		Content: fmt.Sprintf(message, a...),
		Flags:   discord.MessageFlagEphemeral,
	}
}
