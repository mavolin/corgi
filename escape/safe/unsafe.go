package safe

import _ "unsafe" // for go:linkname

type (
	// Unsafe represents a value for an attribute that is marked as unsafe.
	// When used, it must be the complete value of the attribute; it cannot be
	// embedded inside an otherwise not-empty string using interpolation.
	//
	// It must escape any raw double quotes.
	//
	// Use this type with special care.
	// Unsafe attributes are marked unsafe for a reason.
	// If you do need to use this type, place special care in ensuring that the
	// value you are using is safe, e.g. through further validation.
	Unsafe struct {
		attr string // name of the attribute
		val  string
	}

	// UnsafeBool represents a value for a boolean attribute that is marked
	// as unsafe.
	//
	// Use this type with special care.
	// Unsafe attributes are marked unsafe for a reason.
	UnsafeBool struct {
		attr string // name of the attribute
		val  bool
	}
)

// ConstantUnsafe creates a new Unsafe wrapper from the given string constant.
//
// Before using this function, read the documentation of [Unsafe].
func ConstantUnsafe(attr, c constant) Unsafe {
	return trustedUnsafe(string(attr), string(c))
}

// UnsafeTrue creates a new UnsafeBool wrapper with value true.
//
// Before using this function, read the documentation of [UnsafeBool].
func UnsafeTrue(attr constant) UnsafeBool {
	return trustedUnsafeBool(string(attr), true)
}

// UnsafeFalse creates a new UnsafeBool wrapper with value false.
//
// Before using this function, read the documentation of [UnsafeBool].
func UnsafeFalse(attr constant) UnsafeBool {
	return trustedUnsafeBool(string(attr), false)
}

//go:linkname trustedUnsafe
func trustedUnsafe(attr, s string) Unsafe { return Unsafe{val: s, attr: attr} }

//go:linkname trustedUnsafeBool
func trustedUnsafeBool(attr string, val bool) UnsafeBool { return UnsafeBool{val: val, attr: attr} }

func (a Unsafe) Get() string  { return a.val }
func (a Unsafe) Attr() string { return a.attr }

func (a UnsafeBool) Get() bool    { return a.val }
func (a UnsafeBool) Attr() string { return a.attr }
