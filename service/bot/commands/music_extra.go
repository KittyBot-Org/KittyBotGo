package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/disgolink/v4/disgolink"
	"github.com/disgoorg/disgolink/v4/lavalink"
	"github.com/disgoorg/lavaqueue-plugin"

	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

func (c *commands) OnPlayerForward(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return c.seekBy(e, lavalink.Duration(data.Int("seconds"))*lavalink.Second)
}

func (c *commands) OnPlayerRewind(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return c.seekBy(e, -lavalink.Duration(data.Int("seconds"))*lavalink.Second)
}

func (c *commands) seekBy(e *handler.CommandEvent, delta lavalink.Duration) error {
	player := c.Lavalink.Player(*e.GuildID())
	if player.Track == nil {
		return e.CreateMessage(res.CreateError("There is no song playing right now."))
	}
	newPos := player.Position() + delta
	if newPos < 0 {
		newPos = 0
	}
	if newPos >= player.Track.Info.Length && !player.Track.Info.IsStream {
		track, err := lavaqueue.QueueNextTrack(e.Ctx, player.Node, *e.GuildID(), 1)
		if err != nil {
			return e.CreateMessage(res.CreateErr("Failed to skip to the next song", err))
		}
		if track == nil {
			return e.CreateMessage(res.Create("Skipped to the end of the song."))
		}
		return e.CreateMessage(res.CreatePlayerf("▶ Playing: %s", true, res.FormatTrack(*track, 0)))
	}
	if err := player.Update(e.Ctx, disgolink.WithPosition(newPos)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to seek", err))
	}
	return e.CreateMessage(res.CreatePlayerf("Seeked to %s.", false, res.FormatDuration(newPos)))
}

func (c *commands) OnFilterEqualizer(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	var eq lavalink.Equalizer
	if filters.Equalizer != nil {
		eq = *filters.Equalizer
	}
	eq[data.Int("band")] = float32(data.Float("multiplier"))
	filters.Equalizer = &eq
	return c.applyFilters(e, filters, "equalizer")
}

func (c *commands) OnFilterKaraoke(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	filters.Karaoke = &lavalink.Karaoke{
		Level:       float32(data.Float("level")),
		MonoLevel:   float32(data.Float("mono-level")),
		FilterBand:  float32(data.Float("filter-band")),
		FilterWidth: float32(data.Float("filter-width")),
	}
	return c.applyFilters(e, filters, "karaoke")
}

func (c *commands) OnFilterTimescale(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	speed, _ := data.OptFloat("speed")
	pitch, _ := data.OptFloat("pitch")
	rate, _ := data.OptFloat("rate")
	if speed == 0 {
		speed = 1
	}
	if pitch == 0 {
		pitch = 1
	}
	if rate == 0 {
		rate = 1
	}
	filters.Timescale = &lavalink.Timescale{Speed: speed, Pitch: pitch, Rate: rate}
	return c.applyFilters(e, filters, "timescale")
}

func (c *commands) OnFilterTremolo(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	filters.Tremolo = &lavalink.Tremolo{Frequency: float32(data.Float("frequency")), Depth: float32(data.Float("depth"))}
	return c.applyFilters(e, filters, "tremolo")
}

func (c *commands) OnFilterVibrato(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	filters.Vibrato = &lavalink.Vibrato{Frequency: float32(data.Float("frequency")), Depth: float32(data.Float("depth"))}
	return c.applyFilters(e, filters, "vibrato")
}

func (c *commands) OnFilterRotation(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	filters.Rotation = &lavalink.Rotation{RotationHz: data.Int("hertz")}
	return c.applyFilters(e, filters, "rotation")
}

func (c *commands) OnFilterDistortion(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	sinOffset, _ := data.OptFloat("sin-offset")
	sinScale, _ := data.OptFloat("sin-scale")
	cosOffset, _ := data.OptFloat("cos-offset")
	cosScale, _ := data.OptFloat("cos-scale")
	tanOffset, _ := data.OptFloat("tan-offset")
	tanScale, _ := data.OptFloat("tan-scale")
	offset, _ := data.OptFloat("offset")
	scale, _ := data.OptFloat("scale")
	filters.Distortion = &lavalink.Distortion{
		SinOffset: float32(sinOffset),
		SinScale:  float32(sinScale),
		CosOffset: float32(cosOffset),
		CosScale:  float32(cosScale),
		TanOffset: float32(tanOffset),
		TanScale:  float32(tanScale),
		Offset:    float32(offset),
		Scale:     float32(scale),
	}
	return c.applyFilters(e, filters, "distortion")
}

func (c *commands) OnFilterChannelMix(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	filters.ChannelMix = &lavalink.ChannelMix{
		LeftToLeft:   float32(data.Float("left-to-left")),
		LeftToRight:  float32(data.Float("left-to-right")),
		RightToLeft:  float32(data.Float("right-to-left")),
		RightToRight: float32(data.Float("right-to-right")),
	}
	return c.applyFilters(e, filters, "channel mix")
}

func (c *commands) OnFilterLowPass(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	player := c.Lavalink.Player(*e.GuildID())
	filters := player.Filters
	filters.LowPass = &lavalink.LowPass{Smoothing: data.Float("smoothing")}
	return c.applyFilters(e, filters, "low pass")
}

func (c *commands) OnFilterClear(_ discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	return c.applyFilters(e, lavalink.Filters{}, "cleared")
}

func (c *commands) applyFilters(e *handler.CommandEvent, filters lavalink.Filters, name string) error {
	player := c.Lavalink.Player(*e.GuildID())
	if err := player.Update(e.Ctx, disgolink.WithFilters(filters)); err != nil {
		return e.CreateMessage(res.CreateErr("Failed to set the filter", err))
	}
	return e.CreateMessage(res.CreatePlayerf("Set the %s filter.", false, name))
}
