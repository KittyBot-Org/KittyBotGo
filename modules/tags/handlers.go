package tags

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
	"github.com/DisgoOrg/utils/paginator"
	"github.com/KittyBot-Org/KittyBotGo/internal/models"
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"golang.org/x/text/message"
)

func tagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := *data.Options.String("name")
	msg := p.Sprintf("modules.tags.commands.tag.not.found", name)
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID).Scan(context.TODO(), &tag); err == nil {
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

func createTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()

	name := *data.Options.String("name")
	content := *data.Options.String("content")

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
	if _, err := b.DB.NewInsert().Model(&models.Tag{
		GuildID: *e.GuildID,
		Name:    name,
		Content: content,
		OwnerID: e.User.ID,
	}).Exec(context.TODO()); err != nil {
		msg = p.Sprintf("modules.tags.commands.tags.create.error")
	} else {
		msg = p.Sprintf("modules.tags.commands.tags.create.success", name)
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func deleteTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	name := *e.SlashCommandInteractionData().Options.String("name")
	var msg string
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, name).Scan(context.TODO(), &tag); err != nil {
		msg = p.Sprintf("modules.tags.commands.tags.not.found", name)
	} else {
		if tag.OwnerID != e.User.ID && !e.Member.InteractionPermissions().Has(discord.PermissionManageServer) {
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

func editTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	name := *data.Options.String("name")
	var msg string
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, name).Scan(context.TODO(), &tag); err != nil {
		msg = p.Sprintf("modules.tags.commands.tags.not.found", name)
	} else {
		if tag.OwnerID != e.User.ID && !e.Member.InteractionPermissions().Has(discord.PermissionManageServer) {
			msg = p.Sprintf("modules.tags.commands.tags.edit.no.permissions", name)
		} else {
			tag.Content = *data.Options.String("content")
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

func infoTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	name := *e.SlashCommandInteractionData().Options.String("name")
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, name).Scan(context.TODO(), &tag); err != nil {
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
		SetColor(types.KittyBotColor).
		Build()
	return e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
}

func listTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	var tags []models.Tag
	if err := b.DB.NewSelect().Model(&tags).Where("guild_id = ?", *e.GuildID).Order("name ASC").Scan(context.TODO(), &tags); err != nil {
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

	return b.Paginator.Create(e.CreateInteraction, &paginator.Paginator{
		PageFunc: func(page int, embed *discord.EmbedBuilder) discord.Embed {
			return embed.SetTitle(p.Sprintf("modules.tags.commands.tags.list.title", len(tags))).SetDescription(pages[page]).Build()
		},
		MaxPages:        len(pages),
		Expiry:          time.Now(),
		ExpiryLastUsage: true,
	})
}

func autoCompleteTagHandler(b *types.Bot, p *message.Printer, e *events.AutocompleteInteractionEvent) error {
	response := make(map[string]string)
	name := strings.ToLower(*e.Data.Options.String("name"))
	var tags []models.Tag
	if err := b.DB.NewSelect().Model(&tags).Where("guild_id = ?", *e.GuildID).Scan(context.TODO(), &tags); err == nil {
		var options []string
		for _, tag := range tags {
			options = append(options, tag.Name)
		}
		options = fuzzy.FindFold(name, options)
		for _, option := range options {
			if len(response) >= 25 {
				break
			}
			response[option] = option
		}
	}
	return e.ResultMapString(response)
}
