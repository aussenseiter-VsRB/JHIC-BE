package workspace

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
	mux.HandleFunc("GET /api/v1/workspaces", h.List)
	mux.HandleFunc("POST /api/v1/workspaces", h.Create)
	mux.HandleFunc("GET /api/v1/workspaces/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/workspaces/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/workspaces/{id}", h.Delete)
}

type CreateWorkspaceInput struct {
	Name    string `json:"name"`
	OwnerID string `json:"owner_id"`
}

type UpdateWorkspaceInput struct {
	Name string `json:"name"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner_id")
	if ownerID == "" {
		response.Error(w, http.StatusBadRequest, "owner_id query param is required")
		return
	}
	workspaces, err := h.svc.ListByOwner(r.Context(), ownerID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, workspaces)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateWorkspaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Name == "" || input.OwnerID == "" {
		response.Error(w, http.StatusBadRequest, "name and owner_id are required")
		return
	}
	ws, err := h.svc.Create(r.Context(), input.Name, input.OwnerID)
	if err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, ws)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ws, err := h.svc.ByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if ws == nil {
		response.Error(w, http.StatusNotFound, "workspace not found")
		return
	}
	response.JSON(w, http.StatusOK, ws)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input UpdateWorkspaceInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ws, err := h.svc.Update(r.Context(), id, input.Name)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, ws)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
