// Package parser implements the PML (Programming Language Markup) parser.
package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// NewParser creates a new PML parser with specified directories
func NewParser(llm LLMClient, sourcesDir, compiledDir, resultsDir string) *Parser {
	// Cache file is now stored in the .pml directory
	pmlDir := filepath.Join(sourcesDir, ".pml")
	cacheFile := filepath.Join(pmlDir, "cache.json")
	p := &Parser{
		llm:            llm,
		sourcesDir:     sourcesDir,
		compiledDir:    compiledDir, // Keep for compatibility, but will be same as sourcesDir
		rootResultsDir: resultsDir,
		cacheFile:      cacheFile,
		cache:          make(map[string]CacheEntry),
		debug:          os.Getenv("PML_DEBUG") == "1",
		forceProcess:   false,
		usedNames:      make(map[string]bool),
	}

	// Ensure cache directory exists
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		p.debugf("Warning: failed to create cache directory: %v\n", err)
	}

	// If running tests, clear the results directory.
	if os.Getenv("PML_TEST") == "1" {
		os.RemoveAll(p.rootResultsDir)
		os.MkdirAll(p.rootResultsDir, 0755)
		os.MkdirAll(p.rootResultsDir, 0755)
	}
	p.loadCache()

	return p
}

// debugf prints debug messages if debug mode is enabled
func (p *Parser) debugf(format string, args ...interface{}) {
	if p.debug {
		fmt.Printf(format, args...)
	}
}

// SetForceProcess sets whether to force process files regardless of cache
func (p *Parser) SetForceProcess(force bool) {
	p.forceProcess = force
}

// IsPMLFile checks if a file is a PML file
func IsPMLFile(path string) bool {
	// Skip files in .pml/ directory
	if strings.Contains(path, "/.pml/") || strings.Contains(path, "\\.pml\\") {
		return false
	}
	return strings.HasSuffix(strings.ToLower(path), ".pml")
}

// isLiteral checks if a string represents a literal value (number, boolean, null)
func (p *Parser) isLiteral(s string) bool {
	s = strings.TrimSpace(s)

	// Check for boolean and null
	if s == "true" || s == "false" || s == "null" {
		return true
	}

	// Check if it's a number (integer or float)
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}

	return false
}

// formatString properly escapes and quotes a string value
func (p *Parser) formatString(s string) string {
	// First, escape any special characters
	escaped := strings.ReplaceAll(s, "\\", "\\\\")
	escaped = strings.ReplaceAll(escaped, "\"", "\\\"")
	escaped = strings.ReplaceAll(escaped, "\n", "\\n")
	escaped = strings.ReplaceAll(escaped, "\r", "\\r")
	escaped = strings.ReplaceAll(escaped, "\t", "\\t")

	// Wrap in quotes
	return fmt.Sprintf("\"%s\"", escaped)
}
