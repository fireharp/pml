package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fireharp/pml/impl1/llm"
	"github.com/fireharp/pml/impl1/parser"

	"github.com/joho/godotenv"
)

func main() {
	// Set up logging
	log.SetFlags(0)

	// Parse command line flags
	forceProcess := flag.Bool("force", false, "Force processing of all files, ignoring cache")
	targetFile := flag.String("file", "", "Process only this specific file")
	cleanup := flag.Bool("cleanup", false, "Clean up all generated files (*.pml.py and .pml folders)")
	workspaceDirFlag := flag.String("dir", "", "Set workspace directory (defaults to current directory)")
	flag.Parse()

	// Environment variables:
	// PML_DEBUG=1 - Enable debug logging
	// Load .env if exists, but don't warn if missing
	_ = godotenv.Load()

	// Get workspace directory
	workspaceDir := *workspaceDirFlag
	if workspaceDir == "" {
		var err error
		workspaceDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get workspace directory: %v", err)
		}
	} else {
		// Convert to absolute path if relative
		if !filepath.IsAbs(workspaceDir) {
			cwd, err := os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}
			workspaceDir = filepath.Join(cwd, workspaceDir)
		}
	}

	// Handle cleanup if requested
	if *cleanup {
		if err := cleanupGeneratedFiles(workspaceDir); err != nil {
			log.Fatalf("Cleanup failed: %v", err)
		}
		return
	}

	// Setup directory structure
	sourcesDir := filepath.Join(workspaceDir, "sources") // Add sources subdirectory
	resultsDir := filepath.Join(workspaceDir, "results")

	// Create directories if they don't exist
	for _, dir := range []string{sourcesDir, resultsDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create .pml directory for cache
	pmlDir := filepath.Join(sourcesDir, ".pml")
	if err := os.MkdirAll(pmlDir, 0755); err != nil {
		log.Fatalf("Failed to create .pml directory: %v", err)
	}

	// Initialize LLM client
	llmClient, err := llm.NewClient()
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	// Initialize parser - using sourcesDir for both source and compiled files
	pmlParser := parser.NewParser(llmClient, sourcesDir, sourcesDir, resultsDir)
	pmlParser.SetForceProcess(*forceProcess)

	// Initialize file processor
	processor := &FileProcessor{
		parser:       pmlParser,
		forceProcess: *forceProcess,
	}

	if *targetFile != "" {
		// Process only the specified file
		filePath := *targetFile
		if !filepath.IsAbs(filePath) {
			filePath = filepath.Join(workspaceDir, filePath)
		}
		if err := processor.ProcessFile(context.Background(), filePath); err != nil {
			log.Fatalf("Error processing %s: %v\n", filePath, err)
		}
		return
	}

	// Process all PML files
	log.Printf("Processing all PML files in %s\n", sourcesDir)
	if *forceProcess {
		// Use concurrent processing for all files
		if err := processor.ProcessFile(context.Background(), ""); err != nil {
			log.Fatalf("Error processing files: %v\n", err)
		}
	} else {
		// Process files sequentially
		err = filepath.Walk(sourcesDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && parser.IsPMLFile(path) {
				fmt.Printf("Processing file: %s\n", path)
				if err := processor.ProcessFile(context.Background(), path); err != nil {
					log.Printf("Error processing %s: %v\n", path, err)
				}
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking directory: %v", err)
		}
	}
}

// FileProcessor implements the file processing logic
type FileProcessor struct {
	parser       *parser.Parser
	forceProcess bool
}

// ProcessFile processes a file
func (p *FileProcessor) ProcessFile(ctx context.Context, path string) error {
	if !parser.IsPMLFile(path) {
		return nil // Skip non-PML files
	}

	if p.forceProcess {
		fmt.Printf("=== Force processing file: %s ===\n", path)
	} else {
		fmt.Printf("=== Processing file: %s ===\n", path)
	}

	// Use concurrent processing for multiple files
	if p.forceProcess {
		return p.parser.ProcessAllFiles(ctx)
	}
	return p.parser.ProcessFile(ctx, path)
}

// cleanupGeneratedFiles removes all generated PML files and directories
func cleanupGeneratedFiles(workspaceDir string) error {
	// Find and remove all .pml.py files and .pml directories
	err := filepath.Walk(workspaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Skip if file doesn't exist (might have been removed already)
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		// Skip processing files in .pml directories since we'll remove the whole directory
		if strings.Contains(path, "/.pml/") || strings.Contains(path, "\\.pml\\") {
			return nil
		}

		// Clean up result links in .pml files first
		if !info.IsDir() && strings.HasSuffix(path, ".pml") {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read PML file %s: %w", path, err)
			}

			lines := strings.Split(string(content), "\n")
			var newLines []string

			for _, line := range lines {
				trimmed := strings.TrimSpace(line)
				// Keep plain :-- lines, only remove result links
				if strings.HasPrefix(trimmed, ":--(r/") {
					// Replace result link with plain :--
					newLines = append(newLines, ":--")
					continue
				}
				newLines = append(newLines, line)
			}

			newContent := strings.Join(newLines, "\n")
			if newContent != string(content) {
				fmt.Printf("Cleaning result links from: %s\n", path)
				if err := os.WriteFile(path, []byte(newContent), 0644); err != nil {
					return fmt.Errorf("failed to update PML file %s: %w", path, err)
				}
			}
		}

		// Remove .pml.py files and block files
		if !info.IsDir() && (strings.HasSuffix(path, ".pml.py") || strings.Contains(path, ".pml.block_")) {
			fmt.Printf("Removing file: %s\n", path)
			if err := os.Remove(path); err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove %s: %w", path, err)
				}
			}
		}

		// Remove .pml directories last
		if info.IsDir() && info.Name() == ".pml" {
			fmt.Printf("Removing directory: %s\n", path)
			if err := os.RemoveAll(path); err != nil {
				if !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove directory %s: %w", path, err)
				}
			}
			return filepath.SkipDir // Skip processing contents of removed directory
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cleanup walk failed: %w", err)
	}

	fmt.Println("Cleanup completed successfully")
	return nil
}
