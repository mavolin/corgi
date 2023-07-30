package write

import "github.com/mavolin/corgi/file"

// ============================================================================
// Block
// ======================================================================================

func block(ctx *ctx, b file.Block) {
	// todo
}

// ============================================================================
// BlockExpansion
// ======================================================================================

func blockExpansion(ctx *ctx, bexp file.BlockExpansion) {
	scope(ctx, file.Scope{bexp.Item})
}
