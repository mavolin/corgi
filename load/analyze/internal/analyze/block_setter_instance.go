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

type blockSetterInstance_AcceptsAttributes struct{}

func (z *analyzer) BlockSetterInstance_AcceptsAttributes(bsi *file.BlockSetterInstance) {
	defer z.Ran(bsi, blockSetterInstance_AcceptsAttributes{})
	// todo
}

// ============================================================================
// Forwards Received Attributes
// ======================================================================================

type blockSetterInstance_ForwardsReceivedAttributes struct{}

func (z *analyzer) BlockSetterInstance_ForwardsReceivedAttributes(bsi *file.BlockSetterInstance) {
	defer z.Ran(bsi, blockSetterInstance_ForwardsReceivedAttributes{})
	// todo
}

// ============================================================================
// Forwards Attributes
// ======================================================================================

type blockSetterInstance_ForwardsAttributes struct{}

func (z *analyzer) BlockSetterInstance_ForwardsAttributes(bsi *file.BlockSetterInstance) {
	defer z.Ran(bsi, blockSetterInstance_ForwardsAttributes{})
	// todo
}

// ============================================================================
// Writes Content
// ======================================================================================

type blockSetterInstance_WritesContent struct{}

func (z *analyzer) BlockSetterInstance_WritesContent(bsi *file.BlockSetterInstance) {
	defer z.Ran(bsi, blockSetterInstance_WritesContent{})
	// todo
}

// ============================================================================
// Writes Elements
// ======================================================================================

type blockSetterInstance_WritesElements struct{}

func (z *analyzer) BlockSetterInstance_WritesElements(bsi *file.BlockSetterInstance) {
	defer z.Ran(bsi, blockSetterInstance_WritesElements{})
	// todo
}
