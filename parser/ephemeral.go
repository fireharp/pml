package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// IsEphemeral checks if a file is an ephemeral result
func IsEphemeral(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	if fi.IsDir() {
		return false, nil
	}
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
			val, ok := metadata["is_ephemeral"]
			if !ok {
				return false, nil
			}
			isEph, isBool := val.(bool)
			if !isBool {
				return false, fmt.Errorf("is_ephemeral must be bool, but got: %v", val)
			}
			return isEph, nil
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
