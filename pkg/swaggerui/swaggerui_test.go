package swaggerui

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer(8080)
	if server == nil {
		t.Fatal("NewServer() returned nil")
	}
	if server.port != 8080 {
		t.Errorf("port = %d, want 8080", server.port)
	}
}

func TestServer_SetSpecFromData(t *testing.T) {
	server := NewServer(8080)
	spec := []byte(`{"openapi": "3.0.3"}`)

	server.SetSpecFromData(spec)

	if string(server.specData) != string(spec) {
		t.Error("specData not set correctly")
	}
	if server.isRemoteURL {
		t.Error("isRemoteURL should be false")
	}
}

func TestServer_SetSpecFromURL(t *testing.T) {
	server := NewServer(8080)
	url := "https://example.com/spec.yaml"

	server.SetSpecFromURL(url)

	if server.specURL != url {
		t.Errorf("specURL = %q, want %q", server.specURL, url)
	}
	if !server.isRemoteURL {
		t.Error("isRemoteURL should be true")
	}
}

func TestNewEditorServer(t *testing.T) {
	server := NewEditorServer(8081)
	if server == nil {
		t.Fatal("NewEditorServer() returned nil")
	}
	if server.port != 8081 {
		t.Errorf("port = %d, want 8081", server.port)
	}
}

func TestEditorServer_SetSpecFromData(t *testing.T) {
	server := NewEditorServer(8080)
	spec := []byte(`{"openapi": "3.1.0"}`)

	server.SetSpecFromData(spec)

	if string(server.specData) != string(spec) {
		t.Error("specData not set correctly")
	}
	if server.isRemoteURL {
		t.Error("isRemoteURL should be false")
	}
	if !server.hasSpec {
		t.Error("hasSpec should be true")
	}
}

func TestEditorServer_SetSpecFromURL(t *testing.T) {
	server := NewEditorServer(8080)
	url := "https://example.com/spec.yaml"

	server.SetSpecFromURL(url)

	if server.specURL != url {
		t.Errorf("specURL = %q, want %q", server.specURL, url)
	}
	if !server.isRemoteURL {
		t.Error("isRemoteURL should be true")
	}
	if !server.hasSpec {
		t.Error("hasSpec should be true")
	}
}

func TestPatchOpenAPI32To31(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "YAML 3.2.0",
			input: "openapi: 3.2.0\ninfo:\n  title: Test",
			want:  "openapi: 3.1.0\ninfo:\n  title: Test",
		},
		{
			name:  "YAML 3.2.1 quoted",
			input: "openapi: '3.2.1'\ninfo:\n  title: Test",
			want:  "openapi: 3.1.0\ninfo:\n  title: Test",
		},
		{
			name:  "YAML 3.2.0 double quoted",
			input: `openapi: "3.2.0"` + "\ninfo:\n  title: Test",
			want:  "openapi: 3.1.0\ninfo:\n  title: Test",
		},
		{
			name:  "JSON 3.2.0",
			input: `{"openapi": "3.2.0", "info": {"title": "Test"}}`,
			want:  `{"openapi": "3.1.0", "info": {"title": "Test"}}`,
		},
		{
			name:  "JSON 3.2.1",
			input: `{"openapi": "3.2.1", "info": {"title": "Test"}}`,
			want:  `{"openapi": "3.1.0", "info": {"title": "Test"}}`,
		},
		{
			name:  "YAML 3.0.3 unchanged",
			input: "openapi: 3.0.3\ninfo:\n  title: Test",
			want:  "openapi: 3.0.3\ninfo:\n  title: Test",
		},
		{
			name:  "YAML 3.1.0 unchanged",
			input: "openapi: 3.1.0\ninfo:\n  title: Test",
			want:  "openapi: 3.1.0\ninfo:\n  title: Test",
		},
		{
			name:  "JSON 3.0.3 unchanged",
			input: `{"openapi": "3.0.3"}`,
			want:  `{"openapi": "3.0.3"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := patchOpenAPI32To31([]byte(tt.input))
			if string(got) != tt.want {
				t.Errorf("patchOpenAPI32To31() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServer_HandleSpec_LocalSpec(t *testing.T) {
	spec := `{"openapi": "3.0.3", "info": {"title": "Test", "version": "1.0.0"}}`
	server := NewServer(8080)
	server.SetSpecFromData([]byte(spec))

	req := httptest.NewRequest(http.MethodGet, "/spec", nil)
	w := httptest.NewRecorder()

	server.handleSpec(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "3.0.3") {
		t.Error("Response should contain spec content")
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}
}

func TestServer_HandleSpec_YAMLSpec(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"`
	server := NewServer(8080)
	server.SetSpecFromData([]byte(spec))

	req := httptest.NewRequest(http.MethodGet, "/spec", nil)
	w := httptest.NewRecorder()

	server.handleSpec(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/yaml" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/yaml")
	}
}

func TestServer_HandleSpec_PatchesOpenAPI32(t *testing.T) {
	spec := `{"openapi": "3.2.0", "info": {"title": "Test", "version": "1.0.0"}}`
	server := NewServer(8080)
	server.SetSpecFromData([]byte(spec))

	req := httptest.NewRequest(http.MethodGet, "/spec", nil)
	w := httptest.NewRecorder()

	server.handleSpec(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if strings.Contains(string(body), "3.2.0") {
		t.Error("Response should have 3.2.0 patched to 3.1.0")
	}
	if !strings.Contains(string(body), "3.1.0") {
		t.Error("Response should contain patched version 3.1.0")
	}
}

func TestServer_HandleUI_Root(t *testing.T) {
	server := NewServer(8080)
	server.SetSpecFromData([]byte(`{"openapi": "3.0.3"}`))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	server.handleUI(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}
}

func TestServer_HandleUI_NotFound(t *testing.T) {
	server := NewServer(8080)

	req := httptest.NewRequest(http.MethodGet, "/invalid-path", nil)
	w := httptest.NewRecorder()

	server.handleUI(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
	}
}

func TestServer_HandleValidate_LocalSpec(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test API
  version: "1.0.0"
paths: {}`
	server := NewServer(8080)
	server.SetSpecFromData([]byte(spec))

	req := httptest.NewRequest(http.MethodGet, "/validate", nil)
	w := httptest.NewRecorder()

	server.handleValidate(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
	}

	body, _ := io.ReadAll(resp.Body)
	var result ValidationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.Version != "3.0.3" {
		t.Errorf("Version = %q, want %q", result.Version, "3.0.3")
	}
}

func TestValidationResponse_Structure(t *testing.T) {
	response := ValidationResponse{
		Valid:    true,
		Version:  "3.0.3",
		Errors:   []ValidationItem{{Message: "error", Path: "/info"}},
		Warnings: []ValidationItem{{Message: "warning"}},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ValidationResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !decoded.Valid {
		t.Error("Valid should be true")
	}
	if decoded.Version != "3.0.3" {
		t.Errorf("Version = %q, want %q", decoded.Version, "3.0.3")
	}
	if len(decoded.Errors) != 1 {
		t.Errorf("Errors length = %d, want 1", len(decoded.Errors))
	}
	if len(decoded.Warnings) != 1 {
		t.Errorf("Warnings length = %d, want 1", len(decoded.Warnings))
	}
}

func TestValidationItem_Structure(t *testing.T) {
	item := ValidationItem{
		Message: "Field is required",
		Path:    "$.info.title",
		Line:    10,
		Column:  5,
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded ValidationItem
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Message != item.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, item.Message)
	}
	if decoded.Path != item.Path {
		t.Errorf("Path = %q, want %q", decoded.Path, item.Path)
	}
	if decoded.Line != item.Line {
		t.Errorf("Line = %d, want %d", decoded.Line, item.Line)
	}
	if decoded.Column != item.Column {
		t.Errorf("Column = %d, want %d", decoded.Column, item.Column)
	}
}

func TestEditorServer_HandleSpec(t *testing.T) {
	spec := `{"openapi": "3.0.3", "info": {"title": "Test", "version": "1.0.0"}}`
	server := NewEditorServer(8080)
	server.SetSpecFromData([]byte(spec))

	req := httptest.NewRequest(http.MethodGet, "/spec", nil)
	w := httptest.NewRecorder()

	server.handleSpec(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "3.0.3") {
		t.Error("Response should contain spec content")
	}
}

func TestEditorServer_HandleEditorUI(t *testing.T) {
	t.Run("root path", func(t *testing.T) {
		server := NewEditorServer(8080)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		server.handleEditorUI(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server := NewEditorServer(8080)

		req := httptest.NewRequest(http.MethodGet, "/invalid", nil)
		w := httptest.NewRecorder()

		server.handleEditorUI(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusNotFound)
		}
	})
}

func TestServer_HandleSpec_RemoteURL(t *testing.T) {
	// Create a mock remote server
	remoteSpec := `{"openapi": "3.0.3", "info": {"title": "Remote", "version": "1.0.0"}}`
	remoteServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(remoteSpec))
	}))
	defer remoteServer.Close()

	server := NewServer(8080)
	server.SetSpecFromURL(remoteServer.URL)

	req := httptest.NewRequest(http.MethodGet, "/spec", nil)
	w := httptest.NewRecorder()

	server.handleSpec(w, req)

	resp := w.Result()
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Remote") {
		t.Error("Response should contain remote spec content")
	}
}

func TestBuildLocalOnlyResponse(t *testing.T) {
	t.Run("valid result", func(t *testing.T) {
		// This is internal but we can test via handleValidate
		server := NewServer(8080)
		spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths: {}`
		server.SetSpecFromData([]byte(spec))

		req := httptest.NewRequest(http.MethodGet, "/validate", nil)
		w := httptest.NewRecorder()
		server.handleValidate(w, req)

		resp := w.Result()
		defer func() { _ = resp.Body.Close() }()

		body, _ := io.ReadAll(resp.Body)
		var result ValidationResponse
		_ = json.Unmarshal(body, &result)

		if result.Version == "" {
			t.Error("Version should be populated")
		}
	})
}
