package woof

/*
This file contains excerpts from the Go standard library package html/template,
licensed under the below license:

Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"unicode/utf8"
)

var stringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

// Stringify converts the passed value to a string.
//
// It accepts values of type string, all ints, uints and floats, and
// [fmt.Stringer].
// val may also be a pointer to any of the above types, or a type approximation,
// i.e. satisfy interface{ ~string }, interface{ ~int }, etc.
//
// If val is nil or dereferences to nil, Stringify returns "".
//
// If Stringify can't print a value, i.e. val is of an unsupported type,
// Stringify returns a pointer to an [UnprintableValueError].
func Stringify(val any) (string, error) {
	return stringify(val, nil)
}

func stringify(val any, escaper func(string) string) (string, error) {
	// for the types in the switch, this is faster than using reflect directly:
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
	return fmt.Sprintf("woof.Stringify: type %T is not printable, consider using a format verb if you really want to print %[1]T", err.Val)
}

// JSify converts the passed value to a JavaScript value.
//
// It is safe to embed into HTML without further escaping.
func JSify(val any) (string, error) {
	switch t := val.(type) {
	case JS:
		return string(t), nil
	case JSStr:
		return `"` + string(t) + `"`, nil
	case json.Marshaler:
		// Do not treat as a Stringer.
	case fmt.Stringer:
		val = t.String()
	}

	jsonVal, err := json.Marshal(val)
	if err != nil {
		return "", err
	}

	if len(jsonVal) == 0 {
		// In, `x=y/{{.}}*z` a json.Marshaler that produces "" should
		// not cause the output `x=y/*z`.
		return " null ", nil
	}
	first, _ := utf8.DecodeRune(jsonVal)
	last, _ := utf8.DecodeLastRune(jsonVal)
	var buf strings.Builder
	// Prevent IdentifierNames and NumericLiterals from running into
	// keywords: in, instanceof, typeof, void
	pad := isJSIdentPart(first) || isJSIdentPart(last)
	if pad {
		buf.WriteByte(' ')
	}
	written := 0
	// Make sure that json.Marshal escapes codepoints U+2028 & U+2029
	// so it falls within the subset of JSON which is valid JS.
	for i := 0; i < len(jsonVal); {
		r, n := utf8.DecodeRune(jsonVal[i:])
		repl := ""
		if r == 0x2028 {
			repl = `\u2028`
		} else if r == 0x2029 {
			repl = `\u2029`
		}
		if repl != "" {
			buf.Write(jsonVal[written:i])
			buf.WriteString(repl)
			written = i + n
		}
		i += n
	}
	if buf.Len() != 0 {
		buf.Write(jsonVal[written:])
		if pad {
			buf.WriteByte(' ')
		}
		return buf.String(), nil
	}
	return string(jsonVal), nil
}
