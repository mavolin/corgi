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

func (z *analyzer) BlockSetterInstance_AcceptsAttributes(bsi *file.BlockSetterInstance) {
	defer analyzed.BlockSetterInstance.AcceptsAttributes(z, bsi)
	// todo
}

// ============================================================================
// Forwards Received Attributes
// ======================================================================================

func (z *analyzer) BlockSetterInstance_ForwardsReceivedAttributes(bsi *file.BlockSetterInstance) {
	defer analyzed.BlockSetterInstance.ForwardsReceivedAttributes(z, bsi)
	// todo
}

// ============================================================================
// Forwards Attributes
// ======================================================================================

func (z *analyzer) BlockSetterInstance_ForwardsAttributes(bsi *file.BlockSetterInstance) {
	defer analyzed.BlockSetterInstance.ForwardsAttributes(z, bsi)
	// todo
}

// ============================================================================
// Writes Content
// ======================================================================================

func (z *analyzer) BlockSetterInstance_WritesContent(bsi *file.BlockSetterInstance) {
	defer analyzed.BlockSetterInstance.WritesContent(z, bsi)
	// todo
}

// ============================================================================
// Writes Elements
// ======================================================================================

func (z *analyzer) BlockSetterInstance_WritesElements(bsi *file.BlockSetterInstance) {
	defer analyzed.BlockSetterInstance.WritesElements(z, bsi)
	// todo
}
