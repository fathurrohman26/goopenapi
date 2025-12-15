// Package http provides net/http compatible middleware and handlers for OpenAPI specifications.
//
// This package allows you to easily integrate OpenAPI documentation and validation
// into your Go HTTP servers. It is compatible with the standard net/http package
// and works with any router that supports http.Handler.
//
// Basic usage:
//
//	spec := &openapi.Document{...}
//	plugin := http.New(spec, &http.Options{
//	    SpecPath:      "/openapi.json",
//	    SwaggerUIPath: "/docs",
//	    EnableCORS:    true,
//	})
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api/users", usersHandler)
//
//	// Wrap mux with plugin middleware and mount spec handlers
//	handler := plugin.WrapMux(mux)
//	http.ListenAndServe(":8080", handler)
//
// Manual middleware usage:
//
//	mux := http.NewServeMux()
//
//	// Add individual handlers
//	mux.Handle("/openapi.json", plugin.SpecHandler())
//	mux.Handle("/docs", plugin.SwaggerUIHandler())
//
//	// Add middleware
//	handler := http.Chain(
//	    http.CORS(nil),
//	    http.Logging(log.Printf),
//	)(mux)
package yahttp

import (
	"net/http"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

// Handler creates an http.Handler that serves an OpenAPI spec at /openapi.json
// and Swagger UI at /docs, with optional middleware.
func Handler(spec *openapi.Document, opts *Options) http.Handler {
	plugin := New(spec, opts)
	mux := http.NewServeMux()
	return plugin.WrapMux(mux)
}

// MustNew creates a new HTTP plugin and panics if spec is nil.
func MustNew(spec *openapi.Document, opts *Options) *Plugin {
	if spec == nil {
		panic("http: spec cannot be nil")
	}
	return New(spec, opts)
}

// WithSpec creates a plugin builder starting with a spec.
func WithSpec(spec *openapi.Document) *PluginBuilder {
	return &PluginBuilder{
		spec: spec,
		opts: DefaultOptions(),
	}
}

// PluginBuilder provides a fluent API for configuring the HTTP plugin.
type PluginBuilder struct {
	spec *openapi.Document
	opts *Options
}

// SpecPath sets the path for serving the OpenAPI spec.
func (b *PluginBuilder) SpecPath(path string) *PluginBuilder {
	b.opts.SpecPath = path
	return b
}

// SwaggerUIPath sets the path for serving Swagger UI.
func (b *PluginBuilder) SwaggerUIPath(path string) *PluginBuilder {
	b.opts.SwaggerUIPath = path
	return b
}

// EnableValidation enables request validation.
func (b *PluginBuilder) EnableValidation() *PluginBuilder {
	b.opts.EnableValidation = true
	return b
}

// EnableCORS enables CORS with default options.
func (b *PluginBuilder) EnableCORS() *PluginBuilder {
	b.opts.EnableCORS = true
	return b
}

// WithCORS enables CORS with custom options.
func (b *PluginBuilder) WithCORS(opts *CORSOptions) *PluginBuilder {
	b.opts.EnableCORS = true
	b.opts.CORSOptions = opts
	return b
}

// EnableLogging enables request logging.
func (b *PluginBuilder) EnableLogging() *PluginBuilder {
	b.opts.EnableLogging = true
	return b
}

// WithLogger sets a custom logger.
func (b *PluginBuilder) WithLogger(logger func(format string, args ...any)) *PluginBuilder {
	b.opts.EnableLogging = true
	b.opts.Logger = logger
	return b
}

// ValidationErrorHandler sets a custom validation error handler.
func (b *PluginBuilder) ValidationErrorHandler(handler func(http.ResponseWriter, *http.Request, error)) *PluginBuilder {
	b.opts.ValidationErrorHandler = handler
	return b
}

// Build creates the plugin with the configured options.
func (b *PluginBuilder) Build() *Plugin {
	return New(b.spec, b.opts)
}

// Wrap wraps an existing handler with the configured middleware.
func (b *PluginBuilder) Wrap(h http.Handler) http.Handler {
	return b.Build().Handler()(h)
}

// Mount mounts spec handlers on the given mux and returns a wrapped handler.
func (b *PluginBuilder) Mount(mux *http.ServeMux) http.Handler {
	return b.Build().WrapMux(mux)
}

// Recovery returns a middleware that recovers from panics.
func Recovery(handler func(w http.ResponseWriter, r *http.Request, err any)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					if handler != nil {
						handler(w, r, err)
					} else {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestID returns a middleware that adds a request ID to each request.
func RequestID(generator func() string) Middleware {
	if generator == nil {
		counter := int64(0)
		generator = func() string {
			counter++
			return string(rune(counter))
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generator()
			}
			w.Header().Set("X-Request-ID", requestID)
			next.ServeHTTP(w, r)
		})
	}
}

// ContentType returns a middleware that sets the Content-Type header.
func ContentType(contentType string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			next.ServeHTTP(w, r)
		})
	}
}

// JSON returns a middleware that sets Content-Type to application/json.
func JSON() Middleware {
	return ContentType("application/json; charset=utf-8")
}
