package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/Rengoku1926/gpu-inference/internal/config"
)

func ConnectRedis(ctx context.Context, cfg config.RedisConfig, log zerolog.Logger) (*redis.Client, error) {
	redis.SetLogger(redisLogger{log: log.With().Str("component", "redis").Logger()})

	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Address,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to ping redis at %s: %w", cfg.Address, err)
	}

	return client, nil
}

// redisLogger sends go-redis's internal messages (dial failures, pool
// errors) through zerolog instead of the standard library logger.
type redisLogger struct {
	log zerolog.Logger
}

func (l redisLogger) Printf(_ context.Context, format string, v ...any) {
	l.log.Warn().Msg(strings.TrimSpace(fmt.Sprintf(format, v...)))
}
