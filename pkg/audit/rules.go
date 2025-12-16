package audit

import "github.com/fathurrohman26/yaswag/pkg/openapi"

// Rule interface for extensible security checks
type Rule interface {
	ID() string
	Name() string
	Severity() Severity
	Check(doc *openapi.Document) []Finding
}

// DefaultRules returns all built-in security rules
func DefaultRules() []Rule {
	return []Rule{
		&UnprotectedWriteRule{},
		&APIKeyInQueryRule{},
		&OAuthHTTPSRule{},
		&DeprecatedSecurityRule{},
		&ScopeValidationRule{},
	}
}
