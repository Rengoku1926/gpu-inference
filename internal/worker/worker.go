package worker

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// Worker is a long running background loop: the job consumer, the
// reconciler, heartbeats. Run blocks until ctx is cancelled and returns nil
// on a clean stop.
type Worker interface {
	Name() string
	Run(ctx context.Context) error
}

type Runner struct {
	workers []Worker
	logger  zerolog.Logger
}

func NewRunner(logger zerolog.Logger, workers ...Worker) *Runner {
	return &Runner{workers: workers, logger: logger.With().Str("component", "worker_runner").Logger()}
}

// Run starts every worker and waits for all of them. If one fails, the rest
// are cancelled and the first error is returned.
func (r *Runner) Run(ctx context.Context) error {
	if len(r.workers) == 0 {
		r.logger.Debug().Msg("no workers registered")
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)
	for _, w := range r.workers {
		g.Go(func() error { return r.runOne(ctx, w) })
	}
	return g.Wait()
}

func (r *Runner) runOne(ctx context.Context, w Worker) (err error) {
	log := r.logger.With().Str("worker", w.Name()).Logger()

	defer func() {
		if rec := recover(); rec != nil {
			log.Error().Interface("panic", rec).Bytes("stack", debug.Stack()).Msg("worker panicked")
			err = fmt.Errorf("worker %s panicked: %v", w.Name(), rec)
		}
	}()

	log.Info().Msg("worker started")
	err = w.Run(log.WithContext(ctx))
	switch {
	case err == nil, errors.Is(err, context.Canceled):
		log.Info().Msg("worker stopped")
		return nil
	default:
		log.Error().Err(err).Msg("worker failed")
		return fmt.Errorf("worker %s: %w", w.Name(), err)
	}
}
