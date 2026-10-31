package middleware

import "net/http"

type Middleware func(http.Handler) http.Handler

// Chain wraps h so the first middleware listed is the outermost.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
