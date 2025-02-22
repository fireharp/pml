package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestExecutePythonNoVenv ensures we run system Python if venv missing
func TestExecutePythonNoVenv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-python-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pyFile := filepath.Join(tmpDir, "test.py")
	err = os.WriteFile(pyFile, []byte("print('Hello from system Python')"), 0755)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	lines, err := parser.executePython(context.Background(), pyFile)
	if err != nil {
		t.Fatalf("executePython error: %v", err)
	}
	if len(lines) == 0 || lines[0] != "Hello from system Python" {
		t.Errorf("Expected 'Hello from system Python', got: %v", lines)
	}
}

// TestExecutePythonWithVenv tests Python execution in a virtual environment
func TestExecutePythonWithVenv(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-python-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create venv structure
	venvBin := filepath.Join(tmpDir, "venv", "bin")
	err = os.MkdirAll(venvBin, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create test Python file
	pyFile := filepath.Join(tmpDir, "venv_test.py")
	err = os.WriteFile(pyFile, []byte(`
import sys
print('Python path:', sys.executable)
`), 0755)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	lines, err := parser.executePython(context.Background(), pyFile)
	if err != nil {
		t.Fatalf("executePython error: %v", err)
	}

	// Verify output contains Python path
	found := false
	for _, line := range lines {
		if strings.Contains(line, "Python path:") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected Python path in output, got: %v", lines)
	}
}

// TestExecutePythonWithImports tests Python execution with imports
func TestExecutePythonWithImports(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-python-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a Python package
	pkgDir := filepath.Join(tmpDir, "testpkg")
	err = os.MkdirAll(pkgDir, 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create __init__.py
	err = os.WriteFile(filepath.Join(pkgDir, "__init__.py"), []byte(""), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create module.py
	moduleContent := `
def test_function():
    return "Hello from test package"
`
	err = os.WriteFile(filepath.Join(pkgDir, "module.py"), []byte(moduleContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create test file that imports the package
	testContent := `
import sys
sys.path.append('.')
from testpkg.module import test_function
print(test_function())
`
	testFile := filepath.Join(tmpDir, "import_test.py")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))
	lines, err := parser.executePython(context.Background(), testFile)
	if err != nil {
		t.Fatalf("executePython error: %v", err)
	}

	if len(lines) == 0 || lines[0] != "Hello from test package" {
		t.Errorf("Expected 'Hello from test package', got: %v", lines)
	}
}

// TestExecutePythonErrors tests error handling in Python execution
func TestExecutePythonErrors(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-python-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name     string
		content  string
		wantErr  bool
		errMatch string
	}{
		{
			name: "syntax error",
			content: `
if True
    print("Missing colon")
`,
			wantErr:  true,
			errMatch: "SyntaxError",
		},
		{
			name: "import error",
			content: `
import non_existent_module
`,
			wantErr:  true,
			errMatch: "ModuleNotFoundError",
		},
		{
			name: "runtime error",
			content: `
def cause_error():
    raise ValueError("Test error")
cause_error()
`,
			wantErr:  true,
			errMatch: "ValueError",
		},
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tc.name+".py")
			err := os.WriteFile(testFile, []byte(tc.content), 0644)
			if err != nil {
				t.Fatal(err)
			}

			_, err = parser.executePython(context.Background(), testFile)
			if tc.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				} else if !strings.Contains(err.Error(), tc.errMatch) {
					t.Errorf("Expected error containing %q, got %v", tc.errMatch, err)
				}
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestExecutePythonTimeout tests that execution is cancelled after timeout
func TestExecutePythonTimeout(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-python-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a Python script that sleeps
	content := `
import time
time.sleep(10)  # Sleep for 10 seconds
print("Done")
`
	testFile := filepath.Join(tmpDir, "timeout_test.py")
	err = os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), filepath.Join(tmpDir, "results"))

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err = parser.executePython(ctx, testFile)
	if err == nil {
		t.Error("Expected timeout error but got none")
	} else if !strings.Contains(err.Error(), "context deadline exceeded") && !strings.Contains(err.Error(), "signal: killed") {
		t.Errorf("Expected deadline exceeded or killed error, got: %v", err)
	}
}
