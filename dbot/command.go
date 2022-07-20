package dbot

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

type (
	CommandHandler      func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionCreate) error
	ComponentHandler    func(b *Bot, args []string, p *message.Printer, e *events.ComponentInteractionCreate) error
	AutocompleteHandler func(b *Bot, p *message.Printer, e *events.AutocompleteInteractionCreate) error
)

type Command struct {
	Create              discord.ApplicationCommandCreate
	Checks              CommandCheck
	CommandHandler      map[string]CommandHandler
	ComponentHandler    map[string]ComponentHandler
	AutoCompleteHandler map[string]AutocompleteHandler
}

func (b *Bot) LoadCommands(commands ...Command) {
	b.CommandMap = NewCommandMap(b)
	b.CommandMap.AddAll(commands)
}

func (b *Bot) SyncCommands() {
	b.Logger.Info("Syncing commands...")
	var commands []discord.ApplicationCommandCreate
	for _, cmd := range b.CommandMap.commands {
		commands = append(commands, cmd.Create)
	}

	defer b.Logger.Info("Synced CommandMap")
	if b.Config.DevMode {
		for _, guildID := range b.Config.DevGuildIDs {
			b.Logger.Infof("Syncing CommandMap for guild %s...", guildID)
			if _, err := b.Client.Rest().SetGuildCommands(b.Client.ApplicationID(), guildID, commands); err != nil {
				b.Logger.Errorf("Failed to sync commands for guild %s: %s", guildID, err)
			}
		}
		return
	}
	b.Logger.Infof("Syncing CommandMap global...")
	if _, err := b.Client.Rest().SetGlobalCommands(b.Client.ApplicationID(), commands); err != nil {
		b.Logger.Errorf("Failed to sync commands global: %s", err)
	}
}
