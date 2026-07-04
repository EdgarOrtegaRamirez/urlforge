// Package parser provides URL parsing and component extraction.
package parser

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// ParsedURL represents a fully parsed URL with all components.
type ParsedURL struct {
	Raw      string `json:"raw"`
	Scheme   string `json:"scheme"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Host     string `json:"host"`
	Port     string `json:"port,omitempty"`
	Path     string `json:"path"`
	Query    string `json:"query,omitempty"`
	Fragment string `json:"fragment,omitempty"`
	// Derived
	Domain   string            `json:"domain,omitempty"`
	TLD      string            `json:"tld,omitempty"`
	Subdomain string           `json:"subdomain,omitempty"`
	PathSegments []string      `json:"path_segments,omitempty"`
	QueryParams  []QueryParam  `json:"query_params,omitempty"`
	IsHTTPS     bool           `json:"is_https"`
	IsTrailingSlash bool       `json:"is_trailing_slash"`
	TotalLength int            `json:"total_length"`
	Components  map[string]int `json:"components_length"`
}

// QueryParam represents a single query parameter.
type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Parse parses a URL string into its components.
func Parse(raw string) (*ParsedURL, error) {
	if raw == "" {
		return nil, fmt.Errorf("empty URL")
	}

	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	parsed := &ParsedURL{
		Raw:            raw,
		Scheme:         strings.ToLower(u.Scheme),
		Host:           u.Hostname(),
		Path:           u.Path,
		Query:          u.RawQuery,
		Fragment:       u.Fragment,
		IsHTTPS:        strings.ToLower(u.Scheme) == "https",
		IsTrailingSlash: strings.HasSuffix(u.Path, "/") && u.Path != "/",
		TotalLength:    len(raw),
		Components:     make(map[string]int),
	}

	// Default path to "/" if empty
	if parsed.Path == "" {
		parsed.Path = "/"
	}

	if u.User != nil {
		parsed.Username = u.User.Username()
		if p, ok := u.User.Password(); ok {
			parsed.Password = p
		}
	}

	if u.Port() != "" {
		parsed.Port = u.Port()
	}

	// Parse domain components
	if parsed.Host != "" {
		parts := strings.Split(parsed.Host, ".")
		if len(parts) >= 2 {
			// Handle multi-level TLDs
			parsed.TLD = parts[len(parts)-1]
			parsed.Domain = parts[len(parts)-2]

			// Check for multi-level TLDs (co.uk, com.au, etc.)
			twoPartTLDs := map[string]bool{
				"co.uk": true, "co.jp": true, "co.kr": true, "co.in": true,
				"co.nz": true, "co.za": true, "com.au": true, "com.br": true,
				"com.cn": true, "com.mx": true, "com.sg": true, "com.tw": true,
				"org.uk": true, "net.au": true, "or.jp": true, "ne.jp": true,
			}
			if len(parts) >= 3 {
				candidate := parts[len(parts)-2] + "." + parts[len(parts)-1]
				if twoPartTLDs[candidate] {
					parsed.TLD = candidate
					parsed.Domain = parts[len(parts)-3]
					if len(parts) > 3 {
						parsed.Subdomain = strings.Join(parts[:len(parts)-3], ".")
					}
				} else if len(parts) > 2 {
					parsed.Subdomain = strings.Join(parts[:len(parts)-2], ".")
				}
			} else if len(parts) > 2 {
				parsed.Subdomain = strings.Join(parts[:len(parts)-2], ".")
			}
		}
	}

	// Parse path segments
	if parsed.Path != "" && parsed.Path != "/" {
		segments := strings.Split(strings.Trim(parsed.Path, "/"), "/")
		parsed.PathSegments = segments
	}

	// Parse query parameters
	if parsed.Query != "" {
		params, err := url.ParseQuery(parsed.Query)
		if err == nil {
			for key, values := range params {
				for _, val := range values {
					parsed.QueryParams = append(parsed.QueryParams, QueryParam{
						Key:   key,
						Value: val,
					})
				}
			}
		}
	}

	// Calculate component lengths
	if parsed.Scheme != "" {
		parsed.Components["scheme"] = len(parsed.Scheme)
	}
	if parsed.Username != "" {
		parsed.Components["username"] = len(parsed.Username)
	}
	if parsed.Password != "" {
		parsed.Components["password"] = len(parsed.Password)
	}
	if parsed.Host != "" {
		parsed.Components["host"] = len(parsed.Host)
	}
	if parsed.Port != "" {
		parsed.Components["port"] = len(parsed.Port)
	}
	if parsed.Path != "" {
		parsed.Components["path"] = len(parsed.Path)
	}
	if parsed.Query != "" {
		parsed.Components["query"] = len(parsed.Query)
	}
	if parsed.Fragment != "" {
		parsed.Components["fragment"] = len(parsed.Fragment)
	}

	return parsed, nil
}

// ParseBatch parses multiple URLs from a newline-separated string.
func ParseBatch(input string) ([]*ParsedURL, error) {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	var results []*ParsedURL
	var errs []string

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parsed, err := Parse(line)
		if err != nil {
			errs = append(errs, fmt.Sprintf("line %d: %v", i+1, err))
			continue
		}
		results = append(results, parsed)
	}

	if len(errs) > 0 {
		return results, fmt.Errorf("parse errors:\n%s", strings.Join(errs, "\n"))
	}

	return results, nil
}

// ExtractPaths returns all unique path prefixes from a list of parsed URLs.
func ExtractPaths(urls []*ParsedURL) []string {
	pathSet := make(map[string]bool)
	for _, u := range urls {
		if u.Path != "" && u.Path != "/" {
			// Add all prefix paths
			parts := strings.Split(strings.Trim(u.Path, "/"), "/")
			for i := 1; i <= len(parts); i++ {
				prefix := "/" + strings.Join(parts[:i], "/")
				pathSet[prefix] = true
			}
		}
	}
	paths := make([]string, 0, len(pathSet))
	for p := range pathSet {
		paths = append(paths, p)
	}
	sort.Strings(paths)
	return paths
}
