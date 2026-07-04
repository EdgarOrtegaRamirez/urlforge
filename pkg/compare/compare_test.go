package compare

import (
	"testing"
)

func TestCompareIdentical(t *testing.T) {
	result, err := Compare("https://example.com/path?q=1", "https://example.com/path?q=1")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if !result.Identical {
		t.Error("Expected URLs to be identical")
	}
	if result.Similarity != 1.0 {
		t.Errorf("Similarity = %f, want 1.0", result.Similarity)
	}
}

func TestCompareDifferentHost(t *testing.T) {
	result, err := Compare("https://example.com", "https://google.com")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.Identical {
		t.Error("Expected URLs to not be identical")
	}
	if result.SameHost {
		t.Error("Expected hosts to be different")
	}
}

func TestCompareDifferentPath(t *testing.T) {
	result, err := Compare("https://example.com/a", "https://example.com/b")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.SamePath {
		t.Error("Expected paths to be different")
	}
	if !result.SameHost {
		t.Error("Expected hosts to be the same")
	}
}

func TestCompareDifferentQuery(t *testing.T) {
	result, err := Compare("https://example.com?a=1&b=2", "https://example.com?a=1&c=3")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if !result.SameHost {
		t.Error("Expected hosts to be the same")
	}
	if !result.SamePath {
		t.Error("Expected paths to be the same")
	}
	if result.QueryDiff == nil {
		t.Error("Expected query differences")
	}
}

func TestIsSimilar(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		expected bool
	}{
		{"same host and path", "https://example.com/path?q=1", "https://example.com/path?q=2", true},
		{"different host", "https://example.com", "https://google.com", false},
		{"different path", "https://example.com/a", "https://example.com/b", false},
		{"case insensitive", "https://Example.COM/Path", "https://example.com/path", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsSimilar(tt.url1, tt.url2)
			if err != nil {
				t.Fatalf("IsSimilar() error = %v", err)
			}
			if got != tt.expected {
				t.Errorf("IsSimilar(%q, %q) = %v, want %v", tt.url1, tt.url2, got, tt.expected)
			}
		})
	}
}

func TestDeduplicate(t *testing.T) {
	urls := []string{
		"https://example.com/path?q=1",
		"https://example.com/path?q=2",
		"https://example.com/path?q=1", // duplicate (after normalization)
		"https://google.com",
	}

	deduped := Deduplicate(urls)
	if len(deduped) != 3 {
		t.Errorf("Deduplicate() returned %d URLs, want 3", len(deduped))
	}
}

func TestGroupSimilar(t *testing.T) {
	urls := []string{
		"https://example.com/a?q=1",
		"https://example.com/a?q=2",
		"https://example.com/b",
		"https://google.com/search?q=test",
		"https://google.com/search?q=go",
	}

	groups := GroupSimilar(urls)
	if len(groups) != 2 {
		t.Errorf("GroupSimilar() returned %d groups, want 2", len(groups))
	}
}

func TestQueryDifference(t *testing.T) {
	result, err := Compare("https://example.com?a=1&b=2", "https://example.com?b=3&c=4")
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.QueryDiff == nil {
		t.Fatal("Expected query differences")
	}

	// Should have a=1 only in URL1
	found := false
	for _, p := range result.QueryDiff.OnlyIn1 {
		if p.Key == "a" && p.Value == "1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find a=1 only in URL1")
	}

	// Should have c=4 only in URL2
	found = false
	for _, p := range result.QueryDiff.OnlyIn2 {
		if p.Key == "c" && p.Value == "4" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find c=4 only in URL2")
	}

	// Should have b different
	found = false
	for _, d := range result.QueryDiff.Different {
		if d.Key == "b" && d.Value1 == "2" && d.Value2 == "3" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected to find b: 2→3")
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		url1     string
		url2     string
		minSim   float64
	}{
		{"identical", "https://example.com", "https://example.com", 1.0},
		{"same host", "https://example.com/a", "https://example.com/b", 0.5},
		{"completely different", "https://example.com", "https://google.com/path", 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Compare(tt.url1, tt.url2)
			if err != nil {
				t.Fatalf("Compare() error = %v", err)
			}
			if result.Similarity < tt.minSim {
				t.Errorf("Similarity = %f, want >= %f", result.Similarity, tt.minSim)
			}
		})
	}
}
