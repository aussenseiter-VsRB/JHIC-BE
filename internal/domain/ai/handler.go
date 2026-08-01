package ai

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aussenseiter-VsRB/JHIC-BE/internal/infrastructure/response"
)

type Handler struct {
	svc   *Service
	limit func(http.Handler) http.Handler
}

func NewHandler(svc *Service, limit func(http.Handler) http.Handler) *Handler {
	return &Handler{svc: svc, limit: limit}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/ai/chat", h.limit(http.HandlerFunc(h.Chat)))
	mux.Handle("POST /api/v1/ai/nexxa-match", h.limit(http.HandlerFunc(h.NexxaMatch)))
}

func (h *Handler) Chat(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var input ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Chat(r.Context(), input.ChatInput, input.SessionID)
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) NexxaMatch(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var input NexxaRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.NexxaMatch(r.Context(), input.Answers())
	if err != nil {
		h.writeServiceError(w, err)
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

func (h *Handler) writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrChatMessageRequired), errors.Is(err, ErrChatMessageTooLong),
		errors.Is(err, ErrAnswersRequired), errors.Is(err, ErrAnswerTooLong):
		response.Error(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, ErrN8NTimeout):
		response.Error(w, http.StatusGatewayTimeout, err.Error())
	case errors.Is(err, ErrN8NUnavailable):
		response.Error(w, http.StatusBadGateway, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}
