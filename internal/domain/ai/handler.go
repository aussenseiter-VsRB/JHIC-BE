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
	mux.Handle("POST /api/v1/ai/nexxa-match/validate-input", http.HandlerFunc(h.ValidateNexxaInput))
	mux.Handle("POST /api/v1/ai/nexxa-match/normalize-output", http.HandlerFunc(h.NormalizeNexxaOutput))
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
	case errors.Is(err, ErrNexxaOutputInvalid):
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	case errors.Is(err, ErrN8NUnavailable):
		response.Error(w, http.StatusBadGateway, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, "internal server error")
	}
}

func (h *Handler) ValidateNexxaInput(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var raw map[string]json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		logNexxaInput(r, false, nil, nil)
		response.JSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"errors":  []APIError{{Message: "invalid request body"}},
		})
		return
	}

	data, errs := h.svc.ValidateNexxaInput(raw)
	if len(errs) > 0 {
		logNexxaInput(r, false, nil, errs)
		response.JSON(w, http.StatusBadRequest, map[string]any{
			"success": false,
			"errors":  errs,
		})
		return
	}

	logNexxaInput(r, true, data, nil)
	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}

func (h *Handler) NormalizeNexxaOutput(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 32<<10)
	var input NormalizeOutputRequest
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		logNexxaOutput(r, false, "", nil, nil)
		response.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"success": false,
			"errors":  []APIError{{Message: "invalid request body"}},
		})
		return
	}

	data, errs := h.svc.NormalizeNexxaOutput(input.Raw)
	if len(errs) > 0 {
		logNexxaOutput(r, false, input.Raw, nil, errs)
		response.JSON(w, http.StatusUnprocessableEntity, map[string]any{
			"success": false,
			"errors":  errs,
		})
		return
	}

	logNexxaOutput(r, true, input.Raw, data, nil)
	response.JSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    data,
	})
}
