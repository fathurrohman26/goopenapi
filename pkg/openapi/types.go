// Package openapi provides types for OpenAPI 3.x specification.
package openapi

// Document represents the root OpenAPI 3.x document.
// https://spec.openapis.org/oas/v3.1.0#openapi-object
type Document struct {
	OpenAPI      string                 `json:"openapi" yaml:"openapi"`
	Info         Info                   `json:"info" yaml:"info"`
	Servers      []Server               `json:"servers,omitempty" yaml:"servers,omitempty"`
	Paths        Paths                  `json:"paths,omitempty" yaml:"paths,omitempty"`
	Webhooks     map[string]*PathItem   `json:"webhooks,omitempty" yaml:"webhooks,omitempty"`
	Components   *Components            `json:"components,omitempty" yaml:"components,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty" yaml:"security,omitempty"`
	Tags         []Tag                  `json:"tags,omitempty" yaml:"tags,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Info provides metadata about the API.
// https://spec.openapis.org/oas/v3.1.0#info-object
type Info struct {
	Title          string   `json:"title" yaml:"title"`
	Summary        string   `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	TermsOfService string   `json:"termsOfService,omitempty" yaml:"termsOfService,omitempty"`
	Contact        *Contact `json:"contact,omitempty" yaml:"contact,omitempty"`
	License        *License `json:"license,omitempty" yaml:"license,omitempty"`
	Version        string   `json:"version" yaml:"version"`
}

// Contact provides contact information for the API.
// https://spec.openapis.org/oas/v3.1.0#contact-object
type Contact struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	URL   string `json:"url,omitempty" yaml:"url,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// License provides license information for the API.
// https://spec.openapis.org/oas/v3.1.0#license-object
type License struct {
	Name       string `json:"name" yaml:"name"`
	Identifier string `json:"identifier,omitempty" yaml:"identifier,omitempty"`
	URL        string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Server represents a server.
// https://spec.openapis.org/oas/v3.1.0#server-object
type Server struct {
	URL         string                    `json:"url" yaml:"url"`
	Description string                    `json:"description,omitempty" yaml:"description,omitempty"`
	Variables   map[string]ServerVariable `json:"variables,omitempty" yaml:"variables,omitempty"`
}

// ServerVariable represents a server variable for URL template substitution.
// https://spec.openapis.org/oas/v3.1.0#server-variable-object
type ServerVariable struct {
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:"description,omitempty"`
}

// Paths holds the relative paths to the individual endpoints and their operations.
// https://spec.openapis.org/oas/v3.1.0#paths-object
type Paths map[string]*PathItem

// PathItem describes the operations available on a single path.
// https://spec.openapis.org/oas/v3.1.0#path-item-object
type PathItem struct {
	Ref         string       `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string       `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description string       `json:"description,omitempty" yaml:"description,omitempty"`
	Get         *Operation   `json:"get,omitempty" yaml:"get,omitempty"`
	Put         *Operation   `json:"put,omitempty" yaml:"put,omitempty"`
	Post        *Operation   `json:"post,omitempty" yaml:"post,omitempty"`
	Delete      *Operation   `json:"delete,omitempty" yaml:"delete,omitempty"`
	Options     *Operation   `json:"options,omitempty" yaml:"options,omitempty"`
	Head        *Operation   `json:"head,omitempty" yaml:"head,omitempty"`
	Patch       *Operation   `json:"patch,omitempty" yaml:"patch,omitempty"`
	Trace       *Operation   `json:"trace,omitempty" yaml:"trace,omitempty"`
	Servers     []Server     `json:"servers,omitempty" yaml:"servers,omitempty"`
	Parameters  []*Parameter `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

// Operation describes a single API operation on a path.
// https://spec.openapis.org/oas/v3.1.0#operation-object
type Operation struct {
	Tags         []string               `json:"tags,omitempty" yaml:"tags,omitempty"`
	Summary      string                 `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	OperationID  string                 `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   []*Parameter           `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  *RequestBody           `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Responses    Responses              `json:"responses,omitempty" yaml:"responses,omitempty"`
	Callbacks    map[string]*Callback   `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	Deprecated   bool                   `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	Security     []SecurityRequirement  `json:"security,omitempty" yaml:"security,omitempty"`
	Servers      []Server               `json:"servers,omitempty" yaml:"servers,omitempty"`
}

// ExternalDocumentation allows referencing an external resource for extended documentation.
// https://spec.openapis.org/oas/v3.1.0#external-documentation-object
type ExternalDocumentation struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	URL         string `json:"url" yaml:"url"`
}

// Parameter describes a single operation parameter.
// https://spec.openapis.org/oas/v3.1.0#parameter-object
type Parameter struct {
	Ref             string               `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Name            string               `json:"name,omitempty" yaml:"name,omitempty"`
	In              ParameterLocation    `json:"in,omitempty" yaml:"in,omitempty"`
	Description     string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         *bool                `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any                  `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]*Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// ParameterLocation represents where a parameter is expected.
type ParameterLocation string

const (
	ParameterInQuery  ParameterLocation = "query"
	ParameterInHeader ParameterLocation = "header"
	ParameterInPath   ParameterLocation = "path"
	ParameterInCookie ParameterLocation = "cookie"
)

// RequestBody describes a single request body.
// https://spec.openapis.org/oas/v3.1.0#request-body-object
type RequestBody struct {
	Ref         string               `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description string               `json:"description,omitempty" yaml:"description,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Required    bool                 `json:"required,omitempty" yaml:"required,omitempty"`
}

// MediaType provides schema and examples for the media type identified by its key.
// https://spec.openapis.org/oas/v3.1.0#media-type-object
type MediaType struct {
	Schema   *Schema             `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example  any                 `json:"example,omitempty" yaml:"example,omitempty"`
	Examples map[string]*Example `json:"examples,omitempty" yaml:"examples,omitempty"`
	Encoding map[string]Encoding `json:"encoding,omitempty" yaml:"encoding,omitempty"`
}

// Encoding describes a single encoding definition applied to a single schema property.
// https://spec.openapis.org/oas/v3.1.0#encoding-object
type Encoding struct {
	ContentType   string             `json:"contentType,omitempty" yaml:"contentType,omitempty"`
	Headers       map[string]*Header `json:"headers,omitempty" yaml:"headers,omitempty"`
	Style         string             `json:"style,omitempty" yaml:"style,omitempty"`
	Explode       bool               `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved bool               `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
}

// Responses is a container for the expected responses of an operation.
// https://spec.openapis.org/oas/v3.1.0#responses-object
type Responses map[string]*Response

// Response describes a single response from an API Operation.
// https://spec.openapis.org/oas/v3.1.0#response-object
type Response struct {
	Ref         string               `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description string               `json:"description" yaml:"description"`
	Headers     map[string]*Header   `json:"headers,omitempty" yaml:"headers,omitempty"`
	Content     map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
	Links       map[string]*Link     `json:"links,omitempty" yaml:"links,omitempty"`
}

// Header follows the structure of the Parameter Object with some differences.
// https://spec.openapis.org/oas/v3.1.0#header-object
type Header struct {
	Ref             string               `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Description     string               `json:"description,omitempty" yaml:"description,omitempty"`
	Required        bool                 `json:"required,omitempty" yaml:"required,omitempty"`
	Deprecated      bool                 `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	AllowEmptyValue bool                 `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	Style           string               `json:"style,omitempty" yaml:"style,omitempty"`
	Explode         bool                 `json:"explode,omitempty" yaml:"explode,omitempty"`
	AllowReserved   bool                 `json:"allowReserved,omitempty" yaml:"allowReserved,omitempty"`
	Schema          *Schema              `json:"schema,omitempty" yaml:"schema,omitempty"`
	Example         any                  `json:"example,omitempty" yaml:"example,omitempty"`
	Examples        map[string]*Example  `json:"examples,omitempty" yaml:"examples,omitempty"`
	Content         map[string]MediaType `json:"content,omitempty" yaml:"content,omitempty"`
}

// Callback is a map of possible out-of-band callbacks related to the parent operation.
// https://spec.openapis.org/oas/v3.1.0#callback-object
type Callback map[string]*PathItem

// Link represents a possible design-time link for a response.
// https://spec.openapis.org/oas/v3.1.0#link-object
type Link struct {
	Ref          string         `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	OperationRef string         `json:"operationRef,omitempty" yaml:"operationRef,omitempty"`
	OperationID  string         `json:"operationId,omitempty" yaml:"operationId,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	RequestBody  any            `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	Description  string         `json:"description,omitempty" yaml:"description,omitempty"`
	Server       *Server        `json:"server,omitempty" yaml:"server,omitempty"`
}

// Example describes an example of a media type.
// https://spec.openapis.org/oas/v3.1.0#example-object
type Example struct {
	Ref           string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary       string `json:"summary,omitempty" yaml:"summary,omitempty"`
	Description   string `json:"description,omitempty" yaml:"description,omitempty"`
	Value         any    `json:"value,omitempty" yaml:"value,omitempty"`
	ExternalValue string `json:"externalValue,omitempty" yaml:"externalValue,omitempty"`
}

// Tag adds metadata to a single tag that is used by the Operation Object.
// https://spec.openapis.org/oas/v3.1.0#tag-object
type Tag struct {
	Name         string                 `json:"name" yaml:"name"`
	Description  string                 `json:"description,omitempty" yaml:"description,omitempty"`
	ExternalDocs *ExternalDocumentation `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}

// Components holds a set of reusable objects for different aspects of the OAS.
// https://spec.openapis.org/oas/v3.1.0#components-object
type Components struct {
	Schemas         map[string]*Schema         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Responses       map[string]*Response       `json:"responses,omitempty" yaml:"responses,omitempty"`
	Parameters      map[string]*Parameter      `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Examples        map[string]*Example        `json:"examples,omitempty" yaml:"examples,omitempty"`
	RequestBodies   map[string]*RequestBody    `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Headers         map[string]*Header         `json:"headers,omitempty" yaml:"headers,omitempty"`
	SecuritySchemes map[string]*SecurityScheme `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Links           map[string]*Link           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       map[string]*Callback       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
	PathItems       map[string]*PathItem       `json:"pathItems,omitempty" yaml:"pathItems,omitempty"`
}

// SecurityScheme defines a security scheme that can be used by the operations.
// https://spec.openapis.org/oas/v3.1.0#security-scheme-object
type SecurityScheme struct {
	Ref              string      `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Type             string      `json:"type,omitempty" yaml:"type,omitempty"`
	Description      string      `json:"description,omitempty" yaml:"description,omitempty"`
	Name             string      `json:"name,omitempty" yaml:"name,omitempty"`
	In               string      `json:"in,omitempty" yaml:"in,omitempty"`
	Scheme           string      `json:"scheme,omitempty" yaml:"scheme,omitempty"`
	BearerFormat     string      `json:"bearerFormat,omitempty" yaml:"bearerFormat,omitempty"`
	Flows            *OAuthFlows `json:"flows,omitempty" yaml:"flows,omitempty"`
	OpenIDConnectURL string      `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`
}

// OAuthFlows allows configuration of the supported OAuth Flows.
// https://spec.openapis.org/oas/v3.1.0#oauth-flows-object
type OAuthFlows struct {
	Implicit          *OAuthFlow `json:"implicit,omitempty" yaml:"implicit,omitempty"`
	Password          *OAuthFlow `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials *OAuthFlow `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
	AuthorizationCode *OAuthFlow `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
}

// OAuthFlow provides configuration details for a supported OAuth Flow.
// https://spec.openapis.org/oas/v3.1.0#oauth-flow-object
type OAuthFlow struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	RefreshURL       string            `json:"refreshUrl,omitempty" yaml:"refreshUrl,omitempty"`
	Scopes           map[string]string `json:"scopes,omitempty" yaml:"scopes,omitempty"`
}

// SecurityRequirement lists the required security schemes to execute this operation.
// https://spec.openapis.org/oas/v3.1.0#security-requirement-object
type SecurityRequirement map[string][]string
