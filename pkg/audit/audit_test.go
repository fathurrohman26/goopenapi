package audit

import (
	"strings"
	"testing"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

func TestAuditor_Audit_BasicCounting(t *testing.T) {
	doc := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get: &openapi.Operation{
					Security: []openapi.SecurityRequirement{{"bearer": {}}},
				},
				Post: &openapi.Operation{},
			},
			"/public": {
				Get: &openapi.Operation{},
			},
		},
	}

	auditor := New()
	result := auditor.Audit(doc)

	if result.TotalEndpoints != 3 {
		t.Errorf("TotalEndpoints = %d, want 3", result.TotalEndpoints)
	}
	if result.ProtectedEndpoints != 1 {
		t.Errorf("ProtectedEndpoints = %d, want 1", result.ProtectedEndpoints)
	}
	if result.UnprotectedEndpoints != 2 {
		t.Errorf("UnprotectedEndpoints = %d, want 2", result.UnprotectedEndpoints)
	}
}

func TestAuditor_Audit_GlobalSecurity(t *testing.T) {
	doc := &openapi.Document{
		Security: []openapi.SecurityRequirement{{"apiKey": {}}},
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get:  &openapi.Operation{},
				Post: &openapi.Operation{},
			},
		},
	}

	auditor := New()
	result := auditor.Audit(doc)

	if result.ProtectedEndpoints != 2 {
		t.Errorf("ProtectedEndpoints = %d, want 2 (global security)", result.ProtectedEndpoints)
	}
	if result.UnprotectedEndpoints != 0 {
		t.Errorf("UnprotectedEndpoints = %d, want 0", result.UnprotectedEndpoints)
	}
}

func TestUnprotectedWriteRule(t *testing.T) {
	doc := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get:    &openapi.Operation{},
				Post:   &openapi.Operation{},
				Delete: &openapi.Operation{},
			},
		},
	}

	rule := &UnprotectedWriteRule{}
	findings := rule.Check(doc)

	if len(findings) != 2 {
		t.Errorf("got %d findings, want 2 (POST and DELETE)", len(findings))
	}

	for _, f := range findings {
		if f.RuleID != "UNPROTECTED_WRITE" {
			t.Errorf("RuleID = %s, want UNPROTECTED_WRITE", f.RuleID)
		}
		if f.Severity != SeverityWarning {
			t.Errorf("Severity = %s, want WARNING", f.Severity)
		}
	}
}

func TestUnprotectedWriteRule_WithSecurity(t *testing.T) {
	doc := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Post: &openapi.Operation{
					Security: []openapi.SecurityRequirement{{"bearer": {}}},
				},
			},
		},
	}

	rule := &UnprotectedWriteRule{}
	findings := rule.Check(doc)

	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0 (endpoint has security)", len(findings))
	}
}

func TestAPIKeyInQueryRule(t *testing.T) {
	doc := &openapi.Document{
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"apiKeyQuery": {
					Type: "apiKey",
					In:   "query",
					Name: "api_key",
				},
				"apiKeyHeader": {
					Type: "apiKey",
					In:   "header",
					Name: "X-API-Key",
				},
			},
		},
	}

	rule := &APIKeyInQueryRule{}
	findings := rule.Check(doc)

	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1", len(findings))
	}

	if len(findings) > 0 {
		if findings[0].RuleID != "API_KEY_IN_QUERY" {
			t.Errorf("RuleID = %s, want API_KEY_IN_QUERY", findings[0].RuleID)
		}
		if !strings.Contains(findings[0].Location, "apiKeyQuery") {
			t.Errorf("Location should contain 'apiKeyQuery', got %s", findings[0].Location)
		}
	}
}

func TestOAuthHTTPSRule(t *testing.T) {
	doc := &openapi.Document{
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"oauth2": {
					Type: "oauth2",
					Flows: &openapi.OAuthFlows{
						AuthorizationCode: &openapi.OAuthFlow{
							AuthorizationURL: "http://example.com/auth",
							TokenURL:         "https://example.com/token",
						},
					},
				},
			},
		},
	}

	rule := &OAuthHTTPSRule{}
	findings := rule.Check(doc)

	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1 (http authorizationUrl)", len(findings))
	}

	if len(findings) > 0 {
		if findings[0].Severity != SeverityError {
			t.Errorf("Severity = %s, want ERROR", findings[0].Severity)
		}
	}
}

func TestOAuthHTTPSRule_AllHTTPS(t *testing.T) {
	doc := &openapi.Document{
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"oauth2": {
					Type: "oauth2",
					Flows: &openapi.OAuthFlows{
						AuthorizationCode: &openapi.OAuthFlow{
							AuthorizationURL: "https://example.com/auth",
							TokenURL:         "https://example.com/token",
						},
					},
				},
			},
		},
	}

	rule := &OAuthHTTPSRule{}
	findings := rule.Check(doc)

	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0 (all HTTPS)", len(findings))
	}
}

func TestScopeValidationRule(t *testing.T) {
	doc := &openapi.Document{
		Components: &openapi.Components{
			SecuritySchemes: map[string]*openapi.SecurityScheme{
				"oauth2": {
					Type: "oauth2",
					Flows: &openapi.OAuthFlows{
						ClientCredentials: &openapi.OAuthFlow{
							TokenURL: "https://example.com/token",
							Scopes: map[string]string{
								"read":  "Read access",
								"write": "Write access",
							},
						},
					},
				},
			},
		},
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get: &openapi.Operation{
					Security: []openapi.SecurityRequirement{
						{"oauth2": {"read", "admin"}}, // admin is not defined
					},
				},
			},
		},
	}

	rule := &ScopeValidationRule{}
	findings := rule.Check(doc)

	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1 (undefined 'admin' scope)", len(findings))
	}

	if len(findings) > 0 && !strings.Contains(findings[0].Message, "admin") {
		t.Errorf("Message should mention 'admin' scope, got %s", findings[0].Message)
	}
}

func TestDeprecatedSecurityRule(t *testing.T) {
	doc := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/old": {
				Get: &openapi.Operation{
					Deprecated: true,
				},
			},
			"/new": {
				Get: &openapi.Operation{
					Deprecated: false,
				},
			},
		},
	}

	rule := &DeprecatedSecurityRule{}
	findings := rule.Check(doc)

	if len(findings) != 1 {
		t.Errorf("got %d findings, want 1 (deprecated endpoint without security)", len(findings))
	}

	if len(findings) > 0 && findings[0].Severity != SeverityInfo {
		t.Errorf("Severity = %s, want INFO", findings[0].Severity)
	}
}

func TestFormatText(t *testing.T) {
	result := &AuditResult{
		TotalEndpoints:       10,
		ProtectedEndpoints:   8,
		UnprotectedEndpoints: 2,
		Findings: []Finding{
			{
				RuleID:         "TEST_RULE",
				RuleName:       "Test Rule",
				Severity:       SeverityWarning,
				Location:       "POST /users",
				Message:        "Test message",
				Recommendation: "Fix it",
			},
		},
		SecuritySchemes: map[string]SecuritySchemeInfo{
			"bearer": {Type: "http", EndpointCount: 5},
		},
	}

	output := FormatText(result)

	if !strings.Contains(output, "Security Audit Report") {
		t.Error("Output should contain 'Security Audit Report'")
	}
	if !strings.Contains(output, "Total Endpoints: 10") {
		t.Error("Output should contain endpoint count")
	}
	if !strings.Contains(output, "Protected: 8") {
		t.Error("Output should contain protected count")
	}
	if !strings.Contains(output, "[WARNING]") {
		t.Error("Output should contain warning severity")
	}
	if !strings.Contains(output, "TEST_RULE") {
		t.Error("Output should contain rule ID")
	}
}

func TestFormatJSON(t *testing.T) {
	result := &AuditResult{
		TotalEndpoints:       5,
		ProtectedEndpoints:   3,
		UnprotectedEndpoints: 2,
		Findings:             []Finding{},
		EndpointsBySecurity:  map[string][]string{},
		CoverageByTag:        map[string]TagCoverage{},
		SecuritySchemes:      map[string]SecuritySchemeInfo{},
	}

	data, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	if !strings.Contains(string(data), "\"total_endpoints\": 5") {
		t.Error("JSON should contain total_endpoints")
	}
	if !strings.Contains(string(data), "\"protected_endpoints\": 3") {
		t.Error("JSON should contain protected_endpoints")
	}
}

func TestAuditData_JSON(t *testing.T) {
	jsonSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {}
			}
		}
	}`

	auditor := New()
	result, err := auditor.AuditData([]byte(jsonSpec))
	if err != nil {
		t.Fatalf("AuditData error: %v", err)
	}

	if result.TotalEndpoints != 1 {
		t.Errorf("TotalEndpoints = %d, want 1", result.TotalEndpoints)
	}
}

func TestAuditData_YAML(t *testing.T) {
	yamlSpec := `
openapi: "3.0.0"
info:
  title: Test
  version: "1.0.0"
paths:
  /test:
    get: {}
    post: {}
`

	auditor := New()
	result, err := auditor.AuditData([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("AuditData error: %v", err)
	}

	if result.TotalEndpoints != 2 {
		t.Errorf("TotalEndpoints = %d, want 2", result.TotalEndpoints)
	}
}

func TestTagCoverage(t *testing.T) {
	doc := &openapi.Document{
		Paths: map[string]*openapi.PathItem{
			"/users": {
				Get: &openapi.Operation{
					Tags:     []string{"users"},
					Security: []openapi.SecurityRequirement{{"bearer": {}}},
				},
				Post: &openapi.Operation{
					Tags: []string{"users"},
				},
			},
			"/products": {
				Get: &openapi.Operation{
					Tags:     []string{"products"},
					Security: []openapi.SecurityRequirement{{"bearer": {}}},
				},
			},
		},
	}

	auditor := New()
	result := auditor.Audit(doc)

	usersCoverage := result.CoverageByTag["users"]
	if usersCoverage.Total != 2 {
		t.Errorf("users Total = %d, want 2", usersCoverage.Total)
	}
	if usersCoverage.Protected != 1 {
		t.Errorf("users Protected = %d, want 1", usersCoverage.Protected)
	}

	productsCoverage := result.CoverageByTag["products"]
	if productsCoverage.Total != 1 {
		t.Errorf("products Total = %d, want 1", productsCoverage.Total)
	}
	if productsCoverage.Protected != 1 {
		t.Errorf("products Protected = %d, want 1", productsCoverage.Protected)
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()

	if len(rules) != 5 {
		t.Errorf("DefaultRules() returned %d rules, want 5", len(rules))
	}

	expectedIDs := map[string]bool{
		"UNPROTECTED_WRITE":     false,
		"API_KEY_IN_QUERY":      false,
		"OAUTH_HTTP":            false,
		"DEPRECATED_NO_SECURITY": false,
		"SCOPE_NOT_DEFINED":     false,
	}

	for _, rule := range rules {
		if _, ok := expectedIDs[rule.ID()]; !ok {
			t.Errorf("Unexpected rule ID: %s", rule.ID())
		}
		expectedIDs[rule.ID()] = true
	}

	for id, found := range expectedIDs {
		if !found {
			t.Errorf("Missing rule: %s", id)
		}
	}
}
