package query

import (
	"testing"
)

func TestFromURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantLen int
		wantErr bool
	}{
		{
			name:    "with query",
			url:     "https://example.com?foo=bar&baz=qux",
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "without query",
			url:     "https://example.com",
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "duplicate keys",
			url:     "https://example.com?tag=a&tag=b&tag=c",
			wantLen: 3,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb, err := FromURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && qb.Len() != tt.wantLen {
				t.Errorf("Len() = %d, want %d", qb.Len(), tt.wantLen)
			}
		})
	}
}

func TestAddAndSet(t *testing.T) {
	qb := New()

	// Test Add
	qb.Add("foo", "bar")
	if qb.Len() != 1 {
		t.Errorf("After Add: Len() = %d, want 1", qb.Len())
	}

	// Test Set replaces
	qb.Set("foo", "baz")
	if qb.Len() != 1 {
		t.Errorf("After Set: Len() = %d, want 1", qb.Len())
	}
	if qb.Get("foo") != "baz" {
		t.Errorf("Get('foo') = %s, want baz", qb.Get("foo"))
	}

	// Test multiple adds
	qb.Add("foo", "qux")
	if qb.Len() != 2 {
		t.Errorf("After second Add: Len() = %d, want 2", qb.Len())
	}
}

func TestRemove(t *testing.T) {
	qb := New()
	qb.Add("a", "1").Add("b", "2").Add("a", "3")

	qb.Remove("a")
	if qb.Len() != 1 {
		t.Errorf("After Remove: Len() = %d, want 1", qb.Len())
	}
	if qb.Get("b") != "2" {
		t.Errorf("Get('b') = %s, want 2", qb.Get("b"))
	}
}

func TestHas(t *testing.T) {
	qb := New()
	qb.Add("foo", "bar")

	if !qb.Has("foo") {
		t.Error("Has('foo') = false, want true")
	}
	if qb.Has("baz") {
		t.Error("Has('baz') = true, want false")
	}
}

func TestGetAll(t *testing.T) {
	qb := New()
	qb.Add("tag", "a").Add("tag", "b").Add("tag", "c")

	values := qb.GetAll("tag")
	if len(values) != 3 {
		t.Errorf("GetAll('tag') count = %d, want 3", len(values))
	}
	if values[0] != "a" || values[1] != "b" || values[2] != "c" {
		t.Errorf("GetAll('tag') = %v, want [a b c]", values)
	}
}

func TestSort(t *testing.T) {
	qb := New()
	qb.Add("z", "1").Add("a", "2").Add("m", "3")

	sorted := qb.Sort()
	keys := sorted.Keys()

	expected := []string{"a", "m", "z"}
	for i, k := range keys {
		if k != expected[i] {
			t.Errorf("Keys()[%d] = %s, want %s", i, k, expected[i])
		}
	}
}

func TestDeduplicate(t *testing.T) {
	qb := New()
	qb.Add("a", "1").Add("a", "1").Add("b", "2")

	deduped := qb.Deduplicate()
	if deduped.Len() != 2 {
		t.Errorf("After Deduplicate: Len() = %d, want 2", deduped.Len())
	}
}

func TestEncode(t *testing.T) {
	qb := New()
	qb.Add("key", "hello world")
	qb.Add("special", "a&b=c")

	encoded := qb.Encode()
	if encoded == "" {
		t.Error("Encode() returned empty string")
	}
	// Should contain encoded values
	if !contains(encoded, "key=") || !contains(encoded, "special=") {
		t.Errorf("Encode() = %s, missing expected keys", encoded)
	}
}

func TestEncodeSorted(t *testing.T) {
	qb := New()
	qb.Add("z", "1").Add("a", "2")

	encoded := qb.EncodeSorted()
	if encoded != "a=2&z=1" {
		t.Errorf("EncodeSorted() = %s, want a=2&z=1", encoded)
	}
}

func TestMerge(t *testing.T) {
	qb1 := New()
	qb1.Add("a", "1").Add("b", "2")

	qb2 := New()
	qb2.Add("b", "3").Add("c", "4")

	qb1.Merge(qb2)

	if qb1.Get("a") != "1" {
		t.Errorf("Get('a') = %s, want 1", qb1.Get("a"))
	}
	if qb1.Get("b") != "3" {
		t.Errorf("Get('b') = %s, want 3 (should be overridden)", qb1.Get("b"))
	}
	if qb1.Get("c") != "4" {
		t.Errorf("Get('c') = %s, want 4", qb1.Get("c"))
	}
}

func TestDiff(t *testing.T) {
	qb1 := New()
	qb1.Add("a", "1").Add("b", "2")

	qb2 := New()
	qb2.Add("b", "3").Add("c", "4")

	diff := qb1.Diff(qb2)

	if len(diff.Added) != 1 || diff.Added[0].Key != "c" {
		t.Errorf("Added = %v, want [c=4]", diff.Added)
	}
	if len(diff.Removed) != 1 || diff.Removed[0].Key != "a" {
		t.Errorf("Removed = %v, want [a=1]", diff.Removed)
	}
	if len(diff.Modified) != 1 || diff.Modified[0].Key != "b" {
		t.Errorf("Modified = %v, want [b: 2→3]", diff.Modified)
	}
}

func TestClear(t *testing.T) {
	qb := New()
	qb.Add("a", "1").Add("b", "2")

	qb.Clear()
	if qb.Len() != 0 {
		t.Errorf("After Clear: Len() = %d, want 0", qb.Len())
	}
}

func TestFromQueryString(t *testing.T) {
	qb, err := FromString("foo=bar&baz=qux")
	if err != nil {
		t.Fatalf("FromString() error = %v", err)
	}
	if qb.Len() != 2 {
		t.Errorf("Len() = %d, want 2", qb.Len())
	}

	// Test with leading ?
	qb2, err := FromString("?key=value")
	if err != nil {
		t.Fatalf("FromString() error = %v", err)
	}
	if qb2.Get("key") != "value" {
		t.Errorf("Get('key') = %s, want value", qb2.Get("key"))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[0:len(substr)] == substr || contains(s[1:], substr)))
}
