package parser

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// executePython executes a Python file and returns its output
func (p *Parser) executePython(ctx context.Context, pyPath string) ([]string, error) {
	// Get project root directory (where impl1 directory is)
	projectRoot := filepath.Dir(filepath.Dir(p.sourcesDir)) // Go up two levels

	// Add both impl1 and src directories to PYTHONPATH
	env := os.Environ()
	impl1Dir := filepath.Join(projectRoot, "impl1")
	srcDir := filepath.Join(projectRoot, "src")

	pythonPathSet := false
	for i, e := range env {
		if strings.HasPrefix(e, "PYTHONPATH=") {
			env[i] = e + string(os.PathListSeparator) + impl1Dir + string(os.PathListSeparator) + srcDir
			pythonPathSet = true
			break
		}
	}
	if !pythonPathSet {
		env = append(env, fmt.Sprintf("PYTHONPATH=%s%s%s", impl1Dir, string(os.PathListSeparator), srcDir))
	}

	// Use venv Python if it exists, otherwise use system Python
	venvPython := filepath.Join(projectRoot, ".venv", "bin", "python")
	python := "python"
	if _, err := os.Stat(venvPython); err == nil {
		python = venvPython
	}

	if p.debug {
		p.debugf("Executing Python with:\n")
		p.debugf("  Path: %s\n", pyPath)
		p.debugf("  Python: %s\n", python)
		p.debugf("  Project Root: %s\n", projectRoot)
		p.debugf("  Impl1 Dir: %s\n", impl1Dir)
		p.debugf("  Src Dir: %s\n", srcDir)
		for _, e := range env {
			if strings.HasPrefix(e, "PYTHONPATH=") {
				p.debugf("  %s\n", e)
			}
		}
	}

	cmd := exec.CommandContext(ctx, python, pyPath)
	cmd.Env = env

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, context.DeadlineExceeded
		}
		return nil, fmt.Errorf("failed to execute Python: %w\nOutput: %s", err, string(output))
	}

	// Split output into lines
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	return lines, nil
}
