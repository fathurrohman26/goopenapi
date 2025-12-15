package validator

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("New() returned nil")
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  ValidationError
		want string
	}{
		{
			name: "with line and column",
			err:  ValidationError{Line: 10, Column: 5, Message: "invalid field", Path: "$.info.title"},
			want: "[10:5] invalid field (at $.info.title)",
		},
		{
			name: "with path only",
			err:  ValidationError{Message: "missing required field", Path: "$.info"},
			want: "missing required field (at $.info)",
		},
		{
			name: "message only",
			err:  ValidationError{Message: "parse error"},
			want: "parse error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		wantValid   bool
		wantVersion string
		wantErrors  int
	}{
		{
			name: "valid OpenAPI 3.0.3",
			spec: `openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths: {}`,
			wantValid:   true,
			wantVersion: "3.0.3",
			wantErrors:  0,
		},
		{
			name: "valid OpenAPI 3.1.0",
			spec: `openapi: "3.1.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}`,
			wantValid:   true,
			wantVersion: "3.1.0",
			wantErrors:  0,
		},
		{
			name: "OpenAPI 3.2 with warning",
			spec: `openapi: "3.2.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}`,
			wantValid:   true,
			wantVersion: "3.2.0",
			wantErrors:  0,
		},
		{
			name: "Swagger 2.0 unsupported",
			spec: `swagger: "2.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}`,
			wantValid:   false,
			wantVersion: "2.0",
			wantErrors:  1,
		},
		{
			name:        "invalid YAML",
			spec:        `{invalid yaml`,
			wantValid:   false,
			wantVersion: "",
			wantErrors:  1,
		},
		{
			name: "minimal valid spec",
			spec: `openapi: "3.0.3"
info:
  title: "Minimal"
  version: "1.0.0"`,
			wantValid:   true,
			wantVersion: "3.0.3",
			wantErrors:  0,
		},
	}

	v := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := v.Validate([]byte(tt.spec))
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.Version != tt.wantVersion {
				t.Errorf("Version = %q, want %q", result.Version, tt.wantVersion)
			}
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("Errors count = %d, want %d", len(result.Errors), tt.wantErrors)
			}
		})
	}
}

func TestValidator_Validate_OpenAPI32Warning(t *testing.T) {
	v := New()
	spec := `openapi: "3.2.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}`

	result, err := v.Validate([]byte(spec))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(result.Warnings) == 0 {
		t.Error("Expected warning for OpenAPI 3.2.x")
	}
	if len(result.Warnings) > 0 && !strings.Contains(result.Warnings[0].Message, "3.2") {
		t.Errorf("Warning should mention 3.2, got: %s", result.Warnings[0].Message)
	}
}

func TestValidator_ValidateFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("valid file", func(t *testing.T) {
		validSpec := `openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths: {}`
		filePath := filepath.Join(tmpDir, "valid.yaml")
		if err := os.WriteFile(filePath, []byte(validSpec), 0644); err != nil {
			t.Fatal(err)
		}

		v := New()
		result, err := v.ValidateFile(filePath)
		if err != nil {
			t.Fatalf("ValidateFile() error = %v", err)
		}
		if !result.Valid {
			t.Error("Expected valid result")
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		v := New()
		_, err := v.ValidateFile("/nonexistent/path/file.yaml")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})
}

func TestValidator_ValidateURL(t *testing.T) {
	validSpec := `openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths: {}`

	t.Run("valid URL", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			_, _ = w.Write([]byte(validSpec))
		}))
		defer server.Close()

		v := New()
		result, err := v.ValidateURL(server.URL)
		if err != nil {
			t.Fatalf("ValidateURL() error = %v", err)
		}
		if !result.Valid {
			t.Error("Expected valid result")
		}
	})

	t.Run("HTTP error status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		v := New()
		_, err := v.ValidateURL(server.URL)
		if err == nil {
			t.Error("Expected error for HTTP 404")
		}
	})

	t.Run("invalid URL", func(t *testing.T) {
		v := New()
		_, err := v.ValidateURL("http://invalid.localhost.invalid:99999")
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})
}

func TestValidator_ValidateInput(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "validator-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	validSpec := `openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths: {}`

	t.Run("file input", func(t *testing.T) {
		filePath := filepath.Join(tmpDir, "spec.yaml")
		if err := os.WriteFile(filePath, []byte(validSpec), 0644); err != nil {
			t.Fatal(err)
		}

		v := New()
		result, err := v.ValidateInput(filePath)
		if err != nil {
			t.Fatalf("ValidateInput() error = %v", err)
		}
		if !result.Valid {
			t.Error("Expected valid result")
		}
	})

	t.Run("URL input http", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(validSpec))
		}))
		defer server.Close()

		v := New()
		result, err := v.ValidateInput(server.URL)
		if err != nil {
			t.Fatalf("ValidateInput() error = %v", err)
		}
		if !result.Valid {
			t.Error("Expected valid result")
		}
	})
}

func TestFormatResult(t *testing.T) {
	t.Run("valid result", func(t *testing.T) {
		result := &ValidationResult{Valid: true, Version: "3.0.3"}
		output := FormatResult(result)
		assertContains(t, output, "Valid: true")
		assertContains(t, output, "3.0.3")
		assertContains(t, output, "specification is valid")
	})

	t.Run("result with errors", func(t *testing.T) {
		result := &ValidationResult{
			Valid: false, Version: "3.0.3",
			Errors: []ValidationError{{Message: "error 1"}, {Message: "error 2", Path: "$.info"}},
		}
		output := FormatResult(result)
		assertContains(t, output, "Valid: false")
		assertContains(t, output, "Errors (2)")
		assertContains(t, output, "error 1")
	})

	t.Run("result with warnings", func(t *testing.T) {
		result := &ValidationResult{
			Valid: true, Version: "3.2.0",
			Warnings: []ValidationError{{Message: "warning message"}},
		}
		output := FormatResult(result)
		assertContains(t, output, "Warnings (1)")
		assertContains(t, output, "warning message")
	})

	t.Run("result with errors and warnings", func(t *testing.T) {
		result := &ValidationResult{
			Valid: false, Version: "3.0.3",
			Errors:   []ValidationError{{Message: "error message", Line: 5, Column: 10, Path: "$.paths"}},
			Warnings: []ValidationError{{Message: "warning message"}},
		}
		output := FormatResult(result)
		assertContains(t, output, "Errors (1)")
		assertContains(t, output, "Warnings (1)")
	})
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("Expected %q in output", substr)
	}
}

func TestValidator_isOpenAPI3(t *testing.T) {
	v := New()

	tests := []struct {
		version string
		want    bool
	}{
		{"3.0.0", true},
		{"3.0.3", true},
		{"3.1.0", true},
		{"3.1.1", true},
		{"3.2.0", true},
		{"2.0", false},
		{"1.0", false},
		{"4.0.0", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := v.isOpenAPI3(tt.version)
			if got != tt.want {
				t.Errorf("isOpenAPI3(%q) = %v, want %v", tt.version, got, tt.want)
			}
		})
	}
}

func TestValidator_UnsupportedVersion(t *testing.T) {
	v := New()

	spec := `openapi: "4.0.0"
info:
  title: Test API
  version: "1.0.0"
paths: {}`

	result, err := v.Validate([]byte(spec))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if result.Valid {
		t.Error("Expected invalid result for unsupported version")
	}
	if len(result.Errors) == 0 {
		t.Error("Expected error for unsupported version")
	}
	if !strings.Contains(result.Errors[0].Message, "Unsupported") {
		t.Errorf("Expected unsupported version error, got: %s", result.Errors[0].Message)
	}
}
