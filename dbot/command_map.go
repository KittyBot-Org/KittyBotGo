package dbot

import (
	"strings"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func NewCommandMap(bot *Bot) *CommandMap {
	return &CommandMap{
		bot:      bot,
		commands: make(map[string]Command),
	}
}

type CommandMap struct {
	bot      *Bot
	commands map[string]Command
}

func (m *CommandMap) OnEvent(event bot.Event) {
	if e, ok := event.(*events.ApplicationCommandInteractionCreate); ok {
		if cmd, ok := m.commands[e.Data.CommandName()]; ok {
			if cmd.CommandHandler != nil {
				printer := getMessagePrinter(e.BaseInteraction)
				if cmd.Checks != nil && !cmd.Checks(m.bot, printer, e) {
					return
				}

				var path string
				if d, ok := e.Data.(discord.SlashCommandInteractionData); ok {
					path = buildCommandPath(d.SubCommandName, d.SubCommandGroupName)
				}
				if handler, ok := cmd.CommandHandler[path]; ok {
					if err := handler(m.bot, printer, e); err != nil {
						m.bot.Logger.Errorf("Failed to handle command \"%s\": %s", e.Data.CommandName(), err)
					}
					return
				}
			}
			m.bot.Logger.Errorf("No command handler for \"%s\"", e.Data.CommandName())
		}
	} else if e, ok := event.(*events.AutocompleteInteractionCreate); ok {
		if cmd, ok := m.commands[e.Data.CommandName]; ok {
			if cmd.AutoCompleteHandler != nil {
				if handler, ok := cmd.AutoCompleteHandler[buildCommandPath(e.Data.SubCommandName, e.Data.SubCommandGroupName)]; ok {
					if err := handler(m.bot, getMessagePrinter(e.BaseInteraction), e); err != nil {
						m.bot.Logger.Errorf("Failed to handle autocomplete for \"%s\": %s", e.Data.CommandName, err)
					}
					return
				}
			}
			m.bot.Logger.Errorf("No autocomplete handler for command \"%s\"", e.Data.CommandName)
		}
	} else if e, ok := event.(*events.ComponentInteractionCreate); ok {
		customID := e.Data.CustomID().String()
		if !strings.HasPrefix(customID, "cmd:") {
			return
		}
		args := strings.Split(customID, ":")
		cmdName, action := args[1], args[2]
		if cmd, ok := m.commands[cmdName]; ok {
			if cmd.ComponentHandler != nil {
				if handler, ok := cmd.ComponentHandler[action]; ok {
					if err := handler(m.bot, args[2:], getMessagePrinter(e.BaseInteraction), e); err != nil {
						m.bot.Logger.Errorf("Failed to handle component interaction for \"%s\" \"%s\" : %s", cmdName, action, err)
					}
					return
				}
			}
			m.bot.Logger.Errorf("No component handler for action \"%s\" on command \"%s\"", action, cmdName)
		}
	}
}

func getMessagePrinter(i discord.BaseInteraction) *message.Printer {
	lang, err := language.Parse(i.Locale().Code())
	if err != nil && i.GuildLocale() != nil {
		lang, _ = language.Parse(i.GuildLocale().Code())
	}
	if lang == language.Und {
		lang = language.English
	}
	return message.NewPrinter(lang)
}

func buildCommandPath(subcommand *string, subcommandGroup *string) string {
	var path string
	if subcommand != nil {
		path = *subcommand
	}
	if subcommandGroup != nil {
		path = *subcommandGroup + "/" + path
	}
	return path
}

func (m *CommandMap) AddAll(c []Command) {
	for _, cmd := range c {
		m.commands[cmd.Create.Name()] = cmd
	}
}
