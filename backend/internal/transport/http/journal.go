package httptransport

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/journal"
)

type journalHandler struct {
	auth     *auth.Service
	journals *journal.Service
}
type journalInput struct {
	Name string `json:"name"`
}

func (h journalHandler) user(r *http.Request) (auth.User, error) {
	token, ok := cookieHash(r)
	if !ok {
		return auth.User{}, auth.ErrInvalidCredentials
	}
	return h.auth.CurrentUser(r.Context(), token)
}
func (h journalHandler) list(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	items, err := h.journals.List(r.Context(), u.ID)
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, 200, items)
}
func (h journalHandler) get(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	item, err := h.journals.Get(r.Context(), u.ID, mux.Vars(r)["id"])
	h.respond(w, item, err, 200)
}
func (h journalHandler) create(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	var in journalInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.journals.Create(r.Context(), u.ID, in.Name)
	h.respond(w, item, err, 201)
}
func (h journalHandler) rename(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	var in journalInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.journals.Rename(r.Context(), u.ID, mux.Vars(r)["id"], in.Name)
	h.respond(w, item, err, 200)
}
func (h journalHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, err := h.user(r)
	if err != nil {
		writeError(w, 401, "authentication required")
		return
	}
	err = h.journals.Delete(r.Context(), u.ID, mux.Vars(r)["id"])
	if errors.Is(err, journal.ErrNotFound) {
		writeError(w, 404, "journal not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	w.WriteHeader(204)
}
func (h journalHandler) respond(w http.ResponseWriter, item journal.Journal, err error, status int) {
	if errors.Is(err, journal.ErrInvalidName) {
		writeError(w, 400, "invalid journal name")
		return
	}
	if errors.Is(err, journal.ErrNotFound) {
		writeError(w, 404, "journal not found")
		return
	}
	if err != nil {
		writeError(w, 500, "internal error")
		return
	}
	writeJSON(w, status, item)
}
