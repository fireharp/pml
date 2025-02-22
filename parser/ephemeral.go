package parser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// IsEphemeral checks if a file is an ephemeral result
func IsEphemeral(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return false, err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# metadata:") {
			var metadata map[string]interface{}
			jsonStr := strings.TrimPrefix(line, "# metadata:")
			if err := json.Unmarshal([]byte(jsonStr), &metadata); err != nil {
				return false, err
			}
			if isEphemeral, ok := metadata["is_ephemeral"].(bool); ok {
				return isEphemeral, nil
			}
			break
		}
	}
	return false, nil
}

// ListEphemeralBlocks lists all ephemeral blocks in the results directory
func (p *Parser) ListEphemeralBlocks() ([]string, error) {
	var ephemeralBlocks []string

	files, err := os.ReadDir(p.rootResultsDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".pml") {
			path := filepath.Join(p.rootResultsDir, file.Name())
			isEphemeral, err := IsEphemeral(path)
			if err != nil {
				continue // Skip files with errors
			}
			if isEphemeral {
				ephemeralBlocks = append(ephemeralBlocks, path)
			}
		}
	}
	return ephemeralBlocks, nil
}
