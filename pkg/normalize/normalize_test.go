package normalize

import (
	"testing"
)

func TestNormalizeBasic(t *testing.T) {
	n := New()

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "lowercase scheme",
			url:  "HTTP://example.com/",
			want: "http://example.com/",
		},
		{
			name: "lowercase host",
			url:  "https://EXAMPLE.COM/path",
			want: "https://example.com/path",
		},
		{
			name: "remove default http port",
			url:  "http://example.com:80/path",
			want: "http://example.com/path",
		},
		{
			name: "remove default https port",
			url:  "https://example.com:443/path",
			want: "https://example.com/path",
		},
		{
			name: "keep non-default port",
			url:  "https://example.com:8080/path",
			want: "https://example.com:8080/path",
		},
		{
			name: "sort query params",
			url:  "https://example.com/path?z=1&a=2&m=3",
			want: "https://example.com/path?a=2&m=3&z=1",
		},
		{
			name: "normalize path",
			url:  "https://example.com/a/../b",
			want: "https://example.com/b",
		},
		{
			name: "ensure root path",
			url:  "https://example.com",
			want: "https://example.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := n.Normalize(tt.url)
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestNormalizeAggressive(t *testing.T) {
	n := New().WithLevel(LevelAggressive)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "remove fragment",
			url:  "https://example.com/page#section",
			want: "https://example.com/page",
		},
		{
			name: "remove trailing slash",
			url:  "https://example.com/path/",
			want: "https://example.com/path",
		},
		{
			name: "remove www",
			url:  "https://www.example.com/path",
			want: "https://example.com/path",
		},
		{
			name: "remove auth",
			url:  "https://user:pass@example.com/path",
			want: "https://example.com/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := n.Normalize(tt.url)
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestNormalizeWWW(t *testing.T) {
	n := New().WithRemoveWWW(true)

	got, err := n.Normalize("https://www.example.com/path")
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if got != "https://example.com/path" {
		t.Errorf("Normalize() = %q, want https://example.com/path", got)
	}
}

func TestNormalizeTrailingSlash(t *testing.T) {
	n := New().WithRemoveTrailingSlash(true)

	tests := []struct {
		name string
		url  string
		want string
	}{
		{"remove trailing", "https://example.com/path/", "https://example.com/path"},
		{"keep root", "https://example.com/", "https://example.com/"},
		{"no trailing", "https://example.com/path", "https://example.com/path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := n.Normalize(tt.url)
			if err != nil {
				t.Fatalf("Normalize() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.url, got, tt.want)
			}
		})
	}
}

func TestNormalizeStripAuth(t *testing.T) {
	n := New().WithStripAuth(true)

	got, err := n.Normalize("https://user:pass@example.com/path")
	if err != nil {
		t.Fatalf("Normalize() error = %v", err)
	}
	if got != "https://example.com/path" {
		t.Errorf("Normalize() = %q, want https://example.com/path", got)
	}
}

func TestNormalizeBatch(t *testing.T) {
	n := New()
	urls := []string{
		"HTTP://EXAMPLE.COM/path",
		"Https://Google.COM/search",
		"ftp://FTP.EXAMPLE.ORG/files",
	}

	results, err := n.NormalizeBatch(urls)
	if err != nil {
		t.Fatalf("NormalizeBatch() error = %v", err)
	}

	expected := []string{
		"http://example.com/path",
		"https://google.com/search",
		"ftp://ftp.example.org/files",
	}

	for i, r := range results {
		if r != expected[i] {
			t.Errorf("NormalizeBatch()[%d] = %q, want %q", i, r, expected[i])
		}
	}
}

func TestCanonical(t *testing.T) {
	n := New()
	got, err := n.Canonical("HTTPS://Example.COM:443/Path/?B=1&A=2#section")
	if err != nil {
		t.Fatalf("Canonical() error = %v", err)
	}

	// Should be fully normalized - path case is preserved, trailing slash removed
	expected := "https://example.com/Path?a=2&b=1"
	if got != expected {
		t.Errorf("Canonical() = %q, want %q", got, expected)
	}
}
