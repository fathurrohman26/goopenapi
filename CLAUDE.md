# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

YaSwag is a Go CLI tool for generating OpenAPI 3.x specifications from Go source code using custom `!`-prefixed annotations. It includes a built-in Swagger UI server, Swagger Editor, and MCP (Model Context Protocol) server for AI assistant integration.

**Important:** YaSwag uses its own annotation syntax that is NOT compatible with swaggo/swag or other OpenAPI tools.

## Build and Development Commands

```bash
# Build the CLI binary (outputs to ./bin/yaswag)
make build

# Run all tests
make test
go test ./...

# Run a single test
go test -run TestName ./path/to/package

# Run tests in a specific package
go test ./internal/parser

# Format code
make fmt

# Run go vet
make vet

# Run linter (requires golangci-lint or staticcheck)
make lint

# Check cyclomatic complexity (max 10, requires gocyclo)
make gocyclo

# Clean build artifacts
make clean

# Full release test (lint + test + gocyclo + fmt + vet + release check)
make release-test
```

## Architecture

### Package Structure

- **`cmd/yaswag/`** - CLI entrypoint, passes version info to internal/cli
- **`internal/cli/`** - Command implementations (generate, validate, format, serve, editor, mcp)
- **`internal/parser/`** - Go AST parser and annotation processor
  - `parser.go` - Walks directories, parses Go files, extracts annotations
  - `annotations.go` - Regex-based annotation parsing for the `!` syntax
- **`pkg/openapi/`** - OpenAPI 3.x specification types (Document, Schema, Operation, etc.)
- **`pkg/mcp/`** - Model Context Protocol server for AI assistant integration
- **`pkg/output/`** - JSON/YAML output formatting
- **`pkg/validator/`** - OpenAPI spec validation
- **`pkg/swaggerui/`** - Embedded Swagger UI assets and server
- **`pkg/yahttp/`** - HTTP server utilities, middleware, CORS

### Data Flow

1. `Parser.ParseDir()` walks Go files, uses Go AST to find comment blocks
2. `AnnotationParser` extracts `!`-prefixed annotations using regex patterns
3. Annotations are converted to `SpecData` (internal) then `openapi.Document`
4. `output` package serializes to JSON/YAML
5. `validator` validates against OpenAPI 3.x spec
6. `swaggerui` serves the spec with embedded Swagger UI

### Annotation Types

API-level: `!api`, `!info`, `!contact`, `!license`, `!server`, `!tag`, `!tos`, `!security`, `!scope`, `!externalDocs`, `!link`

Operation-level: `!GET/POST/PUT/DELETE/PATCH`, `!query`, `!path`, `!header`, `!body`, `!ok`, `!error`, `!secure`

Schema-level: `!model`, `!field`

### MCP Server

The MCP server (`pkg/mcp/server.go`) exposes OpenAPI spec data to AI assistants via tools like `search_endpoints`, `get_schema`, `analyze_security`. It uses the mcp-go library for protocol handling.

## Key Conventions

- Cyclomatic complexity must stay at 10 or below (enforced by `make gocyclo`)
- Test files use helper structs like `testHelper` for setup/cleanup
- OpenAPI types mirror the spec structure in `pkg/openapi/types.go`
- Annotations are parsed with regex patterns defined in `annotations.go`
