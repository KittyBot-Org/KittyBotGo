package commands

import (
	"fmt"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/go-jet/jet/v2/qrm"
)

func Tag(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "tag",
			Description: "Lets you display a tag",
			Options: []discord.ApplicationCommandOption{

				discord.ApplicationCommandOptionString{
					Name:         "name",
					Description:  "The name of the tag to display",
					Required:     true,
					Autocomplete: true,
				},
			},
		},
		CommandHandlers: map[string]handler.CommandHandler{
			"": tagHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"": autoCompleteTagHandler(b),
		},
	}
}

func tagHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := strings.ToLower(data.String("name"))
		var msg string
		if tag, err := b.DB.Tags().Get(*e.GuildID(), name); err == nil {
			msg = tag.Content
		} else if err == qrm.ErrNoRows {
			msg = fmt.Sprintf("No tag with the name `%s` found", name)
		} else {
			msg = "Failed to get tag."
		}

		if err := b.DB.Tags().IncrementUses(*e.GuildID(), name); err != nil {
			b.Logger.Error("Failed to increment tag usage: ", err)
		}

		return e.CreateMessage(discord.MessageCreate{Content: msg})
	}
}
