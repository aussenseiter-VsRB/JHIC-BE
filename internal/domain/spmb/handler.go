package spmb

import (
	"encoding/json"
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/pkg/id"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux, authMw func(http.Handler) http.Handler, roleCheck middleware.RoleChecker) {
	adminMw := func(next http.Handler) http.Handler {
		return authMw(middleware.RequireRole("admin")(roleCheck)(next))
	}

	mux.Handle("POST /api/v1/spmb", http.HandlerFunc(h.Create))
	mux.Handle("GET /api/v1/spmb", adminMw(http.HandlerFunc(h.List)))
	mux.Handle("GET /api/v1/spmb/{id}", adminMw(http.HandlerFunc(h.Get)))
	mux.Handle("POST /api/v1/spmb/{id}/status", adminMw(http.HandlerFunc(h.SetStatus)))
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	var input SpmbRegistration
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reg, err := h.svc.Create(r.Context(), &input)
	if err != nil {
		switch err.Error() {
		case "nama is required", "nik must be 16 digits", "jenis_kelamin is required",
			"alamat is required", "jurusan is required", "invalid jurusan":
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusCreated, reg)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := id.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	reg, err := h.svc.Get(r.Context(), id)
	if err != nil {
		switch err.Error() {
		case "registration not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusOK, reg)
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	id, err := id.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var input struct {
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	reg, err := h.svc.SetStatus(r.Context(), id, input.Status, input.Reason)
	if err != nil {
		switch err.Error() {
		case "registration not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "invalid status: must be approve or cancel", "cancellation reason is required":
			response.Error(w, http.StatusBadRequest, err.Error())
		case "only proses registrations can be approved", "only proses or approve registrations can be cancelled":
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusOK, reg)
}
