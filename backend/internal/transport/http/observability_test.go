package httptransport

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestObservabilityAddsRequestID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/test", nil)
	response := httptest.NewRecorder()
	handler := observability(slog.New(slog.NewJSONHandler(&bytes.Buffer{}, nil)), &requestMetrics{})(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { writeError(w, http.StatusBadRequest, "invalid request") }))
	handler.ServeHTTP(response, request)
	var body errorResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.RequestID == "" || body.RequestID != response.Header().Get("X-Request-ID") {
		t.Fatalf("request id was not propagated: %#v", body)
	}
	if body.Code != "INVALID_REQUEST" {
		t.Fatalf("unexpected code %q", body.Code)
	}
}

func TestObservabilityRecoversPanic(t *testing.T) {
	var logs bytes.Buffer
	handler := observability(slog.New(slog.NewJSONHandler(&logs, nil)), &requestMetrics{})(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("boom") }))
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", response.Code)
	}
	if !strings.Contains(logs.String(), "request panic") || !strings.Contains(logs.String(), "request_id") {
		t.Fatalf("panic was not logged: %s", logs.String())
	}
}
