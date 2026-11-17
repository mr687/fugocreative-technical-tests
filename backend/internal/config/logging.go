package config

import (
	"log/slog"
	"os"
)

func InitLogging() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}
