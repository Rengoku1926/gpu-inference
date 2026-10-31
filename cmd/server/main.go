package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/Rengoku1926/gpu-inference/internal/app"
	"github.com/Rengoku1926/gpu-inference/internal/config"
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

	application, err := app.New(ctx, cfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to start application")
	}

	err = application.Run(ctx)
	application.Close()
	if err != nil {
		log.Error().Err(err).Msg("application stopped with error")
		os.Exit(1)
	}
}
