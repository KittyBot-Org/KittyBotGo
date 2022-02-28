package types

import (
	"github.com/DisgoOrg/disgo/core/events"
)

type CommandCheck func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool

func (c CommandCheck) Or(check CommandCheck) CommandCheck {
	return func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool {
		return c(b, e) || check(b, e)
	}
}

func (c CommandCheck) And(check CommandCheck) CommandCheck {
	return func(b *Bot, e *events.ApplicationCommandInteractionEvent) bool {
		return c(b, e) && check(b, e)
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
