package modules

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/kbot"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules/metrics"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules/music"
	"github.com/KittyBot-Org/KittyBotGo/internal/modules/tags"
)

var Modules = []kbot.Module{
	music.Module,
	tags.Module,
	metrics.Module,
}
