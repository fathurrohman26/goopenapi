# YaSwag Public Packages

This directory contains public packages that can be imported and used by external projects.

## Available Packages

| Package | Import Path | Description |
|---------|-------------|-------------|
| [openapi](./openapi) | `github.com/fathurrohman26/yaswag/pkg/openapi` | OpenAPI 3.x types and schema builders |
| [yahttp](./yahttp) | `github.com/fathurrohman26/yaswag/pkg/yahttp` | HTTP middleware plugin for net/http |
| [swaggerui](./swaggerui) | `github.com/fathurrohman26/yaswag/pkg/swaggerui` | Swagger UI and Editor server |
| [output](./output) | `github.com/fathurrohman26/yaswag/pkg/output` | Output formatters (JSON/YAML) |
| [validator](./validator) | `github.com/fathurrohman26/yaswag/pkg/validator` | OpenAPI spec validation |

## Package Overview

### openapi

Core OpenAPI 3.x types including `Document`, `Info`, `PathItem`, `Operation`, `Schema`, and more. Also provides schema builder helpers.

```go
import "github.com/fathurrohman26/yaswag/pkg/openapi"

spec := &openapi.Document{
    OpenAPI: "3.0.3",
    Info: openapi.Info{
        Title:   "My API",
        Version: "1.0.0",
    },
}
```

### yahttp

HTTP middleware plugin providing Swagger UI, spec serving, CORS, logging, and request validation.

```go
import "github.com/fathurrohman26/yaswag/pkg/yahttp"

plugin := yahttp.New(spec, &yahttp.Options{
    SpecPath:      "/openapi.json",
    SwaggerUIPath: "/docs",
    EnableCORS:    true,
})
handler := plugin.WrapMux(mux)
```

See [yahttp/README.md](./yahttp/README.md) for full documentation.

### swaggerui

Standalone Swagger UI and Swagger Editor HTTP servers.

```go
import "github.com/fathurrohman26/yaswag/pkg/swaggerui"

server := swaggerui.NewServer(spec, ":8080")
server.Start()
```

### output

Output formatting for OpenAPI specs in JSON or YAML format.

```go
import "github.com/fathurrohman26/yaswag/pkg/output"

formatter := output.NewFormatter(output.JSON, 2)
jsonOutput, err := formatter.Format(spec)
```

### validator

OpenAPI specification validation with detailed error reporting.

```go
import "github.com/fathurrohman26/yaswag/pkg/validator"

v := validator.New()
result := v.Validate(spec)
if !result.Valid {
    for _, err := range result.Errors {
        log.Printf("%s: %s", err.Path, err.Message)
    }
}
```
