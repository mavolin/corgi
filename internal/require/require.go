// Package require provides a type to express optionality.
package require

type Required uint8

const (
	// Never indicates that the required item must not be present.
	Never Required = iota + 1
	// Optional indicates that the required item may be present, but needn't be.
	Optional
	// Always indicates that the required item must always be present.
	Always
)
