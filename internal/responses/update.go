package responses

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

func UpdateError(e *events.ApplicationCommandInteractionEvent, content string) error {
	flags := discord.MessageFlagEphemeral
	_, err := e.Client().Rest().Interactions().UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{Embeds: &[]discord.Embed{
		{
			Description: content,
			Color:       ErrorColor,
		},
	}, Flags: &flags})
	return err
}

func UpdateErrorf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a ...any) error {
	return UpdateError(e, p.Sprintf(languageString, a))
}

func UpdateSuccess(e *events.ApplicationCommandInteractionEvent, content string) error {
	_, err := e.Client().Rest().Interactions().UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{Embeds: &[]discord.Embed{
		{
			Description: content,
			Color:       SuccessColor,
		},
	}})
	return err
}

func UpdateSuccessf(e *events.ApplicationCommandInteractionEvent, p *message.Printer, languageString string, a ...any) error {
	return UpdateSuccess(e, p.Sprintf(languageString, a))
}

func UpdateSuccessEmbed(e *events.ApplicationCommandInteractionEvent, embed discord.Embed) error {
	embed.Color = SuccessColor
	_, err := e.Client().Rest().Interactions().UpdateInteractionResponse(e.ApplicationID(), e.Token(), discord.MessageUpdate{Embeds: &[]discord.Embed{embed}})
	return err
}
