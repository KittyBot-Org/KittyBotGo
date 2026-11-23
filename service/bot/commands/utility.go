package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"github.com/kittybot-org/kittybotgo/service/bot/res"
)

var utilityCommand = discord.SlashCommandCreate{
	Name:        "utility",
	Description: "Small utilities",
	Options: []discord.ApplicationCommandOption{
		discord.ApplicationCommandOptionSubCommand{Name: "hastebin", Description: "Uploads text to hastebin", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "text", Description: "The text to upload", Required: true},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "translate", Description: "Translates text", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "text", Description: "The text to translate", Required: true},
			discord.ApplicationCommandOptionString{Name: "to", Description: "The target language, for example en", Required: true},
			discord.ApplicationCommandOptionString{Name: "from", Description: "The source language, defaults to auto"},
		}},
		discord.ApplicationCommandOptionSubCommand{Name: "steal-emote", Description: "Adds an emote to this server", Options: []discord.ApplicationCommandOption{
			discord.ApplicationCommandOptionString{Name: "name", Description: "The new emote name", Required: true},
			discord.ApplicationCommandOptionString{Name: "emote", Description: "An emote mention or name"},
			discord.ApplicationCommandOptionString{Name: "emote-id", Description: "An emote id"},
			discord.ApplicationCommandOptionString{Name: "url", Description: "A direct image URL"},
		}},
	},
}

func (c *commands) OnHastebin(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	base := strings.TrimRight(c.Config.Dashboard.HastebinURL, "/")
	if base == "" {
		base = "https://hastebin.com"
	}
	response, err := http.Post(base+"/documents", "text/plain", strings.NewReader(data.String("text")))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to upload the text", err))
	}
	defer response.Body.Close()
	var body struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil || body.Key == "" {
		return e.CreateMessage(res.CreateError("Hastebin did not return a key"))
	}
	return e.CreateMessage(res.Create(base + "/" + body.Key))
}

func (c *commands) OnTranslate(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	from, _ := data.OptString("from")
	if from == "" {
		from = "auto"
	}
	endpoint := fmt.Sprintf("https://translate.google.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s&ie=UTF-8&oe=UTF-8", url.QueryEscape(from), url.QueryEscape(data.String("to")), url.QueryEscape(data.String("text")))
	response, err := http.Get(endpoint)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to translate that text", err))
	}
	defer response.Body.Close()
	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to read the translation", err))
	}
	text := googleTranslation(payload)
	if text == "" {
		return e.CreateMessage(res.CreateError("No translation was returned"))
	}
	return e.CreateMessage(res.Create(text))
}

func googleTranslation(payload []byte) string {
	var root []any
	if err := json.Unmarshal(payload, &root); err != nil || len(root) == 0 {
		return ""
	}
	sentences, ok := root[0].([]any)
	if !ok {
		return ""
	}
	var text strings.Builder
	for _, sentence := range sentences {
		parts, ok := sentence.([]any)
		if !ok || len(parts) == 0 {
			continue
		}
		if part, ok := parts[0].(string); ok {
			text.WriteString(part)
		}
	}
	return text.String()
}

func (c *commands) OnStealEmote(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if !hasPerm(e, discord.PermissionManageGuildExpressions) {
		return e.CreateMessage(res.CreateError("You need the Manage Expressions permission"))
	}
	imageURL, _ := data.OptString("url")
	if imageURL == "" {
		if id, ok := data.OptString("emote-id"); ok && id != "" {
			parsed, err := snowflake.Parse(id)
			if err != nil {
				return e.CreateMessage(res.CreateError("That emote id is not valid"))
			}
			imageURL = fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.png", parsed)
		}
	}
	if imageURL == "" {
		if emote, ok := data.OptString("emote"); ok {
			imageURL = emojiURL(emote)
		}
	}
	if imageURL == "" {
		return e.CreateMessage(res.CreateError("Provide an emote, an emote id, or a URL"))
	}
	response, err := http.Get(imageURL)
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to download that emote", err))
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 256*1024))
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to read that emote", err))
	}
	icon, err := discord.ParseIconRaw(body)
	if err != nil {
		return e.CreateMessage(res.CreateErr("That image is not a supported emote format", err))
	}
	emoji, err := c.Discord.Rest.CreateEmoji(*e.GuildID(), discord.EmojiCreate{Name: data.String("name"), Image: *icon})
	if err != nil {
		return e.CreateMessage(res.CreateErr("Failed to create that emote", err))
	}
	return e.CreateMessage(res.Createf("Added %s", emoji.Mention()))
}

func emojiURL(value string) string {
	value = strings.TrimSpace(value)
	start := strings.LastIndex(value, ":")
	end := strings.LastIndex(value, ">")
	if start == -1 || end == -1 || end <= start+1 {
		return ""
	}
	id, err := snowflake.Parse(value[start+1 : end])
	if err != nil {
		return ""
	}
	extension := "png"
	if strings.HasPrefix(value, "<a:") {
		extension = "gif"
	}
	return fmt.Sprintf("https://cdn.discordapp.com/emojis/%s.%s", id, extension)
}
