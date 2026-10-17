package httptransport

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync/atomic"
	"time"
)

type contextKey string

const loggerKey contextKey = "request_logger"

type requestLogContext struct {
	logger    *slog.Logger
	userID    string
	operation string
	route     string
	attrs     []any
}

type responseRecorder struct {
	http.ResponseWriter
	status       int
	bytes        int
	errorCode    string
	errorMessage string
	failed       bool
}

func (w *responseRecorder) Unwrap() http.ResponseWriter { return w.ResponseWriter }

type requestMetrics struct {
	requests   atomic.Uint64
	errors     atomic.Uint64
	panics     atomic.Uint64
	inFlight   atomic.Int64
	durationNS atomic.Uint64
}

func (m *requestMetrics) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	_, _ = fmt.Fprintf(w, "# TYPE trade_diary_http_requests_total counter\ntrade_diary_http_requests_total %d\n# TYPE trade_diary_http_errors_total counter\ntrade_diary_http_errors_total %d\n# TYPE trade_diary_http_panics_total counter\ntrade_diary_http_panics_total %d\n# TYPE trade_diary_http_requests_in_flight gauge\ntrade_diary_http_requests_in_flight %d\n# TYPE trade_diary_http_request_duration_seconds_total counter\ntrade_diary_http_request_duration_seconds_total %f\n", m.requests.Load(), m.errors.Load(), m.panics.Load(), m.inFlight.Load(), float64(m.durationNS.Load())/float64(time.Second))
}

func (w *responseRecorder) WriteHeader(status int) {
	if w.status != 0 {
		return
	}
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseRecorder) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.WriteHeader(http.StatusOK)
	}
	n, err := w.ResponseWriter.Write(body)
	w.bytes += n
	return n, err
}

func observability(logger *slog.Logger, metrics *requestMetrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			metrics.inFlight.Add(1)
			requestID := r.Header.Get("X-Request-ID")
			if !validRequestID(requestID) {
				requestID = newRequestID()
			}
			requestLogger := logger.With("request_id", requestID, "method", r.Method, "path", r.URL.Path)
			info := &requestLogContext{logger: requestLogger, operation: "http.unmatched"}
			if r.Method == http.MethodOptions {
				info.operation = "http.preflight"
			}
			ctx := context.WithValue(r.Context(), loggerKey, info)
			r = r.WithContext(ctx)
			w.Header().Set("X-Request-ID", requestID)
			recorder := &responseRecorder{ResponseWriter: w}
			quiet := r.URL.Path == "/health" || r.URL.Path == "/metrics"
			if !quiet {
				requestLogger.InfoContext(ctx, "Получен HTTP-запрос", "event", "http.request.started", "user_id", "")
			}

			defer func() {
				if recovered := recover(); recovered != nil {
					recorder.failed = true
					metrics.panics.Add(1)
					requestLogger.With("user_id", info.userID).ErrorContext(ctx, "request panic", "event", "http.request.panicked", "operation", info.operation, "panic", recovered, "stack", string(debug.Stack()))
					if recorder.status == 0 {
						writeError(recorder, http.StatusInternalServerError, "internal error")
					}
				}
				status := recorder.status
				if status == 0 {
					status = http.StatusOK
				}
				metrics.inFlight.Add(-1)
				metrics.requests.Add(1)
				metrics.durationNS.Add(uint64(time.Since(started)))
				if status >= 500 || recorder.failed {
					metrics.errors.Add(1)
				}
				if !quiet || status >= 400 || recorder.failed {
					outcome, message := "success", operationMessage(info.operation)
					if status >= 400 {
						outcome, message = "rejected", "Запрос отклонён"
					}
					if status >= 500 || recorder.failed {
						outcome, message = "failed", "Не удалось выполнить запрос"
					}
					args := []any{"event", "http.request.completed", "user_id", info.userID, "operation", info.operation,
						"route", info.route, "outcome", outcome, "status", status, "bytes", recorder.bytes,
						"duration_ms", float64(time.Since(started).Microseconds()) / 1000}
					if recorder.errorCode != "" {
						args = append(args, "error_code", recorder.errorCode, "reason", recorder.errorMessage)
					}
					args = append(args, info.attrs...)
					level := accessLevel(status)
					if recorder.failed {
						level = slog.LevelError
					}
					requestLogger.Log(ctx, level, message, args...)
				}
			}()

			next.ServeHTTP(recorder, r)
		})
	}
}

func requestLogger(r *http.Request) *slog.Logger {
	if info, ok := r.Context().Value(loggerKey).(*requestLogContext); ok {
		return info.logger.With("user_id", info.userID)
	}
	return slog.Default()
}

func internalError(w http.ResponseWriter, r *http.Request, operation string, err error, attrs ...any) {
	args := []any{"event", "http.request.failed", "operation", operation, "error", err}
	for i := 0; i+1 < len(attrs); i += 2 {
		if attrs[i] == "user_id" {
			if userID, ok := attrs[i+1].(string); ok {
				setRequestUser(r, userID)
			}
			continue
		}
		args = append(args, attrs[i], attrs[i+1])
	}
	requestLogger(r).ErrorContext(r.Context(), "Внутренняя ошибка при выполнении операции", args...)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(value)
}

func setRequestUser(r *http.Request, userID string) {
	if info, ok := r.Context().Value(loggerKey).(*requestLogContext); ok {
		info.userID = userID
	}
}

func addRequestAttrs(r *http.Request, attrs ...any) {
	if info, ok := r.Context().Value(loggerKey).(*requestLogContext); ok {
		info.attrs = append(info.attrs, attrs...)
	}
}

func validRequestID(value string) bool {
	if len(value) == 0 || len(value) > 128 {
		return false
	}
	for _, c := range value {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '_' || c == '.') {
			return false
		}
	}
	return true
}

func accessLevel(status int) slog.Level {
	if status >= 500 {
		return slog.LevelError
	}
	if status >= 400 {
		return slog.LevelWarn
	}
	return slog.LevelInfo
}
