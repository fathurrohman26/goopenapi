# YaSwag - Copilot Instructions

## Project Overview

YaSwag generates OpenAPI 3.x specs from Go source code using custom `!`-prefixed annotations. It includes Swagger UI server, editor, and MCP server for AI integration.

**Critical:** YaSwag uses its own annotation syntax—NOT compatible with swaggo/swag or other tools.

## Build & Test Commands

```bash
make build          # Build to ./bin/yaswag
make test           # Run all tests
go test -run TestName ./internal/parser  # Single test
make release-test   # Full CI validation (lint, test, gocyclo, fmt, vet)
```

## Architecture

### Package Structure & Data Flow

1. **`cmd/yaswag/`** → CLI entrypoint, passes version info to `internal/cli`
2. **`internal/cli/`** → Command dispatch (generate, validate, format, serve, editor, mcp)
3. **`internal/parser/`** → Go AST walker + annotation regex parser
   - `parser.go` walks directories, extracts `!`-prefixed annotations from comments
   - `annotations.go` defines regex patterns for each annotation type
4. **`pkg/openapi/`** → OpenAPI 3.x spec types mirroring the official spec
5. **`pkg/validator/`** → Uses `pb33f/libopenapi` for spec validation
6. **`pkg/mcp/`** → MCP server exposing tools like `search_endpoints`, `get_schema`
7. **`pkg/swaggerui/`** → Embedded Swagger UI with templates in `templates/`

### Annotation Syntax (Unique to YaSwag)

API-level annotations go in any comment block:

```go
// !api 3.0.3
// !info "API Title" v1.0.0 "Description"
// !server https://api.example.com "Production"
// !tag users "User operations"
// !security api_key:apiKey:header "API Key auth"
```

Operation annotations go above handler functions:

```go
// !GET /users/{id} -> getUserById "Get user by ID" #users
// !path id:integer "User ID" required
// !ok User "Success"
// !error 404 Error "Not found"
func GetUserById() {}
```

Schema annotations go above struct types:

```go
// !model "User entity"
type User struct {
    // !field id:integer "User ID" required example=123
    ID int `json:"id"`
}
```

## Key Conventions

- **Cyclomatic complexity ≤ 10** — enforced by `make gocyclo`
- **Test helpers** — use `testHelper` struct pattern (see [parser_test.go](internal/parser/parser_test.go#L11-L40))
- **OpenAPI types** — mirror spec structure in [pkg/openapi/types.go](pkg/openapi/types.go)
- **Annotation patterns** — regex-based in [internal/parser/annotations.go](internal/parser/annotations.go#L74-L140)
- **No test file parsing** — parser skips `_test.go` files and `vendor/`, `testdata/`, `.`-prefixed dirs

## Testing Patterns

Tests use temp directories and helper cleanup:

```go
h := newTestHelper(t)
defer h.cleanup()
h.writeFile("api.go", content)
p := h.parse()
```

## Dependencies

- `pb33f/libopenapi` — OpenAPI validation
- `mark3labs/mcp-go` — MCP protocol server
- `gopkg.in/yaml.v3` — YAML serialization

## Examples

See [examples/api/main.go](examples/api/main.go) for a complete Petstore API demonstrating all annotation types.
