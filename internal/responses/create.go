package responses

import (
	"github.com/disgoorg/disgo/discord"
	"golang.org/x/text/message"
)

const (
	ErrorColor   = 0xFF0000
	SuccessColor = 0x5C5FEA
)

func CreateSuccessf(p *message.Printer, languageString string, a ...any) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       SuccessColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
	}
}

func CreateSuccessComponentsf(p *message.Printer, languageString string, a []any, components ...discord.ContainerComponent) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       SuccessColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
		Components: components,
	}
}

func CreateSuccessEmbed(embed discord.Embed) discord.MessageCreate {
	embed.Color = SuccessColor
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			embed,
		},
	}
}

func CreateSuccessEmbedComponents(embed discord.Embed, components ...discord.ContainerComponent) discord.MessageCreate {
	embed.Color = SuccessColor
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			embed,
		},
		Components: components,
	}
}

func CreateErrorf(p *message.Printer, languageString string, a ...any) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       ErrorColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
		Flags: discord.MessageFlagEphemeral,
	}
}

func CreateErrorComponentsf(p *message.Printer, languageString string, a []any, components ...discord.ContainerComponent) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       ErrorColor,
				Description: p.Sprintf(languageString, a...),
			},
		},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	}
}
