package api

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/errs"
	"github.com/joaopsramos/fincon/internal/util"
)

type BaseHandler struct {
	logger *slog.Logger
}

func NewBaseHandler(logger *slog.Logger) *BaseHandler {
	return &BaseHandler{logger: logger}
}

func (h *BaseHandler) getUserIDFromCtx(r *http.Request) uuid.UUID {
	return r.Context().Value(UserIDKey).(uuid.UUID)
}

func (h *BaseHandler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		if h.logger != nil {
			h.logger.Error("Failed to encode response", "error", err)
		}
	}
}

func (h *BaseHandler) sendError(w http.ResponseWriter, status int, message string) {
	h.sendJSON(w, status, util.M{"error": message})
}

func (h *BaseHandler) HandleError(w http.ResponseWriter, err error) {
	if errors.Is(err, errs.ErrNotFound{}) {
		h.sendError(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, errs.ErrValidation{}) {
		h.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	panic(err)
}

func (h *BaseHandler) HandleZodError(w http.ResponseWriter, err util.M) {
	h.sendJSON(w, http.StatusBadRequest, err)
}

func (h *BaseHandler) rateLimiter(limit int, windowLength time.Duration) func(http.Handler) http.Handler {
	rateLimiter := httprate.NewRateLimiter(
		limit,
		windowLength,
		httprate.WithErrorHandler(func(w http.ResponseWriter, r *http.Request, err error) {
			h.sendError(w, http.StatusTooManyRequests, "too many requests")
		}),
	)

	return rateLimiter.Handler
}

func (h *BaseHandler) InvalidJSONBody(w http.ResponseWriter, err error) {
	h.sendError(w, http.StatusBadRequest, "invalid json body")
}
