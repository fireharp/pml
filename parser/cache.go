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
	"time"
)

// loadCache loads the cache from disk
func (p *Parser) loadCache() {
	data, err := os.ReadFile(p.cacheFile)
	if err != nil {
		p.debugf("No cache file found or error reading cache: %v\n", err)
		p.cacheMu.Lock()
		p.cache = make(map[string]CacheEntry)
		p.cacheMu.Unlock()
		return
	}

	var tempCache map[string]CacheEntry
	if err := json.Unmarshal(data, &tempCache); err != nil {
		p.debugf("Error unmarshaling cache: %v\n", err)
		// Start with empty cache if corrupted
		p.cacheMu.Lock()
		p.cache = make(map[string]CacheEntry)
		p.cacheMu.Unlock()
		return
	}

	// Ensure all entries have initialized maps
	p.cacheMu.Lock()
	p.cache = make(map[string]CacheEntry)
	for path, entry := range tempCache {
		if entry.Blocks == nil {
			entry.Blocks = make(map[string]BlockCache)
		}
		// Clean up expired entries (older than 24 hours)
		for blockID, blockCache := range entry.Blocks {
			if time.Since(blockCache.ModTime) > 24*time.Hour {
				delete(entry.Blocks, blockID)
			}
		}
		p.cache[path] = entry
	}
	p.cacheMu.Unlock()
}

// saveCache saves the cache to disk
func (p *Parser) saveCache() error {
	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(p.cacheFile), 0755); err != nil {
		return fmt.Errorf("error creating cache directory: %w", err)
	}

	// Get a copy of the cache under read lock
	p.cacheMu.RLock()
	cacheCopy := make(map[string]CacheEntry)
	for k, v := range p.cache {
		cacheCopy[k] = v
	}
	p.cacheMu.RUnlock()

	// Marshal cache with indentation for readability
	data, err := json.MarshalIndent(cacheCopy, "", "  ")
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
