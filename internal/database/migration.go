package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"github.com/rs/zerolog"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

// Migrate applies all pending migrations. Tern takes an advisory lock, so
// running it from several places at once is safe.
func Migrate(ctx context.Context, pool *pgxpool.Pool, log zerolog.Logger) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	migrator, err := migrate.NewMigrator(ctx, conn.Conn(), "schema_version")
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	migrationsFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("failed to access migrations directory: %w", err)
	}
	if err := migrator.LoadMigrations(migrationsFS); err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	migrator.OnStart = func(seq int32, name, direction, _ string) {
		log.Info().Int32("seq", seq).Str("name", name).Str("direction", direction).Msg("applying migration")
	}

	if err := migrator.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		return fmt.Errorf("failed to read schema version: %w", err)
	}
	log.Info().Int32("version", version).Msg("database schema up to date")
	return nil
}
