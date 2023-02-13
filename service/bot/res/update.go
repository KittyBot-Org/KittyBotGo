package res

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/json"
)

func Update(content string) discord.MessageUpdate {
	return discord.MessageUpdate{
		Content: &content,
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
		Content: json.Ptr(fmt.Sprintf(message, a...)),
		Flags:   json.Ptr(discord.MessageFlagEphemeral),
	}
}
