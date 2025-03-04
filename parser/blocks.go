package parser

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// calculateBlockChecksum calculates SHA-256 checksum of a block's content, ignoring whitespace
func (p *Parser) calculateBlockChecksum(block Block) string {
	// Normalize block content by trimming whitespace and joining with single newlines
	var normalized strings.Builder

	// Always use lowercase for block type to ensure consistency
	normalized.WriteString(strings.ToLower(strings.TrimSpace(block.Type)))
	normalized.WriteString("\n")

	// Normalize content lines
	for _, line := range block.Content {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			normalized.WriteString(trimmed)
			normalized.WriteString("\n")
		}
	}

	hash := sha256.Sum256([]byte(normalized.String()))
	return hex.EncodeToString(hash[:])
}

// parseBlocks parses blocks from PML content
func (p *Parser) parseBlocks(content string) ([]Block, error) {
	var blocks []Block
	lines := strings.Split(content, "\n")
	var currentBlock *Block
	var blockStartPos int
	var currentPos int

	for i, line := range lines {
		lineLen := len(line) + 1 // +1 for newline
		trimmedLine := strings.TrimSpace(line)

		// Handle empty lines
		if trimmedLine == "" {
			if currentBlock != nil {
				currentBlock.Content = append(currentBlock.Content, line)
			}
			currentPos += lineLen
			continue
		}

		// Treat a line exactly equal to ":--" as the end marker.
		if trimmedLine == DirectiveEnd {
			if currentBlock == nil {
				return nil, fmt.Errorf("found end marker without a block at line %d", i+1)
			}
			currentBlock.End = currentPos + len(line)
			blocks = append(blocks, *currentBlock)
			currentBlock = nil
			currentPos += lineLen
			continue
		} else if strings.HasPrefix(trimmedLine, DirectiveEnd) {
			// If the line starts with something like ":--(r/...", skip block termination and treat as normal content.
			if currentBlock != nil {
				currentBlock.Content = append(currentBlock.Content, line)
			}
			currentPos += lineLen
			continue
		}

		switch trimmedLine {
		case DirectiveAsk, DirectiveDo:
			if currentBlock != nil {
				// Found new block without ending previous one
				return nil, fmt.Errorf("found new block without ending previous one at line %d", i+1)
			}
			currentBlock = &Block{
				Type:  trimmedLine,
				Start: currentPos,
			}
			blockStartPos = currentPos
		default:
			if currentBlock != nil {
				currentBlock.Content = append(currentBlock.Content, line)
			}
		}
		currentPos += lineLen
	}

	if currentBlock != nil {
		// File ended without closing block
		return nil, fmt.Errorf("file ended without closing block starting at position %d", blockStartPos)
	}

	// Trim trailing empty lines from each block's content
	for i := range blocks {
		for len(blocks[i].Content) > 0 && strings.TrimSpace(blocks[i].Content[len(blocks[i].Content)-1]) == "" {
			blocks[i].Content = blocks[i].Content[:len(blocks[i].Content)-1]
		}
	}

	return blocks, nil
}

// replaceBlocksInContent replaces blocks in content with their results
func (p *Parser) replaceBlocksInContent(content string, blocks []Block) string {
	var result strings.Builder

	// Add imports at the top
	result.WriteString("# Auto-generated imports for PML blocks\n")
	result.WriteString("from src.pml.directives import process_ask, process_do\n\n")

	lines := strings.Split(content, "\n")
	var currentBlock int
	var inBlock bool

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		switch {
		case trimmedLine == DirectiveAsk || trimmedLine == DirectiveDo:
			inBlock = true
			if currentBlock < len(blocks) {
				block := blocks[currentBlock]
				result.WriteString(fmt.Sprintf("# %s\n", block.Type))
				if block.Type == DirectiveAsk {
					result.WriteString(fmt.Sprintf("result_%d = process_ask('''\n", currentBlock))
				} else {
					result.WriteString(fmt.Sprintf("result_%d = process_do('''\n", currentBlock))
				}
				result.WriteString(strings.Join(block.Content, "\n"))
				result.WriteString("\n''')\n")
			}
		case trimmedLine == DirectiveEnd:
			inBlock = false
			result.WriteString("# :--\n")
			currentBlock++
		default:
			if !inBlock {
				result.WriteString(line + "\n")
			}
		}
	}

	return result.String()
}
