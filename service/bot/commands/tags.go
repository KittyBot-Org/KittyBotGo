package commands

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/kittybot-org/kittybotgo/service/bot/db"
	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var commandNamePattern = regexp.MustCompile(`^[a-z0-9_-]{1,32}$`)

var tagsCommand = discord.SlashCommandCreate{
	Name:        "tags",
	Description: "Create and send saved messages",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{Name: "create", Description: "Creates a tag", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true},
			discord.ApplicationCommandOptionString{Name: "content", Description: "The tag content", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "delete", Description: "Deletes a tag", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true, Autocomplete: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "edit", Description: "Edits a tag", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true, Autocomplete: true},
			discord.ApplicationCommandOptionString{Name: "content", Description: "The new content", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "info", Description: "Shows information about a tag", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true, Autocomplete: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "list", Description: "Lists tags in this server"},
		discord.ApplicationCommandOptionSubCommand{Name: "search", Description: "Searches tags", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "query", Description: "The search query", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "publish", Description: "Publishes a tag as a server command", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true, Autocomplete: true},
			discord.ApplicationCommandOptionString{Name: "description", Description: "The command description", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "unpublish", Description: "Removes a published tag command", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true, Autocomplete: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "send", Description: "Sends a tag", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The tag name", Required: true, Autocomplete: true},
		}},
	},
}

func (c *commands) OnTagCreate(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	name := strings.ToLower(data.String("name"))
	if !commandNamePattern.MatchString(name) {
		return e.CreateMessage(res.CreateError("Tag names must be 1-32 characters of a-z, 0-9, _ or -"))
	}
	if err := c.Database.CreateTag(e.Ctx, *e.GuildID(), e.User().ID, name, data.String("content")); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to create that tag", err))
	}
	return e.CreateMessage(res.Createf("Created tag `%s`", name))
}

func (c *commands) OnTagDelete(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	tag, err := c.Database.DeleteTag(e.Ctx, *e.GuildID(), data.String("name"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to delete that tag", err))
	}
	if tag.CommandID > 0 {
		_ = c.Discord.Rest.DeleteGuildCommand(c.Discord.ApplicationID, *e.GuildID(), snowflake.ID(tag.CommandID))
	}
	return e.CreateMessage(res.Createf("Deleted tag `%s`", tag.Name))
}

func (c *commands) OnTagEdit(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := c.Database.UpdateTag(e.Ctx, *e.GuildID(), data.String("name"), data.String("content")); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to edit that tag", err))
	}
	return e.CreateMessage(res.Createf("Updated tag `%s`", data.String("name")))
}

func (c *commands) OnTagInfo(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	tag, err := c.ownedTag(e, data.String("name"), false)
	if err != nil {
		return e.CreateMessage(res.CreateError(err.Error()))
	}
	return e.CreateMessage(res.Createf("Name: `%s`\nOwner: %s\nCreated: %s\nPublished: `%t`\n\n%s", tag.Name, discord.UserMention(tag.UserID), tag.CreatedAt.Format("2006-01-02 15:04"), tag.CommandID > 0, tag.Content))
}

func (c *commands) OnTagList(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	tags, err := c.Database.ListTags(e.Ctx, *e.GuildID())
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to list tags", err))
	}
	if len(tags) == 0 {
		return e.CreateMessage(res.Create("No tags"))
	}
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return e.CreateMessage(res.Create(strings.Join(names, ", ")))
}

func (c *commands) OnTagSearch(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	tags, err := c.Database.SearchTags(e.Ctx, *e.GuildID(), data.String("query"))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to search tags", err))
	}
	if len(tags) == 0 {
		return e.CreateMessage(res.Create("No tags matched"))
	}
	var names []string
	for _, tag := range tags {
		names = append(names, tag.Name)
	}
	return e.CreateMessage(res.Create(strings.Join(names, ", ")))
}

func (c *commands) OnTagPublish(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionManageGuild) {
		return e.CreateMessage(res.CreateError("You need the Manage Server permission to publish tags"))
	}
	tag, err := c.ownedTag(e, data.String("name"), false)
	if err != nil {
		return e.CreateMessage(res.CreateError(err.Error()))
	}
	if !commandNamePattern.MatchString(tag.Name) {
		return e.CreateMessage(res.CreateError("That tag name cannot be a slash command"))
	}
	command, err := c.Discord.Rest.CreateGuildCommand(c.Discord.ApplicationID, *e.GuildID(), discord.SlashCommandCreate{
		Name:        tag.Name,
		Description: data.String("description"),
	})
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to publish that tag", err))
	}
	if err := c.Database.SetTagCommandID(e.Ctx, tag.ID, int64(command.ID())); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to store the published command", err))
	}
	return e.CreateMessage(res.Createf("Published `/%s`", tag.Name))
}

func (c *commands) OnTagUnpublish(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionManageGuild) {
		return e.CreateMessage(res.CreateError("You need the Manage Server permission to unpublish tags"))
	}
	tag, err := c.ownedTag(e, data.String("name"), false)
	if err != nil {
		return e.CreateMessage(res.CreateError(err.Error()))
	}
	if tag.CommandID <= 0 {
		return e.CreateMessage(res.CreateError("That tag is not published"))
	}
	if err := c.Discord.Rest.DeleteGuildCommand(c.Discord.ApplicationID, *e.GuildID(), snowflake.ID(tag.CommandID)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to unpublish that tag", err))
	}
	if err := c.Database.SetTagCommandID(e.Ctx, tag.ID, -1); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to update that tag", err))
	}
	return e.CreateMessage(res.Createf("Unpublished `%s`", tag.Name))
}

func (c *commands) OnTagSend(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	tag, err := c.ownedTag(e, data.String("name"), false)
	if err != nil {
		return e.CreateMessage(res.CreateError(err.Error()))
	}
	return e.CreateMessage(discord.MessageCreate{Content: tag.Content})
}

func (c *commands) OnTagAutocomplete(e *handler.AutocompleteEvent) error {
	query := e.Data.String("name")
	tags, err := c.Database.SearchTags(e.Ctx, *e.GuildID(), query)
	if err != nil {
		return e.AutocompleteResult(nil)
	}
	var choices []discord.AutocompleteChoice
	for _, tag := range tags {
		choices = append(choices, discord.AutocompleteChoiceString{Name: tag.Name, Value: tag.Name})
	}
	return e.AutocompleteResult(choices)
}

func (c *commands) ownedTag(e *handler.CommandEvent, name string, ownerOnly bool) (db.Tag, error) {
	tag, err := c.Database.GetTag(e.Ctx, *e.GuildID(), name)
	if err != nil {
		return db.Tag{}, errors.New("tag not found")
	}
	if ownerOnly && tag.UserID != e.User().ID && !hasPerm(e, discord.PermissionManageGuild) {
		return db.Tag{}, fmt.Errorf("you do not own that tag")
	}
	return tag, nil
}
