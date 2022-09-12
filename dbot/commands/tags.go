package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/utils/paginator"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/lib/pq"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

var Tags = handler.Command{
	Create: discord.SlashCommandCreate{
		Name:        "tags",
		Description: "lets you create/delete/edit tags",
		Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionSubCommand{
				Name:        "create",
				Description: "lets you create a tag",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "name",
						Description: "the name of the tag to create",
						Required:    true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "content",
						Description: "the content of the new tag",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "delete",
				Description: "lets you delete a tag",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "name",
						Description: "the name of the tag to delete",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "edit",
				Description: "lets you edit a tag",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:        "name",
						Description: "the name of the tag to edit",
						Required:    true,
					},
					discord.ApplicationCommandOptionString{
						Name:        "content",
						Description: "the new content of the new tag",
						Required:    true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "info",
				Description: "lets you view a tag's info",
				Options: []discord.ApplicationCommandOption{
					discord.ApplicationCommandOptionString{
						Name:         "name",
						Description:  "the name of the tag to view",
						Required:     true,
						Autocomplete: true,
					},
				},
			},
			discord.ApplicationCommandOptionSubCommand{
				Name:        "list",
				Description: "lists all tags",
			},
		},
	},
	CommandHandler: map[string]handler.CommandHandler{
		"create": createTagHandler,
		"delete": deleteTagHandler,
		"edit":   editTagHandler,
		"info":   infoTagHandler,
		"list":   listTagHandler,
	},
	AutoCompleteHandler: map[string]dbot.AutocompleteHandler{
		"list": autoCompleteTagHandler,
		"info": autoCompleteTagHandler,
	},
}

func createTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	name := strings.ToLower(data.String("name"))
	content := data.String("content")

	if len(name) >= 64 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.create.name.too.long"))
	}
	if len(content) >= 2048 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.create.content.too.long"))
	}

	if err := b.DB.Tags().Create(*e.GuildID(), e.User().ID, name, content); err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.create.duplicate", name))
		}
		b.Logger.Error("Failed to create tag: ", err)
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.create.error", name))
	}
	return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.create.success", name))
}

func deleteTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	name := strings.ToLower(e.SlashCommandInteractionData().String("name"))
	tag, err := b.DB.Tags().Get(*e.GuildID(), name)
	if err == qrm.ErrNoRows {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.not.found", name))
	} else if err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.error", name))
	}

	if tag.OwnerID != e.User().ID.String() || !e.Member().Permissions.Has(discord.PermissionManageServer) {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.delete.no.permissions", name))
	}
	if err = b.DB.Tags().Delete(*e.GuildID(), name); err != nil {
		b.Logger.Error("Failed to delete tag: ", err)
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.delete.error", name))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.tags.commands.tags.delete.success", name))
}

func editTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	data := e.SlashCommandInteractionData()
	name := strings.ToLower(data.String("name"))
	content := data.String("content")

	tag, err := b.DB.Tags().Get(*e.GuildID(), name)
	if err == qrm.ErrNoRows {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.edit.not.found", name))
	} else if err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.edit.error", name))
	}

	if tag.OwnerID != e.User().ID.String() || !e.Member().Permissions.Has(discord.PermissionManageServer) {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.edit.no.permissions", name))
	}

	if err = b.DB.Tags().Edit(*e.GuildID(), name, content); err != nil {
		b.Logger.Error("Failed to delete tag: ", err)
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.delete.error", name))
	}
	return e.CreateMessage(responses.CreateSuccessf(p, "modules.tags.commands.tags.edit.success", name))
}

func infoTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	name := strings.ToLower(e.SlashCommandInteractionData().String("name"))

	tag, err := b.DB.Tags().Get(*e.GuildID(), name)
	if err == qrm.ErrNoRows {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.info.not.found", name))
	} else if err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.info.error", name))
	}

	return e.CreateMessage(responses.CreateSuccessEmbed(discord.NewEmbedBuilder().
		SetTitle(p.Sprintf("modules.tags.commands.tags.info.title", tag.Name)).
		SetDescription(tag.Content).
		AddField(p.Sprintf("modules.tags.commands.tags.info.owner"), "<@"+tag.OwnerID+">", true).
		AddField(p.Sprintf("modules.tags.commands.tags.info.uses"), strconv.FormatInt(tag.Uses, 10), true).
		AddField(p.Sprintf("modules.tags.commands.tags.info.created.at"), fmt.Sprintf("%s (%s)", discord.NewTimestamp(discord.TimestampStyleNone, tag.CreatedAt), discord.NewTimestamp(discord.TimestampStyleRelative, tag.CreatedAt)), false).
		Build()))
}

func listTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error {
	tags, err := b.DB.Tags().GetAll(*e.GuildID())
	if err != nil {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.list.error"))
	}

	if len(tags) == 0 {
		return e.CreateMessage(responses.CreateErrorf(p, "modules.tags.commands.tags.list.no.tags"))
	}

	var pages []string
	curDesc := ""
	for _, tag := range tags {
		newDesc := fmt.Sprintf("**%s** - <@%s>\n", tag.Name, tag.OwnerID)
		if len(curDesc)+len(newDesc) > 2000 {
			pages = append(pages, curDesc)
			curDesc = ""
		}
		curDesc += newDesc
	}
	if len(curDesc) > 0 {
		pages = append(pages, curDesc)
	}

	return b.Paginator.Create(e.Respond, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) {
			embed.SetTitle(p.Sprintf("modules.tags.commands.tags.list.title", len(tags))).SetDescription(pages[page])
		},
		MaxPages:        len(pages),
		ExpiryLastUsage: true,
	})
}

func autoCompleteTagHandler(b *dbot.Bot, p *message.Printer, e *events.AutocompleteInteractionCreate) error {
	name := strings.ToLower(e.Data.String("name"))

	tags, err := b.DB.Tags().GetAll(*e.GuildID())
	if err != nil {
		return e.Result(nil)
	}
	var response []discord.AutocompleteChoice

	options := make([]string, len(tags))
	for i := range tags {
		options[i] = tags[i].Name
	}
	options = fuzzy.FindFold(name, options)
	for _, option := range options {
		if len(response) >= 25 {
			break
		}
		response = append(response, discord.AutocompleteChoiceString{
			Name:  option,
			Value: option,
		})
	}
	return e.Result(response)
}
