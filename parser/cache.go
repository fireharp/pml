package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// loadCache loads the cache from disk
func (p *Parser) loadCache() {
	data, err := os.ReadFile(p.cacheFile)
	if err != nil {
		p.debugf("No cache file found or error reading cache: %v\n", err)
		p.cache = make(map[string]CacheEntry)
		return
	}

	if err := json.Unmarshal(data, &p.cache); err != nil {
		p.debugf("Error unmarshaling cache: %v\n", err)
		// Start with empty cache if corrupted
		p.cache = make(map[string]CacheEntry)
	}

	// Ensure all entries have initialized maps
	for path, entry := range p.cache {
		if entry.Blocks == nil {
			entry.Blocks = make(map[string]BlockCache)
			p.cache[path] = entry
		}
	}
}

// saveCache saves the cache to disk
func (p *Parser) saveCache() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(p.cacheFile), 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %w", err)
	}

	// Marshal cache with indentation for readability
	data, err := json.MarshalIndent(p.cache, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling cache: %w", err)
	}

	// Write cache file
	if err := os.WriteFile(p.cacheFile, data, 0644); err != nil {
		return fmt.Errorf("error writing cache file: %w", err)
	}

	p.debugf("Cache saved to %s\n", p.cacheFile)
	return nil
}

// calculateChecksum calculates SHA-256 checksum of file content, ignoring result links
func (p *Parser) calculateChecksum(content string) string {
	// Remove result links before calculating checksum
	resultLinkPattern := regexp.MustCompile(`:-+\(r/[a-z]+_[a-z]+\)`)
	contentWithoutLinks := resultLinkPattern.ReplaceAllString(content, ":--")

	// Normalize whitespace
	lines := strings.Split(contentWithoutLinks, "\n")
	var normalized []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized = append(normalized, trimmed)
		}
	}

	hash := sha256.Sum256([]byte(strings.Join(normalized, "\n")))
	return hex.EncodeToString(hash[:])
}
