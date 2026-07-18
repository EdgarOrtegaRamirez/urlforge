package parser

import (
	"testing"
)

func TestParseBasic(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		wantHost string
		wantPath string
		wantErr  bool
	}{
		{
			name:     "simple https",
			url:      "https://example.com",
			wantHost: "example.com",
			wantPath: "/",
			wantErr:  false,
		},
		{
			name:     "with port",
			url:      "https://example.com:8080/api",
			wantHost: "example.com",
			wantPath: "/api",
			wantErr:  false,
		},
		{
			name:     "with query",
			url:      "https://example.com/search?q=hello&page=1",
			wantHost: "example.com",
			wantPath: "/search",
			wantErr:  false,
		},
		{
			name:     "with fragment",
			url:      "https://example.com/page#section",
			wantHost: "example.com",
			wantPath: "/page",
			wantErr:  false,
		},
		{
			name:     "with auth",
			url:      "https://user:pass@example.com/admin",
			wantHost: "example.com",
			wantPath: "/admin",
			wantErr:  false,
		},
		{
			name:     "empty URL",
			url:      "",
			wantHost: "",
			wantPath: "",
			wantErr:  true,
		},
		{
			name:     "complex URL",
			url:      "https://sub.domain.co.uk:443/path/to/resource?foo=bar&baz=qux#top",
			wantHost: "sub.domain.co.uk",
			wantPath: "/path/to/resource",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := Parse(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if parsed.Host != tt.wantHost {
					t.Errorf("Host = %s, want %s", parsed.Host, tt.wantHost)
				}
				if parsed.Path != tt.wantPath {
					t.Errorf("Path = %s, want %s", parsed.Path, tt.wantPath)
				}
			}
		})
	}
}

func TestParseDomain(t *testing.T) {
	tests := []struct {
		name          string
		url           string
		wantDomain    string
		wantTLD       string
		wantSubdomain string
	}{
		{
			name:       "simple domain",
			url:        "https://example.com",
			wantDomain: "example",
			wantTLD:    "com",
		},
		{
			name:          "subdomain",
			url:           "https://api.example.com",
			wantDomain:    "example",
			wantTLD:       "com",
			wantSubdomain: "api",
		},
		{
			name:          "multi-level subdomain",
			url:           "https://v1.api.example.co.uk",
			wantDomain:    "example",
			wantTLD:       "co.uk",
			wantSubdomain: "v1.api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := Parse(tt.url)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if parsed.Domain != tt.wantDomain {
				t.Errorf("Domain = %s, want %s", parsed.Domain, tt.wantDomain)
			}
			if parsed.TLD != tt.wantTLD {
				t.Errorf("TLD = %s, want %s", parsed.TLD, tt.wantTLD)
			}
			if parsed.Subdomain != tt.wantSubdomain {
				t.Errorf("Subdomain = %s, want %s", parsed.Subdomain, tt.wantSubdomain)
			}
		})
	}
}

func TestParseQueryParams(t *testing.T) {
	parsed, err := Parse("https://example.com?foo=bar&baz=qux&foo=123")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(parsed.QueryParams) != 3 {
		t.Errorf("QueryParams count = %d, want 3", len(parsed.QueryParams))
	}

	// Check first param
	if parsed.QueryParams[0].Key != "foo" || parsed.QueryParams[0].Value != "bar" {
		t.Errorf("First param = %s=%s, want foo=bar",
			parsed.QueryParams[0].Key, parsed.QueryParams[0].Value)
	}

	// Check duplicate key
	fooCount := 0
	for _, p := range parsed.QueryParams {
		if p.Key == "foo" {
			fooCount++
		}
	}
	if fooCount != 2 {
		t.Errorf("foo param count = %d, want 2", fooCount)
	}
}

func TestParseHTTPS(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantHTTPS bool
	}{
		{"https", "https://example.com", true},
		{"http", "http://example.com", false},
		{"ftp", "ftp://example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, err := Parse(tt.url)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if parsed.IsHTTPS != tt.wantHTTPS {
				t.Errorf("IsHTTPS = %v, want %v", parsed.IsHTTPS, tt.wantHTTPS)
			}
		})
	}
}

func TestParseBatch(t *testing.T) {
	input := `https://example.com
https://google.com
# comment line
https://github.com

`
	results, err := ParseBatch(input)
	if err != nil {
		t.Fatalf("ParseBatch() error = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("ParseBatch() count = %d, want 3", len(results))
	}
}

func TestExtractPaths(t *testing.T) {
	urls := []*ParsedURL{
		{Path: "/api/v1/users"},
		{Path: "/api/v1/posts"},
		{Path: "/api/v2/comments"},
	}

	paths := ExtractPaths(urls)
	expected := []string{
		"/api",
		"/api/v1",
		"/api/v1/posts",
		"/api/v1/users",
		"/api/v2",
		"/api/v2/comments",
	}

	if len(paths) != len(expected) {
		t.Errorf("ExtractPaths() count = %d, want %d", len(paths), len(expected))
	}

	for i, p := range paths {
		if p != expected[i] {
			t.Errorf("ExtractPaths()[%d] = %s, want %s", i, p, expected[i])
		}
	}
}

func TestParsePathSegments(t *testing.T) {
	parsed, err := Parse("https://example.com/a/b/c/d")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	expected := []string{"a", "b", "c", "d"}
	if len(parsed.PathSegments) != len(expected) {
		t.Errorf("PathSegments count = %d, want %d", len(parsed.PathSegments), len(expected))
	}

	for i, seg := range parsed.PathSegments {
		if seg != expected[i] {
			t.Errorf("PathSegments[%d] = %s, want %s", i, seg, expected[i])
		}
	}
}
