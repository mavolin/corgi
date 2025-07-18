package file

// IsExported reports whether the given identifier, assumed to be valid, would
// be exported.
func IsExported(s string) bool {
	if len(s) == 0 {
		return false
	}
	return 'A' <= s[0] && s[0] <= 'Z'
}

type AnalysisStrategy uint8

const (
	invalidEnd AnalysisStrategy = iota
	// All requires the assertion to be true for all instances.
	All
	// AtLeastOne requires the assertion to be true for at least one instance.
	AtLeastOne
	invalidStart
)

func (s AnalysisStrategy) assertValid() {
	if s >= invalidStart || s <= invalidEnd {
		panic("invalid AnalysisStrategy")
	}
}
