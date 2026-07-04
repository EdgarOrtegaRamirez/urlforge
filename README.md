# UrlForge

Comprehensive URL manipulation toolkit for developers. Parse, build, encode/decode, validate, normalize, and compare URLs with a single CLI tool.

## Features

- **Parse** — Break down URLs into structured components (scheme, host, port, path, query, fragment, auth)
- **Build** — Construct URLs from components
- **Encode/Decode** — URL-encode/decode values, paths, and query strings (RFC 3986 support)
- **Validate** — Check URLs against configurable rules (scheme allow/block, domain block, max length, port block)
- **Normalize** — Canonicalize URLs (lowercase, remove www, sort params, remove fragments, resolve paths)
- **Compare** — Diff two URLs with similarity scoring and detailed component comparison
- **Query** — Manipulate query parameters (add, set, remove, list, diff, merge)
- **Batch** — Process multiple URLs from files or stdin

## Installation

```bash
go install github.com/EdgarOrtegaRamirez/urlforge/cmd/urlforge@latest
```

Or build from source:

```bash
git clone https://github.com/EdgarOrtegaRamirez/urlforge.git
cd urlforge
go build -o urlforge ./cmd/urlforge/
```

## Quick Start

```bash
# Parse a URL into components
urlforge parse "https://user:pass@example.com:8080/path?q=1#top"

# Build a URL from components
urlforge build -s https -H example.com -p 8080 -P /api/v1

# Encode text for URLs
urlforge encode -m query "hello world"      # hello+world
urlforge encode -m path "hello world"       # hello%20world
urlforge encode -m base64 "https://example.com"

# Decode URL-encoded text
urlforge decode -m query "hello+world"      # hello world

# Validate URLs
urlforge validate -l strict "https://example.com" "javascript:alert(1)"
# ✓ https://example.com
# ✗ javascript:alert(1)

# Normalize URLs
urlforge normalize -l aggressive "HTTP://WWW.Example.COM:443/Path/?B=1&A=2#section"
# → https://example.com/Path?a=2&b=1

# Compare two URLs
urlforge compare "https://example.com/a?x=1" "https://example.com/b?x=2"

# Work with query parameters
urlforge query -a list "https://example.com?search=test&page=1"
urlforge query -a get -k page "https://example.com?page=2"
urlforge query -a set -k page -v 5 "https://example.com?page=1"
```

## Commands

| Command | Description |
|---------|-------------|
| `parse [url]` | Parse URL into components |
| `build` | Build URL from components |
| `encode [text]` | Encode text for URLs |
| `decode [text]` | Decode URL-encoded text |
| `validate [url...]` | Validate URLs against rules |
| `normalize [url...]` | Normalize URLs to canonical form |
| `compare <url1> <url2>` | Compare two URLs |
| `query [url]` | Work with query parameters |
| `batch [file]` | Process multiple URLs |
| `info` | Show version and help |

## Output Formats

Most commands support `--output` flag:
- `text` (default) — Human-readable output
- `json` — Structured JSON output

```bash
urlforge parse --output json "https://example.com/path?q=1"
urlforge validate --output json "https://example.com"
```

## Validation Levels

- **lenient** — Basic URL structure validation
- **standard** (default) — Scheme validation, port checks, domain blocking
- **strict** — IP address warnings, auth info warnings, fragment warnings

```bash
urlforge validate -l lenient "ftp://files.example.com"
urlforge validate -l standard "https://example.com"
urlforge validate -l strict "https://user:pass@example.com"
```

## Normalization Levels

- **minimal** — Lowercase scheme/host only
- **standard** (default) — + sort params, remove default ports, resolve paths
- **aggressive** — + remove www, remove auth, remove fragments, remove trailing slashes

## Library Usage

```go
package main

import (
    "fmt"
    "github.com/EdgarOrtegaRamirez/urlforge/pkg/parser"
    "github.com/EdgarOrtegaRamirez/urlforge/pkg/query"
    "github.com/EdgarOrtegaRamirez/urlforge/pkg/validate"
    "github.com/EdgarOrtegaRamirez/urlforge/pkg/normalize"
    "github.com/EdgarOrtegaRamirez/urlforge/pkg/compare"
)

func main() {
    // Parse
    parsed, _ := parser.Parse("https://example.com/path?q=1")
    fmt.Println(parsed.Host)    // example.com
    fmt.Println(parsed.Path)    // /path

    // Query manipulation
    qb, _ := query.FromURL("https://example.com?page=1&search=test")
    qb.Set("page", "5")
    fmt.Println(qb.Encode())    // page=5&search=test

    // Validate
    v := validate.New().WithLevel(validate.LevelStrict)
    result := v.Validate("https://example.com")
    fmt.Println(result.Valid)   // true

    // Normalize
    n := normalize.New().WithLevel(normalize.LevelAggressive)
    normalized, _ := n.Normalize("HTTP://WWW.Example.COM")
    fmt.Println(normalized)     // https://example.com/

    // Compare
    diff, _ := compare.Compare("https://a.com/x", "https://a.com/y")
    fmt.Println(diff.Similarity) // 0.8
}
```

## Architecture

```
urlforge/
├── pkg/
│   ├── parser/    — URL parsing and component extraction
│   ├── query/     — Query parameter manipulation
│   ├── encode/    — URL encoding/decoding utilities
│   ├── validate/  — URL validation with configurable rules
│   ├── normalize/ — URL normalization and canonicalization
│   └── compare/   — URL comparison and diffing
├── cmd/urlforge/  — CLI entry point
└── tests/         — Integration tests
```

## License

MIT
