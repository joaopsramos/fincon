package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	errs "github.com/joaopsramos/fincon/internal/error"
)

func (a *Api) GetUserIDFromCtx(r *http.Request) uuid.UUID {
	return r.Context().Value(UserIDKey).(uuid.UUID)
}

func (a *Api) HandleError(w http.ResponseWriter, err error) {
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

func (a *Api) HandleZodError(w http.ResponseWriter, err map[string]any) {
	a.sendJSON(w, http.StatusBadRequest, err)
}

func (a *Api) InvalidJSONBody(w http.ResponseWriter, err error) {
	if errors.Is(err, &json.InvalidUnmarshalError{}) {
		panic(err)
	}

	a.sendError(w, http.StatusBadRequest, "invalid json body")
}
