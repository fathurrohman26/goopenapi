package yahttp

import (
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware returns a middleware that logs HTTP requests.
func (p *Plugin) LoggingMiddleware() Middleware {
	logger := p.options.Logger
	if logger == nil {
		logger = log.Printf
	}
	return Logging(logger)
}

// Logging returns a standalone logging middleware.
func Logging(logger func(format string, args ...any)) Middleware {
	if logger == nil {
		logger = log.Printf
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)
			logger("[%s] %s %s %d %v",
				r.Method,
				r.URL.Path,
				r.RemoteAddr,
				wrapped.statusCode,
				duration,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Unwrap returns the underlying ResponseWriter for compatibility with
// http.ResponseController and other interfaces.
func (rw *responseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// LoggingHandler wraps a handler with logging using log.Printf.
func LoggingHandler(h http.Handler) http.Handler {
	return Logging(log.Printf)(h)
}

// LogEntry represents a structured log entry.
type LogEntry struct {
	Method     string        `json:"method"`
	Path       string        `json:"path"`
	RemoteAddr string        `json:"remote_addr"`
	StatusCode int           `json:"status_code"`
	Duration   time.Duration `json:"duration"`
	UserAgent  string        `json:"user_agent,omitempty"`
	RequestID  string        `json:"request_id,omitempty"`
}

// StructuredLogging returns a middleware that provides structured log entries.
func StructuredLogging(handler func(entry LogEntry)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(wrapped, r)

			handler(LogEntry{
				Method:     r.Method,
				Path:       r.URL.Path,
				RemoteAddr: r.RemoteAddr,
				StatusCode: wrapped.statusCode,
				Duration:   time.Since(start),
				UserAgent:  r.UserAgent(),
				RequestID:  r.Header.Get("X-Request-ID"),
			})
		})
	}
}
