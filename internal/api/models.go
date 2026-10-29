package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"

	"github.com/Rengoku1926/gpu-inference/internal/store"
)

// Lowercase, digits and dashes, so names are safe in URLs and later as k8s resource names.
var validName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

type createModelRequest struct {
	Name            string `json:"name"`
	Image           string `json:"image"`
	HFRepo          string `json:"hf_repo"`
	DesiredReplicas int    `json:"desired_replicas"`
}

func (req createModelRequest) validate() error {
	switch {
	case !validName.MatchString(req.Name):
		return errors.New("name must be 1-63 chars of lowercase letters, digits and dashes")
	case req.Image == "":
		return errors.New("image is required")
	case req.HFRepo == "":
		return errors.New("hf_repo is required")
	case req.DesiredReplicas < 0:
		return errors.New("desired_replicas must be >= 0")
	}
	return nil
}

func (s *Server) createModel(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	var req createModelRequest
	if err := dec.Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := req.validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	m, err := s.store.CreateModel(r.Context(), store.Model{
		Name:            req.Name,
		Image:           req.Image,
		HFRepo:          req.HFRepo,
		DesiredReplicas: req.DesiredReplicas,
	})
	switch {
	case errors.Is(err, store.ErrConflict):
		writeError(w, http.StatusConflict, "model "+req.Name+" already exists")
	case err != nil:
		s.log.Error("create model", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusCreated, m)
	}
}

func (s *Server) listModels(w http.ResponseWriter, r *http.Request) {
	models, err := s.store.ListModels(r.Context())
	if err != nil {
		s.log.Error("list models", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
		return
	}
	writeJSON(w, http.StatusOK, models)
}

func (s *Server) getModel(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	m, err := s.store.GetModel(r.Context(), name)
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "model "+name+" not found")
	case err != nil:
		s.log.Error("get model", "err", err)
		writeError(w, http.StatusInternalServerError, "internal error")
	default:
		writeJSON(w, http.StatusOK, m)
	}
}
