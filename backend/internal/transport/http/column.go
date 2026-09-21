package httptransport

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/column"
	"net/http"
)

type columnHandler struct {
	auth    *auth.Service
	columns *column.Service
}
type columnInput struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Role     *string  `json:"role"`
	Options  []string `json:"options"`
	Position *int     `json:"position"`
}

func (h columnHandler) user(r *http.Request) (auth.User, error) {
	token, ok := cookieHash(r)
	if !ok {
		return auth.User{}, auth.ErrInvalidCredentials
	}
	return h.auth.CurrentUser(r.Context(), token)
}
func (h columnHandler) list(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	items, err := h.columns.List(r.Context(), u.ID, mux.Vars(r)["journalID"])
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 200, items)
}
func (h columnHandler) create(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	var in columnInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.columns.Create(r.Context(), u.ID, mux.Vars(r)["journalID"], column.Values{Name: in.Name, Type: in.Type, Role: in.Role, Options: in.Options})
	h.respond(w, item, err, 201)
}
func (h columnHandler) update(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	var in columnInput
	if !decode(w, r, &in) {
		return
	}
	vars := mux.Vars(r)
	item, err := h.columns.Update(r.Context(), u.ID, vars["journalID"], vars["id"], column.Values{Name: in.Name, Type: in.Type, Role: in.Role, Options: in.Options, Position: in.Position})
	h.respond(w, item, err, 200)
}
func (h columnHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	vars := mux.Vars(r)
	err = h.columns.Delete(r.Context(), u.ID, vars["journalID"], vars["id"])
	if errors.Is(err, column.ErrNotFound) {
		writeError(w, 404, "column not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	w.WriteHeader(204)
}
func (h columnHandler) respond(w http.ResponseWriter, item column.Column, err error, status int) {
	if errors.Is(err, column.ErrInvalid) {
		writeError(w, 400, "invalid column")
		return
	}
	if errors.Is(err, column.ErrNotFound) {
		writeError(w, 404, "column not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, status, item)
}
