package types

import (
	"github.com/DisgoOrg/disgo/core/events"
	"github.com/DisgoOrg/disgo/discord"
)

type (
	CommandHandler      func(b *Bot, e *events.ApplicationCommandInteractionEvent) error
	ComponentHandler    func(b *Bot, e *events.ComponentInteractionEvent, action string) error
	AutocompleteHandler func(b *Bot, e *events.AutocompleteInteractionEvent) error
)

type Command struct {
	Create              discord.ApplicationCommandCreate
	Checks              CommandCheck
	CommandHandler      map[string]CommandHandler
	ComponentHandler    map[string]ComponentHandler
	AutoCompleteHandler map[string]AutocompleteHandler
}

func (b *Bot) SyncCommands() {
	b.Logger.Info("Syncing commands...")
	var commands []discord.ApplicationCommandCreate
	for _, cmd := range b.Commands.commands {
		commands = append(commands, cmd.Create)
	}

	defer b.Logger.Info("Synced Commands")
	if b.Config.DevMode {
		for _, guildID := range b.Config.DevGuildIDs {
			b.Logger.Infof("Syncing Commands for guild %s...", guildID)
			if _, err := b.Bot.SetGuildCommands(guildID, commands); err != nil {
				b.Logger.Errorf("Failed to sync commands for guild %s: %s", guildID, err)
			}
		}
	}
	b.Logger.Infof("Syncing Commands global...")
	if _, err := b.Bot.SetCommands(commands); err != nil {
		b.Logger.Errorf("Failed to sync commands global: %s", err)
	}
}
