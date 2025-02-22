package parser

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadCacheWhenFileMissing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser.cacheFile = filepath.Join(tmpDir, "non_existent_cache.json")

	// Should gracefully handle no file
	parser.loadCache()
	if len(parser.cache) != 0 {
		t.Errorf("Expected empty cache, got %v", parser.cache)
	}
}

func TestSaveAndLoadCache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cachePath := filepath.Join(tmpDir, "cache.json")
	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser.cacheFile = cachePath

	// Write something to cache
	parser.cache["file1.pml"] = CacheEntry{
		Checksum: "abc123",
		ModTime:  time.Now(),
		Blocks:   make(map[string]BlockCache),
	}
	err = parser.saveCache()
	if err != nil {
		t.Fatalf("saveCache failed: %v", err)
	}

	// Create new parser and load
	parser2 := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser2.cacheFile = cachePath
	parser2.loadCache()

	if len(parser2.cache) != 1 {
		t.Errorf("Expected 1 entry after load, got %d", len(parser2.cache))
	}
}

func TestCorruptCacheFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	cachePath := filepath.Join(tmpDir, "cache.json")
	err = os.WriteFile(cachePath, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser.cacheFile = cachePath
	parser.loadCache()

	if len(parser.cache) != 0 {
		t.Errorf("Expected empty cache after loading corrupt file, got %d entries", len(parser.cache))
	}
}

func TestCacheExpiry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser.cacheFile = filepath.Join(tmpDir, "cache.json")

	// Add an expired entry
	oldTime := time.Now().Add(-24 * time.Hour * 30) // 30 days old
	parser.cache["expired.pml"] = CacheEntry{
		Checksum: "abc123",
		ModTime:  oldTime,
		Blocks:   make(map[string]BlockCache),
	}

	// Add a fresh entry
	parser.cache["fresh.pml"] = CacheEntry{
		Checksum: "def456",
		ModTime:  time.Now(),
		Blocks:   make(map[string]BlockCache),
	}

	err = parser.saveCache()
	if err != nil {
		t.Fatal(err)
	}

	// Load cache in new parser
	parser2 := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser2.cacheFile = parser.cacheFile
	parser2.loadCache()

	// Should only have the fresh entry
	if len(parser2.cache) != 1 {
		t.Errorf("Expected 1 entry after expiry, got %d", len(parser2.cache))
	}
	if _, ok := parser2.cache["fresh.pml"]; !ok {
		t.Error("Fresh entry missing from cache")
	}
	if _, ok := parser2.cache["expired.pml"]; ok {
		t.Error("Expired entry still present in cache")
	}
}

func TestCacheBlockResults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-cache-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	parser := NewParser(&mockLLM{response: "Test response"}, "sources", "compiled", "results")
	parser.cacheFile = filepath.Join(tmpDir, "cache.json")

	// Create a test file
	testFile := filepath.Join(tmpDir, "test.pml")
	content := `:ask
What is 2+2?
:--`
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Process file to populate cache
	err = parser.ProcessFile(nil, testFile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify cache entry exists
	entry, ok := parser.cache[testFile]
	if !ok {
		t.Fatal("Cache entry not created")
	}

	// Verify block cache
	if len(entry.Blocks) == 0 {
		t.Error("No block results cached")
	}
}
