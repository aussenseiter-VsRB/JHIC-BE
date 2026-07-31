package pkl

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/middleware"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(mux *http.ServeMux, authMw func(http.Handler) http.Handler, roleCheck middleware.RoleChecker) {
	userMw := roleMW(authMw, roleCheck, "user")
	guruMw := roleMW(authMw, roleCheck, "guru")
	viewerMw := roleMW(authMw, roleCheck, "user", "guru", "admin")

	mux.Handle("POST /api/v1/approval/pkl", userMw(http.HandlerFunc(h.Create)))
	mux.Handle("GET /api/v1/approval/pkl", viewerMw(http.HandlerFunc(h.List)))
	mux.Handle("GET /api/v1/approval/pkl/{id}", viewerMw(http.HandlerFunc(h.Get)))
	mux.Handle("POST /api/v1/approval/pkl/{id}/decide", guruMw(http.HandlerFunc(h.Decide)))
	mux.Handle("DELETE /api/v1/approval/pkl/{id}", userMw(http.HandlerFunc(h.Cancel)))
}

func roleMW(authMw func(http.Handler) http.Handler, roleCheck middleware.RoleChecker, roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return authMw(middleware.RequireRole(roles...)(roleCheck)(next))
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)

	var input struct {
		Company     string `json:"company"`
		Location    string `json:"location"`
		StartDate   string `json:"start_date"`
		EndDate     string `json:"end_date"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Company == "" || input.Location == "" || input.Description == "" {
		response.Error(w, http.StatusBadRequest, "company, location and description are required")
		return
	}
	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "start_date must be in YYYY-MM-DD format")
		return
	}
	endDate, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "end_date must be in YYYY-MM-DD format")
		return
	}

	req, err := h.svc.Create(r.Context(), userID, input.Company, input.Location, startDate, endDate, input.Description)
	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "no "):
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		case err.Error() == "requester not found":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		case err.Error() == "end date must be on or after start date":
			response.Error(w, http.StatusBadRequest, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusCreated, req)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	role := r.Context().Value(middleware.RoleKey).(string)

	list, err := h.svc.List(r.Context(), userID, role)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, list)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	role := r.Context().Value(middleware.RoleKey).(string)
	id := r.PathValue("id")

	req, err := h.svc.Get(r.Context(), userID, role, id)
	if err != nil {
		switch err.Error() {
		case "pkl request not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "forbidden: not the requester", "forbidden: not an approver on this request":
			response.Error(w, http.StatusForbidden, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusOK, req)
}

func (h *Handler) Decide(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	var input struct {
		Decision string `json:"decision"`
		Note     string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req, err := h.svc.Decide(r.Context(), userID, id, input.Decision, input.Note)
	if err != nil {
		switch err.Error() {
		case "pkl request not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "forbidden: not your step to decide":
			response.Error(w, http.StatusForbidden, err.Error())
		case "invalid decision: must be approve, reject or needs_further_action":
			response.Error(w, http.StatusBadRequest, err.Error())
		case "step already decided", "request already needs further action",
			"request is not awaiting a decision", "approval step already decided", "pkl request status changed":
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusOK, req)
}

func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	var input struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req, err := h.svc.Cancel(r.Context(), userID, id, input.Reason)
	if err != nil {
		switch err.Error() {
		case "pkl request not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "forbidden: not the requester":
			response.Error(w, http.StatusForbidden, err.Error())
		case "cancellation reason is required":
			response.Error(w, http.StatusBadRequest, err.Error())
		case "only pending or needs_further_action requests can be cancelled":
			response.Error(w, http.StatusConflict, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusOK, req)
}
