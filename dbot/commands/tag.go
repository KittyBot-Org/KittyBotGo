package commands

import (
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/go-jet/jet/v2/qrm"
	"golang.org/x/text/message"
)

var Tag = handler.Command{
	Create: discord.SlashCommandCreate{
		Name:        "tag",
		Description: "lets you display a tag",
		Options: []discord.ApplicationCommandOption{

			discord.ApplicationCommandOptionString{
				Name:         "name",
				Description:  "the name of the tag to display",
				Required:     true,
				Autocomplete: true,
			},
		},
	},
	CommandHandlers: map[string]handler.CommandHandler{
		"": tagHandler,
	},
	AutoCompleteHandler: map[string]dbot.AutocompleteHandler{
		"": autoCompleteTagHandler,
	},
}

func tagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	name := strings.ToLower(data.String("name"))
	var msg string
	if tag, err := b.DB.Tags().Get(*e.GuildID(), name); err == nil {
		msg = tag.Content
	} else if err == qrm.ErrNoRows {
		msg = p.Sprintf("modules.tags.commands.tag.not.found", name)
	} else {
		msg = p.Sprintf("modules.tags.commands.tag.error", name)
	}

	if err := b.DB.Tags().IncrementUses(*e.GuildID(), name); err != nil {
		b.Logger.Error("Failed to increment tag usage: ", err)
	}

	return e.CreateMessage(discord.MessageCreate{Content: msg})
}
