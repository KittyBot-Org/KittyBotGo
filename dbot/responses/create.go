package responses

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

const (
	ErrorColor   = 0xFF0000
	SuccessColor = 0x5C5FEA
)

func CreateSuccessf(msg string, a ...any) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       SuccessColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
	}
}

func CreateSuccessComponentsf(msg string, a []any, components ...discord.ContainerComponent) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       SuccessColor,
				Description: fmt.Sprintf(msg, a...),
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

func CreateErrorf(msg string, a ...any) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       ErrorColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
		Flags: discord.MessageFlagEphemeral,
	}
}

func CreateErrorComponentsf(msg string, a []any, components ...discord.ContainerComponent) discord.MessageCreate {
	return discord.MessageCreate{
		Embeds: []discord.Embed{
			{
				Color:       ErrorColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
		Components: components,
		Flags:      discord.MessageFlagEphemeral,
	}
}
