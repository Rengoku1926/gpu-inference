package server

import (
	"log"
	"strings"

	"github.com/rs/zerolog"
)

// newStdLogger routes net/http's internal errors (TLS handshakes, bad
// requests it rejects itself) through zerolog instead of plain stderr.
func newStdLogger(logger zerolog.Logger) *log.Logger {
	return log.New(writerFunc(func(p []byte) (int, error) {
		logger.Warn().Str("source", "net/http").Msg(strings.TrimSpace(string(p)))
		return len(p), nil
	}), "", 0)
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
