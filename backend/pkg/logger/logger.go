package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON structured logger writing to stdout.
// TODO: integrate with a centralized logging stack in production.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
