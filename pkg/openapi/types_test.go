package openapi

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestDocument_JSONSerialization(t *testing.T) {
	doc := &Document{
		OpenAPI: "3.0.3",
		Info: Info{
			Title:       "Test API",
			Version:     "1.0.0",
			Description: "A test API",
		},
		Servers: []Server{
			{URL: "https://api.example.com", Description: "Production"},
		},
		Paths: Paths{
			"/users": &PathItem{
				Get: &Operation{
					Summary:     "List users",
					OperationID: "listUsers",
					Responses: Responses{
						"200": &Response{Description: "Success"},
					},
				},
			},
		},
		Tags: []Tag{
			{Name: "users", Description: "User operations"},
		},
	}

	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Document
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.OpenAPI != doc.OpenAPI {
		t.Errorf("OpenAPI = %q, want %q", decoded.OpenAPI, doc.OpenAPI)
	}
	if decoded.Info.Title != doc.Info.Title {
		t.Errorf("Info.Title = %q, want %q", decoded.Info.Title, doc.Info.Title)
	}
	if len(decoded.Servers) != 1 {
		t.Errorf("Servers length = %d, want 1", len(decoded.Servers))
	}
	if len(decoded.Paths) != 1 {
		t.Errorf("Paths length = %d, want 1", len(decoded.Paths))
	}
	if len(decoded.Tags) != 1 {
		t.Errorf("Tags length = %d, want 1", len(decoded.Tags))
	}
}

func TestDocument_YAMLSerialization(t *testing.T) {
	doc := &Document{
		OpenAPI: "3.1.0",
		Info: Info{
			Title:   "YAML Test API",
			Version: "2.0.0",
			Contact: &Contact{
				Name:  "Support",
				Email: "support@example.com",
				URL:   "https://example.com/support",
			},
			License: &License{
				Name: "MIT",
				URL:  "https://opensource.org/licenses/MIT",
			},
		},
	}

	data, err := yaml.Marshal(doc)
	if err != nil {
		t.Fatalf("yaml.Marshal() error = %v", err)
	}

	var decoded Document
	if err := yaml.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("yaml.Unmarshal() error = %v", err)
	}

	if decoded.OpenAPI != doc.OpenAPI {
		t.Errorf("OpenAPI = %q, want %q", decoded.OpenAPI, doc.OpenAPI)
	}
	if decoded.Info.Contact == nil {
		t.Fatal("Info.Contact should not be nil")
	}
	if decoded.Info.Contact.Email != doc.Info.Contact.Email {
		t.Errorf("Contact.Email = %q, want %q", decoded.Info.Contact.Email, doc.Info.Contact.Email)
	}
	if decoded.Info.License == nil {
		t.Fatal("Info.License should not be nil")
	}
	if decoded.Info.License.Name != doc.Info.License.Name {
		t.Errorf("License.Name = %q, want %q", decoded.Info.License.Name, doc.Info.License.Name)
	}
}

func TestInfo_Complete(t *testing.T) {
	info := Info{
		Title:          "Complete API",
		Summary:        "A complete API example",
		Description:    "Full description",
		TermsOfService: "https://example.com/terms",
		Version:        "1.0.0",
		Contact: &Contact{
			Name:  "API Support",
			URL:   "https://example.com/support",
			Email: "api@example.com",
		},
		License: &License{
			Name:       "Apache 2.0",
			Identifier: "Apache-2.0",
			URL:        "https://www.apache.org/licenses/LICENSE-2.0",
		},
	}

	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Info
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Summary != info.Summary {
		t.Errorf("Summary = %q, want %q", decoded.Summary, info.Summary)
	}
	if decoded.TermsOfService != info.TermsOfService {
		t.Errorf("TermsOfService = %q, want %q", decoded.TermsOfService, info.TermsOfService)
	}
}

func TestServer_WithVariables(t *testing.T) {
	server := Server{
		URL:         "https://{environment}.example.com/v{version}",
		Description: "Server with variables",
		Variables: map[string]ServerVariable{
			"environment": {
				Enum:        []string{"dev", "staging", "prod"},
				Default:     "prod",
				Description: "Environment",
			},
			"version": {
				Default:     "1",
				Description: "API version",
			},
		},
	}

	data, err := json.Marshal(server)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Server
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if len(decoded.Variables) != 2 {
		t.Errorf("Variables length = %d, want 2", len(decoded.Variables))
	}
	envVar := decoded.Variables["environment"]
	if envVar.Default != "prod" {
		t.Errorf("environment.Default = %q, want %q", envVar.Default, "prod")
	}
	if len(envVar.Enum) != 3 {
		t.Errorf("environment.Enum length = %d, want 3", len(envVar.Enum))
	}
}

func TestPathItem_AllMethods(t *testing.T) {
	pathItem := &PathItem{
		Summary:     "Resource operations",
		Description: "All HTTP methods",
		Get:         &Operation{Summary: "Get"},
		Put:         &Operation{Summary: "Put"},
		Post:        &Operation{Summary: "Post"},
		Delete:      &Operation{Summary: "Delete"},
		Options:     &Operation{Summary: "Options"},
		Head:        &Operation{Summary: "Head"},
		Patch:       &Operation{Summary: "Patch"},
		Trace:       &Operation{Summary: "Trace"},
	}

	data, err := json.Marshal(pathItem)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded PathItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	verifyOperation(t, "Get", decoded.Get, "Get")
	verifyOperation(t, "Put", decoded.Put, "Put")
	verifyOperation(t, "Post", decoded.Post, "Post")
	verifyOperation(t, "Delete", decoded.Delete, "Delete")
	verifyOperation(t, "Options", decoded.Options, "Options")
	verifyOperation(t, "Head", decoded.Head, "Head")
	verifyOperation(t, "Patch", decoded.Patch, "Patch")
	verifyOperation(t, "Trace", decoded.Trace, "Trace")
}

func verifyOperation(t *testing.T, name string, op *Operation, expectedSummary string) {
	t.Helper()
	if op == nil || op.Summary != expectedSummary {
		t.Errorf("%s operation not properly decoded", name)
	}
}

func TestOperation_Complete(t *testing.T) {
	op := &Operation{
		Tags:        []string{"users", "admin"},
		Summary:     "Create user",
		Description: "Create a new user",
		OperationID: "createUser",
		Deprecated:  false,
		Parameters: []*Parameter{
			{
				Name:        "X-Request-ID",
				In:          ParameterInHeader,
				Description: "Request ID",
				Required:    false,
				Schema:      StringSchema(),
			},
		},
		RequestBody: &RequestBody{
			Description: "User data",
			Required:    true,
			Content: map[string]MediaType{
				"application/json": {
					Schema: RefTo("CreateUserRequest"),
				},
			},
		},
		Responses: Responses{
			"201": &Response{
				Description: "Created",
				Content: map[string]MediaType{
					"application/json": {
						Schema: RefTo("User"),
					},
				},
			},
			"400": &Response{Description: "Bad Request"},
		},
		Security: []SecurityRequirement{
			{"bearerAuth": {}},
		},
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify JSON contains expected content
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"tags":["users","admin"]`) {
		t.Error("JSON should contain tags")
	}
	if !strings.Contains(jsonStr, `"operationId":"createUser"`) {
		t.Error("JSON should contain operationId")
	}
	if !strings.Contains(jsonStr, `"parameters"`) {
		t.Error("JSON should contain parameters")
	}
	if !strings.Contains(jsonStr, `"requestBody"`) {
		t.Error("JSON should contain requestBody")
	}
	if !strings.Contains(jsonStr, `"responses"`) {
		t.Error("JSON should contain responses")
	}
}

func TestParameter_Locations(t *testing.T) {
	tests := []struct {
		location ParameterLocation
		want     string
	}{
		{ParameterInQuery, "query"},
		{ParameterInHeader, "header"},
		{ParameterInPath, "path"},
		{ParameterInCookie, "cookie"},
	}

	for _, tt := range tests {
		t.Run(string(tt.location), func(t *testing.T) {
			if string(tt.location) != tt.want {
				t.Errorf("ParameterLocation = %q, want %q", tt.location, tt.want)
			}
		})
	}
}

func TestComponents_Complete(t *testing.T) {
	components := &Components{
		Schemas: map[string]*Schema{
			"User": ObjectSchema(),
		},
		Responses: map[string]*Response{
			"NotFound": {Description: "Not found"},
		},
		Parameters: map[string]*Parameter{
			"PageSize": {Name: "pageSize", In: ParameterInQuery},
		},
		RequestBodies: map[string]*RequestBody{
			"CreateUser": {Description: "Create user body"},
		},
		SecuritySchemes: map[string]*SecurityScheme{
			"bearerAuth": {
				Type:         "http",
				Scheme:       "bearer",
				BearerFormat: "JWT",
			},
		},
	}

	data, err := json.Marshal(components)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Verify JSON contains expected content
	jsonStr := string(data)
	if !strings.Contains(jsonStr, `"schemas"`) {
		t.Error("JSON should contain schemas")
	}
	if !strings.Contains(jsonStr, `"User"`) {
		t.Error("JSON should contain User schema")
	}
	if !strings.Contains(jsonStr, `"responses"`) {
		t.Error("JSON should contain responses")
	}
	if !strings.Contains(jsonStr, `"parameters"`) {
		t.Error("JSON should contain parameters")
	}
	if !strings.Contains(jsonStr, `"securitySchemes"`) {
		t.Error("JSON should contain securitySchemes")
	}
	if !strings.Contains(jsonStr, `"bearerAuth"`) {
		t.Error("JSON should contain bearerAuth")
	}
}

func TestSecurityScheme_Types(t *testing.T) {
	t.Run("api key", func(t *testing.T) {
		scheme := &SecurityScheme{
			Type: "apiKey",
			Name: "X-API-Key",
			In:   "header",
		}
		data, err := json.Marshal(scheme)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		var decoded SecurityScheme
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if decoded.Type != "apiKey" {
			t.Errorf("Type = %q, want %q", decoded.Type, "apiKey")
		}
	})

	t.Run("oauth2", func(t *testing.T) {
		scheme := &SecurityScheme{
			Type: "oauth2",
			Flows: &OAuthFlows{
				AuthorizationCode: &OAuthFlow{
					AuthorizationURL: "https://example.com/oauth/authorize",
					TokenURL:         "https://example.com/oauth/token",
					Scopes: map[string]string{
						"read":  "Read access",
						"write": "Write access",
					},
				},
			},
		}
		data, err := json.Marshal(scheme)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		var decoded SecurityScheme
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}
		if decoded.Flows == nil || decoded.Flows.AuthorizationCode == nil {
			t.Error("OAuth flows not properly decoded")
		}
	})
}

func TestExternalDocumentation(t *testing.T) {
	extDocs := &ExternalDocumentation{
		Description: "Find more info here",
		URL:         "https://example.com/docs",
	}

	data, err := json.Marshal(extDocs)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ExternalDocumentation
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.URL != extDocs.URL {
		t.Errorf("URL = %q, want %q", decoded.URL, extDocs.URL)
	}
}

func TestTag(t *testing.T) {
	tag := Tag{
		Name:        "users",
		Description: "User management",
		ExternalDocs: &ExternalDocumentation{
			URL: "https://example.com/docs/users",
		},
	}

	data, err := json.Marshal(tag)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Tag
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Name != tag.Name {
		t.Errorf("Name = %q, want %q", decoded.Name, tag.Name)
	}
	if decoded.ExternalDocs == nil {
		t.Error("ExternalDocs should not be nil")
	}
}

func TestMediaType_WithExample(t *testing.T) {
	mediaType := MediaType{
		Schema:  RefTo("User"),
		Example: map[string]any{"id": 1, "name": "John"},
		Examples: map[string]*Example{
			"basic": {
				Summary: "Basic example",
				Value:   map[string]any{"id": 1, "name": "John"},
			},
		},
	}

	data, err := json.Marshal(mediaType)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded MediaType
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Schema == nil {
		t.Error("Schema should not be nil")
	}
	if decoded.Example == nil {
		t.Error("Example should not be nil")
	}
	if len(decoded.Examples) != 1 {
		t.Errorf("Examples length = %d, want 1", len(decoded.Examples))
	}
}
