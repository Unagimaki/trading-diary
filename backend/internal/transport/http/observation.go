package httptransport

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/observation"
	"net/http"
)

type observationHandler struct {
	auth *auth.Service
	rows *observation.Service
}
type cellInput struct {
	Value json.RawMessage `json:"value"`
}

func (h observationHandler) user(r *http.Request) (auth.User, error) {
	token, ok := cookieHash(r)
	if !ok {
		return auth.User{}, auth.ErrInvalidCredentials
	}
	return h.auth.CurrentUser(r.Context(), token)
}
func (h observationHandler) list(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	items, err := h.rows.List(r.Context(), u.ID, mux.Vars(r)["journalID"])
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 200, items)
}
func (h observationHandler) create(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	item, err := h.rows.Create(r.Context(), u.ID, mux.Vars(r)["journalID"])
	if errors.Is(err, observation.ErrNotFound) {
		writeError(w, 404, "journal not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 201, item)
}
func (h observationHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	v := mux.Vars(r)
	err = h.rows.Delete(r.Context(), u.ID, v["journalID"], v["rowID"])
	h.empty(w, err)
}
func (h observationHandler) setCell(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	var in cellInput
	if !decode(w, r, &in) {
		return
	}
	v := mux.Vars(r)
	err = h.rows.SetCell(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"], in.Value)
	h.empty(w, err)
}
func (h observationHandler) clearCell(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	v := mux.Vars(r)
	err = h.rows.ClearCell(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"])
	h.empty(w, err)
}
func (h observationHandler) empty(w http.ResponseWriter, err error) {
	if errors.Is(err, observation.ErrInvalidValue) {
		writeError(w, 400, "invalid cell value")
		return
	}
	if errors.Is(err, observation.ErrNotFound) {
		writeError(w, 404, "row or column not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	w.WriteHeader(204)
}
