// Package validator provides OpenAPI specification validation using libopenapi.
package validator

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pb33f/libopenapi"
)

// ValidationError represents a validation error.
type ValidationError struct {
	Line    int
	Column  int
	Message string
	Path    string
}

func (e ValidationError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("[%d:%d] %s (at %s)", e.Line, e.Column, e.Message, e.Path)
	}
	if e.Path != "" {
		return fmt.Sprintf("%s (at %s)", e.Message, e.Path)
	}
	return e.Message
}

// ValidationResult holds the results of validation.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationError
	Version  string
}

// Validator validates OpenAPI specifications.
type Validator struct{}

// New creates a new Validator.
func New() *Validator {
	return &Validator{}
}

// ValidateFile validates an OpenAPI specification from a file path.
func (v *Validator) ValidateFile(path string) (*ValidationResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return v.Validate(data)
}

// ValidateURL validates an OpenAPI specification from a URL.
func (v *Validator) ValidateURL(url string) (*ValidationResult, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return v.Validate(data)
}

// Validate validates OpenAPI specification bytes.
func (v *Validator) Validate(data []byte) (*ValidationResult, error) {
	result := &ValidationResult{Valid: true}

	doc, err := libopenapi.NewDocument(data)
	if err != nil {
		return v.parseError(result, err), nil
	}

	result.Version = doc.GetVersion()
	v.validateVersion(result, doc)

	if len(result.Errors) > 0 {
		result.Valid = false
	}
	return result, nil
}

func (v *Validator) parseError(result *ValidationResult, err error) *ValidationResult {
	result.Valid = false
	result.Errors = append(result.Errors, ValidationError{
		Message: fmt.Sprintf("Failed to parse OpenAPI document: %v", err),
	})
	return result
}

func (v *Validator) validateVersion(result *ValidationResult, doc libopenapi.Document) {
	version := result.Version
	if v.isOpenAPI3(version) {
		v.validateOpenAPI3(result, doc, version)
		return
	}
	if strings.HasPrefix(version, "2") {
		v.addError(result, "Swagger 2.0 is not supported. YaSwag only supports OpenAPI 3.x (3.0, 3.1, 3.2). Please upgrade your specification.")
		return
	}
	v.addError(result, fmt.Sprintf("Unsupported OpenAPI version: %s. YaSwag only supports OpenAPI 3.x (3.0, 3.1, 3.2)", version))
}

func (v *Validator) isOpenAPI3(version string) bool {
	return strings.HasPrefix(version, "3.0") || strings.HasPrefix(version, "3.1") || strings.HasPrefix(version, "3.2")
}

func (v *Validator) validateOpenAPI3(result *ValidationResult, doc libopenapi.Document, version string) {
	model, err := doc.BuildV3Model()
	if err != nil {
		v.addError(result, fmt.Sprintf("Failed to build OpenAPI 3.x model: %v", err))
	}
	if model == nil && err == nil {
		v.addError(result, "Failed to build OpenAPI model")
	}
	if strings.HasPrefix(version, "3.2") {
		result.Warnings = append(result.Warnings, ValidationError{
			Message: "OpenAPI 3.2.x will be automatically patched to 3.1.x when served via Swagger UI (Swagger UI does not yet support 3.2)",
		})
	}
}

func (v *Validator) addError(result *ValidationResult, message string) {
	result.Valid = false
	result.Errors = append(result.Errors, ValidationError{Message: message})
}

// ValidateInput validates input from a file path or URL.
func (v *Validator) ValidateInput(input string) (*ValidationResult, error) {
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		return v.ValidateURL(input)
	}
	return v.ValidateFile(input)
}

// FormatResult formats the validation result for display.
func FormatResult(result *ValidationResult) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("OpenAPI Version: %s\n", result.Version))
	sb.WriteString(fmt.Sprintf("Valid: %t\n", result.Valid))

	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\nErrors (%d):\n", len(result.Errors)))
		for i, err := range result.Errors {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, err.Error()))
		}
	}

	if len(result.Warnings) > 0 {
		sb.WriteString(fmt.Sprintf("\nWarnings (%d):\n", len(result.Warnings)))
		for i, warn := range result.Warnings {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, warn.Error()))
		}
	}

	if result.Valid && len(result.Warnings) == 0 {
		sb.WriteString("\nThe OpenAPI specification is valid.\n")
	}

	return sb.String()
}
