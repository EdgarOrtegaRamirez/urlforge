package encode

import (
	"testing"
)

func TestEncodeValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"simple", "hello", "hello"},
		{"spaces", "hello world", "hello+world"},
		{"special chars", "a&b=c", "a%26b%3Dc"},
		{"unicode", "café", "caf%C3%A9"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := New()
			got := enc.EncodeValue(tt.value)
			if got != tt.want {
				t.Errorf("EncodeValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestDecodeValue(t *testing.T) {
	tests := []struct {
		name    string
		encoded string
		want    string
		wantErr bool
	}{
		{"simple", "hello", "hello", false},
		{"spaces", "hello+world", "hello world", false},
		{"percent encoded", "hello%20world", "hello world", false},
		{"unicode", "caf%C3%A9", "café", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enc := New()
			got, err := enc.DecodeValue(tt.encoded)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("DecodeValue(%q) = %q, want %q", tt.encoded, got, tt.want)
			}
		})
	}
}

func TestEncodeDecodeRoundtrip(t *testing.T) {
	enc := New()
	values := []string{
		"hello world",
		"a&b=c",
		"foo/bar",
		"special: chars?#",
		"",
		"unicode: café",
	}

	for _, v := range values {
		encoded := enc.EncodeValue(v)
		decoded, err := enc.DecodeValue(encoded)
		if err != nil {
			t.Errorf("DecodeValue(%q) error = %v", encoded, err)
			continue
		}
		if decoded != v {
			t.Errorf("Roundtrip failed: %q → %q → %q", v, encoded, decoded)
		}
	}
}

func TestEncodePath(t *testing.T) {
	enc := New()

	tests := []struct {
		name string
		path string
		want string
	}{
		{"simple", "/path/to/file", "/path/to/file"},
		{"spaces", "/path/to/my file", "/path/to/my+file"},
		{"special", "/a&b/c=d", "/a%26b/c%3Dd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enc.EncodePath(tt.path)
			if got != tt.want {
				t.Errorf("EncodePath(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestEncodeQueryString(t *testing.T) {
	enc := New()

	tests := []struct {
		name  string
		query string
		want  string
	}{
		{"simple", "foo=bar", "foo=bar"},
		{"spaces", "q=hello world", "q=hello+world"},
		{"multiple", "a=1&b=2", "a=1&b=2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := enc.EncodeQueryString(tt.query)
			if got == "" && tt.query != "" {
				t.Errorf("EncodeQueryString(%q) returned empty", tt.query)
			}
		})
	}
}

func TestIsEncoded(t *testing.T) {
	enc := New()

	tests := []struct {
		value string
		want  bool
	}{
		{"hello", false},
		{"hello%20world", true},
		{"hello+world", true}, // + is URL-encoded for space in query strings
		{"a%26b", true},
		{"caf%C3%A9", true},
		{"plain_text", false},
	}

	for _, tt := range tests {
		got := enc.IsEncoded(tt.value)
		if got != tt.want {
			t.Errorf("IsEncoded(%q) = %v, want %v", tt.value, got, tt.want)
		}
	}
}

func TestDetectEncoding(t *testing.T) {
	enc := New()

	tests := []struct {
		value string
		want  string
	}{
		{"hello", "plain"},
		{"hello%20world", "percent-encoded"},
		{"hello+world", "plus-encoded"},
	}

	for _, tt := range tests {
		got := enc.DetectEncoding(tt.value)
		if got != tt.want {
			t.Errorf("DetectEncoding(%q) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestBase64EncodeDecode(t *testing.T) {
	enc := New()

	tests := []struct {
		name string
		url  string
	}{
		{"simple", "https://example.com"},
		{"complex", "https://example.com/path?q=1#section"},
		{"unicode", "https://example.com/café"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := enc.EncodeBase64(tt.url)
			decoded, err := enc.DecodeBase64(encoded)
			if err != nil {
				t.Errorf("DecodeBase64() error = %v", err)
				return
			}
			if decoded != tt.url {
				t.Errorf("Roundtrip failed: %q → %q → %q", tt.url, encoded, decoded)
			}
		})
	}
}

func TestConvertEncoding(t *testing.T) {
	enc := New()

	// Convert from plain to percent-encoded
	result, err := enc.ConvertEncoding("hello world", "plain", "percent")
	if err != nil {
		t.Fatalf("ConvertEncoding() error = %v", err)
	}
	if result != "hello+world" {
		t.Errorf("ConvertEncoding() = %q, want hello+world", result)
	}

	// Convert from percent-encoded to plain
	result, err = enc.ConvertEncoding("hello+world", "percent", "plain")
	if err != nil {
		t.Fatalf("ConvertEncoding() error = %v", err)
	}
	if result != "hello world" {
		t.Errorf("ConvertEncoding() = %q, want hello world", result)
	}
}

func TestRFC3986Encode(t *testing.T) {
	enc := NewWithType(EncodingRFC3986)

	// RFC 3986 encodes spaces as %20, not +
	got := enc.EncodeValue("hello world")
	if got != "hello%20world" {
		t.Errorf("RFC3986 EncodeValue(%q) = %q, want hello%%20world", "hello world", got)
	}

	// Unreserved characters should not be encoded
	got = enc.EncodeValue("abc-_.~123")
	if got != "abc-_.~123" {
		t.Errorf("RFC3986 EncodeValue(%q) = %q, want abc-_.~123", "abc-_.~123", got)
	}
}
