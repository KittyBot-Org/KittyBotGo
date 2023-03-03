package listeners

import (
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"

	"github.com/KittyBot-Org/KittyBotGo/dbot"
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
			botVoiceState, ok := b.Client.Caches().VoiceState(e.VoiceState.GuildID, e.Client().ID())
			if ok && botVoiceState.ChannelID != nil && *botVoiceState.ChannelID == *e.OldVoiceState.ChannelID {
				var voiceStates []discord.VoiceState
				e.Client().Caches().VoiceStatesForEach(e.VoiceState.GuildID, func(voiceState discord.VoiceState) {
					if voiceState.ChannelID != nil && *voiceState.ChannelID == *botVoiceState.ChannelID {
						voiceStates = append(voiceStates, voiceState)
					}
				})
				if len(voiceStates) == 0 {
					go player.PlanDisconnect()
				}
			}
			return
		}
	})
}
