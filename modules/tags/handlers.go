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
	"github.com/pkg/errors"
	"golang.org/x/text/message"
)

func tagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var msg string
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, *data.Options.String("name")).Scan(context.TODO(), &tag); err != nil {
		msg = "Tag not found."
	} else {
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
			Content: "Tag name must be less than 64 characters.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	if len(content) >= 1024 {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Tag content must be less than 1024 characters.",
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
		msg = "Failed to create tag: " + err.Error()
	} else {
		msg = "Tag created!"
	}
	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func deleteTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var msg string
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, *data.Options.String("name")).Scan(context.TODO(), &tag); err != nil {
		msg = "Tag not found."
	} else {
		if tag.OwnerID != e.User.ID && !e.Member.InteractionPermissions().Has(discord.PermissionManageServer) {
			msg = "You don't have permissions to delete this tag!"
		} else {
			if _, err = b.DB.NewDelete().Model(&tag).WherePK().Exec(context.TODO()); err != nil {
				msg = "Failed to delete tag: " + err.Error()
			} else {
				msg = "Tag deleted!"
			}
		}
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func editTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var msg string
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, *data.Options.String("name")).Scan(context.TODO(), &tag); err != nil {
		msg = "Tag not found."
	} else {
		if tag.OwnerID != e.User.ID && !e.Member.InteractionPermissions().Has(discord.PermissionManageServer) {
			msg = "You don't have permissions to edit this tag!"
		} else {
			tag.Content = *data.Options.String("content")
			if _, err = b.DB.NewUpdate().Model(&tag).Column("content").WherePK().Exec(context.TODO()); err != nil {
				msg = "Failed to edit tag: " + err.Error()
			} else {
				msg = "Tag edited!"
			}
		}
	}

	return e.CreateMessage(discord.MessageCreate{
		Content: msg,
	})
}

func infoTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	data := e.SlashCommandInteractionData()
	var tag models.Tag
	if err := b.DB.NewSelect().Model(&tag).Where("guild_id = ? AND name = ?", *e.GuildID, *data.Options.String("name")).Scan(context.TODO(), &tag); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Tag not found.",
		})
	}

	embed := discord.NewEmbedBuilder().
		SetTitlef("Tag Info: %s", tag.Name).
		SetDescription(tag.Content).
		AddField("Owner", "<@"+tag.OwnerID.String()+">", true).
		AddField("Uses", strconv.Itoa(tag.Uses), true).
		AddField("Created At", "<t:"+strconv.Itoa(int(tag.CreatedAt.Unix()))+"> (<t:"+strconv.Itoa(int(tag.CreatedAt.Unix()))+":R>)", false).
		SetColor(0xe24f96).
		Build()
	return e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
	})
}

func listTagHandler(b *types.Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) error {
	var tags []models.Tag
	if err := b.DB.NewSelect().Model(&tags).Where("guild_id = ?", *e.GuildID).Order("name ASC").Scan(context.TODO(), &tags); err != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Failed to list tags: " + err.Error(),
		})
	}

	if len(tags) == 0 {
		return e.CreateMessage(discord.MessageCreate{
			Content: "No tags have been made in the server.",
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
			return embed.SetTitlef("There are %d total tags:", len(tags)).SetDescription(pages[page]).Build()
		},
		MaxPages:        len(pages),
		Expiry:          time.Now(),
		ExpiryLastUsage: true,
	})
}

func autoCompleteTagHandler(b *types.Bot, p *message.Printer, e *events.AutocompleteInteractionEvent) error {
	response := make(map[string]string)
	opt := e.Data.Options.String("name")
	if opt == nil {
		return errors.New("No autocomplete name provided on required option")
	}
	name := strings.ToLower(*opt)
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
