package httptransport

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/journal"
	"github.com/trade-diary/backend/internal/domain/observation"
)

type journalHandler struct {
	auth     *auth.Service
	journals *journal.Service
	rows     *observation.Service
}
type journalInput struct {
	Name string `json:"name"`
}
type journalSettingsInput struct {
	InitialDeposit float64 `json:"initialDeposit"`
	RiskPercent    float64 `json:"riskPercent"`
	DefaultRR      float64 `json:"defaultRR"`
}

func (h journalHandler) list(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	items, err := h.journals.List(r.Context(), u.ID)
	if err != nil {
		internalError(w, r, "journal.list", err, "user_id", u.ID)
		return
	}
	writeJSON(w, 200, items)
}
func (h journalHandler) get(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	item, err := h.journals.Get(r.Context(), u.ID, mux.Vars(r)["id"])
	h.respond(w, r, item, err, 200)
}
func (h journalHandler) create(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var in journalInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.journals.Create(r.Context(), u.ID, in.Name)
	h.respond(w, r, item, err, 201)
}
func (h journalHandler) rename(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var in journalInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.journals.Rename(r.Context(), u.ID, mux.Vars(r)["id"], in.Name)
	h.respond(w, r, item, err, 200)
}
func (h journalHandler) updateSettings(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	var in journalSettingsInput
	if !decode(w, r, &in) {
		return
	}
	item, err := h.journals.UpdateSettings(r.Context(), u.ID, mux.Vars(r)["id"], in.InitialDeposit, in.RiskPercent, in.DefaultRR)
	if err == nil {
		err = h.rows.RecalculateJournal(r.Context(), u.ID, mux.Vars(r)["id"])
	}
	h.respond(w, r, item, err, http.StatusOK)
}
func (h journalHandler) delete(w http.ResponseWriter, r *http.Request) {
	u, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	err := h.journals.Delete(r.Context(), u.ID, mux.Vars(r)["id"])
	if errors.Is(err, journal.ErrNotFound) {
		writeError(w, 404, "journal not found")
		return
	}
	if err != nil {
		internalError(w, r, "journal.delete", err, "user_id", u.ID, "journal_id", mux.Vars(r)["id"])
		return
	}
	w.WriteHeader(204)
}
func (h journalHandler) respond(w http.ResponseWriter, r *http.Request, item journal.Journal, err error, status int) {
	if errors.Is(err, journal.ErrInvalidName) {
		writeError(w, 400, "invalid journal name")
		return
	}
	if errors.Is(err, journal.ErrInvalidSettings) {
		writeError(w, 400, "invalid journal settings")
		return
	}
	if errors.Is(err, journal.ErrNotFound) {
		writeError(w, 404, "journal not found")
		return
	}
	if err != nil {
		internalError(w, r, "journal.write", err, "journal_id", mux.Vars(r)["id"])
		return
	}
	writeJSON(w, status, item)
}
