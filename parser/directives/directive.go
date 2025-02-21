package directives

import (
	"strings"
)

// Directive represents a PML directive like :ask, :do, :input
type Directive interface {
	// Name returns the directive name (e.g., ":ask")
	Name() string

	// CanGenerateBlocks returns true if this directive can generate new blocks
	CanGenerateBlocks() bool
}

// BaseDirective provides common functionality for directives
type BaseDirective struct {
	name string
}

func (d *BaseDirective) Name() string {
	return d.name
}

func (d *BaseDirective) CanGenerateBlocks() bool {
	return false
}

// DirectiveRegistry maintains a map of available directives
type DirectiveRegistry struct {
	directives map[string]Directive
}

// NewDirectiveRegistry creates a new registry with default directives
func NewDirectiveRegistry() *DirectiveRegistry {
	r := &DirectiveRegistry{
		directives: make(map[string]Directive),
	}
	return r
}

// Register adds a new directive to the registry
func (r *DirectiveRegistry) Register(d Directive) {
	r.directives[d.Name()] = d
}

// Get returns a directive by name
func (r *DirectiveRegistry) Get(name string) (Directive, bool) {
	d, ok := r.directives[strings.TrimSpace(name)]
	return d, ok
}

// List returns all registered directive names
func (r *DirectiveRegistry) List() []string {
	var names []string
	for name := range r.directives {
		names = append(names, name)
	}
	return names
}
