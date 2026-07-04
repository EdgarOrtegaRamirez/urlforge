// Package normalize provides URL normalization and canonicalization.
package normalize

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// NormalizationLevel represents the level of normalization.
type NormalizationLevel int

const (
	LevelMinimal NormalizationLevel = iota
	LevelStandard
	LevelAggressive
)

// Normalizer handles URL normalization.
type Normalizer struct {
	level         NormalizationLevel
	lowercaseHost bool
	lowercaseScheme bool
	removeWWW     bool
	sortParams    bool
	removeFragment bool
	removeTrailingSlash bool
	removeDefaultPorts  bool
	stripAuth     bool
	decodeUnreserved bool
}

// New creates a new Normalizer with standard settings.
func New() *Normalizer {
	return &Normalizer{
		level:              LevelStandard,
		lowercaseHost:      true,
		lowercaseScheme:    true,
		removeWWW:          false,
		sortParams:         true,
		removeFragment:     false,
		removeTrailingSlash: false,
		removeDefaultPorts: true,
		stripAuth:          false,
		decodeUnreserved:   true,
	}
}

// WithLevel sets the normalization level.
func (n *Normalizer) WithLevel(level NormalizationLevel) *Normalizer {
	n.level = level
	if level >= LevelAggressive {
		n.lowercaseHost = true
		n.lowercaseScheme = true
		n.sortParams = true
		n.removeFragment = true
		n.removeTrailingSlash = true
		n.removeDefaultPorts = true
		n.decodeUnreserved = true
		n.removeWWW = true
		n.stripAuth = true
	}
	return n
}

// WithLowercaseHost enables/disables host lowercasing.
func (n *Normalizer) WithLowercaseHost(enable bool) *Normalizer {
	n.lowercaseHost = enable
	return n
}

// WithRemoveWWW enables/disables www removal.
func (n *Normalizer) WithRemoveWWW(enable bool) *Normalizer {
	n.removeWWW = enable
	return n
}

// WithSortParams enables/disables parameter sorting.
func (n *Normalizer) WithSortParams(enable bool) *Normalizer {
	n.sortParams = enable
	return n
}

// WithRemoveFragment enables/disables fragment removal.
func (n *Normalizer) WithRemoveFragment(enable bool) *Normalizer {
	n.removeFragment = enable
	return n
}

// WithRemoveTrailingSlash enables/disables trailing slash removal.
func (n *Normalizer) WithRemoveTrailingSlash(enable bool) *Normalizer {
	n.removeTrailingSlash = enable
	return n
}

// WithStripAuth enables/disables authentication stripping.
func (n *Normalizer) WithStripAuth(enable bool) *Normalizer {
	n.stripAuth = enable
	return n
}

// Normalize applies normalization rules to a URL.
func (n *Normalizer) Normalize(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	// Lowercase scheme
	if n.lowercaseScheme {
		u.Scheme = strings.ToLower(u.Scheme)
	}

	// Lowercase host
	if n.lowercaseHost {
		host := u.Hostname()
		host = strings.ToLower(host)
		if u.Port() != "" {
			u.Host = host + ":" + u.Port()
		} else {
			u.Host = host
		}
	}

	// Remove www prefix
	if n.removeWWW && strings.HasPrefix(u.Hostname(), "www.") {
		host := strings.TrimPrefix(u.Hostname(), "www.")
		if u.Port() != "" {
			u.Host = host + ":" + u.Port()
		} else {
			u.Host = host
		}
	}

	// Remove default ports
	if n.removeDefaultPorts {
		port := u.Port()
		if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
			u.Host = u.Hostname()
		}
	}

	// Strip authentication
	if n.stripAuth {
		u.User = nil
	}

	// Normalize path
	if u.Path == "" {
		u.Path = "/"
	}

	// Remove trailing slash (except for root)
	if n.removeTrailingSlash && u.Path != "/" && strings.HasSuffix(u.Path, "/") {
		u.Path = strings.TrimRight(u.Path, "/")
	}

	// Resolve dot segments in path
	u.Path = resolvePath(u.Path)

	// Sort query parameters
	if n.sortParams && u.RawQuery != "" {
		params, err := url.ParseQuery(u.RawQuery)
		if err == nil {
			u.RawQuery = sortParams(params)
		}
	}

	// Remove fragment
	if n.removeFragment {
		u.Fragment = ""
	}

	// Decode unreserved characters (RFC 3986)
	if n.decodeUnreserved {
		u.Path = decodeUnreserved(u.Path)
		if u.RawQuery != "" {
			u.RawQuery = decodeUnreserved(u.RawQuery)
		}
	}

	return u.String(), nil
}

// NormalizeBatch normalizes multiple URLs.
func (n *Normalizer) NormalizeBatch(urls []string) ([]string, error) {
	var results []string
	var errs []string

	for _, rawURL := range urls {
		normalized, err := n.Normalize(rawURL)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%v: %v", rawURL, err))
			continue
		}
		results = append(results, normalized)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("normalization errors:\n%s", strings.Join(errs, "\n"))
	}

	return results, nil
}

// Canonical returns the canonical form of a URL.
func (n *Normalizer) Canonical(rawURL string) (string, error) {
	// Apply aggressive normalization for canonical form
	canonical := New().WithLevel(LevelAggressive)
	return canonical.Normalize(rawURL)
}

// resolvePath resolves . and .. segments in a path.
func resolvePath(path string) string {
	if !strings.Contains(path, ".") {
		return path
	}

	segments := strings.Split(path, "/")
	var resolved []string

	for _, seg := range segments {
		switch seg {
		case ".", "":
			// Skip current directory references
			continue
		case "..":
			// Go up one level
			if len(resolved) > 0 {
				resolved = resolved[:len(resolved)-1]
			}
		default:
			resolved = append(resolved, seg)
		}
	}

	result := "/" + strings.Join(resolved, "/")
	return result
}

// sortParams sorts query parameters alphabetically by lowercase key.
func sortParams(params url.Values) string {
	// Get all keys, lowercase them, and sort
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)

	// Build sorted query string
	var parts []string
	// Create a map of lowercase keys to their original values
	lowerToOriginal := make(map[string]string)
	for k := range params {
		lowerToOriginal[strings.ToLower(k)] = k
	}
	for _, k := range keys {
		originalKey := lowerToOriginal[k]
		values := params[originalKey]
		sort.Strings(values)
		for _, v := range values {
			parts = append(parts, url.QueryEscape(strings.ToLower(k))+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(parts, "&")
}

// decodeUnreserved decodes percent-encoded unreserved characters.
// Unreserved = ALPHA / DIGIT / "-" / "." / "_" / "~"
func decodeUnreserved(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if i+2 < len(s) && s[i] == '%' {
			var hex byte
			_, err := fmt.Sscanf(s[i+1:i+3], "%02x", &hex)
			if err == nil {
				if (hex >= 'A' && hex <= 'Z') || (hex >= 'a' && hex <= 'z') ||
					(hex >= '0' && hex <= '9') || hex == '-' || hex == '.' || hex == '_' || hex == '~' {
					result.WriteByte(hex)
					i += 2
					continue
				}
			}
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
