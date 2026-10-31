package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Log is usable before Init (for config errors), and replaced by Init.
var Log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.Kitchen}).
	With().Timestamp().Logger()

// Init updates the global Log to a colored console logger in development or a
// JSON logger in production, then returns it. Call this once at startup after
// config is loaded.
func Init(env, level string) zerolog.Logger {
	Log = New(env, level)
	return Log
}

func New(env, level string) zerolog.Logger {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil || level == "" {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if env == "development" {
		return zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "15:04:05.000",
		}).With().Timestamp().Caller().Logger()
	}

	return zerolog.New(os.Stdout).With().
		Timestamp().
		Str("service", "gpu-inference").
		Logger()
}
