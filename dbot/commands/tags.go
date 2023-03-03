package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/handler"
	"github.com/disgoorg/utils/paginator"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/lib/pq"
	"github.com/lithammer/fuzzysearch/fuzzy"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/KittyBot-Org/KittyBotGo/dbot/responses"
)

func Tags(b *dbot.Bot) handler.Command {
	return handler.Command{
		Create: discord.SlashCommandCreate{
			Name:        "tags",
			Description: "Lets you create/delete/edit tags",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionSubCommand{
					Name:        "create",
					Description: "Lets you create a tag",
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
					Description: "Lets you delete a tag",
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
					Description: "Lets you edit a tag",
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
					Description: "Lets you view a tag's info",
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
		CommandHandlers: map[string]handler.CommandHandler{
			"create": createTagHandler(b),
			"delete": deleteTagHandler(b),
			"edit":   editTagHandler(b),
			"info":   infoTagHandler(b),
			"list":   listTagHandler(b),
		},
		AutocompleteHandlers: map[string]handler.AutocompleteHandler{
			"list": autoCompleteTagHandler(b),
			"info": autoCompleteTagHandler(b),
		},
	}
}

func createTagHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := strings.ToLower(data.String("name"))
		content := data.String("content")

		if len(name) >= 64 {
			return e.CreateMessage(responses.CreateErrorf("Tag name must be less than 64 characters."))
		}
		if len(content) >= 2048 {
			return e.CreateMessage(responses.CreateErrorf("Tag content must be less than 2048 characters."))
		}

		if err := b.DB.Tags().Create(*e.GuildID(), e.User().ID, name, content); err != nil {
			if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
				return e.CreateMessage(responses.CreateErrorf("Tag with this name already exists."))
			}
			b.Logger.Error("Failed to create tag: ", err)
			return e.CreateMessage(responses.CreateErrorf("Failed to create tag. Please try again."))
		}
		return e.CreateMessage(responses.CreateErrorf("Created tag with name: `%s`.", name))
	}
}

func deleteTagHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		name := strings.ToLower(e.SlashCommandInteractionData().String("name"))
		tag, err := b.DB.Tags().Get(*e.GuildID(), name)
		if err == qrm.ErrNoRows {
			return e.CreateMessage(responses.CreateErrorf("No tag found with name `%s`.", name))
		} else if err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to check tag. Please try again."))
		}

		if tag.OwnerID != e.User().ID.String() || !e.Member().Permissions.Has(discord.PermissionManageServer) {
			return e.CreateMessage(responses.CreateErrorf("You don't have permissions to delete this tag."))
		}
		if err = b.DB.Tags().Delete(*e.GuildID(), name); err != nil {
			b.Logger.Error("Failed to delete tag: ", err)
			return e.CreateMessage(responses.CreateErrorf("Failed to delete tag. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessf("Deleted tag with name: `%s`.", name))
	}
}

func editTagHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		data := e.SlashCommandInteractionData()
		name := strings.ToLower(data.String("name"))
		content := data.String("content")

		tag, err := b.DB.Tags().Get(*e.GuildID(), name)
		if err == qrm.ErrNoRows {
			return e.CreateMessage(responses.CreateErrorf("No tag found with name `%s`.", name))
		} else if err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to check tag. Please try again."))
		}

		if tag.OwnerID != e.User().ID.String() || !e.Member().Permissions.Has(discord.PermissionManageServer) {
			return e.CreateMessage(responses.CreateErrorf("You don't have permissions to edit this tag."))
		}

		if err = b.DB.Tags().Edit(*e.GuildID(), name, content); err != nil {
			b.Logger.Error("Failed to edit tag: ", err)
			return e.CreateMessage(responses.CreateErrorf("Failed to edit tag. Please try again."))
		}
		return e.CreateMessage(responses.CreateSuccessf("Edited tag with name: `%s`.", name))
	}
}

func infoTagHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		name := strings.ToLower(e.SlashCommandInteractionData().String("name"))

		tag, err := b.DB.Tags().Get(*e.GuildID(), name)
		if err == qrm.ErrNoRows {
			return e.CreateMessage(responses.CreateErrorf("No tag found with name `%s`.", name))
		} else if err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to check tag. Please try again."))
		}

		return e.CreateMessage(responses.CreateSuccessEmbed(discord.NewEmbedBuilder().
			SetTitlef("Tag Info: %s", tag.Name).
			SetDescription(tag.Content).
			AddField("Owner:", "<@"+tag.OwnerID+">", true).
			AddField("Uses:", strconv.FormatInt(tag.Uses, 10), true).
			AddField("Created:", fmt.Sprintf("%s (%s)", discord.NewTimestamp(discord.TimestampStyleNone, tag.CreatedAt), discord.NewTimestamp(discord.TimestampStyleRelative, tag.CreatedAt)), false).
			Build()))
	}
}

func listTagHandler(b *dbot.Bot) handler.CommandHandler {
	return func(e *events.ApplicationCommandInteractionCreate) error {
		tags, err := b.DB.Tags().GetAll(*e.GuildID())
		if err != nil {
			return e.CreateMessage(responses.CreateErrorf("Failed to list tags. Please try again."))
		}

		if len(tags) == 0 {
			return e.CreateMessage(responses.CreateErrorf("No tags found for this server."))
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
				embed.SetTitlef("Tags(%d):", len(tags)).SetDescription(pages[page])
			},
			MaxPages:        len(pages),
			ExpiryLastUsage: true,
		})
	}
}

func autoCompleteTagHandler(b *dbot.Bot) handler.AutocompleteHandler {
	return func(e *events.AutocompleteInteractionCreate) error {
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
}
