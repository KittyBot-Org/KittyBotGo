package bot

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
)

func (b *Bot) onPublishedTag(e *events.InteractionCreate) {
	interaction, ok := e.Interaction.(*discord.ApplicationCommandInteraction)
	if !ok {
		return
	}
	data, ok := interaction.Data.(discord.SlashCommandInteractionData)
	if !ok {
		return
	}
	tag, err := b.Database.GetTagByCommandID(context.Background(), int64(data.CommandID()))
	if err != nil || tag.CommandID <= 0 {
		return
	}
	_ = e.Respond(discord.InteractionResponseTypeCreateMessage, discord.MessageCreate{Content: tag.Content})
}
