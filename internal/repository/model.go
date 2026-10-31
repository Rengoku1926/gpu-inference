package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Rengoku1926/gpu-inference/internal/errs"
	"github.com/Rengoku1926/gpu-inference/internal/model"
)

type ModelRepository struct {
	pool *pgxpool.Pool
}

func NewModelRepository(pool *pgxpool.Pool) *ModelRepository {
	return &ModelRepository{pool: pool}
}

const modelColumns = `id, name, image, hf_repo, desired_replicas, created_at, updated_at`

func scanModel(row pgx.Row) (*model.Model, error) {
	m := &model.Model{}
	err := row.Scan(&m.ID, &m.Name, &m.Image, &m.HFRepo, &m.DesiredReplicas, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *ModelRepository) Create(ctx context.Context, in model.CreateModelInput) (*model.Model, error) {
	query := `INSERT INTO models (name, image, hf_repo, desired_replicas)
	          VALUES ($1, $2, $3, $4)
	          RETURNING ` + modelColumns

	m, err := scanModel(r.pool.QueryRow(ctx, query, in.Name, in.Image, in.HFRepo, in.DesiredReplicas))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, fmt.Errorf("%w: model %q", errs.ErrDuplicate, in.Name)
		}
		return nil, fmt.Errorf("insert model %q: %w", in.Name, err)
	}
	return m, nil
}

func (r *ModelRepository) GetByName(ctx context.Context, name string) (*model.Model, error) {
	query := `SELECT ` + modelColumns + ` FROM models WHERE name = $1`

	m, err := scanModel(r.pool.QueryRow(ctx, query, name))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%w: model %q", errs.ErrNotFound, name)
		}
		return nil, fmt.Errorf("get model %q: %w", name, err)
	}
	return m, nil
}

func (r *ModelRepository) List(ctx context.Context) ([]*model.Model, error) {
	rows, err := r.pool.Query(ctx, `SELECT `+modelColumns+` FROM models ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	defer rows.Close()

	models := []*model.Model{}
	for rows.Next() {
		m, err := scanModel(rows)
		if err != nil {
			return nil, fmt.Errorf("scan model: %w", err)
		}
		models = append(models, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return models, nil
}
