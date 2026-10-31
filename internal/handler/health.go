package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type HealthHandler struct {
	pool  *pgxpool.Pool
	redis *redis.Client
}

func NewHealthHandler(pool *pgxpool.Pool, redisClient *redis.Client) *HealthHandler {
	return &HealthHandler{pool: pool, redis: redisClient}
}

// HandleLiveness says the process is up. It never touches dependencies, so a
// database outage doesn't get every replica restarted.
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// HandleReadiness says whether this replica should receive traffic.
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	checks := map[string]error{
		"postgres": h.pool.Ping(ctx),
		"redis":    h.redis.Ping(ctx).Err(),
	}

	resp := healthResponse{Status: "ok", Checks: map[string]string{}}
	status := http.StatusOK
	for name, err := range checks {
		if err != nil {
			zerolog.Ctx(r.Context()).Warn().Err(err).Str("dependency", name).Msg("readiness check failed")
			resp.Checks[name] = "down"
			resp.Status = "unavailable"
			status = http.StatusServiceUnavailable
			continue
		}
		resp.Checks[name] = "ok"
	}
	writeJSON(w, status, resp)
}
