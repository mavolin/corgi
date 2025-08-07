package runtime

import (
	"fmt"
	"strconv"
)

// Printable is a type approximation for all types that can be printed
// by [Stringify].
type Printable interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 |

		~*string |
		~*int | ~*int8 | ~*int16 | ~*int32 | ~*int64 |
		~*uint | ~*uint8 | ~*uint16 | ~*uint32 | ~*uint64 |
		~*float32 | ~*float64
}

// Stringify converts the passed value to a string.
//
// It accepts values of type string, all ints, uints and floats,
// or a pointer to any of those types.
// Floats use the 'f' format with -1 precision.
//
// If val is nil or dereferences to nil, Stringify returns "".
//
// Stringify does not escape the passed-in value in any way.
func Stringify[P Printable](val P) string {
	return stringify(val, nil)
}

// stringify is like [Stringify], but calls the passed escaper on the value, if
// it is a string.
func stringify[P Printable](val P, escaper func(string) string) string {
	switch val := any(val).(type) {
	case string:
		if escaper != nil {
			return escaper(val)
		}
		return val
	case int:
		return strconv.Itoa(val)
	case int8:
		return strconv.FormatInt(int64(val), 10)
	case int16:
		return strconv.FormatInt(int64(val), 10)
	case int32:
		return strconv.FormatInt(int64(val), 10)
	case int64:
		return strconv.FormatInt(val, 10)
	case uint:
		return strconv.FormatUint(uint64(val), 10)
	case uint8:
		return strconv.FormatUint(uint64(val), 10)
	case uint16:
		return strconv.FormatUint(uint64(val), 10)
	case uint32:
		return strconv.FormatUint(uint64(val), 10)
	case uint64:
		return strconv.FormatUint(val, 10)
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)

	case *string:
		if val == nil {
			return ""
		}
		if escaper != nil {
			return escaper(*val)
		}
		return *val
	case *int:
		if val == nil {
			return ""
		}
		return strconv.Itoa(*val)
	case *int8:
		if val == nil {
			return ""
		}
		return strconv.FormatInt(int64(*val), 10)
	case *int16:
		if val == nil {
			return ""
		}
		return strconv.FormatInt(int64(*val), 10)
	case *int32:
		if val == nil {
			return ""
		}
		return strconv.FormatInt(int64(*val), 10)
	case *int64:
		if val == nil {
			return ""
		}
		return strconv.FormatInt(*val, 10)
	case *uint:
		if val == nil {
			return ""
		}
		return strconv.FormatUint(uint64(*val), 10)
	case *uint8:
		if val == nil {
			return ""
		}
		return strconv.FormatUint(uint64(*val), 10)
	case *uint16:
		if val == nil {
			return ""
		}
		return strconv.FormatUint(uint64(*val), 10)
	case *uint32:
		if val == nil {
			return ""
		}
		return strconv.FormatUint(uint64(*val), 10)
	case *uint64:
		if val == nil {
			return ""
		}
		return strconv.FormatUint(*val, 10)
	case *float32:
		if val == nil {
			return ""
		}
		return strconv.FormatFloat(float64(*val), 'f', -1, 32)
	case *float64:
		if val == nil {
			return ""
		}
		return strconv.FormatFloat(*val, 'f', -1, 64)
	default:
		panic(fmt.Sprintf("runtime: %T is Printable, but not handled by Stringify", val))
	}
}

type UnprintableValueError struct {
	Val any
}

func (err *UnprintableValueError) Error() string {
	return fmt.Sprintf("template: %T is not printable or, if this is a safe.Fragment, not trusted in the current context", err.Val)
}
