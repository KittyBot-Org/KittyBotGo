package gateway

import (
	"github.com/disgoorg/disgo/cache"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/snowflake/v2"
)

// Cache is the gateway's memory of the Discord objects the bots read.
type Cache struct {
	cache.Caches
}

func NewCache() *Cache {
	return &Cache{
		Caches: cache.New(cache.WithCaches(
			cache.FlagGuilds,
			cache.FlagChannels,
			cache.FlagRoles,
			cache.FlagMembers,
			cache.FlagVoiceStates,
		)),
	}
}

func (c *Cache) Handle(event gateway.EventData) {
	switch e := event.(type) {
	case gateway.EventReady:
		c.SetSelfUser(e.User)
	case gateway.EventGuildCreate:
		c.guildCreate(e)
	case gateway.EventGuildUpdate:
		c.guildUpdate(e.Guild)
	case gateway.EventGuildDelete:
		if !e.Unavailable {
			c.removeGuild(e.ID)
		}
	case gateway.EventChannelCreate:
		c.AddChannel(discord.ApplyGuildIDToChannel(e.GuildChannel, e.GuildID()))
	case gateway.EventChannelUpdate:
		c.AddChannel(discord.ApplyGuildIDToChannel(e.GuildChannel, e.GuildID()))
	case gateway.EventChannelDelete:
		c.RemoveChannel(e.ID())
	case gateway.EventGuildRoleCreate:
		c.AddRole(e.Role)
	case gateway.EventGuildRoleUpdate:
		c.AddRole(e.Role)
	case gateway.EventGuildRoleDelete:
		c.RemoveRole(e.GuildID, e.RoleID)
	case gateway.EventGuildMemberAdd:
		c.AddMember(e.Member)
	case gateway.EventGuildMemberUpdate:
		c.AddMember(e.Member)
	case gateway.EventGuildMemberRemove:
		c.RemoveMember(e.GuildID, e.User.ID)
	case gateway.EventVoiceStateUpdate:
		if e.ChannelID == nil {
			c.RemoveVoiceState(e.GuildID, e.UserID)
		} else {
			c.AddVoiceState(e.VoiceState)
		}
		c.AddMember(e.Member)
	}
}

func (c *Cache) guildCreate(e gateway.EventGuildCreate) {
	c.AddGuild(e.CacheGuild)
	c.RemoveChannelsByGuildID(e.ID)
	c.RemoveRolesByGuildID(e.ID)
	c.RemoveMembersByGuildID(e.ID)
	c.RemoveVoiceStatesByGuildID(e.ID)

	for _, channel := range e.Channels {
		c.AddChannel(discord.ApplyGuildIDToChannel(channel, e.ID))
	}
	for _, thread := range e.Threads {
		c.AddChannel(discord.ApplyGuildIDToThread(thread, e.ID))
	}
	for _, role := range e.Roles {
		role.GuildID = e.ID
		c.AddRole(role)
	}
	for _, member := range e.Members {
		member.GuildID = e.ID
		c.AddMember(member)
	}
	for _, state := range e.VoiceStates {
		state.GuildID = e.ID
		if state.ChannelID != nil {
			c.AddVoiceState(state)
		}
	}
}

func (c *Cache) guildUpdate(guild discord.Guild) {
	current, _ := c.Guild(guild.ID)
	current.Guild = guild
	c.AddGuild(current)
}

func (c *Cache) removeGuild(guildID snowflake.ID) {
	c.RemoveGuild(guildID)
	c.RemoveChannelsByGuildID(guildID)
	c.RemoveRolesByGuildID(guildID)
	c.RemoveMembersByGuildID(guildID)
	c.RemoveVoiceStatesByGuildID(guildID)
}
