// Package main provides the CLI entry point for UrlForge.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/EdgarOrtegaRamirez/urlforge/pkg/compare"
	"github.com/EdgarOrtegaRamirez/urlforge/pkg/encode"
	"github.com/EdgarOrtegaRamirez/urlforge/pkg/normalize"
	"github.com/EdgarOrtegaRamirez/urlforge/pkg/parser"
	"github.com/EdgarOrtegaRamirez/urlforge/pkg/query"
	"github.com/EdgarOrtegaRamirez/urlforge/pkg/validate"
	"github.com/spf13/cobra"
)

var version = "0.1.0"

func main() {
	rootCmd := &cobra.Command{
		Use:   "urlforge",
		Short: "Comprehensive URL manipulation toolkit",
		Long: `UrlForge is a command-line tool and library for parsing, building,
encoding/decoding, validating, normalizing, and comparing URLs.

It provides a rich set of operations for working with URLs:
- Parse URLs into structured components
- Build URLs from components
- Encode/decode URL components
- Validate URLs with configurable rules
- Normalize URLs to canonical form
- Compare and diff URLs
- Process URLs in batch from files or stdin`,
		Version: version,
	}

	// Parse command
	parseCmd := &cobra.Command{
		Use:   "parse [url]",
		Short: "Parse a URL into its components",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			return runParse(args, output)
		},
	}
	parseCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	rootCmd.AddCommand(parseCmd)

	// Build command
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build a URL from components",
		RunE: func(cmd *cobra.Command, args []string) error {
			scheme, _ := cmd.Flags().GetString("scheme")
			host, _ := cmd.Flags().GetString("host")
			port, _ := cmd.Flags().GetString("port")
			path, _ := cmd.Flags().GetString("path")
			fragment, _ := cmd.Flags().GetString("fragment")
			return runBuild(scheme, host, port, path, fragment)
		},
	}
	buildCmd.Flags().StringP("scheme", "s", "https", "URL scheme")
	buildCmd.Flags().StringP("host", "H", "", "Hostname (required)")
	buildCmd.Flags().StringP("port", "p", "", "Port number")
	buildCmd.Flags().StringP("path", "P", "/", "URL path")
	buildCmd.Flags().StringP("fragment", "f", "", "Fragment identifier")
	rootCmd.AddCommand(buildCmd)

	// Encode command
	encodeCmd := &cobra.Command{
		Use:   "encode [text]",
		Short: "Encode text for use in URLs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			method, _ := cmd.Flags().GetString("method")
			return runEncode(args[0], method)
		},
	}
	encodeCmd.Flags().StringP("method", "m", "query", "Encoding method: query, path, component, base64, rfc3986")
	rootCmd.AddCommand(encodeCmd)

	// Decode command
	decodeCmd := &cobra.Command{
		Use:   "decode [text]",
		Short: "Decode URL-encoded text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			method, _ := cmd.Flags().GetString("method")
			return runDecode(args[0], method)
		},
	}
	decodeCmd.Flags().StringP("method", "m", "query", "Decoding method: query, path, base64")
	rootCmd.AddCommand(decodeCmd)

	// Validate command
	validateCmd := &cobra.Command{
		Use:   "validate [url...]",
		Short: "Validate URLs against rules",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			level, _ := cmd.Flags().GetString("level")
			output, _ := cmd.Flags().GetString("output")
			return runValidate(args, level, output)
		},
	}
	validateCmd.Flags().StringP("level", "l", "standard", "Validation level: lenient, standard, strict")
	validateCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	rootCmd.AddCommand(validateCmd)

	// Normalize command
	normalizeCmd := &cobra.Command{
		Use:   "normalize [url...]",
		Short: "Normalize URLs to canonical form",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			level, _ := cmd.Flags().GetString("level")
			output, _ := cmd.Flags().GetString("output")
			return runNormalize(args, level, output)
		},
	}
	normalizeCmd.Flags().StringP("level", "l", "standard", "Normalization level: minimal, standard, aggressive")
	normalizeCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	rootCmd.AddCommand(normalizeCmd)

	// Compare command
	compareCmd := &cobra.Command{
		Use:   "compare <url1> <url2>",
		Short: "Compare two URLs",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			return runCompare(args[0], args[1], output)
		},
	}
	compareCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	rootCmd.AddCommand(compareCmd)

	// Query command
	queryCmd := &cobra.Command{
		Use:   "query [url]",
		Short: "Work with URL query parameters",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			action, _ := cmd.Flags().GetString("action")
			key, _ := cmd.Flags().GetString("key")
			value, _ := cmd.Flags().GetString("value")
			output, _ := cmd.Flags().GetString("output")
			return runQuery(args[0], action, key, value, output)
		},
	}
	queryCmd.Flags().StringP("action", "a", "list", "Action: list, get, add, set, remove, count, keys, diff")
	queryCmd.Flags().StringP("key", "k", "", "Parameter key")
	queryCmd.Flags().StringP("value", "v", "", "Parameter value")
	queryCmd.Flags().StringP("output", "o", "text", "Output format: text, json")
	rootCmd.AddCommand(queryCmd)

	// Batch command
	batchCmd := &cobra.Command{
		Use:   "batch [file]",
		Short: "Process multiple URLs from a file or stdin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			operation, _ := cmd.Flags().GetString("operation")
			level, _ := cmd.Flags().GetString("level")
			output, _ := cmd.Flags().GetString("output")
			return runBatch(args, operation, level, output)
		},
	}
	batchCmd.Flags().StringP("operation", "O", "parse", "Operation: parse, validate, normalize, deduplicate")
	batchCmd.Flags().StringP("level", "l", "standard", "Level for validation/normalization")
	batchCmd.Flags().StringP("output", "o", "json", "Output format: text, json")
	rootCmd.AddCommand(batchCmd)

	// Info command
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Show UrlForge version and available operations",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("UrlForge v%s\n", version)
			fmt.Println("\nAvailable operations:")
			fmt.Println("  parse       Parse a URL into components")
			fmt.Println("  build       Build a URL from components")
			fmt.Println("  encode      Encode text for URLs")
			fmt.Println("  decode      Decode URL-encoded text")
			fmt.Println("  validate    Validate URLs")
			fmt.Println("  normalize   Normalize URLs")
			fmt.Println("  compare     Compare two URLs")
			fmt.Println("  query       Work with query parameters")
			fmt.Println("  batch       Process multiple URLs")
			fmt.Println("\nExamples:")
			fmt.Println("  urlforge parse https://example.com/path?q=1#section")
			fmt.Println("  urlforge build -H example.com -P /api/v1 -p 8080")
			fmt.Println("  urlforge encode -m path 'hello world'")
			fmt.Println("  urlforge validate -l strict https://example.com")
			fmt.Println("  urlforge normalize -l aggressive 'HTTP://Example.COM:443/Path/'")
			fmt.Println("  urlforge compare https://example.com/a https://example.com/b")
			fmt.Println("  urlforge query -a get -k page https://example.com?search=test&page=1")
		},
	}
	rootCmd.AddCommand(infoCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func runParse(args []string, output string) error {
	var input string
	if len(args) > 0 {
		input = args[0]
	} else {
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			input = strings.TrimSpace(scanner.Text())
		}
	}

	if input == "" {
		return fmt.Errorf("no URL provided")
	}

	parsed, err := parser.Parse(input)
	if err != nil {
		return err
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(parsed)
	}

	fmt.Printf("URL:       %s\n", parsed.Raw)
	fmt.Printf("Scheme:    %s\n", parsed.Scheme)
	if parsed.Username != "" {
		fmt.Printf("Username:  %s\n", parsed.Username)
	}
	if parsed.Password != "" {
		fmt.Printf("Password:  %s\n", "***")
	}
	fmt.Printf("Host:      %s\n", parsed.Host)
	if parsed.Port != "" {
		fmt.Printf("Port:      %s\n", parsed.Port)
	}
	if parsed.Domain != "" {
		fmt.Printf("Domain:    %s.%s\n", parsed.Domain, parsed.TLD)
	}
	if parsed.Subdomain != "" {
		fmt.Printf("Subdomain: %s\n", parsed.Subdomain)
	}
	fmt.Printf("Path:      %s\n", parsed.Path)
	if parsed.Query != "" {
		fmt.Printf("Query:     %s\n", parsed.Query)
	}
	if parsed.Fragment != "" {
		fmt.Printf("Fragment:  %s\n", parsed.Fragment)
	}
	fmt.Printf("HTTPS:     %v\n", parsed.IsHTTPS)
	fmt.Printf("Length:    %d characters\n", parsed.TotalLength)

	if len(parsed.QueryParams) > 0 {
		fmt.Printf("\nQuery Parameters (%d):\n", len(parsed.QueryParams))
		for _, p := range parsed.QueryParams {
			fmt.Printf("  %s = %s\n", p.Key, p.Value)
		}
	}

	if len(parsed.PathSegments) > 0 {
		fmt.Printf("\nPath Segments (%d):\n", len(parsed.PathSegments))
		for i, seg := range parsed.PathSegments {
			fmt.Printf("  [%d] %s\n", i, seg)
		}
	}

	return nil
}

func runBuild(scheme, host, port, path, fragment string) error {
	if host == "" {
		return fmt.Errorf("host is required (-H flag)")
	}

	var url strings.Builder
	url.WriteString(scheme)
	url.WriteString("://")
	url.WriteString(host)
	if port != "" {
		url.WriteString(":")
		url.WriteString(port)
	}
	if path != "" {
		url.WriteString(path)
	}
	if fragment != "" {
		url.WriteString("#")
		url.WriteString(fragment)
	}

	result := url.String()
	fmt.Println(result)
	return nil
}

func runEncode(text, method string) error {
	enc := encode.New()

	var result string
	switch method {
	case "query":
		result = enc.EncodeValue(text)
	case "path", "component":
		result = enc.EncodeComponent(text)
	case "base64":
		result = enc.EncodeBase64(text)
	case "rfc3986":
		enc := encode.NewWithType(encode.EncodingRFC3986)
		result = enc.EncodeValue(text)
	default:
		return fmt.Errorf("unknown encoding method: %s", method)
	}

	fmt.Println(result)
	return nil
}

func runDecode(text, method string) error {
	enc := encode.New()

	var result string
	var err error
	switch method {
	case "query":
		result, err = enc.DecodeValue(text)
	case "path":
		result, err = enc.DecodePath(text)
	case "base64":
		result, err = enc.DecodeBase64(text)
	default:
		return fmt.Errorf("unknown decoding method: %s", method)
	}

	if err != nil {
		return err
	}

	fmt.Println(result)
	return nil
}

func runValidate(urls []string, level, output string) error {
	v := validate.New()

	switch level {
	case "lenient":
		v.WithLevel(validate.LevelLenient)
	case "strict":
		v.WithLevel(validate.LevelStrict)
	default:
		v.WithLevel(validate.LevelStandard)
	}

	results := v.ValidateBatch(urls)

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(results)
	}

	for _, r := range results {
		if r.Valid {
			fmt.Printf("✓ %s\n", r.URL)
		} else {
			fmt.Printf("✗ %s\n", r.URL)
			for _, e := range r.Errors {
				fmt.Printf("  ERROR [%s]: %s\n", e.Code, e.Message)
			}
		}
		for _, w := range r.Warnings {
			fmt.Printf("  WARN [%s]: %s\n", w.Code, w.Message)
		}
	}

	summary := validate.GetSummary(results)
	fmt.Printf("\nSummary: %d total, %d valid, %d invalid\n",
		summary["total"], summary["valid"], summary["invalid"])

	return nil
}

func runNormalize(urls []string, level, output string) error {
	n := normalize.New()

	switch level {
	case "minimal":
		n.WithLevel(normalize.LevelMinimal)
	case "aggressive":
		n.WithLevel(normalize.LevelAggressive)
	default:
		n.WithLevel(normalize.LevelStandard)
	}

	results, err := n.NormalizeBatch(urls)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	if output == "json" {
		type normResult struct {
			Original  string `json:"original"`
			Normalized string `json:"normalized"`
		}
		var output []normResult
		for i, r := range results {
			output = append(output, normResult{
				Original:  urls[i],
				Normalized: r,
			})
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	for i, r := range results {
		if r != urls[i] {
			fmt.Printf("%s\n  → %s\n", urls[i], r)
		} else {
			fmt.Printf("%s (already normalized)\n", urls[i])
		}
	}

	return nil
}

func runCompare(url1, url2, output string) error {
	result, err := compare.Compare(url1, url2)
	if err != nil {
		return err
	}

	if output == "json" {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	fmt.Printf("URL 1: %s\n", result.URL1)
	fmt.Printf("URL 2: %s\n", result.URL2)
	fmt.Printf("\nIdentical:  %v\n", result.Identical)
	fmt.Printf("Similarity: %.1f%%\n", result.Similarity*100)
	fmt.Printf("Same scheme: %v\n", result.SameScheme)
	fmt.Printf("Same host:   %v\n", result.SameHost)
	fmt.Printf("Same path:   %v\n", result.SamePath)

	if len(result.Differences) > 0 {
		fmt.Printf("\nDifferences:\n")
		for _, d := range result.Differences {
			fmt.Printf("  %s: %s → %s\n", d.Component, d.Value1, d.Value2)
		}
	}

	if result.QueryDiff != nil {
		fmt.Printf("\nQuery Differences:\n")
		for _, p := range result.QueryDiff.OnlyIn1 {
			fmt.Printf("  Only in URL 1: %s=%s\n", p.Key, p.Value)
		}
		for _, p := range result.QueryDiff.OnlyIn2 {
			fmt.Printf("  Only in URL 2: %s=%s\n", p.Key, p.Value)
		}
		for _, d := range result.QueryDiff.Different {
			fmt.Printf("  Different: %s: %s → %s\n", d.Key, d.Value1, d.Value2)
		}
	}

	return nil
}

func runQuery(rawURL, action, key, value, output string) error {
	qb, err := query.FromURL(rawURL)
	if err != nil {
		return err
	}

	switch action {
	case "list":
		params := qb.Sort().Deduplicate()
		if output == "json" {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(params)
		}
		for _, p := range params.GetAll(key) {
			if key == "" || true {
				fmt.Printf("%s=%s\n", p, "")
			}
		}
		// Simpler output
		for _, k := range qb.Keys() {
			vals := qb.GetAll(k)
			for _, v := range vals {
				fmt.Printf("%s=%s\n", k, v)
			}
		}
	case "get":
		if key == "" {
			return fmt.Errorf("key is required (-k flag)")
		}
		val := qb.Get(key)
		if val == "" {
			vals := qb.GetAll(key)
			for _, v := range vals {
				fmt.Println(v)
			}
		} else {
			fmt.Println(val)
		}
	case "add":
		if key == "" {
			return fmt.Errorf("key is required (-k flag)")
		}
		qb.Add(key, value)
		fmt.Println("?" + qb.Encode())
	case "set":
		if key == "" {
			return fmt.Errorf("key is required (-k flag)")
		}
		qb.Set(key, value)
		fmt.Println("?" + qb.Encode())
	case "remove":
		if key == "" {
			return fmt.Errorf("key is required (-k flag)")
		}
		qb.Remove(key)
		if qb.Len() > 0 {
			fmt.Println("?" + qb.Encode())
		} else {
			fmt.Println("(no query parameters)")
		}
	case "count":
		fmt.Println(qb.Len())
	case "keys":
		for _, k := range qb.Keys() {
			fmt.Println(k)
		}
	case "diff":
		// Compare with current URL's query params
		fmt.Printf("Parameters: %d\n", qb.Len())
	default:
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}

func runBatch(args []string, operation, level, output string) error {
	var urls []string

	if len(args) > 0 {
		// Read from file
		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
	} else {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" && !strings.HasPrefix(line, "#") {
				urls = append(urls, line)
			}
		}
	}

	if len(urls) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	switch operation {
	case "parse":
		return runParse(urls, output)
	case "validate":
		return runValidate(urls, level, output)
	case "normalize":
		return runNormalize(urls, level, output)
	case "deduplicate":
		deduped := compare.Deduplicate(urls)
		fmt.Printf("Removed %d duplicates\n", len(urls)-len(deduped))
		for _, u := range deduped {
			fmt.Println(u)
		}
		return nil
	default:
		return fmt.Errorf("unknown operation: %s", operation)
	}
}
