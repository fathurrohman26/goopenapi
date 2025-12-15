package yahttp

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/fathurrohman26/yaswag/pkg/openapi"
)

// ValidationError represents an API validation error.
type ValidationError struct {
	Field   string `json:"field,omitempty"`
	Message string `json:"message"`
	In      string `json:"in,omitempty"` // query, path, header, body
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (in %s)", e.Field, e.Message, e.In)
	}
	return e.Message
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return "no errors"
	}
	if len(e) == 1 {
		return e[0].Error()
	}
	return fmt.Sprintf("%d validation errors", len(e))
}

// ValidationMiddleware returns a middleware that validates requests against the OpenAPI spec.
func (p *Plugin) ValidationMiddleware() Middleware {
	errorHandler := p.options.ValidationErrorHandler
	if errorHandler == nil {
		errorHandler = DefaultValidationErrorHandler
	}
	return RequestValidation(p.spec, errorHandler)
}

// RequestValidation returns a standalone request validation middleware.
func RequestValidation(spec *openapi.Document, errorHandler func(http.ResponseWriter, *http.Request, error)) Middleware {
	if errorHandler == nil {
		errorHandler = DefaultValidationErrorHandler
	}

	validator := newRequestValidator(spec)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if errs := validator.Validate(r); len(errs) > 0 {
				errorHandler(w, r, errs)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// DefaultValidationErrorHandler is the default handler for validation errors.
func DefaultValidationErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := struct {
		Error   string            `json:"error"`
		Details []ValidationError `json:"details,omitempty"`
	}{
		Error: "Validation failed",
	}

	var validationErrs ValidationErrors
	if errors.As(err, &validationErrs) {
		response.Details = validationErrs
	} else {
		response.Details = []ValidationError{{Message: err.Error()}}
	}

	_ = json.NewEncoder(w).Encode(response)
}

// requestValidator validates HTTP requests against an OpenAPI spec.
type requestValidator struct {
	spec       *openapi.Document
	pathRegexs map[string]*pathMatcher
}

type pathMatcher struct {
	regex     *regexp.Regexp
	pathItem  *openapi.PathItem
	paramKeys []string
}

func newRequestValidator(spec *openapi.Document) *requestValidator {
	v := &requestValidator{
		spec:       spec,
		pathRegexs: make(map[string]*pathMatcher),
	}

	if spec != nil && spec.Paths != nil {
		for path, item := range spec.Paths {
			v.pathRegexs[path] = v.compilePath(path, item)
		}
	}

	return v
}

func (v *requestValidator) compilePath(path string, item *openapi.PathItem) *pathMatcher {
	// Convert OpenAPI path params to regex
	var paramKeys []string
	regexPath := regexp.MustCompile(`\{([^}]+)\}`).ReplaceAllStringFunc(path, func(match string) string {
		paramName := match[1 : len(match)-1]
		paramKeys = append(paramKeys, paramName)
		return `([^/]+)`
	})

	regex := regexp.MustCompile("^" + regexPath + "$")
	return &pathMatcher{
		regex:     regex,
		pathItem:  item,
		paramKeys: paramKeys,
	}
}

// Validate validates an HTTP request against the OpenAPI spec.
func (v *requestValidator) Validate(r *http.Request) ValidationErrors {
	var errs ValidationErrors

	if v.spec == nil || v.spec.Paths == nil {
		return errs
	}

	// Find matching path
	matcher, pathParams := v.matchPath(r.URL.Path)
	if matcher == nil {
		// Path not found in spec - skip validation
		return errs
	}

	// Get operation for method
	operation := v.getOperation(matcher.pathItem, r.Method)
	if operation == nil {
		// Method not defined - skip validation
		return errs
	}

	// Validate parameters
	errs = append(errs, v.validateParameters(r, operation, pathParams)...)

	return errs
}

func (v *requestValidator) matchPath(path string) (*pathMatcher, map[string]string) {
	for _, matcher := range v.pathRegexs {
		if matches := matcher.regex.FindStringSubmatch(path); matches != nil {
			params := make(map[string]string)
			for i, key := range matcher.paramKeys {
				if i+1 < len(matches) {
					params[key] = matches[i+1]
				}
			}
			return matcher, params
		}
	}
	return nil, nil
}

func (v *requestValidator) getOperation(pathItem *openapi.PathItem, method string) *openapi.Operation {
	switch strings.ToUpper(method) {
	case "GET":
		return pathItem.Get
	case "POST":
		return pathItem.Post
	case "PUT":
		return pathItem.Put
	case "DELETE":
		return pathItem.Delete
	case "PATCH":
		return pathItem.Patch
	case "OPTIONS":
		return pathItem.Options
	case "HEAD":
		return pathItem.Head
	case "TRACE":
		return pathItem.Trace
	}
	return nil
}

func (v *requestValidator) validateParameters(r *http.Request, op *openapi.Operation, pathParams map[string]string) ValidationErrors {
	var errs ValidationErrors

	for _, param := range op.Parameters {
		if param == nil {
			continue
		}

		value, found := v.extractParamValue(r, param, pathParams)

		if err := v.validateParameter(param, value, found); err != nil {
			errs = append(errs, *err)
		}
	}

	return errs
}

func (v *requestValidator) extractParamValue(r *http.Request, param *openapi.Parameter, pathParams map[string]string) (string, bool) {
	switch param.In {
	case openapi.ParameterInPath:
		val, found := pathParams[param.Name]
		return val, found
	case openapi.ParameterInQuery:
		return r.URL.Query().Get(param.Name), r.URL.Query().Has(param.Name)
	case openapi.ParameterInHeader:
		val := r.Header.Get(param.Name)
		return val, val != ""
	case openapi.ParameterInCookie:
		cookie, err := r.Cookie(param.Name)
		if err != nil {
			return "", false
		}
		return cookie.Value, true
	}
	return "", false
}

func (v *requestValidator) validateParameter(param *openapi.Parameter, value string, found bool) *ValidationError {
	if param.Required && !found {
		return &ValidationError{
			Field:   param.Name,
			Message: "required parameter is missing",
			In:      string(param.In),
		}
	}

	if !found || param.Schema == nil {
		return nil
	}

	return v.validateValue(value, param.Schema, param.Name, string(param.In))
}

func (v *requestValidator) validateValue(value string, schema *openapi.Schema, field, in string) *ValidationError {
	if schema == nil || len(schema.Type) == 0 {
		return nil
	}

	if err := v.validateType(value, schema.Type[0], field, in); err != nil {
		return err
	}

	return v.validateEnum(value, schema.Enum, field, in)
}

func (v *requestValidator) validateType(value, schemaType, field, in string) *ValidationError {
	switch schemaType {
	case openapi.TypeInteger:
		if _, err := strconv.ParseInt(value, 10, 64); err != nil {
			return &ValidationError{Field: field, Message: "must be an integer", In: in}
		}
	case openapi.TypeNumber:
		if _, err := strconv.ParseFloat(value, 64); err != nil {
			return &ValidationError{Field: field, Message: "must be a number", In: in}
		}
	case openapi.TypeBoolean:
		if !isValidBoolean(value) {
			return &ValidationError{Field: field, Message: "must be a boolean", In: in}
		}
	}
	return nil
}

func isValidBoolean(value string) bool {
	return value == "true" || value == "false" || value == "1" || value == "0"
}

func (v *requestValidator) validateEnum(value string, enum []any, field, in string) *ValidationError {
	if len(enum) == 0 {
		return nil
	}

	for _, e := range enum {
		if fmt.Sprintf("%v", e) == value {
			return nil
		}
	}
	return &ValidationError{Field: field, Message: "value not in allowed enum values", In: in}
}

// ValidateRequest validates a single request against an OpenAPI spec.
func ValidateRequest(spec *openapi.Document, r *http.Request) ValidationErrors {
	validator := newRequestValidator(spec)
	return validator.Validate(r)
}
