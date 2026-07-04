// Package query provides URL query parameter manipulation.
package query

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

// QueryBuilder manages URL query parameters.
type QueryBuilder struct {
	params []Param
}

// Param represents a query parameter.
type Param struct {
	Key   string
	Value string
}

// New creates a new QueryBuilder.
func New() *QueryBuilder {
	return &QueryBuilder{}
}

// FromURL parses query parameters from a URL string.
func FromURL(rawURL string) (*QueryBuilder, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	qb := New()
	if u.RawQuery != "" {
		params, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return nil, fmt.Errorf("failed to parse query: %w", err)
		}
		for key, values := range params {
			for _, val := range values {
				qb.params = append(qb.params, Param{Key: key, Value: val})
			}
		}
	}
	return qb, nil
}

// FromString parses query parameters from a query string (without the ?).
func FromString(queryStr string) (*QueryBuilder, error) {
	qb := New()
	if queryStr == "" {
		return qb, nil
	}

	// Remove leading ? if present
	queryStr = strings.TrimPrefix(queryStr, "?")

	params, err := url.ParseQuery(queryStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse query string: %w", err)
	}

	for key, values := range params {
		for _, val := range values {
			qb.params = append(qb.params, Param{Key: key, Value: val})
		}
	}
	return qb, nil
}

// Add adds a parameter. If the key already exists, it appends.
func (qb *QueryBuilder) Add(key, value string) *QueryBuilder {
	qb.params = append(qb.params, Param{Key: key, Value: value})
	return qb
}

// Set sets a parameter, replacing any existing value for the key.
func (qb *QueryBuilder) Set(key, value string) *QueryBuilder {
	// Remove existing
	var filtered []Param
	for _, p := range qb.params {
		if p.Key != key {
			filtered = append(filtered, p)
		}
	}
	qb.params = filtered
	qb.params = append(qb.params, Param{Key: key, Value: value})
	return qb
}

// Remove removes all parameters with the given key.
func (qb *QueryBuilder) Remove(key string) *QueryBuilder {
	var filtered []Param
	for _, p := range qb.params {
		if p.Key != key {
			filtered = append(filtered, p)
		}
	}
	qb.params = filtered
	return qb
}

// Has checks if a parameter key exists.
func (qb *QueryBuilder) Has(key string) bool {
	for _, p := range qb.params {
		if p.Key == key {
			return true
		}
	}
	return false
}

// Get returns the first value for the given key, or empty string.
func (qb *QueryBuilder) Get(key string) string {
	for _, p := range qb.params {
		if p.Key == key {
			return p.Value
		}
	}
	return ""
}

// GetAll returns all values for the given key.
func (qb *QueryBuilder) GetAll(key string) []string {
	var values []string
	for _, p := range qb.params {
		if p.Key == key {
			values = append(values, p.Value)
		}
	}
	return values
}

// Keys returns all unique parameter keys in order.
func (qb *QueryBuilder) Keys() []string {
	seen := make(map[string]bool)
	var keys []string
	for _, p := range qb.params {
		if !seen[p.Key] {
			seen[p.Key] = true
			keys = append(keys, p.Key)
		}
	}
	return keys
}

// Len returns the number of parameters.
func (qb *QueryBuilder) Len() int {
	return len(qb.params)
}

// Clear removes all parameters.
func (qb *QueryBuilder) Clear() *QueryBuilder {
	qb.params = nil
	return qb
}

// Sort sorts parameters alphabetically by key.
func (qb *QueryBuilder) Sort() *QueryBuilder {
	sort.Slice(qb.params, func(i, j int) bool {
		if qb.params[i].Key != qb.params[j].Key {
			return qb.params[i].Key < qb.params[j].Key
		}
		return qb.params[i].Value < qb.params[j].Value
	})
	return qb
}

// Deduplicate removes duplicate key-value pairs.
func (qb *QueryBuilder) Deduplicate() *QueryBuilder {
	seen := make(map[string]bool)
	var unique []Param
	for _, p := range qb.params {
		key := p.Key + "=" + p.Value
		if !seen[key] {
			seen[key] = true
			unique = append(unique, p)
		}
	}
	qb.params = unique
	return qb
}

// Encode encodes special characters in parameter values.
func (qb *QueryBuilder) Encode() string {
	values := url.Values{}
	for _, p := range qb.params {
		values.Add(p.Key, p.Value)
	}
	return values.Encode()
}

// EncodeSorted encodes parameters sorted by key.
func (qb *QueryBuilder) EncodeSorted() string {
	sorted := make([]Param, len(qb.params))
	copy(sorted, qb.params)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Key != sorted[j].Key {
			return sorted[i].Key < sorted[j].Key
		}
		return sorted[i].Value < sorted[j].Value
	})

	var parts []string
	for _, p := range sorted {
		parts = append(parts, url.QueryEscape(p.Key)+"="+url.QueryEscape(p.Value))
	}
	return strings.Join(parts, "&")
}

// Merge adds parameters from another QueryBuilder, overriding existing keys.
func (qb *QueryBuilder) Merge(other *QueryBuilder) *QueryBuilder {
	for _, p := range other.params {
		qb.Set(p.Key, p.Value)
	}
	return qb
}

// Diff returns parameters that are different between two QueryBuilders.
func (qb *QueryBuilder) Diff(other *QueryBuilder) *QueryDiff {
	currentMap := make(map[string]string)
	for _, p := range qb.params {
		currentMap[p.Key] = p.Value
	}

	otherMap := make(map[string]string)
	for _, p := range other.params {
		otherMap[p.Key] = p.Value
	}

	diff := &QueryDiff{}

	// Find added and modified
	for key, val := range otherMap {
		if currentVal, exists := currentMap[key]; !exists {
			diff.Added = append(diff.Added, Param{Key: key, Value: val})
		} else if currentVal != val {
			diff.Modified = append(diff.Modified, ParamChange{
				Key:     key,
				Old:     currentVal,
				New:     val,
			})
		}
	}

	// Find removed
	for key, val := range currentMap {
		if _, exists := otherMap[key]; !exists {
			diff.Removed = append(diff.Removed, Param{Key: key, Value: val})
		}
	}

	return diff
}

// QueryDiff represents differences between two query parameter sets.
type QueryDiff struct {
	Added    []Param       `json:"added,omitempty"`
	Removed  []Param       `json:"removed,omitempty"`
	Modified []ParamChange `json:"modified,omitempty"`
}

// ParamChange represents a changed parameter.
type ParamChange struct {
	Key string `json:"key"`
	Old string `json:"old"`
	New string `json:"new"`
}

// IsEmpty returns true if there are no differences.
func (d *QueryDiff) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Modified) == 0
}
