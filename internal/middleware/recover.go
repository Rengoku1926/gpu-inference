package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Recover turns a panic in a handler into a logged 500 instead of a dropped
// connection. It sits inside Logger so the panic is logged with the request ID.
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			if rec == http.ErrAbortHandler {
				panic(rec) // deliberate abort; let net/http handle it
			}
			zerolog.Ctx(r.Context()).Error().
				Interface("panic", rec).
				Bytes("stack", debug.Stack()).
				Msg("panic recovered")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal server error"}`))
		}()
		next.ServeHTTP(w, r)
	})
}
