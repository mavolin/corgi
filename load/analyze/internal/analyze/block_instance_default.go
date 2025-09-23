package analyze

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (z *analyzer) AnalyzeBlockInstanceDefault(ctx context.Context, parents []*walk.Context, bi *file.BlockInstance) {
	z.AnalyzeBlockInstance_AcceptsAttributes()
	z.AnalyzeBlockInstance_ForwardsReceivedAttributes()

	z.AnalyzeBlockInstance_ElementsWithAndPlaceholder()
	z.AnalyzeBlockInstance_ElementSpecsWithAndPlaceholder()

	z.AnalyzeBlockInstance_ForwardsAttributes()
	z.AnalyzeBlockInstance_WritesContent()
	z.AnalyzeBlockInstance_WritesElements()
}

// ============================================================================
// Accepts Attributes
// ======================================================================================

// AnalyzeBlockInstance_AcceptsAttributes determines whether the given block
// instance accepts attributes.
func (z *analyzer) AnalyzeBlockInstance_AcceptsAttributes() {
	// todo
}

// ============================================================================
// Forwards Received Attributes
// ======================================================================================

// AnalyzeBlockInstance_ForwardsReceivedAttributes determines whether the given
// block instance forwards received attributes.
func (z *analyzer) AnalyzeBlockInstance_ForwardsReceivedAttributes() {
	// todo
}

// ============================================================================
// Elements With &-Placeholder
// ======================================================================================

// AnalyzeBlockInstance_ElementsWithAndPlaceholder collects all elements in the
// given block instance that have a &-placeholder.
func (z *analyzer) AnalyzeBlockInstance_ElementsWithAndPlaceholder() {
	// todo
}

// ============================================================================
// Element Specs With &-Placeholder
// ======================================================================================

// AnalyzeBlockInstance_ElementSpecsWithAndPlaceholder collects all element
// specs in the given block instance that have a &-placeholder, either directly
// or indirectly through a component call.
func (z *analyzer) AnalyzeBlockInstance_ElementSpecsWithAndPlaceholder() {
	// todo
}

// ============================================================================
// Forwards Attributes
// ======================================================================================

// AnalyzeBlockInstance_ForwardsAttributes determines whether the given block
// instance forwards attributes.
func (z *analyzer) AnalyzeBlockInstance_ForwardsAttributes() {
	// todo
}

// ============================================================================
// Writes Content
// ======================================================================================

// AnalyzeBlockInstance_WritesContent determines whether the given block
// instance writes content.
func (z *analyzer) AnalyzeBlockInstance_WritesContent() {
	// todo
}

// ============================================================================
// Writes Elements
// ======================================================================================

// AnalyzeBlockInstance_WritesElements determines whether the given block
// instance writes elements.
func (z *analyzer) AnalyzeBlockInstance_WritesElements() {
	// todo
}
