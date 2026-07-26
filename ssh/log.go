package ssh

import (
	"log/slog"
	"os"
)

var log *slog.Logger

func init() {
	log = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
}
