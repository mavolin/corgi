package analyze

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
)

func (z *analyzer) AnalyzeBlockSetter(ctx context.Context, parents []*context.Context, bsi *file.BlockSetterInstance) {

}

// ============================================================================
// Accepts Attributes
// ======================================================================================

// BlockSetterInstance_AcceptsAttributes checks if the given block setter
// instance accepts attributes.
func (z *analyzer) BlockSetterInstance_AcceptsAttributes(bsi *file.BlockSetterInstance) {
	// todo
}

// ============================================================================
// Forwards Received Attributes
// ======================================================================================

// BlockSetterInstance_ForwardsReceivedAttributes checks if the given block
// setter instance forwards received attributes.
func (z *analyzer) BlockSetterInstance_ForwardsReceivedAttributes(bsi *file.BlockSetterInstance) {
	// todo
}

// ============================================================================
// Forwards Attributes
// ======================================================================================

// BlockSetterInstance_ForwardsAttributes determines if the given block setter
// instance forwards attributes.
func (z *analyzer) BlockSetterInstance_ForwardsAttributes(bsi *file.BlockSetterInstance) {
	// todo
}

// ============================================================================
// Writes Content
// ======================================================================================

// BlockSetterInstance_WritesContent checks if the given block setter instance
// writes content.
func (z *analyzer) BlockSetterInstance_WritesContent(bsi *file.BlockSetterInstance) {
	// todo
}

// ============================================================================
// Writes Elements
// ======================================================================================

// BlockSetterInstance_WritesElements checks if the given block setter instance
// writes elements.
func (z *analyzer) BlockSetterInstance_WritesElements(bsi *file.BlockSetterInstance) {
	// todo
}
