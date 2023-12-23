package safe

// SrcsetAttr represents a known safe srcset attribute fragment, safe to be
// embedded between double quotes and used as a srcset attribute or part of a
// srcset attribute.
type SrcsetAttr struct{ val string }

// TrustedSrcsetAttr creates a new SrcsetAttr fragment from the given trusted
// string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Srcset.
func TrustedSrcsetAttr(s string) SrcsetAttr { return SrcsetAttr{val: s} }

func (s SrcsetAttr) Escaped() string { return s.val }
