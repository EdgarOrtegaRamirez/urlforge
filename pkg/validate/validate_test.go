package validate

import (
	"testing"
)

func TestValidateBasic(t *testing.T) {
	v := New()

	tests := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{"valid https", "https://example.com", true},
		{"valid http", "http://example.com", true},
		{"valid with path", "https://example.com/path/to/page", true},
		{"valid with query", "https://example.com?foo=bar", true},
		{"valid ftp", "ftp://files.example.com", true},
		{"empty URL", "", false},
		{"no scheme", "example.com", false},
		{"javascript scheme", "javascript:alert(1)", false},
		{"blocked port 0", "https://example.com:0", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.url)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q).Valid = %v, want %v", tt.url, result.Valid, tt.wantValid)
				for _, e := range result.Errors {
					t.Logf("  Error: %s - %s", e.Code, e.Message)
				}
			}
		})
	}
}

func TestValidateStrict(t *testing.T) {
	v := New().WithLevel(LevelStrict)

	tests := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{"valid https", "https://example.com", true},
		{"ip address warning", "https://192.168.1.1", true},          // warning, not error
		{"auth info warning", "https://user:pass@example.com", true}, // warning
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.url)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q).Valid = %v, want %v", tt.url, result.Valid, tt.wantValid)
			}
		})
	}
}

func TestValidateBlockedSchemes(t *testing.T) {
	v := New().WithBlockedSchemes("javascript", "data")

	tests := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{"blocked javascript", "javascript:alert(1)", false},
		{"blocked data", "data:text/html,<h1>Hello</h1>", false},
		{"allowed https", "https://example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.url)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q).Valid = %v, want %v", tt.url, result.Valid, tt.wantValid)
			}
		})
	}
}

func TestValidateMaxLength(t *testing.T) {
	v := New().WithMaxLength(50)

	longURL := "https://example.com/" + string(make([]byte, 100))
	for i := 11; i < len(longURL) && i < 61; i++ {
		longURL = longURL[:i] + "a" + longURL[i+1:]
	}

	result := v.Validate(longURL)
	if result.Valid {
		t.Error("Expected long URL to be invalid")
	}
}

func TestValidateBatch(t *testing.T) {
	v := New()
	urls := []string{
		"https://example.com",
		"https://google.com",
		"invalid",
	}

	results := v.ValidateBatch(urls)
	if len(results) != 3 {
		t.Errorf("ValidateBatch() returned %d results, want 3", len(results))
	}

	// First two should be valid
	if !results[0].Valid {
		t.Error("First URL should be valid")
	}
	if !results[1].Valid {
		t.Error("Second URL should be valid")
	}
	// Third should be invalid
	if results[2].Valid {
		t.Error("Third URL should be invalid")
	}
}

func TestGetSummary(t *testing.T) {
	results := []*ValidationResult{
		{Valid: true, Errors: []ValidationError{}, Warnings: []ValidationWarning{{}}},
		{Valid: false, Errors: []ValidationError{{}}, Warnings: []ValidationWarning{}},
		{Valid: true, Errors: []ValidationError{}, Warnings: []ValidationWarning{}},
	}

	summary := GetSummary(results)
	if summary["total"] != 3 {
		t.Errorf("total = %d, want 3", summary["total"])
	}
	if summary["valid"] != 2 {
		t.Errorf("valid = %d, want 2", summary["valid"])
	}
	if summary["invalid"] != 1 {
		t.Errorf("invalid = %d, want 1", summary["invalid"])
	}
	if summary["errors"] != 1 {
		t.Errorf("errors = %d, want 1", summary["errors"])
	}
	if summary["warnings"] != 1 {
		t.Errorf("warnings = %d, want 1", summary["warnings"])
	}
}

func TestValidateWithAllowedSchemes(t *testing.T) {
	v := New().WithLevel(LevelStrict).WithAllowedSchemes("https")

	result := v.Validate("http://example.com")
	if result.Valid {
		t.Error("Expected http to be invalid when only https is allowed in strict mode")
	}

	result = v.Validate("https://example.com")
	if !result.Valid {
		t.Error("Expected https to be valid")
	}
}

func TestValidateBlockedDomains(t *testing.T) {
	v := New().WithBlockedDomains("evil.com", "malware.org")

	tests := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{"blocked exact", "https://evil.com", false},
		{"blocked subdomain", "https://sub.evil.com", false},
		{"allowed domain", "https://good.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.Validate(tt.url)
			if result.Valid != tt.wantValid {
				t.Errorf("Validate(%q).Valid = %v, want %v", tt.url, result.Valid, tt.wantValid)
			}
		})
	}
}
