package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/rs/zerolog"

	"github.com/Rengoku1926/gpu-inference/internal/errs"
	"github.com/Rengoku1926/gpu-inference/internal/model"
	"github.com/Rengoku1926/gpu-inference/internal/repository"
)

// Lowercase, digits and dashes, so names are safe in URLs and later as k8s resource names.
var validModelName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

type ModelService struct {
	repo   *repository.ModelRepository
	logger zerolog.Logger
}

func NewModelService(repo *repository.ModelRepository, logger zerolog.Logger) *ModelService {
	return &ModelService{
		repo:   repo,
		logger: logger.With().Str("component", "model_service").Logger(),
	}
}

func (s *ModelService) RegisterModel(ctx context.Context, in model.CreateModelInput) (*model.Model, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Image = strings.TrimSpace(in.Image)
	in.HFRepo = strings.TrimSpace(in.HFRepo)

	switch {
	case !validModelName.MatchString(in.Name):
		return nil, fmt.Errorf("%w: name must be 1-63 chars of lowercase letters, digits and dashes", errs.ErrValidation)
	case in.Image == "":
		return nil, fmt.Errorf("%w: image is required", errs.ErrValidation)
	case in.HFRepo == "":
		return nil, fmt.Errorf("%w: hf_repo is required", errs.ErrValidation)
	case in.DesiredReplicas < 0:
		return nil, fmt.Errorf("%w: desired_replicas must be >= 0", errs.ErrValidation)
	}

	m, err := s.repo.Create(ctx, in)
	if err != nil {
		return nil, err
	}

	s.logger.Info().
		Int64("model_id", m.ID).
		Str("name", m.Name).
		Int("desired_replicas", m.DesiredReplicas).
		Msg("model registered")
	return m, nil
}

func (s *ModelService) GetModel(ctx context.Context, name string) (*model.Model, error) {
	return s.repo.GetByName(ctx, name)
}

func (s *ModelService) ListModels(ctx context.Context) ([]*model.Model, error) {
	return s.repo.List(ctx)
}
