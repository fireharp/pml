package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsEphemeral(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-ephemeral-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	ephemeralFile := filepath.Join(tmpDir, "ephemeral_result.pml")
	normalFile := filepath.Join(tmpDir, "normal_result.pml")

	// Create ephemeral file
	err = os.WriteFile(ephemeralFile, []byte(`# metadata:{"is_ephemeral":true}
:ask
Some ephemeral content
:--
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create normal file
	err = os.WriteFile(normalFile, []byte(`:ask
Some normal content
:--
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	isEph, err := IsEphemeral(ephemeralFile)
	if err != nil {
		t.Fatalf("IsEphemeral error: %v", err)
	}
	if !isEph {
		t.Errorf("Expected ephemeral to be true for %s", ephemeralFile)
	}

	isEph, err = IsEphemeral(normalFile)
	if err != nil {
		t.Fatalf("IsEphemeral error: %v", err)
	}
	if isEph {
		t.Errorf("Expected ephemeral to be false for %s", normalFile)
	}
}

func TestListEphemeralBlocks(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-ephemeral-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	files := []struct {
		name      string
		content   string
		ephemeral bool
	}{
		{
			name: "ephemeral1.pml",
			content: `# metadata:{"is_ephemeral":true}
:ask
Q ephemeral 1
:--`,
			ephemeral: true,
		},
		{
			name: "ephemeral2.pml",
			content: `# metadata:{"is_ephemeral":true}
:ask
Q ephemeral 2
:--`,
			ephemeral: true,
		},
		{
			name: "normal.pml",
			content: `:ask
Q normal
:--`,
			ephemeral: false,
		},
	}

	for _, f := range files {
		err := os.WriteFile(filepath.Join(tmpDir, f.name), []byte(f.content), 0644)
		if err != nil {
			t.Fatal(err)
		}
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), tmpDir)
	ephList, err := parser.ListEphemeralBlocks()
	if err != nil {
		t.Fatalf("ListEphemeralBlocks failed: %v", err)
	}

	if len(ephList) != 2 {
		t.Errorf("Expected exactly 2 ephemeral blocks, got %d", len(ephList))
	}

	// Verify each ephemeral file is in the list
	foundFiles := make(map[string]bool)
	for _, path := range ephList {
		foundFiles[filepath.Base(path)] = true
	}

	for _, f := range files {
		if f.ephemeral {
			if !foundFiles[f.name] {
				t.Errorf("Expected to find ephemeral file %s in list", f.name)
			}
		} else {
			if foundFiles[f.name] {
				t.Errorf("Found non-ephemeral file %s in ephemeral list", f.name)
			}
		}
	}
}

func TestInvalidMetadata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-ephemeral-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testCases := []struct {
		name    string
		content string
		wantErr bool
		isEphem bool
	}{
		{
			name: "invalid_json.pml",
			content: `# metadata:{invalid json}
:ask
Content
:--`,
			wantErr: true,
			isEphem: false,
		},
		{
			name: "missing_field.pml",
			content: `# metadata:{"other_field":true}
:ask
Content
:--`,
			wantErr: false,
			isEphem: false,
		},
		{
			name: "wrong_type.pml",
			content: `# metadata:{"is_ephemeral":"not_a_bool"}
:ask
Content
:--`,
			wantErr: true,
			isEphem: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath := filepath.Join(tmpDir, tc.name)
			err := os.WriteFile(filePath, []byte(tc.content), 0644)
			if err != nil {
				t.Fatal(err)
			}

			isEph, err := IsEphemeral(filePath)
			if (err != nil) != tc.wantErr {
				t.Errorf("IsEphemeral() error = %v, wantErr %v", err, tc.wantErr)
			}
			if err == nil && isEph != tc.isEphem {
				t.Errorf("IsEphemeral() = %v, want %v", isEph, tc.isEphem)
			}
		})
	}
}

func TestEphemeralBlockProcessing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pml-ephemeral-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an ephemeral file
	content := `# metadata:{"is_ephemeral":true}
:ask
What is 2+2?
:--`

	srcFile := filepath.Join(tmpDir, "ephemeral.pml")
	err = os.WriteFile(srcFile, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	parser := NewParser(&mockLLM{response: "Test response"}, tmpDir, filepath.Join(tmpDir, "compiled"), tmpDir)
	err = parser.ProcessFile(nil, srcFile)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the result file is also marked as ephemeral
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatal(err)
	}

	foundResult := false
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if f.Name() != "ephemeral.pml" {
			// This should be a result file
			resultPath := filepath.Join(tmpDir, f.Name())
			isEph, err := IsEphemeral(resultPath)
			if err != nil {
				t.Errorf("Failed to check if result is ephemeral: %v", err)
				continue
			}
			if !isEph {
				t.Errorf("Result file %s should be marked as ephemeral", f.Name())
			}
			foundResult = true
		}
	}

	if !foundResult {
		t.Error("No result file was created")
	}
}
