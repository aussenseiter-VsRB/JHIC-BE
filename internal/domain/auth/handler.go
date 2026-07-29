package auth

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
	mux.HandleFunc("POST /api/v1/auth/register", h.RegisterUser)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Name     string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Email == "" || input.Password == "" {
		response.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, token, err := h.svc.Register(r.Context(), input.Email, input.Password, input.Name)
	if err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if input.Email == "" || input.Password == "" {
		response.Error(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, token, err := h.svc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"user":  user,
		"token": token,
	})
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	session, err := h.svc.ValidateToken(r.Context(), token)
	if err != nil || session == nil {
		response.Error(w, http.StatusUnauthorized, "invalid or expired token")
		return
	}

	if err := h.svc.Logout(r.Context(), session.UserID); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to logout")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
