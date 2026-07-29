package pipeline

import (
	"encoding/json"
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/pipelines", h.List)
	mux.HandleFunc("POST /api/v1/pipelines", h.Create)
	mux.HandleFunc("GET /api/v1/pipelines/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/pipelines/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/pipelines/{id}", h.Delete)
}

type CreatePipelineInput struct {
	WorkspaceID string `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type UpdatePipelineInput struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Status      string          `json:"status"`
	Config      json.RawMessage `json:"config,omitempty"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		response.Error(w, http.StatusBadRequest, "workspace_id query param is required")
		return
	}
	pipelines, err := h.svc.ListByWorkspace(r.Context(), workspaceID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, pipelines)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreatePipelineInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.WorkspaceID == "" || input.Name == "" {
		response.Error(w, http.StatusBadRequest, "workspace_id and name are required")
		return
	}
	p, err := h.svc.Create(r.Context(), input.WorkspaceID, input.Name, input.Description)
	if err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, p)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, err := h.svc.ByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if p == nil {
		response.Error(w, http.StatusNotFound, "pipeline not found")
		return
	}
	response.JSON(w, http.StatusOK, p)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input UpdatePipelineInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p, err := h.svc.Update(r.Context(), id, input.Name, input.Description, input.Status, input.Config)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, p)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
