package user

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

	mux.Handle("GET /api/v1/users", http.HandlerFunc(h.List))
	mux.Handle("GET /api/v1/users/{id}", http.HandlerFunc(h.Get))
	mux.Handle("POST /api/v1/users", adminMw(http.HandlerFunc(h.Create)))
	mux.Handle("PUT /api/v1/users/{id}", adminMw(http.HandlerFunc(h.Update)))
	mux.Handle("PUT /api/v1/users/{id}/role", adminMw(http.HandlerFunc(h.UpdateRole)))
	mux.Handle("DELETE /api/v1/users/{id}", adminMw(http.HandlerFunc(h.Delete)))
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.svc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, users)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := id.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
		Role     string `json:"role"`
		Class    string `json:"class"`
		Jurusan  string `json:"jurusan"`
		Position string `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	u, err := h.svc.Create(r.Context(), input.Email, input.Password, input.Name, input.Role, input.Class, input.Jurusan, input.Position)
	if err != nil {
		switch err.Error() {
		case "user with email " + input.Email + " already exists":
			response.Error(w, http.StatusConflict, err.Error())
		case "invalid role: must be one of [jurnal guru admin user]",
			"invalid position: must be one of [wali_kelas bk kesiswaan kaprog]",
			"position is only valid for role guru":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.JSON(w, http.StatusCreated, u)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := id.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var input struct {
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
		Class     string `json:"class"`
		Jurusan   string `json:"jurusan"`
		Position  string `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	user, err := h.svc.Update(r.Context(), id, input.Name, input.AvatarURL, input.Class, input.Jurusan, input.Position)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, user)
}

func (h *Handler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	id, err := id.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	var input struct {
		Role string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.UpdateRole(r.Context(), id, input.Role); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := id.Parse(r.PathValue("id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
