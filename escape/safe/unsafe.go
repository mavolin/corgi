package safe

// UnsafeAttr represents a value for an attribute that is marked as unsafe.
//
// It must escape any raw double quotes.
//
// Use this type with special care.
// Unsafe attributes are marked unsafe for a reason.
// If you need to use this type, place special care in ensuring that the
// value you are using is safe, e.g. through validation.
type UnsafeAttr struct{ val string }

// TrustedUnsafe creates a new Unsafe from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an UnsafeAttr.
func TrustedUnsafe(s string) UnsafeAttr { return UnsafeAttr{val: s} }

func (u UnsafeAttr) Escaped() string { return u.val }
