package httptransport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/analytics"
	"github.com/trade-diary/backend/internal/domain/attachment"
	"github.com/trade-diary/backend/internal/domain/auth"
	"github.com/trade-diary/backend/internal/domain/column"
	"github.com/trade-diary/backend/internal/domain/journal"
	"github.com/trade-diary/backend/internal/domain/observation"
)

func NewRouter(logger *slog.Logger, frontendOrigin string, environment string, authService *auth.Service, journalService *journal.Service, columnService *column.Service, observationService *observation.Service, attachmentService *attachment.Service, analyticsService *analytics.Service) http.Handler {
	router := mux.NewRouter()
	metrics := &requestMetrics{}
	router.Use(observability(logger, metrics))
	router.HandleFunc("/health", health).Methods(http.MethodGet)
	router.Handle("/metrics", metrics).Methods(http.MethodGet)
	handler := authHandler{service: authService, production: environment == "production"}
	router.HandleFunc("/auth/register", handler.register).Methods(http.MethodPost)
	router.HandleFunc("/auth/login", handler.login).Methods(http.MethodPost)
	router.HandleFunc("/auth/logout", handler.logout).Methods(http.MethodPost)
	router.HandleFunc("/auth/me", handler.me).Methods(http.MethodGet)
	journals := journalHandler{auth: authService, journals: journalService, rows: observationService}
	router.HandleFunc("/journals", journals.list).Methods(http.MethodGet)
	router.HandleFunc("/journals", journals.create).Methods(http.MethodPost)
	router.HandleFunc("/journals/{id}", journals.get).Methods(http.MethodGet)
	router.HandleFunc("/journals/{id}", journals.rename).Methods(http.MethodPatch)
	router.HandleFunc("/journals/{id}/settings", journals.updateSettings).Methods(http.MethodPatch)
	router.HandleFunc("/journals/{id}", journals.delete).Methods(http.MethodDelete)
	columns := columnHandler{auth: authService, columns: columnService, rows: observationService}
	router.HandleFunc("/journals/{journalID}/columns", columns.list).Methods(http.MethodGet)
	router.HandleFunc("/journals/{journalID}/columns", columns.create).Methods(http.MethodPost)
	router.HandleFunc("/journals/{journalID}/columns/{id}", columns.update).Methods(http.MethodPatch)
	router.HandleFunc("/journals/{journalID}/columns/{id}", columns.delete).Methods(http.MethodDelete)
	rows := observationHandler{auth: authService, rows: observationService}
	router.HandleFunc("/journals/{journalID}/rows", rows.list).Methods(http.MethodGet)
	router.HandleFunc("/journals/{journalID}/rows", rows.create).Methods(http.MethodPost)
	router.HandleFunc("/journals/{journalID}/rows/{rowID}", rows.delete).Methods(http.MethodDelete)
	router.HandleFunc("/journals/{journalID}/rows/{rowID}/cells/{columnID}", rows.setCell).Methods(http.MethodPatch)
	router.HandleFunc("/journals/{journalID}/rows/{rowID}/cells/{columnID}", rows.clearCell).Methods(http.MethodDelete)
	images := attachmentHandler{auth: authService, attachments: attachmentService}
	router.HandleFunc("/journals/{journalID}/rows/{rowID}/cells/{columnID}/image", images.upload).Methods(http.MethodPost)
	router.HandleFunc("/journals/{journalID}/rows/{rowID}/cells/{columnID}/image", images.delete).Methods(http.MethodDelete)
	router.HandleFunc("/attachments/{id}/content", images.content).Methods(http.MethodGet)
	analyticsHandler := analyticsHandler{auth: authService, analytics: analyticsService}
	router.HandleFunc("/journals/{journalID}/analytics", analyticsHandler.get).Methods(http.MethodGet)

	return cors(frontendOrigin)(router)
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func cors(origin string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if origin != "" && origin != "*" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
