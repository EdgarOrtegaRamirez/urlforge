// Package validate provides URL validation with configurable rules.
package validate

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// ValidationLevel represents the strictness of validation.
type ValidationLevel int

const (
	LevelLenient ValidationLevel = iota
	LevelStandard
	LevelStrict
)

// ValidationResult represents the result of URL validation.
type ValidationResult struct {
	Valid    bool               `json:"valid"`
	URL      string             `json:"url"`
	Level    string             `json:"level"`
	Errors   []ValidationError  `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
}

// ValidationError represents a validation error.
type ValidationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field"`
}

// Validator validates URLs with configurable rules.
type Validator struct {
	level            ValidationLevel
	allowedSchemes   map[string]bool
	blockedSchemes   map[string]bool
	blockedDomains   []string
	maxLength        int
	requireScheme    bool
	requireHost      bool
	allowedPorts     map[string]bool
	blockedPorts     map[string]bool
	customPatterns   map[string]*regexp.Regexp
}

// New creates a new Validator with standard settings.
func New() *Validator {
	return &Validator{
		level:          LevelStandard,
		allowedSchemes: map[string]bool{"http": true, "https": true, "ftp": true, "ftps": true, "file": true, "mailto": true},
		blockedSchemes: map[string]bool{},
		maxLength:      2048,
		requireScheme:  true,
		requireHost:    true,
		allowedPorts:   map[string]bool{},
		blockedPorts:   map[string]bool{"0": true},
		customPatterns: make(map[string]*regexp.Regexp),
	}
}

// WithLevel sets the validation level.
func (v *Validator) WithLevel(level ValidationLevel) *Validator {
	v.level = level
	return v
}

// WithAllowedSchemes sets the allowed schemes.
func (v *Validator) WithAllowedSchemes(schemes ...string) *Validator {
	v.allowedSchemes = make(map[string]bool)
	for _, s := range schemes {
		v.allowedSchemes[strings.ToLower(s)] = true
	}
	return v
}

// WithBlockedSchemes sets blocked schemes.
func (v *Validator) WithBlockedSchemes(schemes ...string) *Validator {
	v.blockedSchemes = make(map[string]bool)
	for _, s := range schemes {
		v.blockedSchemes[strings.ToLower(s)] = true
	}
	return v
}

// WithMaxLength sets the maximum URL length.
func (v *Validator) WithMaxLength(max int) *Validator {
	v.maxLength = max
	return v
}

// WithBlockedDomains sets blocked domains.
func (v *Validator) WithBlockedDomains(domains ...string) *Validator {
	v.blockedDomains = domains
	return v
}

// WithBlockedPorts sets blocked ports.
func (v *Validator) WithBlockedPorts(ports ...string) *Validator {
	v.blockedPorts = make(map[string]bool)
	for _, p := range ports {
		v.blockedPorts[p] = true
	}
	return v
}

// Validate validates a URL and returns the result.
func (v *Validator) Validate(rawURL string) *ValidationResult {
	result := &ValidationResult{
		URL:   rawURL,
		Level: v.levelName(),
	}

	// Parse URL
	u, err := url.Parse(rawURL)
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Code:    "PARSE_ERROR",
			Message: fmt.Sprintf("Failed to parse URL: %v", err),
			Field:   "url",
		})
		return result
	}

	// Check length
	if len(rawURL) > v.maxLength {
		result.Errors = append(result.Errors, ValidationError{
			Code:    "TOO_LONG",
			Message: fmt.Sprintf("URL exceeds maximum length of %d characters", v.maxLength),
			Field:   "url",
		})
	}

	// Check scheme
	if v.requireScheme && u.Scheme == "" {
		result.Errors = append(result.Errors, ValidationError{
			Code:    "MISSING_SCHEME",
			Message: "URL must have a scheme (e.g., https://)",
			Field:   "scheme",
		})
	}

	if u.Scheme != "" {
		scheme := strings.ToLower(u.Scheme)
		if v.blockedSchemes[scheme] {
			result.Errors = append(result.Errors, ValidationError{
				Code:    "BLOCKED_SCHEME",
				Message: fmt.Sprintf("Scheme '%s' is not allowed", u.Scheme),
				Field:   "scheme",
			})
		} else if len(v.allowedSchemes) > 0 && !v.allowedSchemes[scheme] {
			if v.level >= LevelStrict {
				result.Errors = append(result.Errors, ValidationError{
					Code:    "UNKNOWN_SCHEME",
					Message: fmt.Sprintf("Scheme '%s' is not in the allowed list", u.Scheme),
					Field:   "scheme",
				})
			} else {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Code:    "UNKNOWN_SCHEME",
					Message: fmt.Sprintf("Scheme '%s' is not commonly used", u.Scheme),
					Field:   "scheme",
				})
			}
		}
	}

	// Check host
	if v.requireHost && u.Host == "" && u.Scheme != "mailto" && u.Scheme != "file" {
		result.Errors = append(result.Errors, ValidationError{
			Code:    "MISSING_HOST",
			Message: "URL must have a host",
			Field:   "host",
		})
	}

	// Check blocked domains
	if u.Hostname() != "" {
		for _, blocked := range v.blockedDomains {
			if strings.EqualFold(u.Hostname(), blocked) || strings.HasSuffix("."+u.Hostname(), "."+blocked) {
				result.Errors = append(result.Errors, ValidationError{
					Code:    "BLOCKED_DOMAIN",
					Message: fmt.Sprintf("Domain '%s' is blocked", u.Hostname()),
					Field:   "host",
				})
			}
		}
	}

	// Check port
	if u.Port() != "" {
		if v.blockedPorts[u.Port()] {
			result.Errors = append(result.Errors, ValidationError{
				Code:    "BLOCKED_PORT",
				Message: fmt.Sprintf("Port '%s' is blocked", u.Port()),
				Field:   "port",
			})
		}
	}

	// Strict mode checks
	if v.level >= LevelStrict {
		// Check for IP addresses (warn)
		if u.Hostname() != "" && net.ParseIP(u.Hostname()) != nil {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    "IP_ADDRESS",
				Message: "URL uses an IP address instead of a domain name",
				Field:   "host",
			})
		}

		// Check for authentication info
		if u.User != nil {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    "AUTH_INFO",
				Message: "URL contains authentication information",
				Field:   "userinfo",
			})
		}

		// Check for fragment
		if u.Fragment != "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Code:    "HAS_FRAGMENT",
				Message: "URL contains a fragment identifier",
				Field:   "fragment",
			})
		}
	}

	result.Valid = len(result.Errors) == 0
	return result
}

// ValidateBatch validates multiple URLs.
func (v *Validator) ValidateBatch(urls []string) []*ValidationResult {
	results := make([]*ValidationResult, len(urls))
	for i, rawURL := range urls {
		results[i] = v.Validate(rawURL)
	}
	return results
}

// GetSummary returns a summary of validation results.
func GetSummary(results []*ValidationResult) map[string]int {
	summary := map[string]int{
		"total":   len(results),
		"valid":   0,
		"invalid": 0,
		"errors":  0,
		"warnings": 0,
	}
	for _, r := range results {
		if r.Valid {
			summary["valid"]++
		} else {
			summary["invalid"]++
		}
		summary["errors"] += len(r.Errors)
		summary["warnings"] += len(r.Warnings)
	}
	return summary
}

func (v *Validator) levelName() string {
	switch v.level {
	case LevelLenient:
		return "lenient"
	case LevelStandard:
		return "standard"
	case LevelStrict:
		return "strict"
	default:
		return "unknown"
	}
}
