package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/joaopsramos/fincon/internal/errs"
)

func (a *App) GetUserIDFromCtx(r *http.Request) uuid.UUID {
	return r.Context().Value(UserIDKey).(uuid.UUID)
}

func (a *App) HandleError(w http.ResponseWriter, err error) {
	if errors.Is(err, errs.ErrNotFound{}) {
		a.sendError(w, http.StatusNotFound, err.Error())
		return
	}

	if errors.Is(err, errs.ErrValidation{}) {
		a.sendError(w, http.StatusBadRequest, err.Error())
		return
	}

	panic(err)
}

func (a *App) HandleZodError(w http.ResponseWriter, err map[string]any) {
	a.sendJSON(w, http.StatusBadRequest, err)
}

func (a *App) InvalidJSONBody(w http.ResponseWriter, err error) {
	if errors.Is(err, &json.InvalidUnmarshalError{}) {
		panic(err)
	}

	a.sendError(w, http.StatusBadRequest, "invalid json body")
}
