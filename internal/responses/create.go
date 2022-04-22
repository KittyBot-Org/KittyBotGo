package responses

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

const (
	ErrorColor   = 0xFF0000
	SuccessColor = 0x5C5FEA
)

func Error(e *events.ApplicationCommandInteractionEvent, content string) error {
	return e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{
		{
			Description: content,
			Color:       ErrorColor,
		},
	}, Flags: discord.MessageFlagEphemeral})
}

func Errorf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a ...any) error {
	return Error(e, p.Sprintf(languageString, a))
}

func SuccessEmbedComponents(e *events.ApplicationCommandInteractionEvent, embed discord.Embed, components ...discord.ContainerComponent) error {
	embed.Color = SuccessColor
	return e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{embed}, Components: components})
}

func SuccessEmbed(e *events.ApplicationCommandInteractionEvent, embed discord.Embed) error {
	return SuccessEmbedComponents(e, embed)
}

func Success(e *events.ApplicationCommandInteractionEvent, content string) error {
	return SuccessEmbed(e, discord.Embed{
		Description: content,
		Color:       SuccessColor,
	})
}

func Successf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a ...any) error {
	return Success(e, p.Sprintf(languageString, a))
}

func SuccessComponentsf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a []any, components ...discord.ContainerComponent) error {
	return Success(e, p.Sprintf(languageString, a))
}
