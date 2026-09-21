package httptransport

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
)

func NewRouter(frontendOrigin string, authService *auth.Service) http.Handler {
	router := mux.NewRouter()
	router.Use(cors(frontendOrigin))
	router.HandleFunc("/health", health).Methods(http.MethodGet)
	handler := authHandler{service: authService}
	router.HandleFunc("/auth/register", handler.register).Methods(http.MethodPost)
	router.HandleFunc("/auth/login", handler.login).Methods(http.MethodPost)
	router.HandleFunc("/auth/logout", handler.logout).Methods(http.MethodPost)
	router.HandleFunc("/auth/me", handler.me).Methods(http.MethodGet)

	return router
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func cors(origin string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
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
