package responses

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
)

func UpdateSuccessf(msg string, a ...any) discord.MessageUpdate {
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       SuccessColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
	}
}

func UpdateSuccessComponentsf(msg string, a []any, components ...discord.ContainerComponent) discord.MessageUpdate {
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       SuccessColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
		Components: &components,
	}
}

func UpdateErrorf(msg string, a ...any) discord.MessageUpdate {
	flags := discord.MessageFlagEphemeral
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       ErrorColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
		Flags: &flags,
	}
}

func UpdateErrorComponentsf(msg string, a []any, components ...discord.ContainerComponent) discord.MessageUpdate {
	flags := discord.MessageFlagEphemeral
	return discord.MessageUpdate{
		Embeds: &[]discord.Embed{
			{
				Color:       ErrorColor,
				Description: fmt.Sprintf(msg, a...),
			},
		},
		Components: &components,
		Flags:      &flags,
	}
}
