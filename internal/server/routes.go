package server

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/Rengoku1926/gpu-inference/internal/config"
	"github.com/Rengoku1926/gpu-inference/internal/handler"
	"github.com/Rengoku1926/gpu-inference/internal/middleware"
)

// Handlers is everything the router needs, built by the app package.
type Handlers struct {
	Health *handler.HealthHandler
	Model  *handler.ModelHandler
}

func NewRouter(cfg config.ServerConfig, logger zerolog.Logger, h Handlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", h.Health.HandleLiveness)
	mux.HandleFunc("GET /readyz", h.Health.HandleReadiness)

	mux.HandleFunc("POST /api/v1/models", h.Model.HandleRegister)
	mux.HandleFunc("GET /api/v1/models", h.Model.HandleList)
	mux.HandleFunc("GET /api/v1/models/{name}", h.Model.HandleGetByName)

	return middleware.Chain(mux,
		middleware.RequestID,
		middleware.Logger(logger),
		middleware.Recover,
		middleware.CORS(cfg.CORSAllowedOrigins),
	)
}
