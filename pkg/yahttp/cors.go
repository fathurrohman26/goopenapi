package yahttp

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSOptions configures CORS behavior.
type CORSOptions struct {
	// AllowedOrigins is a list of allowed origins (default: ["*"])
	AllowedOrigins []string

	// AllowedMethods is a list of allowed HTTP methods (default: common methods)
	AllowedMethods []string

	// AllowedHeaders is a list of allowed headers (default: common headers)
	AllowedHeaders []string

	// ExposedHeaders is a list of headers to expose to the client
	ExposedHeaders []string

	// AllowCredentials indicates whether credentials are allowed
	AllowCredentials bool

	// MaxAge is the max age for preflight cache in seconds (default: 86400)
	MaxAge int
}

// DefaultCORSOptions returns sensible CORS defaults.
func DefaultCORSOptions() *CORSOptions {
	return &CORSOptions{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Requested-With",
			"X-Request-ID",
		},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
	}
}

// CORSMiddleware returns a middleware that handles CORS.
func (p *Plugin) CORSMiddleware() Middleware {
	opts := p.options.CORSOptions
	if opts == nil {
		opts = DefaultCORSOptions()
	}
	return CORS(opts)
}

// CORS returns a standalone CORS middleware with the given options.
func CORS(opts *CORSOptions) Middleware {
	if opts == nil {
		opts = DefaultCORSOptions()
	}

	cfg := newCORSConfig(opts)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			allowOrigin := cfg.getAllowedOrigin(r.Header.Get("Origin"))

			cfg.setCORSHeaders(w, allowOrigin)

			if r.Method == http.MethodOptions {
				cfg.setPreflightHeaders(w, allowOrigin)
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type corsConfig struct {
	allowedOrigins   []string
	allowedMethods   string
	allowedHeaders   string
	exposedHeaders   string
	maxAge           string
	allowCredentials bool
}

func newCORSConfig(opts *CORSOptions) *corsConfig {
	origins := opts.AllowedOrigins
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	return &corsConfig{
		allowedOrigins:   origins,
		allowedMethods:   strings.Join(opts.AllowedMethods, ", "),
		allowedHeaders:   strings.Join(opts.AllowedHeaders, ", "),
		exposedHeaders:   strings.Join(opts.ExposedHeaders, ", "),
		maxAge:           strconv.Itoa(opts.MaxAge),
		allowCredentials: opts.AllowCredentials,
	}
}

func (c *corsConfig) getAllowedOrigin(origin string) string {
	if !isOriginAllowed(origin, c.allowedOrigins) {
		return ""
	}
	if len(c.allowedOrigins) == 1 && c.allowedOrigins[0] == "*" {
		return "*"
	}
	return origin
}

func (c *corsConfig) setCORSHeaders(w http.ResponseWriter, allowOrigin string) {
	if allowOrigin == "" {
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
	if c.allowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
	if c.exposedHeaders != "" {
		w.Header().Set("Access-Control-Expose-Headers", c.exposedHeaders)
	}
}

func (c *corsConfig) setPreflightHeaders(w http.ResponseWriter, allowOrigin string) {
	if allowOrigin == "" {
		return
	}
	w.Header().Set("Access-Control-Allow-Methods", c.allowedMethods)
	w.Header().Set("Access-Control-Allow-Headers", c.allowedHeaders)
	w.Header().Set("Access-Control-Max-Age", c.maxAge)
}

func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
	}
	return false
}

// CORSHandler wraps a handler with CORS support using default options.
func CORSHandler(h http.Handler) http.Handler {
	return CORS(DefaultCORSOptions())(h)
}
