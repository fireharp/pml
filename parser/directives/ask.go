package directives

// AskDirective implements the :ask directive
type AskDirective struct {
	BaseDirective
}

// NewAskDirective creates a new ask directive
func NewAskDirective() *AskDirective {
	return &AskDirective{
		BaseDirective: BaseDirective{name: ":ask"},
	}
}

// CanGenerateBlocks implements Directive
func (d *AskDirective) CanGenerateBlocks() bool {
	return false
}
