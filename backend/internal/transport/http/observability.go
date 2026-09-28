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

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

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
			if requestID == "" {
				requestID = newRequestID()
			}
			requestLogger := logger.With("request_id", requestID, "method", r.Method, "path", r.URL.Path)
			ctx := context.WithValue(r.Context(), loggerKey, requestLogger)
			w.Header().Set("X-Request-ID", requestID)
			recorder := &responseRecorder{ResponseWriter: w}

			defer func() {
				if recovered := recover(); recovered != nil {
					metrics.panics.Add(1)
					requestLogger.Error("request panic", "panic", recovered, "stack", string(debug.Stack()))
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
				if status >= 500 {
					metrics.errors.Add(1)
				}
				if r.URL.Path != "/health" || status >= 400 {
					requestLogger.LogAttrs(ctx, accessLevel(status), "request completed",
						slog.Int("status", status), slog.Int("bytes", recorder.bytes), slog.Duration("duration", time.Since(started)))
				}
			}()

			next.ServeHTTP(recorder, r.WithContext(ctx))
		})
	}
}

func requestLogger(r *http.Request) *slog.Logger {
	if logger, ok := r.Context().Value(loggerKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func internalError(w http.ResponseWriter, r *http.Request, operation string, err error, attrs ...any) {
	args := []any{"operation", operation, "error", err}
	args = append(args, attrs...)
	requestLogger(r).ErrorContext(r.Context(), "request failed", args...)
	writeError(w, http.StatusInternalServerError, "internal error")
}

func newRequestID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return time.Now().UTC().Format("20060102T150405.000000000")
	}
	return hex.EncodeToString(value)
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
