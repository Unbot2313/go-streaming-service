package logger

import (
	"log/slog"
	"os"
	"sync"
)

var once sync.Once

// Setup configures the global slog logger with JSON output and source location.
func Setup() {
	once.Do(func() {
		handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			AddSource: true,
		})

		slog.SetDefault(slog.New(handler))
	})
}
