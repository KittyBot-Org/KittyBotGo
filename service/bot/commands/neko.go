package commands

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var nekoCommand = discord.SlashCommandCreate{
	Name:        "neko",
	Description: "Sends a reaction image",
	Options: []discord.ApplicationCommandOption{
		nekoSub("bite", "bite", false, true),
		nekoSub("cuddle", "cuddle", false, true),
		nekoSub("feed", "feed", false, true),
		nekoSub("hug", "hug", false, true),
		nekoSub("kiss", "kiss", false, true),
		nekoSub("lick", "lick", false, true),
		nekoSub("pat", "pat", false, true),
		nekoSub("poke", "poke", false, true),
		nekoSub("slap", "slap", false, true),
		nekoSub("tickle", "tickle", false, true),
		discord.ApplicationCommandOptionSubCommand{
			Name:        "neko",
			Description: "Sends an image from a neko category",
			Options: []discord.ApplicationCommandOption{
				discord.ApplicationCommandOptionString{
					Name:        "type",
					Description: "The image type",
					Required:    true,
					Choices: []discord.ApplicationCommandOptionChoiceString{
						{Name: "neko", Value: "sfw:neko:img"},
						{Name: "neko gif", Value: "sfw:neko:gif"},
						{Name: "kitsune", Value: "sfw:kitsune:img"},
						{Name: "senko", Value: "sfw:senko:img"},
						{Name: "nsfw neko", Value: "nsfw:neko:img"},
						{Name: "nsfw neko gif", Value: "nsfw:neko:gif"},
					},
				},
			},
		},
	},
}

func nekoSub(name string, _ string, _ bool, _ bool) discord.ApplicationCommandOptionSubCommand {
	return discord.ApplicationCommandOptionSubCommand{
		Name:        name,
		Description: "Sends a " + name + " image",
	}
}

func (c *commands) OnNeko(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	rating, name, kind := "sfw", data.CommandPath(), "gif"
	switch data.CommandPath() {
	case "/neko/bite":
		name = "bite"
	case "/neko/cuddle":
		name = "cuddle"
	case "/neko/feed":
		name = "feed"
	case "/neko/hug":
		name = "hug"
	case "/neko/kiss":
		name = "kiss"
	case "/neko/lick":
		name = "lick"
	case "/neko/pat":
		name = "pat"
	case "/neko/poke":
		name = "poke"
	case "/neko/slap":
		name = "slap"
	case "/neko/tickle":
		name = "tickle"
	case "/neko/neko":
		parts := splitNeko(data.String("type"))
		rating, name, kind = parts[0], parts[1], parts[2]
	}
	if rating == "nsfw" {
		if err := c.Database.EnsureGuild(e.Ctx, int64(*e.GuildID())); err != nil {
			return e.CreateMessage(res.CreateError("NSFW commands are disabled in this guild"))
		}
		settings, err := c.Database.GetGuild(e.Ctx, int64(*e.GuildID()))
		if err != nil || !settings.NsfwEnabled {
			return e.CreateMessage(res.CreateError("NSFW commands are disabled in this guild"))
		}
	}
	link, err := fetchNeko(rating, name, kind)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to fetch an image", err))
	}
	return e.CreateMessage(discord.MessageCreate{Embeds: []discord.Embed{{Image: &discord.EmbedResource{URL: link}, Color: res.PrimaryColor}}})
}

func splitNeko(value string) [3]string {
	var parts [3]string
	n := 0
	start := 0
	for i := 0; i <= len(value); i++ {
		if i == len(value) || value[i] == ':' {
			if n < 3 {
				parts[n] = value[start:i]
				n++
			}
			start = i + 1
		}
	}
	return parts
}

func fetchNeko(rating string, name string, kind string) (string, error) {
	response, err := http.Get(fmt.Sprintf("https://purrbot.site/api/img/%s/%s/%s", rating, name, kind))
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	var body struct {
		Link string `json:"link"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return "", err
	}
	if body.Link == "" {
		return "", fmt.Errorf("empty image link")
	}
	return body.Link, nil
}
