// Command migrate applies database migrations. It runs as its own step
// (make migrate locally, a release job in production) so the server never
// changes the schema on startup.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rengoku1926/gpu-inference/internal/config"
	"github.com/Rengoku1926/gpu-inference/internal/database"
	"github.com/Rengoku1926/gpu-inference/internal/logger"
)

func main() {
	cfg, err := config.Load(".env.local")
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load config")
	}

	log := logger.Init(cfg.Primary.Env, cfg.Primary.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer pool.Close()

	if err := database.Migrate(ctx, pool, log); err != nil {
		log.Error().Err(err).Msg("migration failed")
		pool.Close()
		os.Exit(1)
	}
}
