package cli

import (
	"bytes"
	"encoding/json"
	"strings"

	"gopkg.in/yaml.v3"
)

// yamlUnmarshal unmarshals YAML data into the provided interface.
func yamlUnmarshal(data []byte, v any) error {
	return yaml.Unmarshal(data, v)
}

// jsonMarshalIndent marshals data to JSON with indentation.
func jsonMarshalIndent(v any, indent int) ([]byte, error) {
	indentStr := strings.Repeat(" ", indent)
	return json.MarshalIndent(v, "", indentStr)
}

// yamlMarshalIndent marshals data to YAML with indentation.
func yamlMarshalIndent(v any, indent int) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(indent)

	if err := encoder.Encode(v); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
