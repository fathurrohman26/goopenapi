// Package audit provides security auditing for OpenAPI specifications.
package audit

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
	"gopkg.in/yaml.v3"
)

// Severity levels for audit findings
type Severity string

const (
	SeverityError   Severity = "ERROR"
	SeverityWarning Severity = "WARNING"
	SeverityInfo    Severity = "INFO"
)

// Finding represents a single audit finding
type Finding struct {
	RuleID         string   `json:"rule_id"`
	RuleName       string   `json:"rule_name"`
	Severity       Severity `json:"severity"`
	Location       string   `json:"location"`
	Message        string   `json:"message"`
	Recommendation string   `json:"recommendation"`
}

// TagCoverage tracks security coverage for a tag
type TagCoverage struct {
	Total     int `json:"total"`
	Protected int `json:"protected"`
}

// SecuritySchemeInfo contains info about a security scheme
type SecuritySchemeInfo struct {
	Type          string `json:"type"`
	In            string `json:"in,omitempty"`
	EndpointCount int    `json:"endpoint_count"`
}

// AuditResult contains the complete audit results
type AuditResult struct {
	TotalEndpoints       int                           `json:"total_endpoints"`
	ProtectedEndpoints   int                           `json:"protected_endpoints"`
	UnprotectedEndpoints int                           `json:"unprotected_endpoints"`
	Findings             []Finding                     `json:"findings"`
	EndpointsBySecurity  map[string][]string           `json:"endpoints_by_security"`
	CoverageByTag        map[string]TagCoverage        `json:"coverage_by_tag"`
	SecuritySchemes      map[string]SecuritySchemeInfo `json:"security_schemes"`
}

// Auditor performs security audits on OpenAPI documents
type Auditor struct {
	rules []Rule
}

// New creates a new Auditor with default rules
func New() *Auditor {
	return &Auditor{
		rules: DefaultRules(),
	}
}

// Audit performs a security audit on an OpenAPI document
func (a *Auditor) Audit(doc *openapi.Document) *AuditResult {
	result := &AuditResult{
		EndpointsBySecurity: make(map[string][]string),
		CoverageByTag:       make(map[string]TagCoverage),
		SecuritySchemes:     make(map[string]SecuritySchemeInfo),
	}

	// Analyze security schemes
	a.analyzeSecuritySchemes(doc, result)

	// Analyze endpoints
	a.analyzeEndpoints(doc, result)

	// Analyze tag coverage
	a.analyzeTagCoverage(doc, result)

	// Run all audit rules
	for _, rule := range a.rules {
		findings := rule.Check(doc)
		result.Findings = append(result.Findings, findings...)
	}

	return result
}

// AuditFile audits an OpenAPI specification file
func (a *Auditor) AuditFile(path string) (*AuditResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return a.AuditData(data)
}

// AuditData audits OpenAPI specification bytes (JSON or YAML)
func (a *Auditor) AuditData(data []byte) (*AuditResult, error) {
	var doc openapi.Document
	// yaml.Unmarshal handles both JSON and YAML formats
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("failed to parse spec: %w", err)
	}
	return a.Audit(&doc), nil
}

// AuditURL audits an OpenAPI specification from a URL
func (a *Auditor) AuditURL(url string) (*AuditResult, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return a.AuditData(data)
}

// analyzeSecuritySchemes extracts security scheme information
func (a *Auditor) analyzeSecuritySchemes(doc *openapi.Document, result *AuditResult) {
	if doc.Components == nil || doc.Components.SecuritySchemes == nil {
		return
	}

	for name, scheme := range doc.Components.SecuritySchemes {
		result.SecuritySchemes[name] = SecuritySchemeInfo{
			Type: scheme.Type,
			In:   scheme.In,
		}
	}
}

// analyzeEndpoints analyzes security for all endpoints
func (a *Auditor) analyzeEndpoints(doc *openapi.Document, result *AuditResult) {
	hasGlobalSecurity := len(doc.Security) > 0

	for path, pathItem := range doc.Paths {
		for _, entry := range getOperations(pathItem) {
			result.TotalEndpoints++
			endpoint := fmt.Sprintf("%s %s", entry.method, path)

			security := a.getEndpointSecurity(entry.op, doc.Security, hasGlobalSecurity)
			if len(security) > 0 {
				result.ProtectedEndpoints++
				for _, sec := range security {
					for schemeName := range sec {
						result.EndpointsBySecurity[schemeName] = append(
							result.EndpointsBySecurity[schemeName], endpoint)
						// Update scheme endpoint count
						if info, ok := result.SecuritySchemes[schemeName]; ok {
							info.EndpointCount++
							result.SecuritySchemes[schemeName] = info
						}
					}
				}
			} else {
				result.UnprotectedEndpoints++
				result.EndpointsBySecurity["none"] = append(
					result.EndpointsBySecurity["none"], endpoint)
			}
		}
	}
}

// getEndpointSecurity returns the effective security for an endpoint
func (a *Auditor) getEndpointSecurity(op *openapi.Operation, globalSecurity []openapi.SecurityRequirement, hasGlobalSecurity bool) []openapi.SecurityRequirement {
	if len(op.Security) > 0 {
		return op.Security
	}
	if hasGlobalSecurity {
		return globalSecurity
	}
	return nil
}

// analyzeTagCoverage analyzes security coverage by tag
func (a *Auditor) analyzeTagCoverage(doc *openapi.Document, result *AuditResult) {
	hasGlobalSecurity := len(doc.Security) > 0
	tagStats := make(map[string]*TagCoverage)

	for _, pathItem := range doc.Paths {
		for _, entry := range getOperations(pathItem) {
			security := a.getEndpointSecurity(entry.op, doc.Security, hasGlobalSecurity)
			isProtected := len(security) > 0

			for _, tag := range entry.op.Tags {
				if _, ok := tagStats[tag]; !ok {
					tagStats[tag] = &TagCoverage{}
				}
				tagStats[tag].Total++
				if isProtected {
					tagStats[tag].Protected++
				}
			}
		}
	}

	for tag, stats := range tagStats {
		result.CoverageByTag[tag] = *stats
	}
}

// operationEntry holds method and operation for iteration
type operationEntry struct {
	method string
	op     *openapi.Operation
}

// getOperations returns all non-nil operations from a PathItem
func getOperations(pathItem *openapi.PathItem) []operationEntry {
	entries := []operationEntry{
		{"GET", pathItem.Get},
		{"POST", pathItem.Post},
		{"PUT", pathItem.Put},
		{"DELETE", pathItem.Delete},
		{"PATCH", pathItem.Patch},
	}
	var result []operationEntry
	for _, e := range entries {
		if e.op != nil {
			result = append(result, e)
		}
	}
	return result
}
