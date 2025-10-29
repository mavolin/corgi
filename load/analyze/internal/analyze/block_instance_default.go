package analyze

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/walk"
)

func (z *analyzer) AnalyzeBlockInstanceDefault(ctx context.Context, parents []*walk.Context, bi *file.BlockInstance) {
	z.AnalyzeBlockInstanceDefault_AcceptsAttributes()
	z.AnalyzeBlockInstanceDefault_ForwardsReceivedAttributes()

	z.AnalyzeBlockInstanceDefault_ElementsWithAndPlaceholder()
	z.AnalyzeBlockInstanceDefault_ElementSpecsWithAndPlaceholder()

	z.AnalyzeBlockInstanceDefault_ForwardsAttributes()
	z.AnalyzeBlockInstanceDefault_WritesContent()
	z.AnalyzeBlockInstanceDefault_WritesElements()
}

// ============================================================================
// Accepts Attributes
// ======================================================================================

type blockInstanceDefault_AcceptsAttributes struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_AcceptsAttributes() {
	// defer z.Ran(nil, blockInstanceDefault_AcceptsAttributes{})
	// todo
}

// ============================================================================
// Forwards Received Attributes
// ======================================================================================

type blockInstanceDefault_ForwardsReceivedAttributes struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_ForwardsReceivedAttributes() {
	// defer z.Ran(nil, blockInstanceDefault_ForwardsReceivedAttributes{})
	// todo
}

// ============================================================================
// Elements With &-Placeholder
// ======================================================================================

type blockInstanceDefault_ElementsWithAndPlaceholder struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_ElementsWithAndPlaceholder() {
	// defer z.Ran(nil, blockInstanceDefault_ElementsWithAndPlaceholder{})
	// todo
}

// ============================================================================
// Element Specs With &-Placeholder
// ======================================================================================

type blockInstanceDefault_ElementSpecsWithAndPlaceholder struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_ElementSpecsWithAndPlaceholder() {
	// defer z.Ran(nil, blockInstanceDefault_ElementSpecsWithAndPlaceholder{})
	// todo
}

// ============================================================================
// Forwards Attributes
// ======================================================================================

type blockInstanceDefault_ForwardsAttributes struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_ForwardsAttributes() {
	// defer z.Ran(nil, blockInstanceDefault_ForwardsAttributes{})
	// todo
}

// ============================================================================
// Writes Content
// ======================================================================================

type blockInstanceDefault_WritesContent struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_WritesContent() {
	// defer z.Ran(nil, blockInstanceDefault_WritesContent{})
	// todo
}

// ============================================================================
// Writes Elements
// ======================================================================================

type blockInstanceDefault_WritesElements struct{}

func (z *analyzer) AnalyzeBlockInstanceDefault_WritesElements() {
	// defer z.Ran(nil, blockInstanceDefault_WritesElements{})
	// todo
}
