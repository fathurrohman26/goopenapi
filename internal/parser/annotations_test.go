package parser

import (
	"reflect"
	"testing"
)

func TestAnnotationParser_Parse(t *testing.T) {
	p := NewAnnotationParser()

	tests := []struct {
		name     string
		input    string
		expected []Annotation
	}{
		{
			name:  "parse api annotation",
			input: `!api 3.0.3`,
			expected: []Annotation{
				{Type: AnnotationAPI, RawLine: "!api 3.0.3", Args: map[string]string{"version": "3.0.3"}},
			},
		},
		{
			name:  "parse api annotation with v prefix",
			input: `!api v3.1.0`,
			expected: []Annotation{
				{Type: AnnotationAPI, RawLine: "!api v3.1.0", Args: map[string]string{"version": "3.1.0"}},
			},
		},
		{
			name:  "parse info annotation",
			input: `!info "My API" v1.0.0 "This is a sample API"`,
			expected: []Annotation{
				{Type: AnnotationInfo, RawLine: `!info "My API" v1.0.0 "This is a sample API"`, Args: map[string]string{"title": "My API", "version": "1.0.0", "description": "This is a sample API"}},
			},
		},
		{
			name:  "parse info annotation without description",
			input: `!info "My API" 1.0.0`,
			expected: []Annotation{
				{Type: AnnotationInfo, RawLine: `!info "My API" 1.0.0`, Args: map[string]string{"title": "My API", "version": "1.0.0", "description": ""}},
			},
		},
		{
			name:  "parse contact annotation",
			input: `!contact "API Support" <support@example.com> (https://example.com)`,
			expected: []Annotation{
				{Type: AnnotationContact, RawLine: `!contact "API Support" <support@example.com> (https://example.com)`, Args: map[string]string{"name": "API Support", "email": "support@example.com", "url": "https://example.com"}},
			},
		},
		{
			name:  "parse contact annotation without url",
			input: `!contact "API Support" <support@example.com>`,
			expected: []Annotation{
				{Type: AnnotationContact, RawLine: `!contact "API Support" <support@example.com>`, Args: map[string]string{"name": "API Support", "email": "support@example.com", "url": ""}},
			},
		},
		{
			name:  "parse license annotation",
			input: `!license MIT https://opensource.org/licenses/MIT`,
			expected: []Annotation{
				{Type: AnnotationLicense, RawLine: `!license MIT https://opensource.org/licenses/MIT`, Args: map[string]string{"name": "MIT", "url": "https://opensource.org/licenses/MIT"}},
			},
		},
		{
			name:  "parse server annotation",
			input: `!server https://api.example.com/v1 "Production server"`,
			expected: []Annotation{
				{Type: AnnotationServer, RawLine: `!server https://api.example.com/v1 "Production server"`, Args: map[string]string{"url": "https://api.example.com/v1", "description": "Production server"}},
			},
		},
		{
			name:  "parse tag annotation",
			input: `!tag users "Operations about users"`,
			expected: []Annotation{
				{Type: AnnotationTag, RawLine: `!tag users "Operations about users"`, Args: map[string]string{"name": "users", "description": "Operations about users"}},
			},
		},
		{
			name:  "parse GET route annotation",
			input: `!GET /users -> getUsers "Retrieve users" #users #admin`,
			expected: []Annotation{
				{Type: AnnotationRoute, RawLine: `!GET /users -> getUsers "Retrieve users" #users #admin`, Args: map[string]string{"method": "GET", "path": "/users", "operationId": "getUsers", "summary": "Retrieve users"}, Tags: []string{"users", "admin"}},
			},
		},
		{
			name:  "parse POST route annotation",
			input: `!POST /users -> createUser "Create a user" #users`,
			expected: []Annotation{
				{Type: AnnotationRoute, RawLine: `!POST /users -> createUser "Create a user" #users`, Args: map[string]string{"method": "POST", "path": "/users", "operationId": "createUser", "summary": "Create a user"}, Tags: []string{"users"}},
			},
		},
		{
			name:  "parse query parameter annotation",
			input: `!query limit:integer "The number of results" default=10 required`,
			expected: []Annotation{
				{Type: AnnotationQuery, RawLine: `!query limit:integer "The number of results" default=10 required`, Args: map[string]string{"in": "query", "name": "limit", "type": "integer", "description": "The number of results", "required": "true", "default": "10"}},
			},
		},
		{
			name:  "parse path parameter annotation",
			input: `!path id:integer "The user ID" required`,
			expected: []Annotation{
				{Type: AnnotationPath, RawLine: `!path id:integer "The user ID" required`, Args: map[string]string{"in": "path", "name": "id", "type": "integer", "description": "The user ID", "required": "true"}},
			},
		},
		{
			name:  "parse header parameter annotation",
			input: `!header X-Token:string "Authorization token"`,
			expected: []Annotation{
				{Type: AnnotationHeader, RawLine: `!header X-Token:string "Authorization token"`, Args: map[string]string{"in": "header", "name": "X-Token", "type": "string", "description": "Authorization token"}},
			},
		},
		{
			name:  "parse body annotation",
			input: `!body CreateUserRequest "User data" required`,
			expected: []Annotation{
				{Type: AnnotationBody, RawLine: `!body CreateUserRequest "User data" required`, Args: map[string]string{"schema": "CreateUserRequest", "description": "User data", "required": "true"}},
			},
		},
		{
			name:  "parse ok response annotation with default status",
			input: `!ok User "Successful response"`,
			expected: []Annotation{
				{Type: AnnotationOK, RawLine: `!ok User "Successful response"`, Args: map[string]string{"status": "200", "schema": "User", "description": "Successful response"}},
			},
		},
		{
			name:  "parse ok response annotation with custom status",
			input: `!ok 201 User "User created"`,
			expected: []Annotation{
				{Type: AnnotationOK, RawLine: `!ok 201 User "User created"`, Args: map[string]string{"status": "201", "schema": "User", "description": "User created"}},
			},
		},
		{
			name:  "parse error response annotation",
			input: `!error 404 ErrorResponse "Not found"`,
			expected: []Annotation{
				{Type: AnnotationError, RawLine: `!error 404 ErrorResponse "Not found"`, Args: map[string]string{"status": "404", "schema": "ErrorResponse", "description": "Not found"}},
			},
		},
		{
			name:  "parse error response annotation with default status",
			input: `!error ErrorResponse "Server error"`,
			expected: []Annotation{
				{Type: AnnotationError, RawLine: `!error ErrorResponse "Server error"`, Args: map[string]string{"status": "500", "schema": "ErrorResponse", "description": "Server error"}},
			},
		},
		{
			name:  "parse model annotation",
			input: `!model "A user entity"`,
			expected: []Annotation{
				{Type: AnnotationModel, RawLine: `!model "A user entity"`, Args: map[string]string{"description": "A user entity"}},
			},
		},
		{
			name:  "parse model annotation without description",
			input: `!model`,
			expected: []Annotation{
				{Type: AnnotationModel, RawLine: `!model`, Args: map[string]string{"description": ""}},
			},
		},
		{
			name:  "parse field annotation",
			input: `!field id:integer "User ID" required example=123`,
			expected: []Annotation{
				{Type: AnnotationField, RawLine: `!field id:integer "User ID" required example=123`, Args: map[string]string{"name": "id", "type": "integer", "description": "User ID", "required": "true", "example": "123"}},
			},
		},
		{
			name:  "parse field annotation with quoted example",
			input: `!field name:string "User name" example="John Doe"`,
			expected: []Annotation{
				{Type: AnnotationField, RawLine: `!field name:string "User name" example="John Doe"`, Args: map[string]string{"name": "name", "type": "string", "description": "User name", "example": "John Doe"}},
			},
		},
		{
			name: "parse multiple annotations",
			input: `!GET /users -> getUsers "Get users" #users
!query limit:integer "Limit results"
!ok User[] "Success"`,
			expected: []Annotation{
				{Type: AnnotationRoute, RawLine: `!GET /users -> getUsers "Get users" #users`, Args: map[string]string{"method": "GET", "path": "/users", "operationId": "getUsers", "summary": "Get users"}, Tags: []string{"users"}},
				{Type: AnnotationQuery, RawLine: `!query limit:integer "Limit results"`, Args: map[string]string{"in": "query", "name": "limit", "type": "integer", "description": "Limit results"}},
				{Type: AnnotationOK, RawLine: `!ok User[] "Success"`, Args: map[string]string{"status": "200", "schema": "User[]", "description": "Success"}},
			},
		},
		{
			name:     "no annotations",
			input:    "This is just a comment without annotations",
			expected: nil,
		},
		{
			name:     "mixed content with annotations",
			input:    "Some description\n!api 3.0.3\nMore text",
			expected: []Annotation{{Type: AnnotationAPI, RawLine: "!api 3.0.3", Args: map[string]string{"version": "3.0.3"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := p.Parse(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Parse() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func TestGetAPI(t *testing.T) {
	a := Annotation{Type: AnnotationAPI, Args: map[string]string{"version": "3.0.3"}}
	api := GetAPI(a)
	if api.Version != "3.0.3" {
		t.Errorf("Version = %v, want %v", api.Version, "3.0.3")
	}
}

func TestGetInfo(t *testing.T) {
	a := Annotation{Type: AnnotationInfo, Args: map[string]string{"title": "My API", "version": "1.0.0", "description": "Description"}}
	info := GetInfo(a)
	if info.Title != "My API" {
		t.Errorf("Title = %v, want %v", info.Title, "My API")
	}
	if info.Version != "1.0.0" {
		t.Errorf("Version = %v, want %v", info.Version, "1.0.0")
	}
	if info.Description != "Description" {
		t.Errorf("Description = %v, want %v", info.Description, "Description")
	}
}

func TestGetContact(t *testing.T) {
	a := Annotation{Type: AnnotationContact, Args: map[string]string{"name": "Support", "email": "support@example.com", "url": "https://example.com"}}
	contact := GetContact(a)
	if contact.Name != "Support" {
		t.Errorf("Name = %v, want %v", contact.Name, "Support")
	}
	if contact.Email != "support@example.com" {
		t.Errorf("Email = %v, want %v", contact.Email, "support@example.com")
	}
	if contact.URL != "https://example.com" {
		t.Errorf("URL = %v, want %v", contact.URL, "https://example.com")
	}
}

func TestGetLicense(t *testing.T) {
	a := Annotation{Type: AnnotationLicense, Args: map[string]string{"name": "MIT", "url": "https://opensource.org/licenses/MIT"}}
	license := GetLicense(a)
	if license.Name != "MIT" {
		t.Errorf("Name = %v, want %v", license.Name, "MIT")
	}
	if license.URL != "https://opensource.org/licenses/MIT" {
		t.Errorf("URL = %v, want %v", license.URL, "https://opensource.org/licenses/MIT")
	}
}

func TestGetServer(t *testing.T) {
	a := Annotation{Type: AnnotationServer, Args: map[string]string{"url": "https://api.example.com", "description": "Production"}}
	server := GetServer(a)
	if server.URL != "https://api.example.com" {
		t.Errorf("URL = %v, want %v", server.URL, "https://api.example.com")
	}
	if server.Description != "Production" {
		t.Errorf("Description = %v, want %v", server.Description, "Production")
	}
}

func TestGetTag(t *testing.T) {
	a := Annotation{Type: AnnotationTag, Args: map[string]string{"name": "users", "description": "User operations"}}
	tag := GetTag(a)
	if tag.Name != "users" {
		t.Errorf("Name = %v, want %v", tag.Name, "users")
	}
	if tag.Description != "User operations" {
		t.Errorf("Description = %v, want %v", tag.Description, "User operations")
	}
}

func TestGetRoute(t *testing.T) {
	a := Annotation{Type: AnnotationRoute, Args: map[string]string{"method": "GET", "path": "/users", "operationId": "getUsers", "summary": "Get all users"}, Tags: []string{"users", "admin"}}
	route := GetRoute(a)
	if route.Method != "GET" {
		t.Errorf("Method = %v, want %v", route.Method, "GET")
	}
	if route.Path != "/users" {
		t.Errorf("Path = %v, want %v", route.Path, "/users")
	}
	if route.OperationID != "getUsers" {
		t.Errorf("OperationID = %v, want %v", route.OperationID, "getUsers")
	}
	if route.Summary != "Get all users" {
		t.Errorf("Summary = %v, want %v", route.Summary, "Get all users")
	}
	if len(route.Tags) != 2 || route.Tags[0] != "users" || route.Tags[1] != "admin" {
		t.Errorf("Tags = %v, want %v", route.Tags, []string{"users", "admin"})
	}
}

func TestGetParam(t *testing.T) {
	a := Annotation{Type: AnnotationQuery, Args: map[string]string{"in": "query", "name": "limit", "type": "integer", "description": "Limit results", "required": "true", "default": "10"}}
	param := GetParam(a)
	if param.In != "query" {
		t.Errorf("In = %v, want %v", param.In, "query")
	}
	if param.Name != "limit" {
		t.Errorf("Name = %v, want %v", param.Name, "limit")
	}
	if param.Type != "integer" {
		t.Errorf("Type = %v, want %v", param.Type, "integer")
	}
	if param.Description != "Limit results" {
		t.Errorf("Description = %v, want %v", param.Description, "Limit results")
	}
	if !param.Required {
		t.Errorf("Required = %v, want %v", param.Required, true)
	}
	if param.Default != "10" {
		t.Errorf("Default = %v, want %v", param.Default, "10")
	}
}

func TestGetBody(t *testing.T) {
	a := Annotation{Type: AnnotationBody, Args: map[string]string{"schema": "CreateUser", "description": "User data", "required": "true"}}
	body := GetBody(a)
	if body.Schema != "CreateUser" {
		t.Errorf("Schema = %v, want %v", body.Schema, "CreateUser")
	}
	if body.Description != "User data" {
		t.Errorf("Description = %v, want %v", body.Description, "User data")
	}
	if !body.Required {
		t.Errorf("Required = %v, want %v", body.Required, true)
	}
}

func TestGetResponse(t *testing.T) {
	a := Annotation{Type: AnnotationOK, Args: map[string]string{"status": "200", "schema": "User", "description": "Success"}}
	resp := GetResponse(a)
	if resp.Status != "200" {
		t.Errorf("Status = %v, want %v", resp.Status, "200")
	}
	if resp.Schema != "User" {
		t.Errorf("Schema = %v, want %v", resp.Schema, "User")
	}
	if resp.Description != "Success" {
		t.Errorf("Description = %v, want %v", resp.Description, "Success")
	}
	if resp.IsError {
		t.Errorf("IsError = %v, want %v", resp.IsError, false)
	}
}

func TestGetResponseError(t *testing.T) {
	a := Annotation{Type: AnnotationError, Args: map[string]string{"status": "404", "schema": "Error", "description": "Not found"}}
	resp := GetResponse(a)
	if resp.Status != "404" {
		t.Errorf("Status = %v, want %v", resp.Status, "404")
	}
	if !resp.IsError {
		t.Errorf("IsError = %v, want %v", resp.IsError, true)
	}
}

func TestGetModel(t *testing.T) {
	a := Annotation{Type: AnnotationModel, Args: map[string]string{"description": "A user entity"}}
	model := GetModel(a)
	if model.Description != "A user entity" {
		t.Errorf("Description = %v, want %v", model.Description, "A user entity")
	}
}

func TestGetField(t *testing.T) {
	a := Annotation{Type: AnnotationField, Args: map[string]string{"name": "id", "type": "integer", "description": "User ID", "required": "true", "example": "123"}}
	field := GetField(a)
	if field.Name != "id" {
		t.Errorf("Name = %v, want %v", field.Name, "id")
	}
	if field.Type != "integer" {
		t.Errorf("Type = %v, want %v", field.Type, "integer")
	}
	if field.Description != "User ID" {
		t.Errorf("Description = %v, want %v", field.Description, "User ID")
	}
	if !field.Required {
		t.Errorf("Required = %v, want %v", field.Required, true)
	}
	if field.Example != "123" {
		t.Errorf("Example = %v, want %v", field.Example, "123")
	}
}

func TestParseValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected any
	}{
		{"integer value", "123", int64(123)},
		{"float value", "19.99", float64(19.99)},
		{"string value", "John Doe", "John Doe"},
		{"boolean true", "true", true},
		{"boolean false", "false", false},
		{"quoted string", `"hello"`, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseValue(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("parseValue(%q) = %v (%T), want %v (%T)", tt.input, result, result, tt.expected, tt.expected)
			}
		})
	}
}
