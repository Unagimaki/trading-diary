package httptransport

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/column"
	"github.com/trade-diary/backend/internal/domain/observation"
	"net/http"
)

type columnHandler struct {
	auth    *auth.Service
	columns *column.Service
	rows    *observation.Service
}
type columnInput struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Role     *string  `json:"role"`
	Options  []string `json:"options"`
	Position *int     `json:"position"`
}

func (h columnHandler) list(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	items, err := h.columns.List(r.Context(), u.ID, mux.Vars(r)["journalID"])
	if err != nil {
		internalError(w, r, "column.list", err, "user_id", u.ID, "journal_id", mux.Vars(r)["journalID"])
		return
	}
	writeJSON(w, 200, items)
}
func (h columnHandler) create(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var in columnInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.columns.Create(r.Context(), u.ID, mux.Vars(r)["journalID"], column.Values{Name: in.Name, Type: in.Type, Role: in.Role, Options: in.Options})
	if err == nil {
		err = h.rows.RecalculateJournal(r.Context(), u.ID, mux.Vars(r)["journalID"])
	}
	h.respond(w, r, item, err, 201)
}
func (h columnHandler) update(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var in columnInput
	if !decode(w, r, &in) {
		return
	}
	vars := mux.Vars(r)
	item, err := h.columns.Update(r.Context(), u.ID, vars["journalID"], vars["id"], column.Values{Name: in.Name, Type: in.Type, Role: in.Role, Options: in.Options, Position: in.Position})
	if err == nil {
		err = h.rows.RecalculateJournal(r.Context(), u.ID, vars["journalID"])
	}
	h.respond(w, r, item, err, 200)
}
func (h columnHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	vars := mux.Vars(r)
	err := h.columns.Delete(r.Context(), u.ID, vars["journalID"], vars["id"])
	if errors.Is(err, column.ErrNotFound) {
		writeError(w, 404, "column not found")
		return
	}
	if err != nil {
		internalError(w, r, "column.delete", err, "user_id", u.ID, "journal_id", vars["journalID"], "column_id", vars["id"])
		return
	}
	if err = h.rows.RecalculateJournal(r.Context(), u.ID, vars["journalID"]); err != nil {
		internalError(w, r, "column.recalculate", err, "user_id", u.ID, "journal_id", vars["journalID"])
		return
	}
	w.WriteHeader(204)
}
func (h columnHandler) respond(w http.ResponseWriter, r *http.Request, item column.Column, err error, status int) {
	if errors.Is(err, column.ErrInvalid) {
		writeError(w, 400, "invalid column")
		return
	}
	if errors.Is(err, column.ErrNotFound) {
		writeError(w, 404, "column not found")
		return
	}
	if err != nil {
		internalError(w, r, "column.write", err, "journal_id", mux.Vars(r)["journalID"], "column_id", mux.Vars(r)["id"])
		return
	}
	writeJSON(w, status, item)
}
