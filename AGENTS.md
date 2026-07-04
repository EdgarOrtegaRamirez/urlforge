# UrlForge

Comprehensive URL manipulation toolkit for developers.

## What It Does

UrlForge is a Go CLI tool and library for working with URLs. It provides:
- URL parsing into structured components
- URL building from components
- URL encoding/decoding (query, path, base64, RFC 3986)
- URL validation with configurable rules
- URL normalization to canonical form
- URL comparison with similarity scoring
- Query parameter manipulation
- Batch processing from files/stdin

## Project Structure

- `pkg/parser/` — URL parsing, component extraction, batch parsing
- `pkg/query/` — Query parameter CRUD, sorting, diffing, merging
- `pkg/encode/` — URL encoding/decoding (multiple formats)
- `pkg/validate/` — Validation rules (scheme, domain, port, length)
- `pkg/normalize/` — Normalization levels (minimal, standard, aggressive)
- `pkg/compare/` — URL comparison, similarity scoring, deduplication
- `cmd/urlforge/` — CLI with Cobra commands

## Build & Test

```bash
go build ./cmd/urlforge/
go test ./pkg/... -v
```

## Key APIs

### Parser
```go
parsed, err := parser.Parse("https://example.com/path?q=1#top")
// parsed.Host, parsed.Path, parsed.QueryParams, parsed.Domain, parsed.TLD
```

### Query
```go
qb, _ := query.FromURL("https://example.com?a=1&b=2")
qb.Set("a", "10").Remove("b").Add("c", "3")
fmt.Println(qb.EncodeSorted()) // a=10&c=3
```

### Validate
```go
v := validate.New().WithLevel(validate.LevelStrict).WithBlockedDomains("evil.com")
result := v.Validate("https://evil.com")
// result.Valid == false, result.Errors[0].Code == "BLOCKED_DOMAIN"
```

### Normalize
```go
n := normalize.New().WithLevel(normalize.LevelAggressive)
normalized, _ := n.Normalize("HTTP://WWW.Example.COM:443/Path/?A=1#top")
// https://example.com/Path?a=1
```

### Compare
```go
diff, _ := compare.Compare("https://a.com/x?q=1", "https://a.com/y?q=2")
// diff.Similarity, diff.Differences, diff.QueryDiff
```

## CLI Usage

```bash
urlforge parse "https://example.com/path?q=1"
urlforge build -H example.com -P /api -p 8080
urlforge encode -m query "hello world"
urlforge validate -l strict "https://example.com"
urlforge normalize -l aggressive "HTTP://WWW.Example.COM"
urlforge compare "https://a.com" "https://b.com"
urlforge query -a set -k page -v 2 "https://example.com?page=1"
urlforge batch -O normalize < urls.txt
```

## Dependencies

- `github.com/spf13/cobra` — CLI framework
- Standard library: `net/url`, `encoding/base64`, `sort`, `strings`, `fmt`
