package tags

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/dbot"
	"github.com/disgoorg/disgo/discord"
)

var Module = dbot.DefaultCommandsModule{
	Cmds: []dbot.Command{
		{
			Create: discord.SlashCommandCreate{
				CommandName: "tag",
				Description: "lets you display a tag",
				Options: []discord.ApplicationCommandOption{

					discord.ApplicationCommandOptionString{
						Name:         "name",
						Description:  "the name of the tag to display",
						Required:     true,
						Autocomplete: true,
					},
				},
				DefaultPermission: true,
			},
			CommandHandler: map[string]dbot.CommandHandler{
				"": tagHandler,
			},
			AutoCompleteHandler: map[string]dbot.AutocompleteHandler{
				"": autoCompleteTagHandler,
			},
		},
		{
			Create: discord.SlashCommandCreate{
				CommandName: "tags",
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
				DefaultPermission: true,
			},
			CommandHandler: map[string]dbot.CommandHandler{
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
		},
	},
}
