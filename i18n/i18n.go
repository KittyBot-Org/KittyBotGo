package i18n

import (
	"embed"
	"encoding/json"
	"strings"

	"github.com/disgoorg/log"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

//go:embed languages/*.json
var languages embed.FS

func Setup(logger log.Logger) error {
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
			logger.Errorf("Failed to parse tag %s: %s", tag, err)
			continue
		}

		var (
			data    map[string]interface{}
			rawData []byte
		)
		rawData, err = languages.ReadFile("languages/" + entry.Name())
		if err != nil {
			logger.Error("Failed to read language file: ", err)
			continue
		}
		if err = json.Unmarshal(rawData, &data); err != nil {
			logger.Error("Failed to parse language file: ", err)
			continue
		}
		parseData(logger, lang, data, "")
	}
	return nil
}

func parseData(logger log.Logger, lang language.Tag, data map[string]interface{}, path string) {
	for key, value := range data {
		if value == nil {
			continue
		}
		newPath := strings.TrimPrefix(path+"."+key, ".")
		switch v := value.(type) {
		case string:
			if err := message.SetString(lang, newPath, v); err != nil {
				logger.Errorf("Failed to set string with path %s and value %s: %s", newPath, v, err)
			}
		case map[string]interface{}:
			parseData(logger, lang, v, newPath)
		}
	}
}
