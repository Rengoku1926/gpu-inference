// Package app builds the whole application from config: connections,
// repositories, services, handlers, HTTP server and background workers.
// It keeps cmd/server/main.go down to "load config, build, run".
package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/Rengoku1926/gpu-inference/internal/config"
	"github.com/Rengoku1926/gpu-inference/internal/database"
	"github.com/Rengoku1926/gpu-inference/internal/handler"
	"github.com/Rengoku1926/gpu-inference/internal/repository"
	"github.com/Rengoku1926/gpu-inference/internal/server"
	"github.com/Rengoku1926/gpu-inference/internal/service"
	"github.com/Rengoku1926/gpu-inference/internal/worker"
)

type App struct {
	logger  zerolog.Logger
	pool    *pgxpool.Pool
	redis   *redis.Client
	server  *server.Server
	workers *worker.Runner
}

func New(ctx context.Context, cfg *config.Config, logger zerolog.Logger) (*App, error) {
	pool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}
	logger.Info().Str("host", cfg.Database.Host).Int("port", cfg.Database.Port).Msg("connected to postgres")

	redisClient, err := database.ConnectRedis(ctx, cfg.Redis, logger)
	if err != nil {
		pool.Close()
		return nil, err
	}
	logger.Info().Str("addr", cfg.Redis.Address).Msg("connected to redis")

	modelRepo := repository.NewModelRepository(pool)
	modelService := service.NewModelService(modelRepo, logger)

	router := server.NewRouter(cfg.Server, logger, server.Handlers{
		Health: handler.NewHealthHandler(pool, redisClient),
		Model:  handler.NewModelHandler(modelService),
	})

	return &App{
		logger:  logger,
		pool:    pool,
		redis:   redisClient,
		server:  server.New(cfg.Server, logger, router),
		workers: worker.NewRunner(logger), // queue consumer and reconciler register here
	}, nil
}

// Run blocks until ctx is cancelled or a component fails. Either way every
// component is stopped before it returns.
func (a *App) Run(ctx context.Context) error {
	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error { return a.server.Start(ctx) })
	g.Go(func() error { return a.workers.Run(ctx) })
	return g.Wait()
}

// Close releases connections. Call it after Run returns.
func (a *App) Close() {
	if err := a.redis.Close(); err != nil {
		a.logger.Warn().Err(err).Msg("failed to close redis")
	}
	a.pool.Close()
	a.logger.Info().Msg("connections closed")
}
