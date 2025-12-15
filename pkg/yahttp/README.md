# HTTP Middleware Plugin

The HTTP middleware plugin provides `net/http` compatible middleware and handlers for integrating OpenAPI specifications into your Go HTTP servers.

## Installation

```bash
go get github.com/fathurrohman26/yaswag/pkg/yahttp
```

## Features

- Serve OpenAPI specs in JSON/YAML format with auto-detection
- Swagger UI and ReDoc documentation handlers
- CORS middleware with configurable options
- Request logging (standard and structured)
- Request validation against OpenAPI spec
- Panic recovery middleware
- Request ID middleware
- Fluent builder API for easy configuration
- Compatible with `net/http.ServeMux` and any `http.Handler`-based router

## Quick Start

```go
package main

import (
    "net/http"

    "github.com/fathurrohman26/yaswag/pkg/openapi"
    "github.com/fathurrohman26/yaswag/pkg/yahttp"
)

func main() {
    // Create your OpenAPI spec
    spec := &openapi.Document{
        OpenAPI: "3.0.3",
        Info: openapi.Info{
            Title:   "My API",
            Version: "1.0.0",
        },
        Paths: map[string]*openapi.PathItem{
            "/users": {
                Get: &openapi.Operation{
                    Summary: "List users",
                    Responses: map[string]*openapi.Response{
                        "200": {Description: "Success"},
                    },
                },
            },
        },
    }

    // Create plugin with options
    plugin := yahttp.New(spec, &yahttp.Options{
        SpecPath:      "/openapi.json",
        SwaggerUIPath: "/docs",
        EnableCORS:    true,
        EnableLogging: true,
    })

    // Create your mux and add routes
    mux := http.NewServeMux()
    mux.HandleFunc("/users", usersHandler)

    // Wrap mux with plugin middleware and mount spec handlers
    handler := plugin.WrapMux(mux)

    http.ListenAndServe(":8080", handler)
}

func usersHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte(`[{"id": 1, "name": "John"}]`))
}
```

## Fluent Builder API

For a more expressive configuration, use the builder pattern:

```go
handler := yahttp.WithSpec(spec).
    SpecPath("/openapi.json").
    SwaggerUIPath("/docs").
    EnableCORS().
    EnableLogging().
    EnableValidation().
    Mount(mux)
```

## Configuration Options

```go
type Options struct {
    // SpecPath is the URL path for serving the OpenAPI spec (default: "/openapi.json")
    SpecPath string

    // SwaggerUIPath is the URL path for Swagger UI (default: "/docs")
    SwaggerUIPath string

    // EnableValidation enables request validation against the OpenAPI spec
    EnableValidation bool

    // EnableCORS enables CORS middleware
    EnableCORS bool

    // CORSOptions configures CORS behavior (uses defaults if nil)
    CORSOptions *CORSOptions

    // EnableLogging enables request logging
    EnableLogging bool

    // Logger is a custom logger function (defaults to log.Printf)
    Logger func(format string, args ...any)

    // ValidationErrorHandler handles validation errors (uses default JSON response if nil)
    ValidationErrorHandler func(w http.ResponseWriter, r *http.Request, err error)
}
```

## Middleware

### CORS

```go
// Use default CORS options
handler := yahttp.CORS(nil)(mux)

// Or with custom options
handler := yahttp.CORS(&yahttp.CORSOptions{
    AllowedOrigins:   []string{"https://example.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
    AllowedHeaders:   []string{"Authorization", "Content-Type"},
    AllowCredentials: true,
    MaxAge:           86400,
})(mux)
```

### Logging

```go
// Standard logging with log.Printf
handler := yahttp.Logging(log.Printf)(mux)

// Structured logging
handler := yahttp.StructuredLogging(func(entry yahttp.LogEntry) {
    log.Printf("%s %s %d %v", entry.Method, entry.Path, entry.StatusCode, entry.Duration)
})(mux)
```

### Request Validation

```go
// Validate requests against OpenAPI spec
handler := yahttp.RequestValidation(spec, nil)(mux)

// With custom error handler
handler := yahttp.RequestValidation(spec, func(w http.ResponseWriter, r *http.Request, err error) {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
})(mux)
```

### Recovery

```go
// Recover from panics with custom handler
handler := yahttp.Recovery(func(w http.ResponseWriter, r *http.Request, err any) {
    log.Printf("Panic recovered: %v", err)
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
})(mux)
```

### Request ID

```go
// Add request IDs to responses
handler := yahttp.RequestID(nil)(mux) // Uses default generator

// With custom generator
handler := yahttp.RequestID(func() string {
    return uuid.New().String()
})(mux)
```

### Content-Type

```go
// Set Content-Type header
handler := yahttp.ContentType("application/json")(mux)

// Shorthand for JSON
handler := yahttp.JSON()(mux)
```

## Chaining Middleware

```go
handler := yahttp.Chain(
    yahttp.Recovery(nil),
    yahttp.RequestID(nil),
    yahttp.CORS(nil),
    yahttp.Logging(log.Printf),
    yahttp.JSON(),
)(mux)
```

## Handlers

### Spec Handler

Serves the OpenAPI spec with automatic format detection:

```go
plugin := yahttp.New(spec, nil)

// Auto-detects format from URL extension, Accept header, or query param
mux.Handle("/openapi", plugin.SpecHandler())
mux.Handle("/openapi.json", plugin.SpecHandler())
mux.Handle("/openapi.yaml", plugin.SpecHandler())

// Or explicit format handlers
mux.Handle("/openapi.json", plugin.JSONSpecHandler())
mux.Handle("/openapi.yaml", plugin.YAMLSpecHandler())
```

Format detection priority:

1. URL extension (`.json`, `.yaml`, `.yml`)
2. `format` query parameter (`?format=yaml`)
3. `Accept` header (`application/yaml`, `text/yaml`)
4. Default: JSON

### Swagger UI Handler

```go
mux.Handle("/docs", plugin.SwaggerUIHandler())

// With custom options
mux.Handle("/docs", plugin.SwaggerUIHandlerWithOptions(&yahttp.SwaggerUIOptions{
    Title:   "My API Documentation",
    SpecURL: "/openapi.json",
}))
```

### ReDoc Handler

```go
mux.Handle("/redoc", plugin.RedocHandler())
```

## Standalone Functions

For simple use cases without creating a plugin:

```go
// Serve spec directly
mux.Handle("/openapi.json", yahttp.ServeSpec(spec))

// Simple handler with all features
handler := yahttp.Handler(spec, &yahttp.Options{
    SpecPath:      "/openapi.json",
    SwaggerUIPath: "/docs",
    EnableCORS:    true,
})
http.ListenAndServe(":8080", handler)
```

## Validation Errors

When validation fails, the default error handler returns JSON:

```json
{
    "error": "Validation failed",
    "details": [
        {
            "field": "id",
            "message": "required parameter is missing",
            "in": "path"
        }
    ]
}
```

Supported validations:

- Required parameters (path, query, header)
- Type validation (integer, number, boolean)
- Enum validation

## Integration with Routers

The plugin works with any router that implements `http.Handler`:

### chi

```go
r := chi.NewRouter()
r.Use(yahttp.CORS(nil))
r.Use(yahttp.Logging(log.Printf))
r.Mount("/", plugin.SpecHandler())
```

### gorilla/mux

```go
r := mux.NewRouter()
r.Use(yahttp.CORS(nil))
r.Handle("/openapi.json", plugin.SpecHandler())
```

### Standard net/http

```go
mux := http.NewServeMux()
handler := plugin.WrapMux(mux)
```
