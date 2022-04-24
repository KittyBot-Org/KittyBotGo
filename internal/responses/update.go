package responses

import (
	"github.com/disgoorg/disgo/discord"
	"golang.org/x/text/message"
)

func UpdateSuccessf(p *message.Printer, languageString string, a ...any) discord.MessageUpdate {
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       SuccessColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
	}
}

func UpdateSuccessComponentsf(p *message.Printer, languageString string, a []any, components ...discord.ContainerComponent) discord.MessageUpdate {
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       SuccessColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
		Components: &components,
	}
}

func UpdateErrorf(p *message.Printer, languageString string, a ...any) discord.MessageUpdate {
	flags := discord.MessageFlagEphemeral
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       ErrorColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
		Flags: &flags,
	}
}

func UpdateErrorComponentsf(p *message.Printer, languageString string, a []any, components ...discord.ContainerComponent) discord.MessageUpdate {
	flags := discord.MessageFlagEphemeral
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       ErrorColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
		Components: &components,
		Flags:      &flags,
	}
}
