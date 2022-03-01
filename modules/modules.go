package modules

import (
	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"github.com/KittyBot-Org/KittyBotGo/modules/music"
	"github.com/KittyBot-Org/KittyBotGo/modules/tags"
)

var Modules = []types.Module{
	music.Module,
	tags.Module,
}
