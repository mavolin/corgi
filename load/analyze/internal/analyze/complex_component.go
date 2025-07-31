package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
)

// ============================================================================
// Analyze First Permanent Top-Level &-Placeholder
// ======================================================================================

// FindFirstPermanentTopLevelAndPlaceholder attempts to find the first
// permanent top-level &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.FirstPermanentTopLevelAndPlaceholder
//
// Depends on Fields: None
func (z *analyzer) FindFirstPermanentTopLevelAndPlaceholder(logger *slog.Logger, c *file.Component) {
	// todo
}

// ============================================================================
// First Permanent &-Placeholder
// ======================================================================================

// FindFirstPermanentAndPlaceholder attempts to find the first permanent
// &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.FirstPermanentAndPlaceholder
//
// Depends on Fields:
//   - Components.FirstPermanentTopLevelAndPlaceholder
func (z *analyzer) FindFirstPermanentAndPlaceholder(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("first_permanent_and_placeholder")

	if c.FirstPermanentTopLevelAndPlaceholder != nil {
		c.FirstPermanentAndPlaceholder = c.FirstPermanentTopLevelAndPlaceholder
		return
	}

	// todo
}
