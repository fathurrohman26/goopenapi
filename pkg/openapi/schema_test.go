package openapi

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestNewSchemaType(t *testing.T) {
	tests := []struct {
		input string
		want  SchemaType
	}{
		{TypeString, SchemaType{"string"}},
		{TypeInteger, SchemaType{"integer"}},
		{TypeNumber, SchemaType{"number"}},
		{TypeBoolean, SchemaType{"boolean"}},
		{TypeArray, SchemaType{"array"}},
		{TypeObject, SchemaType{"object"}},
		{TypeNull, SchemaType{"null"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := NewSchemaType(tt.input)
			if len(got) != len(tt.want) || got[0] != tt.want[0] {
				t.Errorf("NewSchemaType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSchemaType_MarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input SchemaType
		want  string
	}{
		{
			name:  "empty type",
			input: SchemaType{},
			want:  "null",
		},
		{
			name:  "single type string",
			input: SchemaType{"string"},
			want:  `"string"`,
		},
		{
			name:  "single type integer",
			input: SchemaType{"integer"},
			want:  `"integer"`,
		},
		{
			name:  "multiple types",
			input: SchemaType{"string", "null"},
			want:  `["string","null"]`,
		},
		{
			name:  "three types",
			input: SchemaType{"string", "integer", "null"},
			want:  `["string","integer","null"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.MarshalJSON()
			if err != nil {
				t.Fatalf("MarshalJSON() error = %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("MarshalJSON() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSchemaType_MarshalYAML(t *testing.T) {
	tests := []struct {
		name  string
		input SchemaType
		want  any
	}{
		{
			name:  "empty type",
			input: SchemaType{},
			want:  nil,
		},
		{
			name:  "single type",
			input: SchemaType{"string"},
			want:  "string",
		},
		{
			name:  "multiple types",
			input: SchemaType{"string", "null"},
			want:  []string{"string", "null"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.input.MarshalYAML()
			if err != nil {
				t.Fatalf("MarshalYAML() error = %v", err)
			}
			switch want := tt.want.(type) {
			case nil:
				if got != nil {
					t.Errorf("MarshalYAML() = %v, want nil", got)
				}
			case string:
				if got != want {
					t.Errorf("MarshalYAML() = %v, want %v", got, want)
				}
			case []string:
				gotSlice, ok := got.([]string)
				if !ok {
					t.Fatalf("MarshalYAML() returned %T, want []string", got)
				}
				if len(gotSlice) != len(want) {
					t.Errorf("MarshalYAML() length = %d, want %d", len(gotSlice), len(want))
				}
			}
		})
	}
}

func TestRefTo(t *testing.T) {
	schema := RefTo("User")
	if schema.Ref != "#/components/schemas/User" {
		t.Errorf("RefTo() = %q, want %q", schema.Ref, "#/components/schemas/User")
	}
}

func TestRefToResponse(t *testing.T) {
	resp := RefToResponse("NotFound")
	if resp.Ref != "#/components/responses/NotFound" {
		t.Errorf("RefToResponse() = %q, want %q", resp.Ref, "#/components/responses/NotFound")
	}
}

func TestRefToParameter(t *testing.T) {
	param := RefToParameter("PageSize")
	if param.Ref != "#/components/parameters/PageSize" {
		t.Errorf("RefToParameter() = %q, want %q", param.Ref, "#/components/parameters/PageSize")
	}
}

func TestRefToRequestBody(t *testing.T) {
	reqBody := RefToRequestBody("CreateUser")
	if reqBody.Ref != "#/components/requestBodies/CreateUser" {
		t.Errorf("RefToRequestBody() = %q, want %q", reqBody.Ref, "#/components/requestBodies/CreateUser")
	}
}

func TestStringSchema(t *testing.T) {
	schema := StringSchema()
	if len(schema.Type) != 1 || schema.Type[0] != TypeString {
		t.Errorf("StringSchema().Type = %v, want [%q]", schema.Type, TypeString)
	}
}

func TestIntegerSchema(t *testing.T) {
	schema := IntegerSchema()
	if len(schema.Type) != 1 || schema.Type[0] != TypeInteger {
		t.Errorf("IntegerSchema().Type = %v, want [%q]", schema.Type, TypeInteger)
	}
}

func TestNumberSchema(t *testing.T) {
	schema := NumberSchema()
	if len(schema.Type) != 1 || schema.Type[0] != TypeNumber {
		t.Errorf("NumberSchema().Type = %v, want [%q]", schema.Type, TypeNumber)
	}
}

func TestBooleanSchema(t *testing.T) {
	schema := BooleanSchema()
	if len(schema.Type) != 1 || schema.Type[0] != TypeBoolean {
		t.Errorf("BooleanSchema().Type = %v, want [%q]", schema.Type, TypeBoolean)
	}
}

func TestArraySchema(t *testing.T) {
	items := StringSchema()
	schema := ArraySchema(items)
	if len(schema.Type) != 1 || schema.Type[0] != TypeArray {
		t.Errorf("ArraySchema().Type = %v, want [%q]", schema.Type, TypeArray)
	}
	if schema.Items != items {
		t.Error("ArraySchema().Items does not match provided items schema")
	}
}

func TestObjectSchema(t *testing.T) {
	schema := ObjectSchema()
	if len(schema.Type) != 1 || schema.Type[0] != TypeObject {
		t.Errorf("ObjectSchema().Type = %v, want [%q]", schema.Type, TypeObject)
	}
	if schema.Properties == nil {
		t.Error("ObjectSchema().Properties should be initialized")
	}
}

func TestSchema_JSONSerialization(t *testing.T) {
	schema := &Schema{
		Type:        NewSchemaType(TypeObject),
		Description: "A test schema",
		Properties: map[string]*Schema{
			"id":   IntegerSchema(),
			"name": StringSchema(),
		},
		Required: []string{"id"},
	}

	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify JSON contains expected content
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"type":"object"`) {
		t.Error("JSON should contain type:object")
	}
	if !strings.Contains(jsonStr, `"description":"A test schema"`) {
		t.Error("JSON should contain description")
	}
	if !strings.Contains(jsonStr, `"required":["id"]`) {
		t.Error("JSON should contain required field")
	}
	if !strings.Contains(jsonStr, `"properties"`) {
		t.Error("JSON should contain properties")
	}
}

func TestSchema_YAMLSerialization(t *testing.T) {
	schema := &Schema{
		Type:        NewSchemaType(TypeString),
		Format:      "email",
		Description: "Email address",
		Example:     "test@example.com",
	}

	data, err := yaml.Marshal(schema)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	// Verify YAML contains expected content
	yamlStr := string(data)
	if !strings.Contains(yamlStr, "type: string") {
		t.Error("YAML should contain type: string")
	}
	if !strings.Contains(yamlStr, "format: email") {
		t.Error("YAML should contain format: email")
	}
	if !strings.Contains(yamlStr, "description: Email address") {
		t.Error("YAML should contain description")
	}
}

func TestSchema_WithValidation(t *testing.T) {
	minLen := int64(1)
	maxLen := int64(100)
	min := float64(0)
	max := float64(1000)

	schema := &Schema{
		Type:      NewSchemaType(TypeString),
		MinLength: &minLen,
		MaxLength: &maxLen,
		Minimum:   &min,
		Maximum:   &max,
		Pattern:   "^[a-z]+$",
	}

	if *schema.MinLength != 1 {
		t.Errorf("MinLength = %d, want 1", *schema.MinLength)
	}
	if *schema.MaxLength != 100 {
		t.Errorf("MaxLength = %d, want 100", *schema.MaxLength)
	}
	if schema.Pattern != "^[a-z]+$" {
		t.Errorf("Pattern = %q, want %q", schema.Pattern, "^[a-z]+$")
	}
}

func TestSchema_Composition(t *testing.T) {
	t.Run("allOf", func(t *testing.T) {
		schema := &Schema{
			AllOf: []*Schema{
				RefTo("Base"),
				{
					Type: NewSchemaType(TypeObject),
					Properties: map[string]*Schema{
						"extra": StringSchema(),
					},
				},
			},
		}
		if len(schema.AllOf) != 2 {
			t.Errorf("AllOf length = %d, want 2", len(schema.AllOf))
		}
	})

	t.Run("oneOf", func(t *testing.T) {
		schema := &Schema{
			OneOf: []*Schema{
				RefTo("TypeA"),
				RefTo("TypeB"),
			},
		}
		if len(schema.OneOf) != 2 {
			t.Errorf("OneOf length = %d, want 2", len(schema.OneOf))
		}
	})

	t.Run("anyOf", func(t *testing.T) {
		schema := &Schema{
			AnyOf: []*Schema{
				StringSchema(),
				IntegerSchema(),
			},
		}
		if len(schema.AnyOf) != 2 {
			t.Errorf("AnyOf length = %d, want 2", len(schema.AnyOf))
		}
	})
}

func TestSchema_Enum(t *testing.T) {
	schema := &Schema{
		Type: NewSchemaType(TypeString),
		Enum: []any{"pending", "active", "inactive"},
	}
	if len(schema.Enum) != 3 {
		t.Errorf("Enum length = %d, want 3", len(schema.Enum))
	}
}

func TestSchema_Discriminator(t *testing.T) {
	schema := &Schema{
		OneOf: []*Schema{
			RefTo("Cat"),
			RefTo("Dog"),
		},
		Discriminator: &Discriminator{
			PropertyName: "petType",
			Mapping: map[string]string{
				"cat": "#/components/schemas/Cat",
				"dog": "#/components/schemas/Dog",
			},
		},
	}
	if schema.Discriminator.PropertyName != "petType" {
		t.Errorf("Discriminator.PropertyName = %q, want %q", schema.Discriminator.PropertyName, "petType")
	}
	if len(schema.Discriminator.Mapping) != 2 {
		t.Errorf("Discriminator.Mapping length = %d, want 2", len(schema.Discriminator.Mapping))
	}
}
