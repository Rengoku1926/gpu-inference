package config

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Load reads the env file at path if it exists, then builds the config from
// the environment. Real environment variables win over the file, so the same
// binary works locally with .env.local and in production with injected env.
func Load(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("failed to load env file %s: %w", path, err)
	}

	cfg := &Config{
		Primary: Primary{
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		Server: ServerConfig{
			Port:               getEnv("SERVER_PORT", "8080"),
			ReadTimeout:        getEnvInt("SERVER_READ_TIMEOUT", 15),
			WriteTimeout:       getEnvInt("SERVER_WRITE_TIMEOUT", 30),
			IdleTimeout:        getEnvInt("SERVER_IDLE_TIMEOUT", 60),
			ShutdownTimeout:    getEnvInt("SERVER_SHUTDOWN_TIMEOUT", 15),
			CORSAllowedOrigins: strings.Split(getEnv("SERVER_CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnvInt("DB_PORT", 5433),
			User:            getEnv("DB_USER", "gpu"),
			Password:        getEnv("DB_PASSWORD", ""),
			Name:            getEnv("DB_NAME", "gpu"),
			SSLMode:         getEnv("DB_SSL_MODE", "disable"),
			MaxConns:        getEnvInt("DB_MAX_CONNS", 25),
			MinConns:        getEnvInt("DB_MIN_CONNS", 2),
			ConnMaxLifetime: getEnvInt("DB_CONN_MAX_LIFETIME", 300),
			ConnMaxIdleTime: getEnvInt("DB_CONN_MAX_IDLE_TIME", 60),
		},
		Redis: RedisConfig{
			Address:  getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return cfg, nil
}

func (c *Config) IsProduction() bool { return c.Primary.Env == "production" }

func (c *Config) validate() error {
	var problems []string
	if c.Primary.Env != "development" && c.Primary.Env != "production" {
		problems = append(problems, "APP_ENV must be development or production")
	}
	if c.Server.Port == "" {
		problems = append(problems, "SERVER_PORT is required")
	}
	if c.IsProduction() && c.Database.Password == "" {
		problems = append(problems, "DB_PASSWORD is required in production")
	}
	if c.Database.MinConns > c.Database.MaxConns {
		problems = append(problems, "DB_MIN_CONNS cannot exceed DB_MAX_CONNS")
	}
	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func (d DatabaseConfig) DSN() string {
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(d.User, d.Password),
		Host:     fmt.Sprintf("%s:%d", d.Host, d.Port),
		Path:     d.Name,
		RawQuery: "sslmode=" + d.SSLMode,
	}
	return u.String()
}

func getEnv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
