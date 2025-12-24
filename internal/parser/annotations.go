// Package parser provides a custom annotation parser for YaSwag.
// YaSwag uses its own eccentric annotation syntax that is NOT compatible with swag or other tools.
package parser

import (
	"regexp"
	"strconv"
	"strings"
)

// Common annotation argument values.
const (
	argTrue = "true"
)

// AnnotationType represents the type of YaSwag annotation.
type AnnotationType string

const (
	// API-level annotations
	AnnotationAPI          AnnotationType = "api"          // !api 3.0.3
	AnnotationInfo         AnnotationType = "info"         // !info "Title" v1.0.0 "Description"
	AnnotationContact      AnnotationType = "contact"      // !contact "Name" <email> (url)
	AnnotationLicense      AnnotationType = "license"      // !license MIT https://...
	AnnotationServer       AnnotationType = "server"       // !server https://... "Description"
	AnnotationTag          AnnotationType = "tag"          // !tag users "Description"
	AnnotationTOS          AnnotationType = "tos"          // !tos https://example.com/tos
	AnnotationSecurity     AnnotationType = "security"     // !security apiKey:header:api_key "API Key Auth"
	AnnotationScope        AnnotationType = "scope"        // !scope petstore_auth write:pets "modify pets in your account"
	AnnotationExternalDocs AnnotationType = "externalDocs" // !externalDocs https://... "Description"
	AnnotationLink         AnnotationType = "link"         // !link "Label" https://...

	// Webhook annotations (OpenAPI 3.1+)
	AnnotationWebhook         AnnotationType = "webhook"          // !webhook name:method "Description"
	AnnotationWebhookBody     AnnotationType = "webhook-body"     // !webhook-body SchemaRef "description" required
	AnnotationWebhookResponse AnnotationType = "webhook-response" // !webhook-response 200 SchemaRef "description"

	// Operation annotations
	AnnotationRoute  AnnotationType = "route"  // !GET /path -> operationId "summary" #tag1 #tag2
	AnnotationQuery  AnnotationType = "query"  // !query name:type "description" default=value required
	AnnotationPath   AnnotationType = "path"   // !path id:integer "description" required
	AnnotationHeader AnnotationType = "header" // !header X-Token:string "description"
	AnnotationBody   AnnotationType = "body"   // !body SchemaRef "description" required
	AnnotationOK     AnnotationType = "ok"     // !ok SchemaRef "description" or !ok 201 SchemaRef "description"
	AnnotationError  AnnotationType = "error"  // !error 404 SchemaRef "description"
	AnnotationSecure AnnotationType = "secure" // !secure api_key oauth2

	// Schema annotations
	AnnotationModel AnnotationType = "model" // !model "Description"
	AnnotationField AnnotationType = "field" // !field name:type "description" required example=value
)

// Annotation represents a parsed YaSwag annotation.
type Annotation struct {
	Type    AnnotationType
	RawLine string
	Args    map[string]string
	Tags    []string
}

// AnnotationParser parses YaSwag's eccentric annotation syntax.
type AnnotationParser struct {
	// Patterns for different annotation types
	apiPattern          *regexp.Regexp
	infoPattern         *regexp.Regexp
	contactPattern      *regexp.Regexp
	licensePattern      *regexp.Regexp
	serverPattern       *regexp.Regexp
	tagPattern          *regexp.Regexp
	tosPattern          *regexp.Regexp
	securityPattern     *regexp.Regexp
	scopePattern        *regexp.Regexp
	externalDocsPattern *regexp.Regexp
	linkPattern         *regexp.Regexp
	webhookPattern      *regexp.Regexp
	webhookBodyPattern  *regexp.Regexp
	webhookRespPattern  *regexp.Regexp
	routePattern        *regexp.Regexp
	paramPattern        *regexp.Regexp
	bodyPattern         *regexp.Regexp
	responsePattern     *regexp.Regexp
	securePattern       *regexp.Regexp
	modelPattern        *regexp.Regexp
	fieldPattern        *regexp.Regexp
}

// NewAnnotationParser creates a new annotation parser for YaSwag's eccentric syntax.
func NewAnnotationParser() *AnnotationParser {
	return &AnnotationParser{
		// !api 3.0.3 or !api v3.0.3
		apiPattern: regexp.MustCompile(`^!api\s+v?([\d.]+)`),

		// !info "Title" v1.0.0 "Description"
		infoPattern: regexp.MustCompile(`^!info\s+"([^"]+)"\s+v?([\d.]+)(?:\s+"([^"]*)")?`),

		// !contact "Name" <email> or !contact "Name" <email> (url) or !contact "" <email>
		contactPattern: regexp.MustCompile(`^!contact\s+"([^"]*)"(?:\s+<([^>]+)>)?(?:\s+\(([^)]+)\))?`),

		// !license Name URL
		licensePattern: regexp.MustCompile(`^!license\s+(\S+)(?:\s+(\S+))?`),

		// !server URL "Description"
		serverPattern: regexp.MustCompile(`^!server\s+(\S+)(?:\s+"([^"]*)")?`),

		// !tag name "description"
		tagPattern: regexp.MustCompile(`^!tag\s+(\S+)(?:\s+"([^"]*)")?`),

		// !tos URL
		tosPattern: regexp.MustCompile(`^!tos\s+(\S+)`),

		// !security name:type:location "description"
		// Examples:
		//   !security api_key:apiKey:header "API Key authentication"
		//   !security petstore_auth:oauth2:implicit "OAuth2 authentication" https://petstore3.swagger.io/oauth/authorize
		securityPattern: regexp.MustCompile(`^!security\s+(\w+):(apiKey|oauth2|http|openIdConnect):?(\w*)(?:\s+"([^"]*)")?(?:\s+(\S+))?`),

		// !scope security_name scope_name "description"
		// Example: !scope petstore_auth write:pets "modify pets in your account"
		scopePattern: regexp.MustCompile(`^!scope\s+(\w+)\s+([\w:]+)(?:\s+"([^"]*)")?`),

		// !externalDocs URL "Description"
		// Example: !externalDocs https://swagger.io "Find out more about Swagger"
		externalDocsPattern: regexp.MustCompile(`^!externalDocs\s+(\S+)(?:\s+"([^"]*)")?`),

		// !link "Label" URL
		// Example: !link "The Pet Store repository" https://github.com/swagger-api/swagger-petstore
		linkPattern: regexp.MustCompile(`^!link\s+"([^"]+)"\s+(\S+)`),

		// !webhook name:method "Description"
		// Example: !webhook newOrder:post "New order notification"
		// Methods: get, post, put, delete, patch
		webhookPattern: regexp.MustCompile(`^!webhook\s+(\w+):(get|post|put|delete|patch)\s*(?:"([^"]*)")?`),

		// !webhook-body SchemaRef "description" required
		// Example: !webhook-body OrderWebhook "Order data" required
		webhookBodyPattern: regexp.MustCompile(`^!webhook-body\s+(\S+)(?:\s+"([^"]*)")?`),

		// !webhook-response status SchemaRef "description"
		// Example: !webhook-response 200 SuccessResponse "Webhook received"
		webhookRespPattern: regexp.MustCompile(`^!webhook-response\s+(\d+)\s+(\S+)(?:\s+"([^"]*)")?`),

		// !GET /path -> operationId "summary" #tag1 #tag2
		// !POST /path -> operationId "summary" #tag
		routePattern: regexp.MustCompile(`^!(GET|POST|PUT|DELETE|PATCH|OPTIONS|HEAD)\s+(\S+)\s+->\s+(\S+)(?:\s+"([^"]*)")?`),

		// !query name:type "description" default=value required
		// !path id:integer "description" required
		// !header X-Token:string "description"
		paramPattern: regexp.MustCompile(`^!(query|path|header|cookie)\s+([\w-]+):(\w+)\??\s*(?:"([^"]*)")?`),

		// !body SchemaRef "description" required
		bodyPattern: regexp.MustCompile(`^!body\s+(\S+)(?:\s+"([^"]*)")?`),

		// !ok SchemaRef "description" or !ok 201 SchemaRef "description"
		// !error 404 SchemaRef "description"
		responsePattern: regexp.MustCompile(`^!(ok|error)\s+(?:(\d+)\s+)?(\S+)(?:\s+"([^"]*)")?`),

		// !secure securityName1 securityName2
		securePattern: regexp.MustCompile(`^!secure\s+(.+)`),

		// !model "Description"
		modelPattern: regexp.MustCompile(`^!model(?:\s+"([^"]*)")?`),

		// !field name:type "description" required example=value
		fieldPattern: regexp.MustCompile(`^!field\s+(\w+):(\w+)\??\s*(?:"([^"]*)")?`),
	}
}

// Parse extracts all YaSwag annotations from comment text.
func (p *AnnotationParser) Parse(text string) []Annotation {
	var annotations []Annotation

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "!") {
			continue
		}

		if a := p.parseLine(line); a != nil {
			annotations = append(annotations, *a)
		}
	}

	return annotations
}

func (p *AnnotationParser) parseLine(line string) *Annotation {
	if a := p.parseSimplePatterns(line); a != nil {
		return a
	}
	if a := p.parseRoutePattern(line); a != nil {
		return a
	}
	if a := p.parseParamPattern(line); a != nil {
		return a
	}
	if a := p.parseBodyPattern(line); a != nil {
		return a
	}
	if a := p.parseResponsePattern(line); a != nil {
		return a
	}
	if a := p.parseSecurePattern(line); a != nil {
		return a
	}
	if a := p.parseWebhookPattern(line); a != nil {
		return a
	}
	if a := p.parseModelPattern(line); a != nil {
		return a
	}
	return p.parseFieldPattern(line)
}

func (p *AnnotationParser) parseSimplePatterns(line string) *Annotation {
	// Simple patterns that just extract matched groups
	matchers := []struct {
		pattern *regexp.Regexp
		aType   AnnotationType
		keys    []string
	}{
		{p.apiPattern, AnnotationAPI, []string{"version"}},
		{p.infoPattern, AnnotationInfo, []string{"title", "version", "description"}},
		{p.contactPattern, AnnotationContact, []string{"name", "email", "url"}},
		{p.licensePattern, AnnotationLicense, []string{"name", "url"}},
		{p.serverPattern, AnnotationServer, []string{"url", "description"}},
		{p.tagPattern, AnnotationTag, []string{"name", "description"}},
		{p.tosPattern, AnnotationTOS, []string{"url"}},
		{p.securityPattern, AnnotationSecurity, []string{"name", "type", "location", "description", "url"}},
		{p.scopePattern, AnnotationScope, []string{"security", "name", "description"}},
		{p.externalDocsPattern, AnnotationExternalDocs, []string{"url", "description"}},
		{p.linkPattern, AnnotationLink, []string{"label", "url"}},
	}

	for _, m := range matchers {
		if match := m.pattern.FindStringSubmatch(line); match != nil {
			args := make(map[string]string)
			for i, key := range m.keys {
				if i+1 < len(match) {
					args[key] = match[i+1]
				}
			}
			return &Annotation{Type: m.aType, RawLine: line, Args: args}
		}
	}
	return nil
}

func (p *AnnotationParser) parseRoutePattern(line string) *Annotation {
	match := p.routePattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	return &Annotation{
		Type:    AnnotationRoute,
		RawLine: line,
		Args: map[string]string{
			"method":      strings.ToUpper(match[1]),
			"path":        match[2],
			"operationId": match[3],
			"summary":     match[4],
		},
		Tags: extractTags(line),
	}
}

func (p *AnnotationParser) parseParamPattern(line string) *Annotation {
	match := p.paramPattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	args := map[string]string{
		"in":          match[1],
		"name":        match[2],
		"type":        match[3],
		"description": match[4],
	}
	if strings.Contains(line, " required") {
		args["required"] = argTrue
	}
	if defMatch := regexp.MustCompile(`default=(\S+)`).FindStringSubmatch(line); defMatch != nil {
		args["default"] = strings.Trim(defMatch[1], `"'`)
	}

	aType := AnnotationQuery
	switch match[1] {
	case "path":
		aType = AnnotationPath
	case "header":
		aType = AnnotationHeader
	}
	return &Annotation{Type: aType, RawLine: line, Args: args}
}

func (p *AnnotationParser) parseBodyPattern(line string) *Annotation {
	match := p.bodyPattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	args := map[string]string{"schema": match[1], "description": match[2]}
	if strings.Contains(line, " required") {
		args["required"] = argTrue
	}
	return &Annotation{Type: AnnotationBody, RawLine: line, Args: args}
}

func (p *AnnotationParser) parseResponsePattern(line string) *Annotation {
	match := p.responsePattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	statusCode, schema := match[2], match[3]
	if match[1] == "ok" && statusCode == "" {
		statusCode = "200"
	}
	if match[1] == "error" && statusCode == "" {
		statusCode = "500"
	}
	aType := AnnotationOK
	if match[1] == "error" {
		aType = AnnotationError
	}
	return &Annotation{
		Type:    aType,
		RawLine: line,
		Args:    map[string]string{"status": statusCode, "schema": schema, "description": match[4]},
	}
}

func (p *AnnotationParser) parseSecurePattern(line string) *Annotation {
	match := p.securePattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	names := strings.Fields(match[1])
	return &Annotation{
		Type:    AnnotationSecure,
		RawLine: line,
		Args:    map[string]string{"names": strings.Join(names, ",")},
		Tags:    names,
	}
}

func (p *AnnotationParser) parseWebhookPattern(line string) *Annotation {
	// Try webhook pattern
	if match := p.webhookPattern.FindStringSubmatch(line); match != nil {
		return &Annotation{
			Type:    AnnotationWebhook,
			RawLine: line,
			Args: map[string]string{
				"name":        match[1],
				"method":      match[2],
				"description": match[3],
			},
		}
	}

	// Try webhook-body pattern
	if match := p.webhookBodyPattern.FindStringSubmatch(line); match != nil {
		return &Annotation{
			Type:    AnnotationWebhookBody,
			RawLine: line,
			Args: map[string]string{
				"schema":      match[1],
				"description": match[2],
			},
		}
	}

	// Try webhook-response pattern
	if match := p.webhookRespPattern.FindStringSubmatch(line); match != nil {
		return &Annotation{
			Type:    AnnotationWebhookResponse,
			RawLine: line,
			Args: map[string]string{
				"status":      match[1],
				"schema":      match[2],
				"description": match[3],
			},
		}
	}

	return nil
}

func (p *AnnotationParser) parseModelPattern(line string) *Annotation {
	match := p.modelPattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	return &Annotation{Type: AnnotationModel, RawLine: line, Args: map[string]string{"description": match[1]}}
}

func (p *AnnotationParser) parseFieldPattern(line string) *Annotation {
	match := p.fieldPattern.FindStringSubmatch(line)
	if match == nil {
		return nil
	}
	args := map[string]string{"name": match[1], "type": match[2], "description": match[3]}
	if strings.Contains(line, " required") {
		args["required"] = argTrue
	}
	if exMatch := regexp.MustCompile(`example=("[^"]*"|\S+)`).FindStringSubmatch(line); exMatch != nil {
		args["example"] = strings.Trim(exMatch[1], `"'`)
	}
	return &Annotation{Type: AnnotationField, RawLine: line, Args: args}
}

// extractTags extracts hashtag-style tags from a line (e.g., #users #admin)
func extractTags(line string) []string {
	var tags []string
	tagPattern := regexp.MustCompile(`#(\w+)`)
	matches := tagPattern.FindAllStringSubmatch(line, -1)
	for _, match := range matches {
		tags = append(tags, match[1])
	}
	return tags
}

// Helper functions for parsed data

// ParsedAPI holds parsed !api data.
type ParsedAPI struct {
	Version string
}

// GetAPI extracts API info from annotation.
func GetAPI(a Annotation) ParsedAPI {
	return ParsedAPI{Version: a.Args["version"]}
}

// ParsedInfo holds parsed !info data.
type ParsedInfo struct {
	Title       string
	Version     string
	Description string
}

// GetInfo extracts info from annotation.
func GetInfo(a Annotation) ParsedInfo {
	return ParsedInfo{
		Title:       a.Args["title"],
		Version:     a.Args["version"],
		Description: a.Args["description"],
	}
}

// ParsedContact holds parsed !contact data.
type ParsedContact struct {
	Name  string
	Email string
	URL   string
}

// GetContact extracts contact from annotation.
func GetContact(a Annotation) ParsedContact {
	return ParsedContact{
		Name:  a.Args["name"],
		Email: a.Args["email"],
		URL:   a.Args["url"],
	}
}

// ParsedLicense holds parsed !license data.
type ParsedLicense struct {
	Name string
	URL  string
}

// GetLicense extracts license from annotation.
func GetLicense(a Annotation) ParsedLicense {
	return ParsedLicense{
		Name: a.Args["name"],
		URL:  a.Args["url"],
	}
}

// ParsedServer holds parsed !server data.
type ParsedServer struct {
	URL         string
	Description string
}

// GetServer extracts server from annotation.
func GetServer(a Annotation) ParsedServer {
	return ParsedServer{
		URL:         a.Args["url"],
		Description: a.Args["description"],
	}
}

// ParsedTag holds parsed !tag data.
type ParsedTag struct {
	Name        string
	Description string
}

// GetTag extracts tag from annotation.
func GetTag(a Annotation) ParsedTag {
	return ParsedTag{
		Name:        a.Args["name"],
		Description: a.Args["description"],
	}
}

// ParsedRoute holds parsed route (!GET, !POST, etc.) data.
type ParsedRoute struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Tags        []string
}

// GetRoute extracts route from annotation.
func GetRoute(a Annotation) ParsedRoute {
	return ParsedRoute{
		Method:      a.Args["method"],
		Path:        a.Args["path"],
		OperationID: a.Args["operationId"],
		Summary:     a.Args["summary"],
		Tags:        a.Tags,
	}
}

// ParsedParam holds parsed parameter (!query, !path, !header) data.
type ParsedParam struct {
	In          string
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
}

// GetParam extracts parameter from annotation.
func GetParam(a Annotation) ParsedParam {
	return ParsedParam{
		In:          a.Args["in"],
		Name:        a.Args["name"],
		Type:        a.Args["type"],
		Description: a.Args["description"],
		Required:    a.Args["required"] == argTrue,
		Default:     a.Args["default"],
	}
}

// ParsedBody holds parsed !body data.
type ParsedBody struct {
	Schema      string
	Description string
	Required    bool
}

// GetBody extracts body from annotation.
func GetBody(a Annotation) ParsedBody {
	return ParsedBody{
		Schema:      a.Args["schema"],
		Description: a.Args["description"],
		Required:    a.Args["required"] == argTrue,
	}
}

// ParsedResponse holds parsed response (!ok, !error) data.
type ParsedResponse struct {
	Status      string
	Schema      string
	Description string
	IsError     bool
}

// GetResponse extracts response from annotation.
func GetResponse(a Annotation) ParsedResponse {
	return ParsedResponse{
		Status:      a.Args["status"],
		Schema:      a.Args["schema"],
		Description: a.Args["description"],
		IsError:     a.Type == AnnotationError,
	}
}

// ParsedModel holds parsed !model data.
type ParsedModel struct {
	Description string
}

// GetModel extracts model from annotation.
func GetModel(a Annotation) ParsedModel {
	return ParsedModel{
		Description: a.Args["description"],
	}
}

// ParsedField holds parsed !field data.
type ParsedField struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Example     string
}

// GetField extracts field from annotation.
func GetField(a Annotation) ParsedField {
	return ParsedField{
		Name:        a.Args["name"],
		Type:        a.Args["type"],
		Description: a.Args["description"],
		Required:    a.Args["required"] == argTrue,
		Example:     a.Args["example"],
	}
}

// ParsedTOS holds parsed !tos data.
type ParsedTOS struct {
	URL string
}

// GetTOS extracts terms of service from annotation.
func GetTOS(a Annotation) ParsedTOS {
	return ParsedTOS{
		URL: a.Args["url"],
	}
}

// ParsedSecurity holds parsed !security data.
type ParsedSecurity struct {
	Name        string
	Type        string // apiKey, oauth2, http, openIdConnect
	Location    string // header, query, cookie (for apiKey); implicit, password, clientCredentials, authorizationCode (for oauth2)
	Description string
	URL         string // Authorization URL for OAuth2
}

// GetSecurity extracts security from annotation.
func GetSecurity(a Annotation) ParsedSecurity {
	return ParsedSecurity{
		Name:        a.Args["name"],
		Type:        a.Args["type"],
		Location:    a.Args["location"],
		Description: a.Args["description"],
		URL:         a.Args["url"],
	}
}

// ParsedSecure holds parsed !secure data (security requirements for operations).
type ParsedSecure struct {
	Names []string
}

// GetSecure extracts secure from annotation.
func GetSecure(a Annotation) ParsedSecure {
	return ParsedSecure{
		Names: a.Tags,
	}
}

// ParsedScope holds parsed !scope data (OAuth2 scopes for security schemes).
type ParsedScope struct {
	Security    string // The security scheme name (e.g., petstore_auth)
	Name        string // The scope name (e.g., write:pets)
	Description string // The scope description
}

// GetScope extracts scope from annotation.
func GetScope(a Annotation) ParsedScope {
	return ParsedScope{
		Security:    a.Args["security"],
		Name:        a.Args["name"],
		Description: a.Args["description"],
	}
}

// ParsedExternalDocs holds parsed !externalDocs data.
type ParsedExternalDocs struct {
	URL         string
	Description string
}

// GetExternalDocs extracts external docs from annotation.
func GetExternalDocs(a Annotation) ParsedExternalDocs {
	return ParsedExternalDocs{
		URL:         a.Args["url"],
		Description: a.Args["description"],
	}
}

// ParsedLink holds parsed !link data.
type ParsedLink struct {
	Label string
	URL   string
}

// GetLink extracts link from annotation.
func GetLink(a Annotation) ParsedLink {
	return ParsedLink{
		Label: a.Args["label"],
		URL:   a.Args["url"],
	}
}

// ParsedWebhook holds parsed !webhook data.
type ParsedWebhook struct {
	Name        string
	Method      string
	Description string
}

// GetWebhook extracts webhook from annotation.
func GetWebhook(a Annotation) ParsedWebhook {
	return ParsedWebhook{
		Name:        a.Args["name"],
		Method:      strings.ToUpper(a.Args["method"]),
		Description: a.Args["description"],
	}
}

// ParsedWebhookBody holds parsed !webhook-body data.
type ParsedWebhookBody struct {
	Schema      string
	Description string
	Required    bool
}

// GetWebhookBody extracts webhook body from annotation.
func GetWebhookBody(a Annotation) ParsedWebhookBody {
	return ParsedWebhookBody{
		Schema:      a.Args["schema"],
		Description: a.Args["description"],
		Required:    strings.Contains(a.RawLine, "required"),
	}
}

// ParsedWebhookResponse holds parsed !webhook-response data.
type ParsedWebhookResponse struct {
	Status      string
	Schema      string
	Description string
}

// GetWebhookResponse extracts webhook response from annotation.
func GetWebhookResponse(a Annotation) ParsedWebhookResponse {
	return ParsedWebhookResponse{
		Status:      a.Args["status"],
		Schema:      a.Args["schema"],
		Description: a.Args["description"],
	}
}

// parseValue attempts to parse a string value into its appropriate type.
func parseValue(s string) any {
	s = strings.Trim(s, `"'`)
	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}
	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	// Try boolean
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	// Return as string
	return s
}
