package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/trade-diary/backend/internal/domain/auth"
)

type logAuthRepository struct{}

func (logAuthRepository) CreateUser(context.Context, string, string, string) (auth.User, error) {
	return auth.User{ID: "user-42"}, nil
}
func (logAuthRepository) UserByEmail(context.Context, string) (auth.User, string, error) {
	return auth.User{}, "", auth.ErrInvalidCredentials
}
func (logAuthRepository) CreateSession(context.Context, string, string, time.Time) error { return nil }
func (logAuthRepository) UserBySession(context.Context, string) (auth.User, error) {
	return auth.User{ID: "user-42"}, nil
}
func (logAuthRepository) DeleteSession(context.Context, string) error { return nil }

func parsedLogs(t *testing.T, logs *bytes.Buffer) []map[string]any {
	t.Helper()
	var entries []map[string]any
	for _, line := range strings.Split(strings.TrimSpace(logs.String()), "\n") {
		if line == "" {
			continue
		}
		var entry map[string]any
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			t.Fatal(err)
		}
		if strings.Count(line, `"user_id":`) > 1 {
			t.Fatalf("duplicate user_id: %s", line)
		}
		entries = append(entries, entry)
	}
	return entries
}

func loggedRouter(logs *bytes.Buffer) http.Handler {
	return NewRouter(slog.New(slog.NewJSONHandler(logs, nil)), "http://localhost:5173", "development", auth.NewService(logAuthRepository{}), nil, nil, nil, nil, nil)
}

func TestLogsCorrelateAuthenticatedRequest(t *testing.T) {
	var logs bytes.Buffer
	r := httptest.NewRequest(http.MethodGet, "/auth/me?secret=DO-NOT-LOG", nil)
	r.Header.Set("X-Request-ID", "demo-request-1")
	r.AddCookie(&http.Cookie{Name: sessionCookie, Value: "COOKIE-SECRET"})
	w := httptest.NewRecorder()
	loggedRouter(&logs).ServeHTTP(w, r)
	entries := parsedLogs(t, &logs)
	if len(entries) != 2 {
		t.Fatalf("unexpected entries: %s", logs.String())
	}
	for _, entry := range entries {
		if entry["request_id"] != "demo-request-1" {
			t.Fatalf("missing correlation: %#v", entry)
		}
	}
	completed := entries[1]
	if completed["user_id"] != "user-42" || completed["event"] != "http.request.completed" || completed["operation"] != "auth.current_user" || completed["outcome"] != "success" {
		t.Fatalf("missing request context: %#v", completed)
	}
	if _, ok := completed["duration_ms"].(float64); !ok {
		t.Fatal("duration_ms must be numeric")
	}
	if strings.Contains(logs.String(), "DO-NOT-LOG") || strings.Contains(logs.String(), "COOKIE-SECRET") {
		t.Fatalf("secrets in logs: %s", logs.String())
	}
	if w.Header().Get("X-Request-ID") != "demo-request-1" {
		t.Fatal("request ID not returned")
	}
}

func TestRejectedCellWriteLogsUserIdentifiersAndReason(t *testing.T) {
	var logs bytes.Buffer
	r := httptest.NewRequest(http.MethodPatch, "/journals/j1/rows/r1/cells/c1", strings.NewReader(`{"value": "SECRET-CELL", "unknown": true}`))
	r.AddCookie(&http.Cookie{Name: sessionCookie, Value: "token"})
	w := httptest.NewRecorder()
	loggedRouter(&logs).ServeHTTP(w, r)
	entries := parsedLogs(t, &logs)
	entry := entries[len(entries)-1]
	for key, want := range map[string]any{"user_id": "user-42", "journal_id": "j1", "row_id": "r1", "column_id": "c1", "operation": "cell.update", "error_code": "INVALID_REQUEST", "outcome": "rejected", "level": "WARN"} {
		if entry[key] != want {
			t.Fatalf("%s: want %v, got %v", key, want, entry[key])
		}
	}
	if strings.Contains(logs.String(), "SECRET-CELL") {
		t.Fatal("cell contents logged")
	}
}

func TestRouterLogsUnmatchedAndPreflightRequests(t *testing.T) {
	for _, tc := range []struct {
		method, path string
		status       int
	}{{"GET", "/missing", 404}, {"PUT", "/auth/login", 405}, {"OPTIONS", "/auth/login", 204}} {
		var logs bytes.Buffer
		w := httptest.NewRecorder()
		loggedRouter(&logs).ServeHTTP(w, httptest.NewRequest(tc.method, tc.path, nil))
		if w.Code != tc.status || w.Header().Get("X-Request-ID") == "" {
			t.Fatalf("unexpected response: %d %v", w.Code, w.Header())
		}
		entries := parsedLogs(t, &logs)
		if entries[len(entries)-1]["event"] != "http.request.completed" {
			t.Fatal("missing access log")
		}
	}
}

func TestInvalidRequestIDsAreReplaced(t *testing.T) {
	for _, value := range []string{"", strings.Repeat("x", 129), "untrusted value", "bad\nvalue"} {
		var logs bytes.Buffer
		r := httptest.NewRequest("GET", "/missing", nil)
		r.Header.Set("X-Request-ID", value)
		w := httptest.NewRecorder()
		loggedRouter(&logs).ServeHTTP(w, r)
		if got := w.Header().Get("X-Request-ID"); !validRequestID(got) || got == value {
			t.Fatalf("bad generated id %q", got)
		}
	}
}

func TestInternalErrorsKeepUserContextWithoutDuplicateFields(t *testing.T) {
	var logs bytes.Buffer
	h := observability(slog.New(slog.NewJSONHandler(&logs, nil)), &requestMetrics{})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		internalError(w, r, "cell.write", errors.New("database unavailable"), "user_id", "user-42", "column_id", "c1")
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("PATCH", "/test", nil))
	entries := parsedLogs(t, &logs)
	if entries[1]["event"] != "http.request.failed" || entries[1]["user_id"] != "user-42" || entries[2]["outcome"] != "failed" {
		t.Fatalf("bad errors: %s", logs.String())
	}
}

func TestLogoutLogsUserBeforeDeletingSession(t *testing.T) {
	var logs bytes.Buffer
	r := httptest.NewRequest("POST", "/auth/logout", nil)
	r.AddCookie(&http.Cookie{Name: sessionCookie, Value: "token"})
	loggedRouter(&logs).ServeHTTP(httptest.NewRecorder(), r)
	entries := parsedLogs(t, &logs)
	if entries[1]["user_id"] != "user-42" || entries[1]["operation"] != "auth.logout" {
		t.Fatalf("missing logout user: %s", logs.String())
	}
}

func TestSuccessfulHealthAndMetricsDoNotFloodLogs(t *testing.T) {
	var logs bytes.Buffer
	h := loggedRouter(&logs)
	for _, path := range []string{"/health", "/metrics"} {
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", path, nil))
	}
	if logs.Len() != 0 {
		t.Fatalf("monitoring noise: %s", logs.String())
	}
}

func TestSuccessfulCellUpdateHasReadableOperation(t *testing.T) {
	var logs bytes.Buffer
	router := mux.NewRouter()
	router.Use(logRoute)
	router.HandleFunc("/journals/{journalID}/rows/{rowID}/cells/{columnID}", func(w http.ResponseWriter, r *http.Request) {
		setRequestUser(r, "user-42")
		w.WriteHeader(http.StatusNoContent)
	}).Methods(http.MethodPatch)
	h := observability(slog.New(slog.NewJSONHandler(&logs, nil)), &requestMetrics{})(router)
	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("PATCH", "/journals/j1/rows/r1/cells/c1", nil))
	entries := parsedLogs(t, &logs)
	if entry := entries[1]; entry["msg"] != "Значение ячейки сохранено" || entry["operation"] != "cell.update" || entry["column_id"] != "c1" {
		t.Fatalf("bad success log: %#v", entry)
	}
}

func TestRegisterLogsUserWithoutCredentials(t *testing.T) {
	var logs bytes.Buffer
	r := httptest.NewRequest("POST", "/auth/register", strings.NewReader(`{"name":"SECRET-NAME","email":"SECRET-EMAIL@example.com","password":"SECRET-PASSWORD"}`))
	w := httptest.NewRecorder()
	loggedRouter(&logs).ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register failed: %d", w.Code)
	}
	entries := parsedLogs(t, &logs)
	if entries[1]["user_id"] != "user-42" || entries[1]["operation"] != "auth.register" {
		t.Fatal("missing new user")
	}
	if strings.Contains(logs.String(), "SECRET-") {
		t.Fatal("credentials in logs")
	}
}
