package directives

import (
	"context"
	"fmt"
	"strings"
)

// DoDirective implements the :do directive
type DoDirective struct {
	BaseDirective
}

// NewDoDirective creates a new do directive
func NewDoDirective() *DoDirective {
	return &DoDirective{
		BaseDirective: BaseDirective{name: ":do"},
	}
}

// Process implements Directive
func (d *DoDirective) Process(ctx context.Context, content []string) (string, error) {
	// For now, just return the action as a string
	// In the future, this could actually execute the action
	return fmt.Sprintf("Executed action: %s", strings.Join(content, "\n")), nil
}

// CompileToPython implements Directive
func (d *DoDirective) CompileToPython(index int, content []string, resultsDir string) string {
	var sb strings.Builder

	// Generate the function
	sb.WriteString(fmt.Sprintf("def process_do_%d():\n", index))
	sb.WriteString(fmt.Sprintf("    action = '''\n%s\n'''\n", strings.Join(content, "\n")))
	sb.WriteString("    # Execute the action and potentially generate new blocks\n")
	sb.WriteString("    print(f'Executing action: {action}')\n")

	// Add block generation capability
	sb.WriteString("    # Example actions that can generate blocks:\n")
	sb.WriteString("    if 'generate_block' in action.lower():\n")
	sb.WriteString(fmt.Sprintf("        return write_ephemeral_block(':ask', 'Generated question?', %q)\n", resultsDir))
	sb.WriteString("    elif 'create_input' in action.lower():\n")
	sb.WriteString(fmt.Sprintf("        return write_ephemeral_block(':input', 'Please provide input:', %q)\n", resultsDir))
	sb.WriteString("    return None\n\n")

	return sb.String()
}

// CanGenerateBlocks implements Directive
func (d *DoDirective) CanGenerateBlocks() bool {
	return true // Do directive can generate new blocks
}
