package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// generateUniqueResultName generates a friendly name for a result file that is guaranteed to be unique
var uniqueNameCounters sync.Map // maps "sourceFile_blockIndex_blockType" (string) to int

func (p *Parser) generateUniqueResultName(sourceFile string, blockIndex int, blockType string, localResultsDir string) string {
	p.usedNamesMu.Lock()
	defer p.usedNamesMu.Unlock()

	// Initialize usedNames if nil
	if p.usedNames == nil {
		p.usedNames = make(map[string]bool)
	}

	key := fmt.Sprintf("%s_%d_%s", sourceFile, blockIndex, blockType)
	var counter int
	if cnt, ok := uniqueNameCounters.Load(key); ok {
		counter = cnt.(int)
	} else {
		counter = 0
	}

	var resultName string
	for {
		// Compute a hash index from the source file for variation.
		hash := 0
		for _, c := range sourceFile {
			hash = (hash*31 + int(c)) % len(nouns)
		}
		adjIndex := (blockIndex + hash + counter) % len(adjectives)
		nounIndex := ((blockIndex + hash + counter) * 7) % len(nouns)

		prefix := ""
		switch blockType {
		case DirectiveAsk:
			prefix = "ask_"
		case DirectiveDo:
			prefix = "do_"
		default:
			prefix = "result_"
		}

		// Ensure consistent naming by using a deterministic pattern
		resultName = fmt.Sprintf("%s%s_%s_block%d_%d.pml", prefix, adjectives[adjIndex], nouns[nounIndex], blockIndex, counter)

		// Check both in-memory and on disk for uniqueness
		if _, exists := p.usedNames[resultName]; exists {
			counter++
			continue
		}

		// Check if file exists in the local results directory
		if _, err := os.Stat(filepath.Join(localResultsDir, resultName)); err == nil {
			counter++
			continue
		}

		// Mark as used and update counter
		p.usedNames[resultName] = true
		uniqueNameCounters.Store(key, counter+1)
		break
	}
	return resultName
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
