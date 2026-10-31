package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Flush keeps streaming (SSE) working through the recorder.
func (r *responseRecorder) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (r *responseRecorder) Unwrap() http.ResponseWriter { return r.ResponseWriter }

// Logger puts a request scoped logger (carrying the request ID) in the
// context, so handlers and services can use zerolog.Ctx(ctx), then logs one
// line per request with the level chosen by status code.
func Logger(base zerolog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqLog := base.With().Str("request_id", GetRequestID(r.Context())).Logger()
			rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

			next.ServeHTTP(rec, r.WithContext(reqLog.WithContext(r.Context())))

			var event *zerolog.Event
			switch {
			case rec.status >= 500:
				event = reqLog.Error()
			case rec.status >= 400:
				event = reqLog.Warn()
			default:
				event = reqLog.Info()
			}
			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", rec.status).
				Int("bytes", rec.bytes).
				Dur("duration", time.Since(start)).
				Str("remote_ip", r.RemoteAddr).
				Msg("request")
		})
	}
}
