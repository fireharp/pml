package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// generateUniqueResultName generates a friendly name for a result file that is guaranteed to be unique
func (p *Parser) generateUniqueResultName(sourceFile string, blockIndex int, blockType string, localResultsDir string) string {
	counter := 0
	var resultName string
	for {
		// Generate base name using file hash and block index
		hash := 0
		for _, c := range sourceFile {
			hash = (hash*31 + int(c)) % len(nouns)
		}

		// Use blockIndex, counter and file hash to ensure uniqueness
		adjIndex := (blockIndex + hash + counter) % len(adjectives)
		nounIndex := ((blockIndex + hash + counter) * 7) % len(nouns)

		// Add prefix based on block type
		prefix := ""
		switch blockType {
		case DirectiveAsk:
			prefix = "ask_"
		case DirectiveDo:
			prefix = "do_"
		default:
			prefix = "result_"
		}
		resultName = fmt.Sprintf("%s%s_%s_%d", prefix, adjectives[adjIndex], nouns[nounIndex], counter)

		// Check if this name is already taken
		resultFile := fmt.Sprintf("%s.pml", resultName)
		resultPath := filepath.Join(localResultsDir, resultFile)
		if _, err := os.Stat(resultPath); os.IsNotExist(err) {
			// Name is available
			return resultName
		}
		// Name is taken, try next combination
		counter++

		// Safety check to prevent infinite loop (though extremely unlikely)
		if counter > len(adjectives)*len(nouns) {
			// If we somehow exhausted all combinations, use timestamp as fallback
			return fmt.Sprintf("result_%d", time.Now().UnixNano())
		}
	}
}

// formatResult formats a result value as valid PML
func (p *Parser) formatResult(result string) string {
	// If it looks like a number, boolean, or null, keep it as is
	if p.isLiteral(result) {
		return result
	}
	// Otherwise treat as string and properly escape
	return p.formatString(result)
}
