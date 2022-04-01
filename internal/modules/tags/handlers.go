package tags

import (
	"context"
	"fmt"
	"github.com/KittyBot-Org/KittyBotGo/internal/dbot"
	"strconv"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/internal/db"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/utils/paginator"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

func tagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := data.String("name")
	msg := p.Sprintf("modules.tags.commands.tag.not.found", name)
	var tag db.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name like ?", *e.GuildID(), name).Scan(context.TODO(), &tag); err == nil {
		msg = tag.Content
	}
	if err := e.CreateMessage(discord.MessageCreate{
		Content: msg,
	}); err != nil {
		return err
	}
	if len(tag.Content) > 0 {
		if _, err := b.DB.NewUpdate().Model(&tag).Set("uses = uses + 1").WherePK().Exec(context.TODO()); err != nil {
			return err
		}
	}
	return nil
}

func createTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	name := data.String("name")
	content := data.String("content")

	if len(name) >= 64 {
		return e.CreateMessage(discord.MessageCreate{
			Content: p.Sprintf("modules.tags.commands.tags.create.name.too.long"),
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	if len(content) >= 2048 {
		return e.CreateMessage(discord.MessageCreate{
			Content: p.Sprintf("modules.tags.commands.tags.create.content.too.long"),
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	var msg string
	if _, err := b.DB.NewInsert().Model(&db.Tag{
		GuildID: *e.GuildID(),
		Name:    name,
		Content: content,
		OwnerID: e.User().ID,
	}).Exec(context.TODO()); err != nil {
		msg = p.Sprintf("modules.tags.commands.tags.create.error")
	} else {
		msg = p.Sprintf("modules.tags.commands.tags.create.success", name)
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func deleteTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	name := e.SlashCommandInteractionData().String("name")
	var msg string
	var tag db.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID(), name).Scan(context.TODO(), &tag); err != nil {
		msg = p.Sprintf("modules.tags.commands.tags.not.found", name)
	} else {
		if tag.OwnerID != e.User().ID && !e.Member().Permissions.Has(discord.PermissionManageServer) {
			msg = p.Sprintf("modules.tags.commands.tags.delete.no.permissions", name)
		} else {
			if _, err = b.DB.NewDelete().Model(&tag).WherePK().Exec(context.TODO()); err != nil {
				msg = p.Sprintf("modules.tags.commands.tags.delete.error")
			} else {
				msg = p.Sprintf("modules.tags.commands.tags.delete.success", name)
			}
		}
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func editTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := data.String("name")
	var msg string
	var tag db.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID(), name).Scan(context.TODO(), &tag); err != nil {
		msg = p.Sprintf("modules.tags.commands.tags.not.found", name)
	} else {
		if tag.OwnerID != e.User().ID && !e.Member().Permissions.Has(discord.PermissionManageServer) {
			msg = p.Sprintf("modules.tags.commands.tags.edit.no.permissions", name)
		} else {
			tag.Content = data.String("content")
			if _, err = b.DB.NewUpdate().Model(&tag).Column("content").WherePK().Exec(context.TODO()); err != nil {
				msg = p.Sprintf("modules.tags.commands.tags.edit.error")
			} else {
				msg = p.Sprintf("modules.tags.commands.tags.edit.success", name)
			}
		}
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func infoTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	name := e.SlashCommandInteractionData().String("name")
	var tag db.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID(), name).Scan(context.TODO(), &tag); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: p.Sprintf("modules.tags.commands.tags.not.found", name),
		})
	}

	embed := discord.NewEmbedBuilder().
		SetTitle(p.Sprintf("modules.tags.commands.tags.info.title", tag.Name)).
		SetDescription(tag.Content).
		AddField(p.Sprintf("modules.tags.commands.tags.info.owner"), "<@"+tag.OwnerID.String()+">", true).
		AddField(p.Sprintf("modules.tags.commands.tags.info.uses"), strconv.Itoa(tag.Uses), true).
		AddField(p.Sprintf("modules.tags.commands.tags.info.created.at"), fmt.Sprintf("%s (%s)", discord.NewTimestamp(discord.TimestampStyleNone, tag.CreatedAt), discord.NewTimestamp(discord.TimestampStyleRelative, tag.CreatedAt)), false).
		SetColor(dbot.KittyBotColor).
		Build()
	return e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
}

func listTagHandler(b *dbot.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	var tags []db.Tag
	if err := b.DB.NewSelect().Model(&tags).Where("guild_id = ?", *e.GuildID()).Order("name ASC").Scan(context.TODO(), &tags); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: p.Sprintf("modules.tags.commands.tags.list.error"),
		})
	}

	if len(tags) == 0 {
		return e.CreateMessage(discord.MessageCreate{
			Content: p.Sprintf("modules.tags.commands.tags.list.no.tags"),
		})
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
	var (
		tags     []db.Tag
		response []discord.AutocompleteChoice
	)
	if err := b.DB.NewSelect().Model(&tags).Where("guild_id = ?", *e.GuildID()).Scan(context.TODO(), &tags); err == nil {
		var options []string
		for _, tag := range tags {
			options = append(options, tag.Name)
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
	}
	return e.Result(response)
}
