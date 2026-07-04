// Package encode provides URL encoding/decoding utilities.
package encode

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"
)

// EncodingType represents the type of URL encoding.
type EncodingType int

const (
	EncodingStandard EncodingType = iota
	EncodingRFC3986
	EncodingComponent
)

// Encoder handles URL encoding operations.
type Encoder struct {
	encodingType EncodingType
}

// New creates a new Encoder with standard encoding.
func New() *Encoder {
	return &Encoder{encodingType: EncodingStandard}
}

// NewWithType creates a new Encoder with the specified encoding type.
func NewWithType(t EncodingType) *Encoder {
	return &Encoder{encodingType: t}
}

// EncodeValue encodes a single value for use in URL query parameters.
func (e *Encoder) EncodeValue(value string) string {
	switch e.encodingType {
	case EncodingRFC3986:
		return rfc3986Encode(value)
	case EncodingComponent:
		return url.QueryEscape(value)
	default:
		return url.QueryEscape(value)
	}
}

// DecodeValue decodes a single URL-encoded value.
func (e *Encoder) DecodeValue(encoded string) (string, error) {
	return url.QueryUnescape(encoded)
}

// EncodePath encodes a URL path component.
func (e *Encoder) EncodePath(path string) string {
	// Split path into segments and encode each
	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if seg != "" {
			// Use QueryEscape for more aggressive encoding of path segments
			// This encodes &, =, and other special characters
			segments[i] = url.QueryEscape(seg)
		}
	}
	return strings.Join(segments, "/")
}

// DecodePath decodes a URL path component.
func (e *Encoder) DecodePath(encoded string) (string, error) {
	return url.PathUnescape(encoded)
}

// EncodeComponent encodes a URL component (does not encode /).
func (e *Encoder) EncodeComponent(value string) string {
	return url.PathEscape(value)
}

// DecodeComponent decodes a URL component.
func (e *Encoder) DecodeComponent(encoded string) (string, error) {
	return url.PathUnescape(encoded)
}

// EncodeFull encodes an entire URL string.
func (e *Encoder) EncodeFull(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}
	return u.String(), nil
}

// EncodeQueryString encodes all parameters in a query string.
func (e *Encoder) EncodeQueryString(query string) string {
	if query == "" {
		return ""
	}

	// Parse the query string
	params, err := url.ParseQuery(query)
	if err != nil {
		return query
	}

	// Re-encode
	return params.Encode()
}

// DecodeQueryString decodes a query string.
func (e *Encoder) DecodeQueryString(encoded string) (string, error) {
	params, err := url.ParseQuery(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode query string: %w", err)
	}

	// Reconstruct decoded query string
	var parts []string
	for key, values := range params {
		for _, val := range values {
			parts = append(parts, key+"="+val)
		}
	}
	return strings.Join(parts, "&"), nil
}

// EncodeBase64 encodes a URL to base64 (useful for embedding).
func (e *Encoder) EncodeBase64(rawURL string) string {
	return base64.URLEncoding.EncodeToString([]byte(rawURL))
}

// DecodeBase64 decodes a base64-encoded URL.
func (e *Encoder) DecodeBase64(encoded string) (string, error) {
	data, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}
	return string(data), nil
}

// IsEncoded checks if a string appears to be URL-encoded.
func (e *Encoder) IsEncoded(value string) bool {
	decoded, err := url.QueryUnescape(value)
	if err != nil {
		return false
	}
	return decoded != value
}

// DetectEncoding attempts to detect the encoding of a value.
func (e *Encoder) DetectEncoding(value string) string {
	if strings.Contains(value, "%") {
		return "percent-encoded"
	}
	if strings.Contains(value, "+") {
		return "plus-encoded"
	}
	return "plain"
}

// ConvertEncoding converts between encoding formats.
func (e *Encoder) ConvertEncoding(value, from, to string) (string, error) {
	var decoded string
	var err error

	switch from {
	case "percent":
		decoded, err = url.QueryUnescape(value)
	case "base64":
		decoded, err = e.DecodeBase64(value)
	case "plain":
		decoded = value
	default:
		return "", fmt.Errorf("unknown encoding: %s", from)
	}

	if err != nil {
		return "", err
	}

	switch to {
	case "percent":
		return url.QueryEscape(decoded), nil
	case "base64":
		return e.EncodeBase64(decoded), nil
	case "plain":
		return decoded, nil
	default:
		return "", fmt.Errorf("unknown encoding: %s", to)
	}
}

// rfc3986Encode encodes a string following RFC 3986.
// RFC 3986 encodes everything except unreserved characters.
func rfc3986Encode(s string) string {
	// unreserved = ALPHA / DIGIT / "-" / "." / "_" / "~"
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~' {
			result.WriteByte(c)
		} else {
			result.WriteString(fmt.Sprintf("%%%02X", c))
		}
	}
	return result.String()
}
