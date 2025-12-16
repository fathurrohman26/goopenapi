package audit

import (
	"fmt"
	"strings"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

// UnprotectedWriteRule warns on POST/PUT/DELETE/PATCH without security
type UnprotectedWriteRule struct{}

func (r *UnprotectedWriteRule) ID() string       { return "UNPROTECTED_WRITE" }
func (r *UnprotectedWriteRule) Name() string     { return "Unprotected write operation" }
func (r *UnprotectedWriteRule) Severity() Severity { return SeverityWarning }

func (r *UnprotectedWriteRule) Check(doc *openapi.Document) []Finding {
	var findings []Finding
	hasGlobalSecurity := len(doc.Security) > 0
	writeMethods := map[string]bool{"POST": true, "PUT": true, "DELETE": true, "PATCH": true}

	for path, pathItem := range doc.Paths {
		for _, entry := range getOperations(pathItem) {
			if !writeMethods[entry.method] {
				continue
			}

			if !hasEndpointSecurity(entry.op, hasGlobalSecurity) {
				findings = append(findings, Finding{
					RuleID:         r.ID(),
					RuleName:       r.Name(),
					Severity:       r.Severity(),
					Location:       fmt.Sprintf("%s %s", entry.method, path),
					Message:        fmt.Sprintf("%s endpoint has no security requirement", entry.method),
					Recommendation: "Add authentication/authorization requirement to protect write operations",
				})
			}
		}
	}
	return findings
}

// APIKeyInQueryRule warns when API keys use query params instead of headers
type APIKeyInQueryRule struct{}

func (r *APIKeyInQueryRule) ID() string       { return "API_KEY_IN_QUERY" }
func (r *APIKeyInQueryRule) Name() string     { return "API key in query parameter" }
func (r *APIKeyInQueryRule) Severity() Severity { return SeverityWarning }

func (r *APIKeyInQueryRule) Check(doc *openapi.Document) []Finding {
	var findings []Finding
	if doc.Components == nil || doc.Components.SecuritySchemes == nil {
		return findings
	}

	for name, scheme := range doc.Components.SecuritySchemes {
		if scheme.Type == "apiKey" && scheme.In == "query" {
			findings = append(findings, Finding{
				RuleID:         r.ID(),
				RuleName:       r.Name(),
				Severity:       r.Severity(),
				Location:       fmt.Sprintf("SecurityScheme '%s'", name),
				Message:        fmt.Sprintf("API key '%s' is passed in query parameter", name),
				Recommendation: "Use header-based API key for better security (prevents logging in URLs)",
			})
		}
	}
	return findings
}

// OAuthHTTPSRule warns when OAuth URLs don't use HTTPS
type OAuthHTTPSRule struct{}

func (r *OAuthHTTPSRule) ID() string       { return "OAUTH_HTTP" }
func (r *OAuthHTTPSRule) Name() string     { return "OAuth URL not using HTTPS" }
func (r *OAuthHTTPSRule) Severity() Severity { return SeverityError }

func (r *OAuthHTTPSRule) Check(doc *openapi.Document) []Finding {
	var findings []Finding
	if doc.Components == nil || doc.Components.SecuritySchemes == nil {
		return findings
	}

	for name, scheme := range doc.Components.SecuritySchemes {
		if scheme.Type != "oauth2" || scheme.Flows == nil {
			continue
		}
		findings = append(findings, r.checkOAuthFlows(name, scheme.Flows)...)
	}
	return findings
}

func (r *OAuthHTTPSRule) checkOAuthFlows(schemeName string, flows *openapi.OAuthFlows) []Finding {
	var findings []Finding

	checkURL := func(urlType, url string) {
		if url != "" && strings.HasPrefix(url, "http://") {
			findings = append(findings, Finding{
				RuleID:         r.ID(),
				RuleName:       r.Name(),
				Severity:       r.Severity(),
				Location:       fmt.Sprintf("SecurityScheme '%s' %s", schemeName, urlType),
				Message:        fmt.Sprintf("OAuth %s uses HTTP instead of HTTPS", urlType),
				Recommendation: "Use HTTPS for all OAuth URLs to protect tokens in transit",
			})
		}
	}

	if flows.Implicit != nil {
		checkURL("authorizationUrl", flows.Implicit.AuthorizationURL)
	}
	if flows.Password != nil {
		checkURL("tokenUrl", flows.Password.TokenURL)
	}
	if flows.ClientCredentials != nil {
		checkURL("tokenUrl", flows.ClientCredentials.TokenURL)
	}
	if flows.AuthorizationCode != nil {
		checkURL("authorizationUrl", flows.AuthorizationCode.AuthorizationURL)
		checkURL("tokenUrl", flows.AuthorizationCode.TokenURL)
	}

	return findings
}

// DeprecatedSecurityRule checks deprecated endpoints still have security
type DeprecatedSecurityRule struct{}

func (r *DeprecatedSecurityRule) ID() string       { return "DEPRECATED_NO_SECURITY" }
func (r *DeprecatedSecurityRule) Name() string     { return "Deprecated endpoint without security" }
func (r *DeprecatedSecurityRule) Severity() Severity { return SeverityInfo }

func (r *DeprecatedSecurityRule) Check(doc *openapi.Document) []Finding {
	var findings []Finding
	hasGlobalSecurity := len(doc.Security) > 0

	for path, pathItem := range doc.Paths {
		for _, entry := range getOperations(pathItem) {
			if !entry.op.Deprecated {
				continue
			}

			if !hasEndpointSecurity(entry.op, hasGlobalSecurity) {
				findings = append(findings, Finding{
					RuleID:         r.ID(),
					RuleName:       r.Name(),
					Severity:       r.Severity(),
					Location:       fmt.Sprintf("%s %s", entry.method, path),
					Message:        "Deprecated endpoint has no security requirement",
					Recommendation: "Consider adding security or removing the deprecated endpoint",
				})
			}
		}
	}
	return findings
}

// ScopeValidationRule validates OAuth scopes are defined and used
type ScopeValidationRule struct{}

func (r *ScopeValidationRule) ID() string       { return "SCOPE_NOT_DEFINED" }
func (r *ScopeValidationRule) Name() string     { return "OAuth scope not defined" }
func (r *ScopeValidationRule) Severity() Severity { return SeverityWarning }

func (r *ScopeValidationRule) Check(doc *openapi.Document) []Finding {
	var findings []Finding

	// Collect all defined scopes from security schemes
	definedScopes := r.collectDefinedScopes(doc)

	// Check scopes used in operations
	for path, pathItem := range doc.Paths {
		for _, entry := range getOperations(pathItem) {
			findings = append(findings, r.checkOperationScopes(path, entry, definedScopes)...)
		}
	}

	return findings
}

func (r *ScopeValidationRule) collectDefinedScopes(doc *openapi.Document) map[string]map[string]bool {
	scopes := make(map[string]map[string]bool)

	if doc.Components == nil || doc.Components.SecuritySchemes == nil {
		return scopes
	}

	for name, scheme := range doc.Components.SecuritySchemes {
		if scheme.Type != "oauth2" || scheme.Flows == nil {
			continue
		}
		scopes[name] = make(map[string]bool)

		collectFlowScopes := func(flow *openapi.OAuthFlow) {
			if flow == nil {
				return
			}
			for scope := range flow.Scopes {
				scopes[name][scope] = true
			}
		}

		collectFlowScopes(scheme.Flows.Implicit)
		collectFlowScopes(scheme.Flows.Password)
		collectFlowScopes(scheme.Flows.ClientCredentials)
		collectFlowScopes(scheme.Flows.AuthorizationCode)
	}

	return scopes
}

func (r *ScopeValidationRule) checkOperationScopes(path string, entry operationEntry, definedScopes map[string]map[string]bool) []Finding {
	var findings []Finding

	for _, secReq := range entry.op.Security {
		for schemeName, requiredScopes := range secReq {
			schemeScopes, schemeExists := definedScopes[schemeName]
			if !schemeExists {
				continue // Not an OAuth scheme
			}

			for _, scope := range requiredScopes {
				if !schemeScopes[scope] {
					findings = append(findings, Finding{
						RuleID:         r.ID(),
						RuleName:       r.Name(),
						Severity:       r.Severity(),
						Location:       fmt.Sprintf("%s %s", entry.method, path),
						Message:        fmt.Sprintf("Scope '%s' used but not defined in security scheme '%s'", scope, schemeName),
						Recommendation: "Define the scope in the security scheme or remove from operation",
					})
				}
			}
		}
	}

	return findings
}

// hasEndpointSecurity checks if an endpoint has security (operation or global)
func hasEndpointSecurity(op *openapi.Operation, hasGlobalSecurity bool) bool {
	return len(op.Security) > 0 || hasGlobalSecurity
}
