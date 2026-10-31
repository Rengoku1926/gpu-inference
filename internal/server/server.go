package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/Rengoku1926/gpu-inference/internal/config"
)

type Server struct {
	cfg    config.ServerConfig
	logger zerolog.Logger
	http   *http.Server
}

func New(cfg config.ServerConfig, logger zerolog.Logger, handler http.Handler) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger.With().Str("component", "http_server").Logger(),
		http: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           handler,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout:      time.Duration(cfg.WriteTimeout) * time.Second,
			IdleTimeout:       time.Duration(cfg.IdleTimeout) * time.Second,
			ErrorLog:          newStdLogger(logger),
		},
	}
}

// Start serves until ctx is cancelled, then drains in flight requests for up
// to ShutdownTimeout before forcing connections closed.
func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.http.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.http.Addr, err)
	}

	serverErr := make(chan error, 1)
	go func() {
		s.logger.Info().Str("addr", s.http.Addr).Msg("server starting")
		serverErr <- s.http.Serve(ln)
	}()

	select {
	case err := <-serverErr:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	s.logger.Info().Msg("shutdown signal received, draining connections")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(s.cfg.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := s.http.Shutdown(shutdownCtx); err != nil {
		s.logger.Error().Err(err).Msg("graceful shutdown failed, forcing close")
		s.http.Close()
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	s.logger.Info().Msg("server shutdown gracefully")
	return nil
}
