package responses

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

func Error(e *events.ApplicationCommandInteractionEvent, content string) error {
	return e.CreateMessage(discord.MessageCreate{Content: content, Flags: discord.MessageFlagEphemeral})
}

func Errorf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a ...any) error {
	return Error(e, p.Sprintf(languageString, a))
}

func Success(e *events.ApplicationCommandInteractionEvent, content string) error {
	return e.CreateMessage(discord.MessageCreate{Content: content})
}

func Successf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a ...any) error {
	return Success(e, p.Sprintf(languageString, a))
}
