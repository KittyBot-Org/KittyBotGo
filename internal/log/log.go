package log

import (
	"log/slog"
	"os"

	"github.com/mattn/go-colorable"
	"github.com/topi314/tint"
)

type Config struct {
	Level     slog.Level `toml:"level"`
	Format    string     `toml:"format"`
	AddSource bool       `toml:"add_source"`
	NoColor   bool       `toml:"no_color"`
}

const (
	ansiFaint         = "\033[2m"
	ansiWhiteBold     = "\033[37;1m"
	ansiYellowBold    = "\033[33;1m"
	ansiCyanBold      = "\033[36;1m"
	ansiCyanBoldFaint = "\033[36;1;2m"
	ansiRedFaint      = "\033[31;2m"
	ansiRedBold       = "\033[31;1m"

	ansiRed     = "\033[31m"
	ansiYellow  = "\033[33m"
	ansiGreen   = "\033[32m"
	ansiMagenta = "\033[35m"
)

func Setup(cfg Config) {
	var h slog.Handler
	switch cfg.Format {
	case "json":
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: cfg.AddSource,
			Level:     cfg.Level,
		})

	case "text":
		h = tint.NewHandler(colorable.NewColorable(os.Stdout), &tint.Options{
			AddSource: cfg.AddSource,
			Level:     cfg.Level,
			NoColor:   cfg.NoColor,
			LevelColors: map[slog.Level]string{
				slog.LevelDebug: ansiMagenta,
				slog.LevelInfo:  ansiGreen,
				slog.LevelWarn:  ansiYellow,
				slog.LevelError: ansiRed,
			},
			Colors: map[tint.Kind]string{
				tint.KindTime:            ansiYellowBold,
				tint.KindSourceFile:      ansiCyanBold,
				tint.KindSourceSeparator: ansiCyanBoldFaint,
				tint.KindSourceLine:      ansiCyanBold,
				tint.KindMessage:         ansiWhiteBold,
				tint.KindKey:             ansiFaint,
				tint.KindSeparator:       ansiFaint,
				tint.KindValue:           ansiWhiteBold,
				tint.KindErrorKey:        ansiRedFaint,
				tint.KindErrorSeparator:  ansiFaint,
				tint.KindErrorValue:      ansiRedBold,
			},
		})
	default:
		slog.Error("Unknown log format", slog.String("format", cfg.Format))
		os.Exit(-1)
	}
	slog.SetDefault(slog.New(h))
}
