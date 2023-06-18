package woof

import (
	"encoding"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// WriteBytes writes the passed bs to w.
// It returns any errors that occur.
func WriteBytes(w io.Writer, bs []byte) error {
	_, err := w.Write(bs)
	return err
}

// Write writes the passed string s to w, utilizing w.WriteString, if w
// implements it.
func Write(w io.Writer, s string) error {
	_, err := io.WriteString(w, s)
	return err
}

// ============================================================================
// Contextual Text
// ======================================================================================

// WriteAnyUnescaped calls [Stringify] on a and writes the result to w.
// It uses no escaper for its call to [Stringify].
func WriteAnyUnescaped(w io.Writer, a any) error {
	s, err := Stringify(a, nil)
	if err != nil {
		return err
	}

	return Write(w, s)
}

// WriteCSS writes CSS to w.
//
// If a is of type [CSS], it writes a directly to w.
// Otherwise, it calls [Stringify] on a, escaping the value with [EscapeCSS],
// and then writes it to w.
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

// WriteHTML writes HTML to w.
//
// If a is of type [HTML], it writes a directly to w.
// Otherwise, it calls [Stringify] on a, escaping the value with [EscapeHTML],
// and then writes it to w.
func WriteHTML(w io.Writer, a any) error {
	html, ok := a.(HTML)
	if ok {
		return Write(w, string(html))
	}

	s, err := Stringify(a, func(s string) string {
		return string(EscapeHTML(s))
	})
	if err != nil {
		return err
	}

	return Write(w, s)
}

// WriteJS calls [JSify] on a and writes the result to w.
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

// WriteAttr writes the passed name-val-pair to w, preceded by a single space.
//
// For that it checks if val is of type bool.
// If so, it either writes ` #{name}="#{name}"` if mirror is true, or just
// ` #{name}, otherwise.
//
// If val is of type [HTMLAttr], it writes the attributes, not escaping val.
// val will be enclosed in double quotes.
//
// In any other case, WriteAttr first calls [Stringify] on val using
// [EscapeHTML] to escape stringified version.
// It then writes the attribute, enclosing val in double quotes.
func WriteAttr(w io.Writer, name string, val any, mirror bool) error {
	switch val := val.(type) {
	case bool:
		if !val {
			return nil
		}

		if mirror {
			return Write(w, ` `+name+`="`+name+`"`)
		}

		return Write(w, " "+name)
	case HTMLAttr:
		return Write(w, ` `+name+`="`+string(val)+`"`)
	default:
		s, err := Stringify(val, func(s string) string {
			return string(EscapeHTML(s))
		})
		if err != nil {
			return err
		}

		return Write(w, ` `+name+`="`+s+`"`)
	}
}

// WriteAttrUnescaped writes the passed name-val-pair to w, preceded by a
// single space.
//
// For that, it checks if val is of type bool, and if so correctly mirrors it
// according to the mirror parameter.
//
// In any other case, WriteAttrUnescaped writes the attribute, using the
// unescaped return of [Stringify] as value.
func WriteAttrUnescaped(w io.Writer, name string, val any, mirror bool) error {
	switch val := val.(type) {
	case bool:
		if val {
			if mirror {
				return Write(w, ` `+name+`="`+name+`"`)
			}

			return Write(w, " "+name)
		}

		return nil
	default:
		s, err := Stringify(val, nil)
		if err != nil {
			return err
		}

		return Write(w, ` `+name+`="`+s+`"`)
	}
}

var (
	runeSliceType      = reflect.TypeOf(([]rune)(nil))
	stringerType       = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()
	textMarshallerType = reflect.TypeOf((*encoding.TextMarshaler)(nil)).Elem()
)

// Stringify converts the passed value to a string.
//
// It accepts values of type string, all ints, all uints, all floats, bool,
// []rune, [fmt.Stringer], and [encoding.TextUnmarshaler].
// val may also be a pointer to any of the above types.
//
// If val is nil or dereferences to nil, it will return "".
//
// If escaper is not nil, it will call it on val, if val is a string,
// []rune, implements [fmt.Stringer], or [encoding.TextMarshaler].
func Stringify(val any, escaper func(string) string) (string, error) {
	if val == nil {
		return "", nil
	}

	rval := reflect.ValueOf(val)
	rtyp := rval.Type()
	for rval.Kind() == reflect.Ptr {
		if rval.IsNil() {
			return "", nil
		}

		switch {
		case rtyp.Implements(stringerType):
			val := rval.Interface().(fmt.Stringer)
			if escaper != nil {
				return escaper(val.String()), nil
			}

			return val.String(), nil
		case rtyp.Implements(textMarshallerType):
			data, err := rval.Interface().(encoding.TextMarshaler).MarshalText()
			if err != nil {
				return "", err
			}

			if escaper != nil {
				return escaper(string(data)), nil
			}

			return string(data), nil
		}

		rval = rval.Elem()
		rtyp = rval.Type()
	}

	switch {
	case rtyp == runeSliceType:
		val := rval.Interface().([]rune)
		if escaper != nil {
			return escaper(string(val)), nil
		}

		return string(val), nil
	case rtyp.Implements(stringerType):
		val := rval.Interface().(fmt.Stringer)
		if escaper != nil {
			return escaper(val.String()), nil
		}

		return val.String(), nil
	case rtyp.Implements(textMarshallerType):
		data, err := rval.Interface().(encoding.TextMarshaler).MarshalText()
		if err != nil {
			return "", err
		}

		if escaper != nil {
			return escaper(string(data)), nil
		}

		return string(data), nil
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

	return "", fmt.Errorf("woof.Stringify: unsupported type %T", rval.Interface())
}
