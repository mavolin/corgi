package escape

import "github.com/mavolin/corgi/escape/safe"

// Unsafe is used to escape unsafe attributes.
// Unless val is of type [safe.UnsafeAttr], Unsafe will always return the
// [safe.UnsafeReplacement] value.
func Unsafe(val any) (safe.UnsafeAttr, error) {
	if u, ok := val.(safe.UnsafeAttr); ok {
		return u, nil
	}
	return safe.TrustedUnsafe(safe.UnsafeReplacement), nil
}
