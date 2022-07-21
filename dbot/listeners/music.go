package listeners

import (
	"github.com/KittyBot-Org/KittyBotGo/dbot"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
)

func Music(b *dbot.Bot) bot.EventListener {
	return bot.NewListenerFunc(func(e *events.GuildVoiceLeave) {
		player := b.MusicPlayers.Get(e.VoiceState.GuildID)
		if player == nil {
			return
		}
		if e.VoiceState.UserID == b.Client.ID() {
			if err := player.Destroy(); err != nil {
				b.Logger.Error("Failed to destroy music player: ", err)
			}
			b.MusicPlayers.Delete(e.VoiceState.GuildID)
			return
		}
		if e.VoiceState.ChannelID == nil && e.OldVoiceState.ChannelID != nil {
			botVoiceState, ok := b.Client.Caches().VoiceStates().Get(e.VoiceState.GuildID, e.Client().ID())
			if ok && botVoiceState.ChannelID != nil && *botVoiceState.ChannelID == *e.OldVoiceState.ChannelID {
				voiceStates := e.Client().Caches().VoiceStates().FindAll(func(groupID snowflake.ID, voiceState discord.VoiceState) bool {
					return voiceState.ChannelID != nil && *voiceState.ChannelID == *botVoiceState.ChannelID
				})
				if len(voiceStates) == 0 {
					go player.PlanDisconnect()
				}
			}
			return
		}
	})
}
