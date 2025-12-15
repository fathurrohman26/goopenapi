// Package swaggerui provides a Swagger UI server for viewing OpenAPI specifications.
package swaggerui

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/fathurrohman26/yaswag/pkg/validator"
)

//go:embed templates/*.html
var templates embed.FS

// Server serves OpenAPI specifications with Swagger UI.
type Server struct {
	specData    []byte
	specURL     string
	isRemoteURL bool
	port        int
}

// NewServer creates a new Swagger UI server.
func NewServer(port int) *Server {
	return &Server{port: port}
}

// SetSpecFromFile loads the OpenAPI specification from a file.
func (s *Server) SetSpecFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}
	s.specData = data
	s.isRemoteURL = false
	return nil
}

// SetSpecFromURL sets a remote URL for the OpenAPI specification.
func (s *Server) SetSpecFromURL(url string) {
	s.specURL = url
	s.isRemoteURL = true
}

// SetSpecFromData sets the OpenAPI specification from raw data.
func (s *Server) SetSpecFromData(data []byte) {
	s.specData = data
	s.isRemoteURL = false
}

// Serve starts the HTTP server and serves the Swagger UI.
func (s *Server) Serve() error {
	mux := http.NewServeMux()

	// Serve the spec
	mux.HandleFunc("/spec", s.handleSpec)

	// Serve validation endpoint
	mux.HandleFunc("/validate", s.handleValidate)

	// Serve the Swagger UI HTML
	mux.HandleFunc("/", s.handleUI)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Swagger UI is available at http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop the server")

	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleSpec(w http.ResponseWriter, r *http.Request) {
	var specData []byte

	if s.isRemoteURL {
		// Proxy the remote URL
		resp, err := http.Get(s.specURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch remote spec: %v", err), http.StatusInternalServerError)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		specData, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read remote spec: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		specData = s.specData
	}

	// Patch OpenAPI 3.2.x to 3.1.x for Swagger UI compatibility
	// Swagger UI does not yet support OpenAPI 3.2.x rendering
	specData = patchOpenAPI32To31(specData)

	// Determine content type
	contentType := "application/json"
	if len(specData) > 0 && (specData[0] == '-' || specData[0] == '#' || strings.HasPrefix(string(specData), "openapi:") || strings.HasPrefix(string(specData), "swagger:")) {
		contentType = "application/yaml"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(specData)
}

// ValidationResponse represents the JSON response for validation endpoint.
type ValidationResponse struct {
	Valid    bool             `json:"valid"`
	Version  string           `json:"version"`
	Errors   []ValidationItem `json:"errors,omitempty"`
	Warnings []ValidationItem `json:"warnings,omitempty"`
}

// ValidationItem represents a single validation error or warning.
type ValidationItem struct {
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	specData, err := s.getSpecData()
	if err != nil {
		writeValidationError(w, err.Error())
		return
	}

	localResult, version := s.runLocalValidation(specData)
	swaggerResult, err := callSwaggerValidator(specData)

	var response ValidationResponse
	if err != nil {
		response = buildLocalOnlyResponse(localResult, version)
	} else {
		response = buildMergedResponse(swaggerResult, localResult, version)
	}

	writeJSONResponse(w, response)
}

func (s *Server) getSpecData() ([]byte, error) {
	if !s.isRemoteURL {
		return s.specData, nil
	}
	resp, err := http.Get(s.specURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch remote spec: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote spec: %w", err)
	}
	return data, nil
}

func (s *Server) runLocalValidation(specData []byte) (*validator.ValidationResult, string) {
	v := validator.New()
	result, _ := v.Validate(specData)
	version := ""
	if result != nil {
		version = result.Version
	}
	return result, version
}

func buildLocalOnlyResponse(localResult *validator.ValidationResult, version string) ValidationResponse {
	response := ValidationResponse{Valid: localResult.Valid, Version: version}
	for _, e := range localResult.Errors {
		response.Errors = append(response.Errors, ValidationItem{Message: e.Message, Path: e.Path, Line: e.Line, Column: e.Column})
	}
	for _, warn := range localResult.Warnings {
		response.Warnings = append(response.Warnings, ValidationItem{Message: warn.Message, Path: warn.Path, Line: warn.Line, Column: warn.Column})
	}
	return response
}

func buildMergedResponse(swaggerResult *SwaggerValidatorResponse, localResult *validator.ValidationResult, version string) ValidationResponse {
	response := ValidationResponse{
		Valid:   len(swaggerResult.Errors) == 0 && len(swaggerResult.SchemaValidationMessages) == 0,
		Version: version,
	}

	for _, msg := range swaggerResult.SchemaValidationMessages {
		response.Errors = append(response.Errors, ValidationItem{Message: msg})
	}
	addSwaggerMessages(&response, swaggerResult.Messages)
	mergeLocalResults(&response, localResult)
	return response
}

func addSwaggerMessages(response *ValidationResponse, messages []string) {
	for _, msg := range messages {
		item := ValidationItem{Message: msg}
		if response.Valid {
			response.Warnings = append(response.Warnings, item)
		} else {
			response.Errors = append(response.Errors, item)
		}
	}
}

func mergeLocalResults(response *ValidationResponse, localResult *validator.ValidationResult) {
	if localResult == nil {
		return
	}
	if response.Valid && !localResult.Valid {
		response.Valid = false
		for _, e := range localResult.Errors {
			response.Errors = append(response.Errors, ValidationItem{Message: e.Message, Path: e.Path})
		}
	}
	for _, warn := range localResult.Warnings {
		response.Warnings = append(response.Warnings, ValidationItem{Message: warn.Message, Path: warn.Path})
	}
}

func writeJSONResponse(w http.ResponseWriter, response ValidationResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_ = json.NewEncoder(w).Encode(response)
}

// SwaggerValidatorResponse represents the response from Swagger.io validator API.
type SwaggerValidatorResponse struct {
	Messages                 []string `json:"messages"`
	SchemaValidationMessages []string `json:"schemaValidationMessages"`
	Errors                   []string `json:"errors"`
}

// callSwaggerValidator calls the Swagger.io validator API.
func callSwaggerValidator(specData []byte) (*SwaggerValidatorResponse, error) {
	// Determine content type
	contentType := "application/json"
	if len(specData) > 0 && (specData[0] == '-' || specData[0] == '#' || strings.HasPrefix(string(specData), "openapi:") || strings.HasPrefix(string(specData), "swagger:")) {
		contentType = "application/yaml"
	}

	// POST to Swagger.io validator
	req, err := http.NewRequest("POST", "https://validator.swagger.io/validator/debug", strings.NewReader(string(specData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call validator: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("validator returned status %d", resp.StatusCode)
	}

	var result SwaggerValidatorResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

func writeValidationError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(ValidationResponse{
		Valid:  false,
		Errors: []ValidationItem{{Message: message}},
	})
}

// patchOpenAPI32To31 patches OpenAPI 3.2.x versions to 3.1.0 for Swagger UI compatibility.
// Swagger UI does not yet support rendering OpenAPI 3.2.x specifications.
func patchOpenAPI32To31(data []byte) []byte {
	content := string(data)

	// Pattern to match openapi: 3.2.x in YAML or "openapi": "3.2.x" in JSON
	yamlPattern := regexp.MustCompile(`(?m)^openapi:\s*['"]?(3\.2\.\d+)['"]?`)
	jsonPattern := regexp.MustCompile(`"openapi"\s*:\s*"(3\.2\.\d+)"`)

	// Check if patching is needed
	if yamlPattern.MatchString(content) {
		content = yamlPattern.ReplaceAllString(content, "openapi: 3.1.0")
	}
	if jsonPattern.MatchString(content) {
		content = jsonPattern.ReplaceAllString(content, `"openapi": "3.1.0"`)
	}

	return []byte(content)
}

func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(templates, "templates/index.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load template: %v", err), http.StatusInternalServerError)
		return
	}

	specURL := "/spec"
	if s.isRemoteURL {
		specURL = s.specURL
	}

	data := struct {
		SpecURL string
	}{
		SpecURL: specURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}

// EditorServer serves the Swagger Editor for editing OpenAPI specifications.
type EditorServer struct {
	specData    []byte
	specURL     string
	isRemoteURL bool
	hasSpec     bool
	port        int
}

// NewEditorServer creates a new Swagger Editor server.
func NewEditorServer(port int) *EditorServer {
	return &EditorServer{port: port}
}

// SetSpecFromFile loads the OpenAPI specification from a file.
func (s *EditorServer) SetSpecFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read spec file: %w", err)
	}
	s.specData = data
	s.isRemoteURL = false
	s.hasSpec = true
	return nil
}

// SetSpecFromURL sets a remote URL for the OpenAPI specification.
func (s *EditorServer) SetSpecFromURL(url string) {
	s.specURL = url
	s.isRemoteURL = true
	s.hasSpec = true
}

// SetSpecFromData sets the OpenAPI specification from raw data.
func (s *EditorServer) SetSpecFromData(data []byte) {
	s.specData = data
	s.isRemoteURL = false
	s.hasSpec = true
}

// Serve starts the HTTP server and serves the Swagger Editor.
func (s *EditorServer) Serve() error {
	mux := http.NewServeMux()

	// Serve the spec (if provided)
	if s.hasSpec {
		mux.HandleFunc("/spec", s.handleSpec)
	}

	// Serve the Swagger Editor HTML
	mux.HandleFunc("/", s.handleEditorUI)

	addr := fmt.Sprintf(":%d", s.port)
	fmt.Printf("Swagger Editor is available at http://localhost%s\n", addr)
	fmt.Println("Press Ctrl+C to stop the server")

	return http.ListenAndServe(addr, mux)
}

func (s *EditorServer) handleSpec(w http.ResponseWriter, r *http.Request) {
	var specData []byte

	if s.isRemoteURL {
		// Proxy the remote URL
		resp, err := http.Get(s.specURL)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to fetch remote spec: %v", err), http.StatusInternalServerError)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		specData, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read remote spec: %v", err), http.StatusInternalServerError)
			return
		}
	} else {
		specData = s.specData
	}

	// Determine content type
	contentType := "application/json"
	if len(specData) > 0 && (specData[0] == '-' || specData[0] == '#' || strings.HasPrefix(string(specData), "openapi:") || strings.HasPrefix(string(specData), "swagger:")) {
		contentType = "application/yaml"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	_, _ = w.Write(specData)
}

func (s *EditorServer) handleEditorUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}

	tmpl, err := template.ParseFS(templates, "templates/editor.html")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to load template: %v", err), http.StatusInternalServerError)
		return
	}

	data := struct {
		HasSpec bool
	}{
		HasSpec: s.hasSpec,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = tmpl.Execute(w, data)
}
