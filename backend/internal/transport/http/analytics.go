package httptransport

import (
	"errors"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/analytics"
	"github.com/trade-diary/backend/internal/domain/auth"
)

type analyticsHandler struct {
	auth      *auth.Service
	analytics *analytics.Service
}

func (h analyticsHandler) get(w http.ResponseWriter, r *http.Request) {
	user, ok := authenticatedUser(w, r, h.auth)
	if !ok {
		return
	}
	journalID := mux.Vars(r)["journalID"]
	report, err := h.analytics.Get(r.Context(), user.ID, journalID)
	if errors.Is(err, analytics.ErrNotFound) {
		writeError(w, http.StatusNotFound, "journal not found")
		return
	}
	if err != nil {
		internalError(w, r, "analytics.get", err, "user_id", user.ID, "journal_id", journalID)
		return
	}
	writeJSON(w, http.StatusOK, report)
}
