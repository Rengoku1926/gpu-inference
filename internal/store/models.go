package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Model struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Image           string    `json:"image"`
	HFRepo          string    `json:"hf_repo"`
	DesiredReplicas int       `json:"desired_replicas"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

const modelColumns = `id, name, image, hf_repo, desired_replicas, created_at, updated_at`

func scanModel(row pgx.Row) (Model, error) {
	var m Model
	err := row.Scan(&m.ID, &m.Name, &m.Image, &m.HFRepo, &m.DesiredReplicas, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (s *Store) CreateModel(ctx context.Context, m Model) (Model, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO models (name, image, hf_repo, desired_replicas)
		VALUES ($1, $2, $3, $4)
		RETURNING `+modelColumns,
		m.Name, m.Image, m.HFRepo, m.DesiredReplicas)
	created, err := scanModel(row)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
		return Model{}, ErrConflict
	}
	return created, err
}

func (s *Store) GetModel(ctx context.Context, name string) (Model, error) {
	row := s.pool.QueryRow(ctx, `SELECT `+modelColumns+` FROM models WHERE name = $1`, name)
	m, err := scanModel(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return Model{}, ErrNotFound
	}
	return m, err
}

func (s *Store) ListModels(ctx context.Context) ([]Model, error) {
	rows, err := s.pool.Query(ctx, `SELECT `+modelColumns+` FROM models ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	models := []Model{}
	for rows.Next() {
		m, err := scanModel(rows)
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, rows.Err()
}
