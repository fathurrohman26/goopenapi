package yahttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

// SpecHandler returns an http.Handler that serves the OpenAPI specification.
// It supports both JSON and YAML formats based on Accept header or file extension.
func (p *Plugin) SpecHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		format := p.detectFormat(r)
		p.serveSpec(w, format)
	})
}

// SpecHandlerFunc returns an http.HandlerFunc that serves the OpenAPI specification.
func (p *Plugin) SpecHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		format := p.detectFormat(r)
		p.serveSpec(w, format)
	}
}

// JSONSpecHandler returns a handler that always serves the spec as JSON.
func (p *Plugin) JSONSpecHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.serveSpec(w, "json")
	})
}

// YAMLSpecHandler returns a handler that always serves the spec as YAML.
func (p *Plugin) YAMLSpecHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p.serveSpec(w, "yaml")
	})
}

func (p *Plugin) detectFormat(r *http.Request) string {
	// Check URL path extension
	path := r.URL.Path
	if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") {
		return "yaml"
	}
	if strings.HasSuffix(path, ".json") {
		return "json"
	}

	// Check Accept header
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/yaml") || strings.Contains(accept, "text/yaml") {
		return "yaml"
	}

	// Check query parameter
	if r.URL.Query().Get("format") == "yaml" {
		return "yaml"
	}

	// Default to JSON
	return "json"
}

func (p *Plugin) serveSpec(w http.ResponseWriter, format string) {
	var data []byte
	var err error
	var contentType string

	switch format {
	case "yaml":
		data, err = yaml.Marshal(p.spec)
		contentType = "application/yaml; charset=utf-8"
	default:
		data, err = json.MarshalIndent(p.spec, "", "  ")
		contentType = "application/json; charset=utf-8"
	}

	if err != nil {
		http.Error(w, "Failed to serialize OpenAPI spec", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(data)
}

// ServeSpec is a standalone function to serve an OpenAPI spec.
func ServeSpec(spec *openapi.Document) http.Handler {
	p := New(spec, nil)
	return p.SpecHandler()
}
