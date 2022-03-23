package types

import (
	"github.com/disgoorg/disgo/events"
	"golang.org/x/text/message"
)

type CommandCheck func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool

func (c CommandCheck) Or(check CommandCheck) CommandCheck {
	return func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		return c(b, p, e) || check(b, p, e)
	}
}

func (c CommandCheck) And(check CommandCheck) CommandCheck {
	return func(b *Bot, p *message.Printer, e *events.ApplicationCommandInteractionEvent) bool {
		return c(b, p, e) && check(b, p, e)
	}
}

func CommandCheckAnyOf(checks ...CommandCheck) CommandCheck {
	var check CommandCheck
	for _, c := range checks {
		check = check.Or(c)
	}
	return check
}

func CommandCheckAllOf(checks ...CommandCheck) CommandCheck {
	var check CommandCheck
	for _, c := range checks {
		check = check.And(c)
	}
	return check
}
