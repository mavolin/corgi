package writeutil

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

func WriteBytes(w io.Writer, bs []byte) error {
	_, err := w.Write(bs)
	return err
}

func Write(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}

// ============================================================================
// Contextual Text
// ======================================================================================

func WriteAnyUnescaped(w io.Writer, a any) error {
	s, err := Stringify(a, nil)
	if err != nil {
		return err
	}

	return Write(w, s)
}

func WriteCSS(w io.Writer, a any) error {
	css, ok := a.(CSS)
	if ok {
		return Write(w, string(css))
	}

	s, err := Stringify(a, func(s string) string {
		return string(EscapeCSS(s))
	})
	if err != nil {
		return err
	}

	return Write(w, s)
}

func WriteHTML(w io.Writer, a any) error {
	html, ok := a.(HTML)
	if ok {
		return Write(w, string(html))
	}

	s, err := Stringify(a, func(s string) string {
		return `"` + string(EscapeHTML(s)) + `"`
	})
	if err != nil {
		return err
	}

	return Write(w, s)
}

func WriteJS(w io.Writer, a any) error {
	s, err := JSify(a)
	if err != nil {
		return err
	}

	return Write(w, s)
}

// ============================================================================
// Attributes
// ======================================================================================

func WriteAttr(w io.Writer, name string, val any, mirror bool) error {
	switch val := val.(type) {
	case bool:
		if val {
			if mirror {
				return Write(w, " "+name+`="`+name+`"`)
			}

			return Write(w, " "+name)
		}

		return nil
	case HTMLAttr:
		return Write(w, " "+name+`="`+string(val)+`"`)
	default:
		s, err := Stringify(val, func(s string) string {
			return string(EscapeHTML(s))
		})
		if err != nil {
			return err
		}

		return Write(w, " "+name+`="`+s+`"`)
	}
}

func WriteAttrUnescaped(w io.Writer, name string, val any, mirror bool) error {
	switch val := val.(type) {
	case bool:
		if val {
			if mirror {
				return Write(w, " "+name+`="`+name+`"`)
			}

			return Write(w, " "+name)
		}

		return nil
	default:
		s, err := Stringify(val, nil)
		if err != nil {
			return err
		}

		return Write(w, " "+name+`="`+s+`"`)
	}
}

// Stringify converts the passed value to a string.
//
// If escaper is not nil, it will call it on val, if val is a string, []rune,
// implements fmt.Stringer, or implements encoding.TextMarshaler.
//
// If val is nil, it will return "".
func Stringify(val any, escaper func(string) string) (string, error) {
	if val == nil {
		return "", nil
	}

	rval := reflect.ValueOf(val)
	for rval.Kind() == reflect.Ptr {
		if rval.IsNil() {
			return "", nil
		}

		rval = rval.Elem()
	}

	switch rval.Kind() {
	case reflect.String:
		if escaper != nil {
			return escaper(rval.String()), nil
		}

		return rval.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rval.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rval.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rval.Float(), 'f', -1, 64), nil
	case reflect.Bool:
		return strconv.FormatBool(rval.Bool()), nil
	}

	switch val := rval.Interface().(type) {
	case []rune:
		if escaper != nil {
			return escaper(string(val)), nil
		}

		return string(val), nil
	case fmt.Stringer:
		if escaper != nil {
			return escaper(val.String()), nil
		}

		return val.String(), nil
	case encoding.TextMarshaler:
		bytes, err := val.MarshalText()
		if err != nil {
			return "", err
		}

		if escaper != nil {
			return escaper(string(bytes)), nil
		}

		return string(bytes), nil
	}

	return "", fmt.Errorf("writeutil.Stringify: unsupported type %T", val)
}
