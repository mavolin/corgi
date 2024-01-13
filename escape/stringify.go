package escape

import (
	"fmt"
	"reflect"
	"strconv"
)

var stringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

// Stringify converts the passed value to a string.
//
// It accepts values of type string, all ints, uints and floats, and
// [fmt.Stringer].
// val may also be a pointer to any of the above types, or a type approximation,
// i.e. satisfy interface{ ~string }, interface{ ~int }, etc.
//
// Booleans are printed as the empty string.
//
// If val is nil or dereferences to nil, Stringify returns "".
//
// If Stringify can't print a value, i.e. val is of an unsupported type,
// Stringify returns a pointer to an [UnprintableValueError].
//
// Stringify does not escape the passed-in value in any way.
func Stringify(val any) (string, error) {
	return stringify(val, nil)
}

func stringify(val any, escaper func(string) string) (string, error) {
	// although negligible small, the switch is a bit faster than using reflect
	// string: 2ns vs 13ns
	// uint64: 12ns vs 15ns
	// *string: 7ns vs 24ns
	//
	// for non-switch types, the perf loss is marginal, e.g. 16ns vs 13ns for a
	// type S string
	switch val := val.(type) {
	// most common elem types
	case string:
		if escaper != nil {
			return escaper(val), nil
		}
		return val, nil
	case int:
		return strconv.Itoa(val), nil
	case float32:
		return strconv.FormatFloat(float64(val), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64), nil
	case fmt.Stringer:
		if val == nil {
			return "", nil
		}

		if escaper != nil {
			return escaper(val.String()), nil
		}
		return val.String(), nil

	// remaining elem types
	case bool:
		return strconv.FormatBool(val), nil
	case int8:
		return strconv.FormatInt(int64(val), 10), nil
	case int16:
		return strconv.FormatInt(int64(val), 10), nil
	case int32:
		return strconv.FormatInt(int64(val), 10), nil
	case int64:
		return strconv.FormatInt(val, 10), nil
	case uint16:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(val), 10), nil
	case uint64:
		return strconv.FormatUint(val, 10), nil
	case uint:
		return strconv.FormatUint(uint64(val), 10), nil

	// most common elem types as ptrs (excl. fmt.Stringer)
	case *string:
		if val == nil {
			return "", nil
		}

		if escaper != nil {
			return escaper(*val), nil
		}
		return *val, nil
	case *int:
		if val == nil {
			return "", nil
		}
		return strconv.Itoa(*val), nil
	case *float32:
		if val == nil {
			return "", nil
		}
		return strconv.FormatFloat(float64(*val), 'f', -1, 32), nil
	case *float64:
		if val == nil {
			return "", nil
		}
		return strconv.FormatFloat(*val, 'f', -1, 64), nil
	}

	rt := reflect.TypeOf(val)
	if rt == nil {
		return "", nil
	}
	rv := reflect.ValueOf(val)
	for {
		switch rt.Kind() { //nolint:exhaustive
		case reflect.Pointer:
			if rv.IsNil() {
				return "", nil
			}

			rv = rv.Elem()
			rt = rt.Elem()
		case reflect.String:
			if escaper != nil {
				return escaper(rv.String()), nil
			}
			return rv.String(), nil
		case reflect.Bool:
			return strconv.FormatBool(rv.Bool()), nil
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(rv.Int(), 10), nil
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(rv.Uint(), 10), nil
		case reflect.Float32, reflect.Float64:
			return strconv.FormatFloat(rv.Float(), 'f', -1, 64), nil
		case reflect.Interface:
			if rv.IsNil() {
				return "", nil
			}

			if rt.Implements(stringerType) {
				if escaper != nil {
					return escaper(rv.Interface().(fmt.Stringer).String()), nil
				}
				return rv.Interface().(fmt.Stringer).String(), nil
			}

			fallthrough
		default:
			return "", &UnprintableValueError{Val: val}
		}
	}
}

type UnprintableValueError struct {
	Val any
}

func (err *UnprintableValueError) Error() string {
	return fmt.Sprintf("template: %T is not printable or, if this is a safe.Fragment, not trusted in the current context", err.Val)
}
