package writeutil

import (
	"encoding"
	"errors"
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

func WriteEscaped(w io.Writer, s string) error {
	return Write(w, string(EscapeHTML(s)))
}

// ============================================================================
// Contextual Text
// ======================================================================================

func WriteAnyUnescaped(w io.Writer, a any) error {
	s, err := Stringify(a, false)
	if err != nil {
		return err
	}

	return Write(w, s)
}

func WriteCSS(w io.Writer, a any) error {
	css, ok := a.(CSS)
	if !ok {
		return errors.New("unsafe interpolation of CSS")
	}

	return Write(w, string(css))
}

func WriteHTML(w io.Writer, a any) error {
	html, ok := a.(HTML)
	if ok {
		return Write(w, string(html))
	}

	s, err := Stringify(a, true)
	if err != nil {
		return err
	}

	return Write(w, s)
}

func WriteJS(w io.Writer, a any) error {
	a, ok := a.(JS)
	if !ok {
		return errors.New("unsafe interpolation in JavaScript code")
	}

	s, err := Stringify(a, false)
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
		s, err := Stringify(val, true)
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
		s, err := Stringify(val, false)
		if err != nil {
			return err
		}

		return Write(w, " "+name+`="`+s+`"`)
	}
}

func WriteUnsafeAttr(w io.Writer, name string, val any) error {
	attr, ok := val.(HTMLAttr)
	if !ok {
		return errors.New("unsafe interpolation in attribute")
	}

	return Write(w, " "+name+`="`+string(EscapeCSS(string(attr)))+`"`)
}

func WriteCSSAttr(w io.Writer, name string, val any) error {
	css, ok := val.(CSS)
	if !ok {
		return errors.New("unsafe interpolation in CSS attribute")
	}

	return Write(w, " "+name+`="`+string(EscapeCSS(string(css)))+`"`)
}

func WriteHTMLAttr(w io.Writer, name string, val any) error {
	html, ok := val.(HTML)
	if !ok {
		return errors.New("unsafe interpolation in HTML attribute")
	}

	return Write(w, " "+name+`="`+string(EscapeHTML(string(html)))+`"`)
}

func WriteJSAttr(w io.Writer, name string, val any) error {
	js, ok := val.(JS)
	if !ok {
		return errors.New("unsafe interpolation in JavaScript attribute")
	}

	return Write(w, " "+name+`="`+string(EscapeHTML(string(js)))+`"`)
}

func WriteURLAttr(w io.Writer, name string, val any) error {
	switch val := val.(type) {
	case string:
		if IsSafeURL(val) {
			return Write(w, " "+name+`="`+string(EscapeHTML(val))+`"`)
		}

		return errors.New("unsafe interpolation in URL attribute")
	case URL:
		return Write(w, " "+name+`="`+string(val)+`"`)
	}

	return errors.New("unsafe interpolation in URL attribute")
}

func WriteSrcsetAttr(w io.Writer, name string, val any) error {
	srcset, ok := val.(CSS)
	if !ok {
		return errors.New("unsafe interpolation in srcset attribute")
	}

	return Write(w, " "+name+`="`+string(EscapeHTML(string(srcset)))+`"`)
}

// Stringify converts the passed value to an escaped string.
func Stringify(val any, escaped bool) (string, error) {
	if val == nil {
		return "", nil
	}

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
		if escaped {
			return string(EscapeHTML(rval.String())), nil
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
		if escaped {
			return string(EscapeHTML(string(val))), nil
		}

		return string(val), nil
	case fmt.Stringer:
		if escaped {
			return string(EscapeHTML(val.String())), nil
		}

		return val.String(), nil
	case encoding.TextMarshaler:
		bytes, err := val.MarshalText()
		if err != nil {
			return "", err
		}

		if escaped {
			return string(EscapeHTML(string(bytes))), nil
		}

		return string(bytes), nil
	}

	return "", fmt.Errorf("writeutil.Stringify: unsupported type %T", val)
}
