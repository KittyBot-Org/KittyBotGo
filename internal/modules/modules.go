package modules

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/dbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules/music"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules/tags"
)

var Modules = []dbot.Module{
	music.Module,
	tags.Module,
	metrics.Module,
}
