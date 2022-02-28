package i18n

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/KittyBot-Org/KittyBotGo/internal/types"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed languages/*.json
var languages embed.FS

func Setup(bot *types.Bot) error {
	entries, err := languages.ReadDir("languages")
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		tag := strings.TrimSuffix(entry.Name(), ".json")
		lang, err := language.Parse(tag)
		if err != nil {
			bot.Logger.Errorf("Failed to parse tag %s: %s", tag, err)
			continue
		}

		var (
			data    map[string]interface{}
			rawData []byte
		)
		rawData, err = languages.ReadFile("languages/" + entry.Name())
		if err != nil {
			bot.Logger.Error("Failed to read language file: ", err)
			continue
		}
		if err = json.Unmarshal(rawData, &data); err != nil {
			bot.Logger.Error("Failed to parse language file: ", err)
			continue
		}
		parseData(bot, lang, data, "")
	}
	return nil
}

func parseData(bot *types.Bot, lang language.Tag, data map[string]interface{}, path string) {
	for key, value := range data {
		if value == nil {
			continue
		}
		newPath := strings.TrimPrefix(path+"."+key, ".")
		switch v := value.(type) {
		case string:
			if err := message.SetString(lang, newPath, v); err != nil {
				bot.Logger.Errorf("Failed to set string with path %s and value %s: %s", newPath, v, err)
			}
		case map[string]interface{}:
			parseData(bot, lang, v, newPath)
		}
	}
}
