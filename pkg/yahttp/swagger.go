package yahttp

import (
	"fmt"
	"html/template"
	"net/http"
)

const swaggerUITemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{.Title}} - Swagger UI</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
        .swagger-ui .topbar { display: none; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "{{.SpecURL}}",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIBundle.SwaggerUIStandalonePreset
                ],
                layout: "BaseLayout",
                defaultModelsExpandDepth: 1,
                defaultModelExpandDepth: 1,
                docExpansion: "list",
                filter: true,
                showExtensions: true,
                showCommonExtensions: true
            });
        };
    </script>
</body>
</html>`

// SwaggerUIOptions configures Swagger UI rendering.
type SwaggerUIOptions struct {
	// Title is the page title (default: API title from spec)
	Title string

	// SpecURL is the URL to the OpenAPI spec (default: plugin's SpecPath)
	SpecURL string

	// CustomCSS is optional custom CSS to inject
	CustomCSS string

	// CustomJS is optional custom JavaScript to inject
	CustomJS string
}

// SwaggerUIHandler returns an http.Handler that serves Swagger UI.
func (p *Plugin) SwaggerUIHandler() http.Handler {
	return p.SwaggerUIHandlerWithOptions(nil)
}

// SwaggerUIHandlerWithOptions returns a Swagger UI handler with custom options.
func (p *Plugin) SwaggerUIHandlerWithOptions(opts *SwaggerUIOptions) http.Handler {
	title, specURL := p.resolveDocOptions(opts.getTitle(), opts.getSpecURL())
	return p.createDocHandler("swagger", swaggerUITemplate, title, specURL, "Swagger UI")
}

// SwaggerUIHandlerFunc returns an http.HandlerFunc that serves Swagger UI.
func (p *Plugin) SwaggerUIHandlerFunc() http.HandlerFunc {
	handler := p.SwaggerUIHandler()
	return func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, r)
	}
}

// RedocHandler returns an http.Handler that serves ReDoc documentation.
func (p *Plugin) RedocHandler() http.Handler {
	return p.RedocHandlerWithOptions(nil)
}

// RedocOptions configures ReDoc rendering.
type RedocOptions struct {
	Title   string
	SpecURL string
}

const redocTemplate = `<!DOCTYPE html>
<html>
<head>
    <title>{{.Title}} - API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>body { margin: 0; padding: 0; }</style>
</head>
<body>
    <redoc spec-url='{{.SpecURL}}'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

// RedocHandlerWithOptions returns a ReDoc handler with custom options.
func (p *Plugin) RedocHandlerWithOptions(opts *RedocOptions) http.Handler {
	title, specURL := p.resolveDocOptions(opts.getTitle(), opts.getSpecURL())
	return p.createDocHandler("redoc", redocTemplate, title, specURL, "ReDoc")
}

// Helper methods for nil-safe option access
func (o *SwaggerUIOptions) getTitle() string {
	if o == nil {
		return ""
	}
	return o.Title
}

func (o *SwaggerUIOptions) getSpecURL() string {
	if o == nil {
		return ""
	}
	return o.SpecURL
}

func (o *RedocOptions) getTitle() string {
	if o == nil {
		return ""
	}
	return o.Title
}

func (o *RedocOptions) getSpecURL() string {
	if o == nil {
		return ""
	}
	return o.SpecURL
}

// resolveDocOptions resolves title and specURL with defaults from plugin.
func (p *Plugin) resolveDocOptions(title, specURL string) (string, string) {
	if title == "" && p.spec != nil {
		title = p.spec.Info.Title
	}
	if title == "" {
		title = "API Documentation"
	}
	if specURL == "" {
		specURL = p.options.SpecPath
	}
	return title, specURL
}

// createDocHandler creates an HTTP handler that renders a documentation template.
func (p *Plugin) createDocHandler(name, tmplContent, title, specURL, docType string) http.Handler {
	tmpl := template.Must(template.New(name).Parse(tmplContent))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Title   string
			SpecURL string
		}{
			Title:   title,
			SpecURL: specURL,
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := tmpl.Execute(w, data); err != nil {
			http.Error(w, fmt.Sprintf("Failed to render %s: %v", docType, err), http.StatusInternalServerError)
		}
	})
}
