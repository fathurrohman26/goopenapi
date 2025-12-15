package yahttp

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

func createTestSpec() *openapi.Document {
	return &openapi.Document{
		OpenAPI: "3.0.3",
		Info: openapi.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi.Paths{
			"/users": &openapi.PathItem{
				Get: &openapi.Operation{
					Summary:     "List users",
					OperationID: "listUsers",
					Parameters: []*openapi.Parameter{
						{
							Name:     "limit",
							In:       openapi.ParameterInQuery,
							Required: false,
							Schema:   openapi.IntegerSchema(),
						},
						{
							Name:     "page",
							In:       openapi.ParameterInQuery,
							Required: true,
							Schema:   openapi.IntegerSchema(),
						},
					},
					Responses: openapi.Responses{
						"200": &openapi.Response{Description: "Success"},
					},
				},
			},
			"/users/{id}": &openapi.PathItem{
				Get: &openapi.Operation{
					Summary:     "Get user",
					OperationID: "getUser",
					Parameters: []*openapi.Parameter{
						{
							Name:     "id",
							In:       openapi.ParameterInPath,
							Required: true,
							Schema:   openapi.IntegerSchema(),
						},
					},
					Responses: openapi.Responses{
						"200": &openapi.Response{Description: "Success"},
					},
				},
			},
		},
	}
}

func TestNew(t *testing.T) {
	spec := createTestSpec()
	plugin := New(spec, nil)

	if plugin == nil {
		t.Fatal("New() returned nil")
	}
	if plugin.Spec() != spec {
		t.Error("Spec() should return the provided spec")
	}
	if plugin.Options() == nil {
		t.Error("Options() should not be nil")
	}
}

func TestDefaultOptions(t *testing.T) {
	opts := DefaultOptions()

	if opts.SpecPath != "/openapi.json" {
		t.Errorf("SpecPath = %q, want %q", opts.SpecPath, "/openapi.json")
	}
	if opts.SwaggerUIPath != "/docs" {
		t.Errorf("SwaggerUIPath = %q, want %q", opts.SwaggerUIPath, "/docs")
	}
	if opts.EnableValidation {
		t.Error("EnableValidation should be false by default")
	}
	if opts.EnableCORS {
		t.Error("EnableCORS should be false by default")
	}
}

func TestMustNew(t *testing.T) {
	t.Run("with spec", func(t *testing.T) {
		plugin := MustNew(createTestSpec(), nil)
		if plugin == nil {
			t.Error("MustNew() should return a plugin")
		}
	})

	t.Run("without spec panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("MustNew(nil) should panic")
			}
		}()
		MustNew(nil, nil)
	})
}

func TestPluginBuilder(t *testing.T) {
	spec := createTestSpec()
	plugin := WithSpec(spec).
		SpecPath("/api/spec.json").
		SwaggerUIPath("/api/docs").
		EnableValidation().
		EnableCORS().
		EnableLogging().
		Build()

	opts := plugin.Options()
	if opts.SpecPath != "/api/spec.json" {
		t.Errorf("SpecPath = %q, want %q", opts.SpecPath, "/api/spec.json")
	}
	if opts.SwaggerUIPath != "/api/docs" {
		t.Errorf("SwaggerUIPath = %q, want %q", opts.SwaggerUIPath, "/api/docs")
	}
	if !opts.EnableValidation {
		t.Error("EnableValidation should be true")
	}
	if !opts.EnableCORS {
		t.Error("EnableCORS should be true")
	}
	if !opts.EnableLogging {
		t.Error("EnableLogging should be true")
	}
}

func TestChain(t *testing.T) {
	var order []int

	m1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 1)
			next.ServeHTTP(w, r)
		})
	}
	m2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 2)
			next.ServeHTTP(w, r)
		})
	}
	m3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, 3)
			next.ServeHTTP(w, r)
		})
	}

	final := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		order = append(order, 0)
	})

	handler := Chain(m1, m2, m3)(final)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if len(order) != 4 {
		t.Errorf("Expected 4 calls, got %d", len(order))
	}
	expected := []int{1, 2, 3, 0}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %d, want %d", i, order[i], v)
		}
	}
}

func TestMiddlewareFunc(t *testing.T) {
	called := false
	mf := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		called = true
		next.ServeHTTP(w, r)
	})

	handler := mf.Then(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if !called {
		t.Error("MiddlewareFunc was not called")
	}
}

func TestRecovery(t *testing.T) {
	t.Run("recovers from panic", func(t *testing.T) {
		var recovered any
		recovery := Recovery(func(w http.ResponseWriter, r *http.Request, err any) {
			recovered = err
			w.WriteHeader(http.StatusInternalServerError)
		})

		handler := recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if recovered != "test panic" {
			t.Errorf("recovered = %v, want %q", recovered, "test panic")
		}
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})

	t.Run("default handler", func(t *testing.T) {
		recovery := Recovery(nil)
		handler := recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			panic("test panic")
		}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusInternalServerError)
		}
	})
}

func TestRequestID(t *testing.T) {
	t.Run("generates request ID", func(t *testing.T) {
		middleware := RequestID(func() string { return "test-id-123" })
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("X-Request-ID"); got != "test-id-123" {
			t.Errorf("X-Request-ID = %q, want %q", got, "test-id-123")
		}
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		middleware := RequestID(func() string { return "new-id" })
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Request-ID", "existing-id")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("X-Request-ID"); got != "existing-id" {
			t.Errorf("X-Request-ID = %q, want %q", got, "existing-id")
		}
	})
}

func TestContentType(t *testing.T) {
	middleware := ContentType("application/xml")
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Content-Type"); got != "application/xml" {
		t.Errorf("Content-Type = %q, want %q", got, "application/xml")
	}
}

func TestJSON(t *testing.T) {
	middleware := JSON()
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if got := w.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json; charset=utf-8")
	}
}

func TestHandler(t *testing.T) {
	spec := createTestSpec()
	handler := Handler(spec, nil)

	if handler == nil {
		t.Fatal("Handler() returned nil")
	}
}

func TestPlugin_Mount(t *testing.T) {
	spec := createTestSpec()
	plugin := New(spec, nil)
	mux := http.NewServeMux()

	plugin.Mount(mux)

	// Test spec endpoint
	req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Spec endpoint status = %d, want %d", w.Code, http.StatusOK)
	}

	// Test docs endpoint
	req = httptest.NewRequest(http.MethodGet, "/docs", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Docs endpoint status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestPlugin_WrapMux(t *testing.T) {
	spec := createTestSpec()
	opts := &Options{
		SpecPath:      "/openapi.json",
		SwaggerUIPath: "/docs",
		EnableCORS:    true,
	}
	plugin := New(spec, opts)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := plugin.WrapMux(mux)

	// Test API endpoint with CORS
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("API endpoint status = %d, want %d", w.Code, http.StatusOK)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Error("Expected CORS header to be set")
	}
}

func TestSpecHandler(t *testing.T) {
	spec := createTestSpec()
	plugin := New(spec, nil)
	handler := plugin.SpecHandler()

	t.Run("JSON format default", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi.json", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}
		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "application/json") {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		// Verify JSON is valid and contains expected content
		body := w.Body.String()
		if !strings.Contains(body, `"openapi"`) {
			t.Error("Response should contain openapi field")
		}
		if !strings.Contains(body, `"Test API"`) {
			t.Error("Response should contain API title")
		}
	})

	t.Run("YAML format via extension", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi.yaml", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "yaml") {
			t.Errorf("Content-Type = %q, want yaml", ct)
		}
	})

	t.Run("YAML format via Accept header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi", nil)
		req.Header.Set("Accept", "application/yaml")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "yaml") {
			t.Errorf("Content-Type = %q, want yaml", ct)
		}
	})

	t.Run("YAML format via query param", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/openapi?format=yaml", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "yaml") {
			t.Errorf("Content-Type = %q, want yaml", ct)
		}
	})
}

func TestSwaggerUIHandler(t *testing.T) {
	spec := createTestSpec()
	plugin := New(spec, nil)
	handler := plugin.SwaggerUIHandler()

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}

	body := w.Body.String()
	if !strings.Contains(body, "swagger-ui") {
		t.Error("Response should contain swagger-ui")
	}
	if !strings.Contains(body, "Test API") {
		t.Error("Response should contain API title")
	}
}

func TestRedocHandler(t *testing.T) {
	spec := createTestSpec()
	plugin := New(spec, nil)
	handler := plugin.RedocHandler()

	req := httptest.NewRequest(http.MethodGet, "/redoc", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if !strings.Contains(body, "redoc") {
		t.Error("Response should contain redoc")
	}
}

func TestCORSMiddleware(t *testing.T) {
	opts := &CORSOptions{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           3600,
	}
	middleware := CORS(opts)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("simple request with allowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://example.com" {
			t.Errorf("Allow-Origin = %q, want %q", got, "http://example.com")
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Errorf("Allow-Credentials = %q, want %q", got, "true")
		}
	})

	t.Run("preflight request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusNoContent)
		}
		if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Error("Expected Allow-Methods header")
		}
	})

	t.Run("disallowed origin", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://evil.com")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("Allow-Origin should be empty for disallowed origin, got %q", got)
		}
	})
}

func TestLoggingMiddleware(t *testing.T) {
	var logged string
	logger := func(format string, args ...any) {
		logged = format
	}

	middleware := Logging(logger)
	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if logged == "" {
		t.Error("Expected log output")
	}
}

func TestStructuredLogging(t *testing.T) {
	var entry LogEntry
	middleware := StructuredLogging(func(e LogEntry) {
		entry = e
	})

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("X-Request-ID", "test-123")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if entry.Method != "POST" {
		t.Errorf("Method = %q, want %q", entry.Method, "POST")
	}
	if entry.Path != "/users" {
		t.Errorf("Path = %q, want %q", entry.Path, "/users")
	}
	if entry.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want %d", entry.StatusCode, http.StatusCreated)
	}
	if entry.RequestID != "test-123" {
		t.Errorf("RequestID = %q, want %q", entry.RequestID, "test-123")
	}
}

func TestValidationMiddleware(t *testing.T) {
	spec := createTestSpec()
	plugin := New(spec, &Options{EnableValidation: true})
	middleware := plugin.ValidationMiddleware()

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	t.Run("valid request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users?page=1", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("missing required parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid parameter type", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users?page=abc", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("path parameter validation", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/abc", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Status = %d, want %d for non-integer path param", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid path parameter", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/123", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestValidationError(t *testing.T) {
	err := ValidationError{
		Field:   "limit",
		Message: "must be an integer",
		In:      "query",
	}

	if !strings.Contains(err.Error(), "limit") {
		t.Error("Error should contain field name")
	}
	if !strings.Contains(err.Error(), "must be an integer") {
		t.Error("Error should contain message")
	}
}

func TestValidateRequest(t *testing.T) {
	spec := createTestSpec()

	t.Run("valid request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users?page=1", nil)
		errs := ValidateRequest(spec, req)

		if len(errs) != 0 {
			t.Errorf("Expected no errors, got %d: %v", len(errs), errs)
		}
	})

	t.Run("invalid request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		errs := ValidateRequest(spec, req)

		if len(errs) == 0 {
			t.Error("Expected validation errors")
		}
	})
}

func TestServeSpec(t *testing.T) {
	spec := createTestSpec()
	handler := ServeSpec(spec)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status = %d, want %d", w.Code, http.StatusOK)
	}

	body, _ := io.ReadAll(w.Body)
	if !strings.Contains(string(body), "Test API") {
		t.Error("Response should contain spec content")
	}
}
