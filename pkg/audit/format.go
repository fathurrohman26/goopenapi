package audit

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// FormatText formats audit result as human-readable text
func FormatText(result *AuditResult) string {
	var sb strings.Builder

	// Header
	sb.WriteString("Security Audit Report\n")
	sb.WriteString("=====================\n\n")

	// Summary
	writeSummary(&sb, result)

	// Findings
	writeFindings(&sb, result)

	// Security Schemes
	writeSecuritySchemes(&sb, result)

	// Coverage by Tag
	writeCoverageByTag(&sb, result)

	return sb.String()
}

func writeSummary(sb *strings.Builder, result *AuditResult) {
	sb.WriteString("Summary\n")
	sb.WriteString("-------\n")

	protectedPct := 0
	if result.TotalEndpoints > 0 {
		protectedPct = (result.ProtectedEndpoints * 100) / result.TotalEndpoints
	}

	sb.WriteString(fmt.Sprintf("Total Endpoints: %d\n", result.TotalEndpoints))
	sb.WriteString(fmt.Sprintf("Protected: %d (%d%%)\n", result.ProtectedEndpoints, protectedPct))
	sb.WriteString(fmt.Sprintf("Unprotected: %d (%d%%)\n\n", result.UnprotectedEndpoints, 100-protectedPct))
}

func writeFindings(sb *strings.Builder, result *AuditResult) {
	if len(result.Findings) == 0 {
		sb.WriteString("Findings\n")
		sb.WriteString("--------\n")
		sb.WriteString("No issues found.\n\n")
		return
	}

	sb.WriteString(fmt.Sprintf("Findings (%d issues)\n", len(result.Findings)))
	sb.WriteString("-------------------\n\n")

	for _, f := range result.Findings {
		sb.WriteString(fmt.Sprintf("[%s] %s - %s\n", f.Severity, f.RuleID, f.RuleName))
		sb.WriteString(fmt.Sprintf("  Location: %s\n", f.Location))
		sb.WriteString(fmt.Sprintf("  Message: %s\n", f.Message))
		sb.WriteString(fmt.Sprintf("  Recommendation: %s\n\n", f.Recommendation))
	}
}

func writeSecuritySchemes(sb *strings.Builder, result *AuditResult) {
	if len(result.SecuritySchemes) == 0 {
		return
	}

	sb.WriteString("Security Schemes\n")
	sb.WriteString("----------------\n")

	// Sort scheme names for consistent output
	names := make([]string, 0, len(result.SecuritySchemes))
	for name := range result.SecuritySchemes {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		info := result.SecuritySchemes[name]
		sb.WriteString(fmt.Sprintf("- %s (%s): %d endpoints\n", name, info.Type, info.EndpointCount))
	}
	sb.WriteString("\n")
}

func writeCoverageByTag(sb *strings.Builder, result *AuditResult) {
	if len(result.CoverageByTag) == 0 {
		return
	}

	sb.WriteString("Coverage by Tag\n")
	sb.WriteString("---------------\n")

	// Sort tag names for consistent output
	tags := make([]string, 0, len(result.CoverageByTag))
	for tag := range result.CoverageByTag {
		tags = append(tags, tag)
	}
	sort.Strings(tags)

	for _, tag := range tags {
		coverage := result.CoverageByTag[tag]
		pct := 0
		if coverage.Total > 0 {
			pct = (coverage.Protected * 100) / coverage.Total
		}
		sb.WriteString(fmt.Sprintf("- %s: %d/%d protected (%d%%)\n", tag, coverage.Protected, coverage.Total, pct))
	}
	sb.WriteString("\n")
}

// FormatJSON formats audit result as JSON
func FormatJSON(result *AuditResult) ([]byte, error) {
	return json.MarshalIndent(result, "", "  ")
}
