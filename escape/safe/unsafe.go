package safe

type (
	// Unsafe represents a value for an attribute that is marked as unsafe.
	// When used, it must be the complete value of the attribute; it cannot be
	// embedded inside an otherwise not-empty string using interpolation.
	//
	// It must escape any raw double quotes.
	//
	// Use this type with special care.
	// Unsafe attributes are marked unsafe for a reason.
	// If you don't understand why a value is marked as unsafe, simply wrapping
	// it in this type to still include it in the runtime output is _not_ the
	// solution.
	// If you do need to use this type, place special care in ensuring that the
	// value you are using is safe, e.g. through validation.
	Unsafe struct {
		attr string // name of the attribute
		val  string
	}

	// UnsafeBool represents a value for a boolean attribute that is marked
	// as unsafe.
	//
	// Use this type with special care.
	// Unsafe attributes are marked unsafe for a reason.
	// If you don't understand why a value is marked as unsafe, simply wrapping
	// it in this type to still include it in the runtime output is _not_ the
	// solution.
	// If you need to use this type, place special care in ensuring that the
	// value you are using is safe and doesn't pose any security implications.
	UnsafeBool struct {
		attr string // name of the attribute
		val  bool
	}
)

// TrustedUnsafe creates a new Unsafe value from the given trusted string,
// considered safe only for the passed attribute.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an Unsafe.
func TrustedUnsafe(attr, val string) Unsafe { return Unsafe{attr: attr, val: val} }

// TrustedUnsafeBool creates a new Unsafe boolean from the given value,
// considered safe only for the passed attribute.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for an Unsafe.
func TrustedUnsafeBool(attr string, val bool) UnsafeBool {
	return UnsafeBool{attr: attr, val: val}
}

func (a Unsafe) Get() string  { return a.val }
func (a Unsafe) Attr() string { return a.attr }

func (a UnsafeBool) Escaped() bool { return a.val }
func (a UnsafeBool) Attr() string  { return a.attr }
