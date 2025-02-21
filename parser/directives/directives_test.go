package directives

import (
	"context"
	"testing"
)

// mockDirective implements Directive for testing
type mockDirective struct {
	BaseDirective
	processFunc func(ctx context.Context, content []string) (string, error)
}

func (d *mockDirective) Process(ctx context.Context, content []string) (string, error) {
	if d.processFunc != nil {
		return d.processFunc(ctx, content)
	}
	return "", nil
}

func (d *mockDirective) CompileToPython(index int, content []string, resultsDir string) string {
	return "# Mock Python code"
}

func TestDirectiveRegistry(t *testing.T) {
	registry := NewDirectiveRegistry()

	// Test registering and retrieving directives
	mockDir := &mockDirective{BaseDirective: BaseDirective{name: ":mock"}}
	registry.Register(mockDir)

	// Test getting registered directive
	if dir, ok := registry.Get(":mock"); !ok {
		t.Error("Failed to get registered directive")
	} else if dir.Name() != ":mock" {
		t.Errorf("Wrong directive name, got %s, want :mock", dir.Name())
	}

	// Test getting non-existent directive
	if _, ok := registry.Get(":nonexistent"); ok {
		t.Error("Got non-existent directive")
	}

	// Test listing directives
	directives := registry.List()
	if len(directives) != 1 {
		t.Errorf("Wrong number of directives, got %d, want 1", len(directives))
	}
	if directives[0] != ":mock" {
		t.Errorf("Wrong directive in list, got %s, want :mock", directives[0])
	}
}

func TestBaseDirective(t *testing.T) {
	base := BaseDirective{name: ":test"}

	if base.Name() != ":test" {
		t.Errorf("Wrong name, got %s, want :test", base.Name())
	}

	if base.CanGenerateBlocks() {
		t.Error("Base directive should not generate blocks by default")
	}
}
