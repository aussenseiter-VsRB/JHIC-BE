package user

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
	mux.HandleFunc("GET /api/v1/users", h.List)
	mux.HandleFunc("POST /api/v1/users", h.Create)
	mux.HandleFunc("GET /api/v1/users/{id}", h.Get)
	mux.HandleFunc("PUT /api/v1/users/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/users/{id}", h.Delete)
}

type CreateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type UpdateUserInput struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Email == "" || input.Name == "" {
		response.Error(w, http.StatusBadRequest, "email and name are required")
		return
	}
	user, err := h.svc.Create(r.Context(), input.Email, input.Name)
	if err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, user)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	user, err := h.svc.ByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	if user == nil {
		response.Error(w, http.StatusNotFound, "user not found")
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var input UpdateUserInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.Update(r.Context(), id, input.Email, input.Name)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
