// Package output provides formatters for OpenAPI specifications.
package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

// Format represents the output format type.
type Format string

const (
	FormatJSON Format = "json"
	FormatYAML Format = "yaml"
)

// Options configures the output formatting.
type Options struct {
	Format Format
	Indent int
	Pretty bool
}

// DefaultOptions returns default output options.
func DefaultOptions() Options {
	return Options{
		Format: FormatYAML,
		Indent: 2,
		Pretty: true,
	}
}

// Formatter handles OpenAPI document output formatting.
type Formatter struct {
	opts Options
}

// NewFormatter creates a new formatter with the given options.
func NewFormatter(opts Options) *Formatter {
	return &Formatter{opts: opts}
}

// Format formats an OpenAPI document to the configured format.
func (f *Formatter) Format(doc *openapi.Document) ([]byte, error) {
	switch f.opts.Format {
	case FormatJSON:
		return f.toJSON(doc)
	case FormatYAML:
		return f.toYAML(doc)
	default:
		return nil, fmt.Errorf("unsupported format: %s", f.opts.Format)
	}
}

// FormatTo formats an OpenAPI document and writes to the given writer.
func (f *Formatter) FormatTo(doc *openapi.Document, w io.Writer) error {
	data, err := f.Format(doc)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// FormatToFile formats an OpenAPI document and writes to a file.
func (f *Formatter) FormatToFile(doc *openapi.Document, filename string) error {
	data, err := f.Format(doc)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0644)
}

func (f *Formatter) toJSON(doc *openapi.Document) ([]byte, error) {
	if f.opts.Pretty {
		indent := strings.Repeat(" ", f.opts.Indent)
		return json.MarshalIndent(doc, "", indent)
	}
	return json.Marshal(doc)
}

func (f *Formatter) toYAML(doc *openapi.Document) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(f.opts.Indent)

	if err := encoder.Encode(doc); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ParseFormat parses a format string into a Format type.
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON, nil
	case "yaml", "yml":
		return FormatYAML, nil
	default:
		return "", fmt.Errorf("unknown format: %s (supported: json, yaml)", s)
	}
}

// DetectFormat detects the format from a filename extension.
func DetectFormat(filename string) Format {
	lower := strings.ToLower(filename)
	if strings.HasSuffix(lower, ".json") {
		return FormatJSON
	}
	return FormatYAML
}

// ToJSON converts an OpenAPI document to JSON with the given indentation.
func ToJSON(doc *openapi.Document, indent int) ([]byte, error) {
	f := NewFormatter(Options{
		Format: FormatJSON,
		Indent: indent,
		Pretty: indent > 0,
	})
	return f.Format(doc)
}

// ToYAML converts an OpenAPI document to YAML with the given indentation.
func ToYAML(doc *openapi.Document, indent int) ([]byte, error) {
	f := NewFormatter(Options{
		Format: FormatYAML,
		Indent: indent,
		Pretty: true,
	})
	return f.Format(doc)
}
