package safe

// Srcset represents a known safe srcset attribute fragment, safe to be
// embedded between double quotes and used as a srcset attribute or part of a
// srcset attribute.
type Srcset struct{ val string }

// TrustedSrcset creates a new Srcset fragment from the given trusted
// string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Srcset.
func TrustedSrcset(s string) Srcset { return Srcset{val: s} }

func (s Srcset) Get() string { return s.val }
