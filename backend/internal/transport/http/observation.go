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

func (h observationHandler) list(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	items, err := h.rows.List(r.Context(), u.ID, mux.Vars(r)["journalID"])
	if err != nil {
		internalError(w, r, "row.list", err, "user_id", u.ID, "journal_id", mux.Vars(r)["journalID"])
		return
	}
	writeJSON(w, 200, items)
}
func (h observationHandler) create(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	item, err := h.rows.Create(r.Context(), u.ID, mux.Vars(r)["journalID"], r.URL.Query().Get("localDate"))
	if errors.Is(err, observation.ErrNotFound) {
		writeError(w, 404, "journal not found")
		return
	}
	if err != nil {
		internalError(w, r, "row.create", err, "user_id", u.ID, "journal_id", mux.Vars(r)["journalID"])
		return
	}
	writeJSON(w, 201, item)
}
func (h observationHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	v := mux.Vars(r)
	err := h.rows.Delete(r.Context(), u.ID, v["journalID"], v["rowID"])
	h.empty(w, r, err)
}
func (h observationHandler) setCell(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var in cellInput
	if !decode(w, r, &in) {
		return
	}
	v := mux.Vars(r)
	err := h.rows.SetCell(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"], in.Value)
	h.empty(w, r, err)
}
func (h observationHandler) clearCell(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	v := mux.Vars(r)
	err := h.rows.ClearCell(r.Context(), u.ID, v["journalID"], v["rowID"], v["columnID"])
	h.empty(w, r, err)
}
func (h observationHandler) empty(w http.ResponseWriter, r *http.Request, err error) {
	if errors.Is(err, observation.ErrInvalidValue) {
		writeError(w, 400, "invalid cell value")
		return
	}
	if errors.Is(err, observation.ErrNotFound) {
		writeError(w, 404, "row or column not found")
		return
	}
	if err != nil {
		internalError(w, r, "cell.write", err, "journal_id", mux.Vars(r)["journalID"], "row_id", mux.Vars(r)["rowID"], "column_id", mux.Vars(r)["columnID"])
		return
	}
	w.WriteHeader(204)
}
