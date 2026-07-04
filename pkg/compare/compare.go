// Package compare provides URL comparison and diffing utilities.
package compare

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// ComparisonResult represents the comparison of two URLs.
type ComparisonResult struct {
	URL1          string            `json:"url1"`
	URL2          string            `json:"url2"`
	Identical     bool              `json:"identical"`
	Similarity    float64           `json:"similarity"`
	Differences   []Difference      `json:"differences,omitempty"`
	SameHost      bool              `json:"same_host"`
	SamePath      bool              `json:"same_path"`
	SameScheme    bool              `json:"same_scheme"`
	SameQueryKeys bool              `json:"same_query_keys"`
	QueryDiff     *QueryDifference  `json:"query_diff,omitempty"`
}

// Difference represents a specific difference between two URLs.
type Difference struct {
	Component string `json:"component"`
	Value1    string `json:"value1"`
	Value2    string `json:"value2"`
}

// QueryDifference represents differences in query parameters.
type QueryDifference struct {
	OnlyIn1   []QueryParam `json:"only_in_url1,omitempty"`
	OnlyIn2   []QueryParam `json:"only_in_url2,omitempty"`
	Different []ParamDiff  `json:"different_values,omitempty"`
}

// QueryParam represents a query parameter.
type QueryParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// ParamDiff represents a difference in a query parameter value.
type ParamDiff struct {
	Key   string `json:"key"`
	Value1 string `json:"value1"`
	Value2 string `json:"value2"`
}

// Compare compares two URLs and returns detailed comparison.
func Compare(rawURL1, rawURL2 string) (*ComparisonResult, error) {
	u1, err := url.Parse(rawURL1)
	if err != nil {
		return nil, fmt.Errorf("failed to parse first URL: %w", err)
	}
	u2, err := url.Parse(rawURL2)
	if err != nil {
		return nil, fmt.Errorf("failed to parse second URL: %w", err)
	}

	result := &ComparisonResult{
		URL1:      rawURL1,
		URL2:      rawURL2,
		Identical: rawURL1 == rawURL2,
	}

	// Compare components
	result.SameScheme = strings.EqualFold(u1.Scheme, u2.Scheme)
	result.SameHost = strings.EqualFold(u1.Hostname(), u2.Hostname()) && u1.Port() == u2.Port()
	result.SamePath = u1.Path == u2.Path

	// Compare query parameters
	q1 := parseQueryParams(u1.RawQuery)
	q2 := parseQueryParams(u2.RawQuery)
	result.SameQueryKeys = haveSameKeys(q1, q2)

	// Calculate similarity
	result.Similarity = calculateSimilarity(u1, u2)

	// Collect differences
	if !result.SameScheme {
		result.Differences = append(result.Differences, Difference{
			Component: "scheme",
			Value1:    u1.Scheme,
			Value2:    u2.Scheme,
		})
	}
	if !result.SameHost {
		result.Differences = append(result.Differences, Difference{
			Component: "host",
			Value1:    u1.Host,
			Value2:    u2.Host,
		})
	}
	if u1.Port() != u2.Port() {
		result.Differences = append(result.Differences, Difference{
			Component: "port",
			Value1:    u1.Port(),
			Value2:    u2.Port(),
		})
	}
	if !result.SamePath {
		result.Differences = append(result.Differences, Difference{
			Component: "path",
			Value1:    u1.Path,
			Value2:    u2.Path,
		})
	}

	// Query parameter differences
	if u1.RawQuery != u2.RawQuery {
		qDiff := compareQueryParams(q1, q2)
		if !qDiff.isEmpty() {
			result.QueryDiff = qDiff
		}
	}

	if !result.SameScheme || !result.SameHost || !result.SamePath || result.QueryDiff != nil {
		result.Identical = false
	}

	return result, nil
}

// CompareNormalized compares two URLs after normalization.
func CompareNormalized(rawURL1, rawURL2 string) (*ComparisonResult, error) {
	// Simple normalization: lowercase, remove fragments, sort params
	norm1 := normalizeSimple(rawURL1)
	norm2 := normalizeSimple(rawURL2)

	return Compare(norm1, norm2)
}

// IsSimilar checks if two URLs are similar (same host and path, different query).
func IsSimilar(rawURL1, rawURL2 string) (bool, error) {
	u1, err := url.Parse(rawURL1)
	if err != nil {
		return false, err
	}
	u2, err := url.Parse(rawURL2)
	if err != nil {
		return false, err
	}

	return strings.EqualFold(u1.Hostname(), u2.Hostname()) &&
		strings.EqualFold(u1.Path, u2.Path), nil
}

// GroupSimilar groups a list of URLs by similarity.
func GroupSimilar(urls []string) [][]string {
	groups := make(map[string][]string)

	for _, rawURL := range urls {
		u, err := url.Parse(rawURL)
		if err != nil {
			continue
		}
		key := strings.ToLower(u.Hostname() + u.Path)
		groups[key] = append(groups[key], rawURL)
	}

	var result [][]string
	for _, group := range groups {
		if len(group) > 1 {
			result = append(result, group)
		}
	}
	return result
}

// Deduplicate removes duplicate URLs from a list.
func Deduplicate(urls []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, rawURL := range urls {
		normalized := normalizeSimple(rawURL)
		if !seen[normalized] {
			seen[normalized] = true
			result = append(result, rawURL)
		}
	}
	return result
}

// parseQueryParams parses query parameters into a map.
func parseQueryParams(query string) map[string][]string {
	params := make(map[string][]string)
	if query == "" {
		return params
	}

	parsed, err := url.ParseQuery(query)
	if err != nil {
		return params
	}

	for k, v := range parsed {
		params[k] = v
	}
	return params
}

// haveSameKeys checks if two parameter maps have the same keys.
func haveSameKeys(m1, m2 map[string][]string) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k := range m1 {
		if _, ok := m2[k]; !ok {
			return false
		}
	}
	return true
}

// compareQueryParams compares two parameter maps.
func compareQueryParams(q1, q2 map[string][]string) *QueryDifference {
	diff := &QueryDifference{}

	// Find params only in q1
	for k, v1 := range q1 {
		if v2, ok := q2[k]; ok {
			// Check if values are the same
			if !stringSliceEqual(v1, v2) {
				for i, val := range v1 {
					if i < len(v2) && val != v2[i] {
						diff.Different = append(diff.Different, ParamDiff{
							Key:   k,
							Value1: val,
							Value2: v2[i],
						})
					}
				}
			}
		} else {
			for _, v := range v1 {
				diff.OnlyIn1 = append(diff.OnlyIn1, QueryParam{Key: k, Value: v})
			}
		}
	}

	// Find params only in q2
	for k, v2 := range q2 {
		if _, ok := q1[k]; !ok {
			for _, v := range v2 {
				diff.OnlyIn2 = append(diff.OnlyIn2, QueryParam{Key: k, Value: v})
			}
		}
	}

	return diff
}

func (d *QueryDifference) isEmpty() bool {
	return len(d.OnlyIn1) == 0 && len(d.OnlyIn2) == 0 && len(d.Different) == 0
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// calculateSimilarity calculates a similarity score between two URLs.
func calculateSimilarity(u1, u2 *url.URL) float64 {
	score := 0.0
	total := 0.0

	// Scheme match
	total++
	if strings.EqualFold(u1.Scheme, u2.Scheme) {
		score++
	}

	// Host match
	total++
	if strings.EqualFold(u1.Hostname(), u2.Hostname()) {
		score++
	}

	// Port match
	total++
	if u1.Port() == u2.Port() {
		score++
	}

	// Path similarity
	total++
	sim := pathSimilarity(u1.Path, u2.Path)
	score += sim

	// Query similarity
	total++
	qSim := querySimilarity(u1.RawQuery, u2.RawQuery)
	score += qSim

	return score / total
}

// pathSimilarity calculates similarity between two paths.
func pathSimilarity(p1, p2 string) float64 {
	if p1 == p2 {
		return 1.0
	}

	s1 := strings.Split(strings.Trim(p1, "/"), "/")
	s2 := strings.Split(strings.Trim(p2, "/"), "/")

	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}

	matches := 0
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	// Count matching segments at same positions
	for i := 0; i < len(s1) && i < len(s2); i++ {
		if s1[i] == s2[i] {
			matches++
		}
	}

	if maxLen == 0 {
		return 1.0
	}
	return float64(matches) / float64(maxLen)
}

// querySimilarity calculates similarity between two query strings.
func querySimilarity(q1, q2 string) float64 {
	if q1 == q2 {
		return 1.0
	}

	if q1 == "" || q2 == "" {
		return 0.0
	}

	p1 := parseQueryParams(q1)
	p2 := parseQueryParams(q2)

	// Collect all keys
	allKeys := make(map[string]bool)
	for k := range p1 {
		allKeys[k] = true
	}
	for k := range p2 {
		allKeys[k] = true
	}

	if len(allKeys) == 0 {
		return 1.0
	}

	matches := 0
	for k := range allKeys {
		v1, ok1 := p1[k]
		v2, ok2 := p2[k]
		if ok1 && ok2 && stringSliceEqual(v1, v2) {
			matches++
		}
	}

	return float64(matches) / float64(len(allKeys))
}

// normalizeSimple performs basic URL normalization.
func normalizeSimple(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	host := strings.ToLower(u.Hostname())
	if u.Port() != "" {
		u.Host = host + ":" + u.Port()
	} else {
		u.Host = host
	}

	// Remove fragment
	u.Fragment = ""

	// Sort query params
	if u.RawQuery != "" {
		params, err := url.ParseQuery(u.RawQuery)
		if err == nil {
			keys := make([]string, 0, len(params))
			for k := range params {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			var parts []string
			for _, k := range keys {
				vals := params[k]
				sort.Strings(vals)
				for _, v := range vals {
					parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
				}
			}
			u.RawQuery = strings.Join(parts, "&")
		}
	}

	return u.String()
}
