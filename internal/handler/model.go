package handler

import (
	"net/http"

	"github.com/Rengoku1926/gpu-inference/internal/model"
	"github.com/Rengoku1926/gpu-inference/internal/service"
)

type ModelHandler struct {
	modelService *service.ModelService
}

func NewModelHandler(modelService *service.ModelService) *ModelHandler {
	return &ModelHandler{modelService: modelService}
}

func (h *ModelHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req model.CreateModelInput
	if err := decodeJSON(w, r, &req); err != nil {
		writeError(w, r, err)
		return
	}

	m, err := h.modelService.RegisterModel(r.Context(), req)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusCreated, m)
}

func (h *ModelHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	models, err := h.modelService.ListModels(r.Context())
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, models)
}

func (h *ModelHandler) HandleGetByName(w http.ResponseWriter, r *http.Request) {
	m, err := h.modelService.GetModel(r.Context(), r.PathValue("name"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
