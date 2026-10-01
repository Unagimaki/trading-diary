package httptransport

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), "http://localhost:5173", "development", nil, nil, nil, nil, nil, nil).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if response.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("unexpected body: %s", response.Body.String())
	}
}

func TestCredentialedCORS(t *testing.T) {
	for _, origin := range []string{"https://trading-diary-frontend.vercel.app", "*"} {
		response := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodOptions, "/auth/login", nil)
		request.Header.Set("Origin", "https://trading-diary-frontend.vercel.app")
		cors(origin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
			t.Fatal("preflight reached handler")
		})).ServeHTTP(response, request)
		if response.Code != http.StatusNoContent {
			t.Fatalf("unexpected preflight status: %d", response.Code)
		}
		if origin == "*" {
			if response.Header().Get("Access-Control-Allow-Origin") != "" || response.Header().Get("Access-Control-Allow-Credentials") != "" {
				t.Fatal("wildcard must not enable credentialed CORS")
			}
		} else if response.Header().Get("Access-Control-Allow-Origin") != origin || response.Header().Get("Access-Control-Allow-Credentials") != "true" {
			t.Fatal("missing credentialed CORS headers")
		}
	}
}
