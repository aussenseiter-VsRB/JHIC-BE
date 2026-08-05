package spmb

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/analytics"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/domain/nexxa"
	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

type Handler struct {
	svc       *Service
	limit     func(http.Handler) http.Handler
	analytics *analytics.Service
}

func NewHandler(svc *Service, limit func(http.Handler) http.Handler, tracking ...*analytics.Service) *Handler {
	var a *analytics.Service
	if len(tracking) > 0 {
		a = tracking[0]
	}
	return &Handler{svc: svc, limit: limit, analytics: a}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/nexxa/spmb/parse-kk", h.limit(http.HandlerFunc(h.ParseKk)))
	mux.Handle("POST /api/v1/nexxa/spmb/ask", h.limit(http.HandlerFunc(h.Ask)))
}

func (h *Handler) ParseKk(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, KkFileMaxSize+(64<<10))
	if err := r.ParseMultipartForm(KkFileMaxSize + (64 << 10)); err != nil {
		response.Error(w, http.StatusBadRequest, "file too large or invalid multipart form")
		return
	}

	childName := r.FormValue("child_name")
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	if header.Size > KkFileMaxSize {
		response.Error(w, http.StatusBadRequest, ErrKkTooLarge.Error())
		return
	}

	raw, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to read file")
		return
	}
	mime := detectKkType(raw, header.Header.Get("Content-Type"))
	if !allowedKkType(mime) {
		response.Error(w, http.StatusBadRequest, "unsupported file type: jpeg, png, webp or pdf allowed")
		return
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	if h.analytics != nil {
		h.analytics.Record(r.Context(), "spmb.parse_kk", "", nil, map[string]any{
			"file_size": header.Size,
			"mime":      mime,
			"success":   true,
		})
	}

	res, err := h.svc.ParseKk(r.Context(), encoded, mime, childName)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	log.Printf("spmb parse-kk: ok size=%d sha=%s child=%s", len(raw), shortSHA(string(raw)), sanitizeForLog(childName))
	response.JSON(w, http.StatusOK, map[string]any{"data": res})
}

func (h *Handler) Ask(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var input AskRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Ask(r.Context(), input.Question, input.SessionID)
	if h.analytics != nil {
		h.analytics.Record(r.Context(), "spmb.ask", input.SessionID, nil, map[string]any{
			"question_length": len(input.Question),
			"success":         err == nil,
		})
	}
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrKkFileRequired), errors.Is(err, ErrKkTooLarge),
		errors.Is(err, ErrChildNameRequired), errors.Is(err, ErrChildNameTooLong),
		errors.Is(err, ErrQuestionRequired), errors.Is(err, ErrQuestionTooLong):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, nexxa.ErrN8NTimeout):
		response.Error(w, http.StatusGatewayTimeout, err.Error())
	case errors.Is(err, nexxa.ErrN8NUnavailable):
		response.Error(w, http.StatusBadGateway, err.Error())
	case errors.Is(err, ErrOutputInvalid):
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func allowedKkType(mime string) bool {
	switch strings.ToLower(mime) {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
		return true
	}
	return false
}

func detectKkType(data []byte, declared string) string {
	sniff := http.DetectContentType(data)
	switch sniff {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
		return sniff
	}
	switch strings.ToLower(declared) {
	case "image/jpeg", "image/png", "image/webp", "application/pdf":
		return declared
	}
	return sniff
}

func sanitizeForLog(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 60 {
		return s[:60] + "..."
	}
	return s
}

func shortSHA(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:4])
}
