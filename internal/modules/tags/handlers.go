package tags

import (
	"database/sql"
	"fmt"
	"github.com/KittyBot-Org/KittyBotGo/internal/responses"
	"strconv"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/internal/dbot"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/utils/paginator"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

func tagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := strings.ToLower(data.String("name"))
	var msg string
	if tag, err := b.DB.Tags().Get(*e.GuildID(), name); err == nil {
		msg = tag.Content
	} else if err == sql.ErrNoRows {
		msg = p.Sprintf("modules.tags.commands.tag.not.found", name)
	} else {
		msg = p.Sprintf("modules.tags.commands.tag.error", name)
	}

	if err := b.DB.Tags().IncrementUses(*e.GuildID(), name); err != nil {
		b.Logger.Error("Failed to increment tag usage: ", err)
	}

	return e.CreateMessage(discord.MessageCreate{Content: msg})
}

func createTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := strings.ToLower(data.String("name"))
	content := data.String("content")

	if len(name) >= 64 {
		return responses.Errorf(e, p, "modules.tags.commands.tags.create.name.too.long")
	}
	if len(content) >= 2048 {
		return responses.Errorf(e, p, "modules.tags.commands.tags.create.content.too.long")
	}

	if err := b.DB.Tags().Create(*e.GuildID(), e.User().ID, name, content); err == nil {
		b.Logger.Error("Failed to create tag: ", err)
		return responses.Errorf(e, p, "modules.tags.commands.tags.create.error", name)
	}
	// TODO: handle duplicate tag name
	return responses.Errorf(e, p, "modules.tags.commands.tags.create.success", name)
}

func deleteTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	name := strings.ToLower(e.SlashCommandInteractionData().String("name"))
	tag, err := b.DB.Tags().Get(*e.GuildID(), name)
	if err == sql.ErrNoRows {
		return responses.Errorf(e, p, "modules.tags.commands.tags.not.found", name)
	} else if err != nil {
		return responses.Errorf(e, p, "modules.tags.commands.tags.error", name)
	}

	if tag.OwnerID != e.User().ID.String() || !e.Member().Permissions.Has(discord.PermissionManageServer) {
		return responses.Errorf(e, p, "modules.tags.commands.tags.delete.no.permissions", name)
	}
	if err = b.DB.Tags().Delete(*e.GuildID(), name); err != nil {
		b.Logger.Error("Failed to delete tag: ", err)
		return responses.Errorf(e, p, "modules.tags.commands.tags.delete.error", name)
	}
	return responses.Successf(e, p, "modules.tags.commands.tags.delete.success", name)
}

func editTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := strings.ToLower(data.String("name"))
	content := data.String("content")

	tag, err := b.DB.Tags().Get(*e.GuildID(), name)
	if err == sql.ErrNoRows {
		return responses.Errorf(e, p, "modules.tags.commands.tags.edit.not.found", name)
	} else if err != nil {
		return responses.Errorf(e, p, "modules.tags.commands.tags.edit.error", name)
	}

	if tag.OwnerID != e.User().ID.String() || !e.Member().Permissions.Has(discord.PermissionManageServer) {
		return responses.Errorf(e, p, "modules.tags.commands.tags.edit.no.permissions", name)
	}

	if err = b.DB.Tags().Edit(*e.GuildID(), name, content); err != nil {
		b.Logger.Error("Failed to delete tag: ", err)
		return responses.Errorf(e, p, "modules.tags.commands.tags.delete.error", name)
	}
	return responses.Successf(e, p, "modules.tags.commands.tags.edit.success", name)
}

func infoTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	name := strings.ToLower(e.SlashCommandInteractionData().String("name"))

	tag, err := b.DB.Tags().Get(*e.GuildID(), name)
	if err == sql.ErrNoRows {
		return responses.Errorf(e, p, "modules.tags.commands.tags.info.not.found", name)
	} else if err != nil {
		return responses.Errorf(e, p, "modules.tags.commands.tags.info.error", name)
	}

	embed := discord.NewEmbedBuilder().
		SetTitle(p.Sprintf("modules.tags.commands.tags.info.title", tag.Name)).
		SetDescription(tag.Content).
		AddField(p.Sprintf("modules.tags.commands.tags.info.owner"), "<@"+tag.OwnerID+">", true).
		AddField(p.Sprintf("modules.tags.commands.tags.info.uses"), strconv.FormatInt(tag.Uses, 10), true).
		AddField(p.Sprintf("modules.tags.commands.tags.info.created.at"), fmt.Sprintf("%s (%s)", discord.NewTimestamp(discord.TimestampStyleNone, tag.CreatedAt), discord.NewTimestamp(discord.TimestampStyleRelative, tag.CreatedAt)), false).
		SetColor(dbot.KittyBotColor).
		Build()
	return e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
}

func listTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	tags, err := b.DB.Tags().GetAll(*e.GuildID())
	if err != nil {
		return responses.Errorf(e, p, "modules.tags.commands.tags.list.error")
	}

	if len(tags) == 0 {
		return responses.Errorf(e, p, "modules.tags.commands.tags.list.no.tags")
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

func autoCompleteTagHandler(b *dbot.Bot, p *message.Printer, e *events.AutocompleteInteractionEvent) error {
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
