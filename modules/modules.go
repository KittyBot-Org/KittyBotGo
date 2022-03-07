package modules

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/bot/types"
	"github.com/KittyBot-Org/KittyBotGo/modules/metrics"
	"github.com/KittyBot-Org/KittyBotGo/modules/music"
	"github.com/KittyBot-Org/KittyBotGo/modules/tags"
)

var Modules = []types.Module{
	music.Module,
	tags.Module,
	metrics.Module,
}
