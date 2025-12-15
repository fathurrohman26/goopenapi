// Package http provides net/http compatible middleware for OpenAPI specifications.
package yahttp

import (
	"net/http"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

// Middleware represents a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// MiddlewareFunc is an adapter to allow ordinary functions to be used as middleware.
type MiddlewareFunc func(http.ResponseWriter, *http.Request, http.Handler)

// Then wraps a handler with the middleware function.
func (m MiddlewareFunc) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m(w, r, next)
	})
}

// Chain chains multiple middleware together.
func Chain(middlewares ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Plugin provides OpenAPI-aware HTTP middleware.
type Plugin struct {
	spec    *openapi.Document
	options *Options
}

// Options configures the HTTP plugin behavior.
type Options struct {
	// SpecPath is the path to serve the OpenAPI spec (default: "/openapi.json")
	SpecPath string

	// SwaggerUIPath is the path to serve Swagger UI (default: "/docs")
	SwaggerUIPath string

	// EnableValidation enables request validation (default: false)
	EnableValidation bool

	// EnableCORS enables CORS headers (default: false)
	EnableCORS bool

	// CORSOptions configures CORS behavior
	CORSOptions *CORSOptions

	// EnableLogging enables request logging (default: false)
	EnableLogging bool

	// Logger is the logging function (default: log.Printf)
	Logger func(format string, args ...any)

	// ValidationErrorHandler handles validation errors
	ValidationErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}

// DefaultOptions returns default plugin options.
func DefaultOptions() *Options {
	return &Options{
		SpecPath:         "/openapi.json",
		SwaggerUIPath:    "/docs",
		EnableValidation: false,
		EnableCORS:       false,
		EnableLogging:    false,
	}
}

// New creates a new HTTP plugin with the given OpenAPI specification.
func New(spec *openapi.Document, opts *Options) *Plugin {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &Plugin{
		spec:    spec,
		options: opts,
	}
}

// Spec returns the OpenAPI specification.
func (p *Plugin) Spec() *openapi.Document {
	return p.spec
}

// Options returns the plugin options.
func (p *Plugin) Options() *Options {
	return p.options
}

// Handler returns a middleware chain based on the configured options.
func (p *Plugin) Handler() Middleware {
	var middlewares []Middleware

	if p.options.EnableLogging {
		middlewares = append(middlewares, p.LoggingMiddleware())
	}

	if p.options.EnableCORS {
		middlewares = append(middlewares, p.CORSMiddleware())
	}

	if p.options.EnableValidation {
		middlewares = append(middlewares, p.ValidationMiddleware())
	}

	if len(middlewares) == 0 {
		return func(h http.Handler) http.Handler { return h }
	}

	return Chain(middlewares...)
}

// Mount mounts the OpenAPI spec and Swagger UI handlers on the given mux.
func (p *Plugin) Mount(mux *http.ServeMux) {
	if p.options.SpecPath != "" {
		mux.Handle(p.options.SpecPath, p.SpecHandler())
	}
	if p.options.SwaggerUIPath != "" {
		mux.Handle(p.options.SwaggerUIPath, p.SwaggerUIHandler())
		mux.Handle(p.options.SwaggerUIPath+"/", p.SwaggerUIHandler())
	}
}

// WrapMux wraps an existing ServeMux with the plugin middleware and mounts spec handlers.
func (p *Plugin) WrapMux(mux *http.ServeMux) http.Handler {
	p.Mount(mux)
	return p.Handler()(mux)
}
