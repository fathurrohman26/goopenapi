package output

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

func createTestDocument() *openapi.Document {
	return &openapi.Document{
		OpenAPI: "3.0.3",
		Info: openapi.Info{
			Title:       "Test API",
			Version:     "1.0.0",
			Description: "A test API",
		},
		Servers: []openapi.Server{
			{URL: "https://api.example.com"},
		},
		Paths: openapi.Paths{
			"/users": &openapi.PathItem{
				Get: &openapi.Operation{
					Summary: "List users",
					Responses: openapi.Responses{
						"200": &openapi.Response{Description: "Success"},
					},
				},
			},
		},
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.Format != FormatYAML {
		t.Errorf("Default Format = %q, want %q", opts.Format, FormatYAML)
	}
	if opts.Indent != 2 {
		t.Errorf("Default Indent = %d, want 2", opts.Indent)
	}
	if !opts.Pretty {
		t.Error("Default Pretty = false, want true")
	}
}

func TestNewFormatter(t *testing.T) {
	opts := Options{
		Format: FormatJSON,
		Indent: 4,
		Pretty: true,
	}

	f := NewFormatter(opts)
	if f == nil {
		t.Fatal("NewFormatter() returned nil")
	}
	if f.opts.Format != FormatJSON {
		t.Errorf("Formatter.opts.Format = %q, want %q", f.opts.Format, FormatJSON)
	}
}

func TestFormatter_Format_JSON(t *testing.T) {
	doc := createTestDocument()

	t.Run("pretty JSON", func(t *testing.T) {
		f := NewFormatter(Options{
			Format: FormatJSON,
			Indent: 2,
			Pretty: true,
		})

		data, err := f.Format(doc)
		if err != nil {
			t.Fatalf("Format() error = %v", err)
		}

		// Verify it's valid JSON
		var decoded openapi.Document
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Verify it's indented
		if !strings.Contains(string(data), "\n") {
			t.Error("Pretty JSON should contain newlines")
		}
	})

	t.Run("compact JSON", func(t *testing.T) {
		f := NewFormatter(Options{
			Format: FormatJSON,
			Indent: 0,
			Pretty: false,
		})

		data, err := f.Format(doc)
		if err != nil {
			t.Fatalf("Format() error = %v", err)
		}

		// Verify it's valid JSON
		var decoded openapi.Document
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid JSON output: %v", err)
		}

		// Compact JSON should be single line
		lines := strings.Split(string(data), "\n")
		if len(lines) > 1 {
			t.Error("Compact JSON should be single line")
		}
	})
}

func TestFormatter_Format_YAML(t *testing.T) {
	doc := createTestDocument()

	f := NewFormatter(Options{
		Format: FormatYAML,
		Indent: 2,
		Pretty: true,
	})

	data, err := f.Format(doc)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify it's valid YAML
	var decoded openapi.Document
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Invalid YAML output: %v", err)
	}

	if decoded.OpenAPI != doc.OpenAPI {
		t.Errorf("OpenAPI = %q, want %q", decoded.OpenAPI, doc.OpenAPI)
	}
	if decoded.Info.Title != doc.Info.Title {
		t.Errorf("Info.Title = %q, want %q", decoded.Info.Title, doc.Info.Title)
	}
}

func TestFormatter_Format_UnsupportedFormat(t *testing.T) {
	f := NewFormatter(Options{
		Format: Format("xml"),
	})

	_, err := f.Format(createTestDocument())
	if err == nil {
		t.Error("Expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("Error should mention unsupported format, got: %v", err)
	}
}

func TestFormatter_FormatTo(t *testing.T) {
	doc := createTestDocument()
	f := NewFormatter(Options{
		Format: FormatJSON,
		Indent: 2,
		Pretty: true,
	})

	var buf bytes.Buffer
	err := f.FormatTo(doc, &buf)
	if err != nil {
		t.Fatalf("FormatTo() error = %v", err)
	}

	// Verify output
	var decoded openapi.Document
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("Invalid JSON output: %v", err)
	}
}

func TestFormatter_FormatToFile(t *testing.T) {
	doc := createTestDocument()
	tmpDir, err := os.MkdirTemp("", "output-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	t.Run("JSON file", func(t *testing.T) {
		f := NewFormatter(Options{
			Format: FormatJSON,
			Indent: 2,
			Pretty: true,
		})

		filename := filepath.Join(tmpDir, "spec.json")
		err := f.FormatToFile(doc, filename)
		if err != nil {
			t.Fatalf("FormatToFile() error = %v", err)
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		var decoded openapi.Document
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid JSON in file: %v", err)
		}
	})

	t.Run("YAML file", func(t *testing.T) {
		f := NewFormatter(Options{
			Format: FormatYAML,
			Indent: 2,
			Pretty: true,
		})

		filename := filepath.Join(tmpDir, "spec.yaml")
		err := f.FormatToFile(doc, filename)
		if err != nil {
			t.Fatalf("FormatToFile() error = %v", err)
		}

		data, err := os.ReadFile(filename)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		var decoded openapi.Document
		if err := yaml.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid YAML in file: %v", err)
		}
	})
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		input   string
		want    Format
		wantErr bool
	}{
		{"json", FormatJSON, false},
		{"JSON", FormatJSON, false},
		{"Json", FormatJSON, false},
		{"yaml", FormatYAML, false},
		{"YAML", FormatYAML, false},
		{"yml", FormatYAML, false},
		{"YML", FormatYAML, false},
		{"xml", "", true},
		{"", "", true},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseFormat(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFormat(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseFormat(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		filename string
		want     Format
	}{
		{"spec.json", FormatJSON},
		{"spec.JSON", FormatJSON},
		{"openapi.json", FormatJSON},
		{"spec.yaml", FormatYAML},
		{"spec.YAML", FormatYAML},
		{"spec.yml", FormatYAML},
		{"spec.YML", FormatYAML},
		{"spec.txt", FormatYAML}, // Default to YAML
		{"spec", FormatYAML},     // No extension, default to YAML
		{"/path/to/spec.json", FormatJSON},
		{"/path/to/spec.yaml", FormatYAML},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := DetectFormat(tt.filename)
			if got != tt.want {
				t.Errorf("DetectFormat(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

func TestToJSON(t *testing.T) {
	doc := createTestDocument()

	t.Run("with indentation", func(t *testing.T) {
		data, err := ToJSON(doc, 2)
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}

		var decoded openapi.Document
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		// Check it's formatted
		if !strings.Contains(string(data), "\n") {
			t.Error("Expected formatted JSON with newlines")
		}
	})

	t.Run("without indentation", func(t *testing.T) {
		data, err := ToJSON(doc, 0)
		if err != nil {
			t.Fatalf("ToJSON() error = %v", err)
		}

		var decoded openapi.Document
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid JSON: %v", err)
		}

		// Should be compact
		lines := strings.Split(string(data), "\n")
		if len(lines) > 1 {
			t.Error("Expected compact JSON without newlines")
		}
	})
}

func TestToYAML(t *testing.T) {
	doc := createTestDocument()

	t.Run("default indentation", func(t *testing.T) {
		data, err := ToYAML(doc, 2)
		if err != nil {
			t.Fatalf("ToYAML() error = %v", err)
		}

		var decoded openapi.Document
		if err := yaml.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid YAML: %v", err)
		}

		if decoded.OpenAPI != doc.OpenAPI {
			t.Errorf("OpenAPI = %q, want %q", decoded.OpenAPI, doc.OpenAPI)
		}
	})

	t.Run("custom indentation", func(t *testing.T) {
		data, err := ToYAML(doc, 4)
		if err != nil {
			t.Fatalf("ToYAML() error = %v", err)
		}

		var decoded openapi.Document
		if err := yaml.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Invalid YAML: %v", err)
		}
	})
}

func TestFormat_Constants(t *testing.T) {
	if FormatJSON != "json" {
		t.Errorf("FormatJSON = %q, want %q", FormatJSON, "json")
	}
	if FormatYAML != "yaml" {
		t.Errorf("FormatYAML = %q, want %q", FormatYAML, "yaml")
	}
}

func TestFormatter_Format_ComplexDocument(t *testing.T) {
	doc := createComplexTestDocument()

	t.Run("JSON format", func(t *testing.T) {
		verifyComplexDocJSON(t, doc)
	})

	t.Run("YAML format", func(t *testing.T) {
		verifyComplexDocYAML(t, doc)
	})
}

func createComplexTestDocument() *openapi.Document {
	return &openapi.Document{
		OpenAPI: "3.1.0",
		Info: openapi.Info{
			Title: "Complex API", Version: "2.0.0", Description: "A complex API with all features",
			Contact: &openapi.Contact{Name: "Support", Email: "support@example.com"},
			License: &openapi.License{Name: "MIT"},
		},
		Servers: []openapi.Server{
			{URL: "https://api.example.com/v2", Description: "Production"},
			{URL: "https://staging.example.com/v2", Description: "Staging"},
		},
		Paths: openapi.Paths{"/users": createUsersPathItem()},
		Components: &openapi.Components{
			Schemas: map[string]*openapi.Schema{"User": createUserSchema()},
		},
		Tags: []openapi.Tag{{Name: "users", Description: "User operations"}},
	}
}

func createUsersPathItem() *openapi.PathItem {
	return &openapi.PathItem{
		Get: &openapi.Operation{
			Tags: []string{"users"}, Summary: "List users", OperationID: "listUsers",
			Parameters: []*openapi.Parameter{{Name: "limit", In: openapi.ParameterInQuery, Schema: openapi.IntegerSchema()}},
			Responses:  openapi.Responses{"200": {Description: "Success"}},
		},
		Post: &openapi.Operation{
			Tags: []string{"users"}, Summary: "Create user", OperationID: "createUser",
			RequestBody: &openapi.RequestBody{Required: true},
			Responses:   openapi.Responses{"201": {Description: "Created"}, "400": {Description: "Bad Request"}},
		},
	}
}

func createUserSchema() *openapi.Schema {
	return &openapi.Schema{
		Type: openapi.NewSchemaType(openapi.TypeObject),
		Properties: map[string]*openapi.Schema{
			"id": openapi.IntegerSchema(), "name": openapi.StringSchema(), "email": openapi.StringSchema(),
		},
		Required: []string{"id", "name"},
	}
}

func verifyComplexDocJSON(t *testing.T, doc *openapi.Document) {
	t.Helper()
	f := NewFormatter(Options{Format: FormatJSON, Indent: 2, Pretty: true})
	data, err := f.Format(doc)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	var generic map[string]any
	if err := json.Unmarshal(data, &generic); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	jsonStr := string(data)
	assertContainsStr(t, jsonStr, `"openapi": "3.1.0"`)
	assertContainsStr(t, jsonStr, `"contact"`)
	assertContainsStr(t, jsonStr, `"servers"`)
	assertContainsStr(t, jsonStr, `"components"`)
}

func verifyComplexDocYAML(t *testing.T, doc *openapi.Document) {
	t.Helper()
	f := NewFormatter(Options{Format: FormatYAML, Indent: 2, Pretty: true})
	data, err := f.Format(doc)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	yamlStr := string(data)
	if !strings.Contains(yamlStr, "openapi: \"3.1.0\"") && !strings.Contains(yamlStr, "openapi: 3.1.0") {
		t.Error("YAML should contain openapi version")
	}
	assertContainsStr(t, yamlStr, "title: Complex API")
}

func assertContainsStr(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("Expected %q in output", substr)
	}
}
