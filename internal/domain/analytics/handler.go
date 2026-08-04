package analytics

import (
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
	"net/http"
	"strconv"
	"time"
)

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo: repo} }
func (h *Handler) Register(mux *http.ServeMux, adminMw func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/analytics/summary", adminMw(http.HandlerFunc(h.Summary)))
}
func (h *Handler) RegisterBerita(mux *http.ServeMux, jurnalMw func(http.Handler) http.Handler) {
	mux.Handle("GET /api/v1/analytics/berita/summary", jurnalMw(http.HandlerFunc(h.BeritaSummary)))
}
func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	h.writeSummary(w, r, "")
}
func (h *Handler) BeritaSummary(w http.ResponseWriter, r *http.Request) {
	h.writeSummary(w, r, "berita.")
}
func (h *Handler) writeSummary(w http.ResponseWriter, r *http.Request, prefix string) {
	days, _ := strconv.Atoi(r.URL.Query().Get("days"))
	if days < 1 || days > 90 {
		days = 30
	}
	data, err := h.repo.Summary(r.Context(), time.Now().UTC().AddDate(0, 0, -days), prefix)
	if err != nil {
		response.Error(w, 500, "internal server error")
		return
	}
	response.JSON(w, 200, map[string]any{"days": days, "events": data})
}
