// Package woof provides serialization of Go values to be used in HTML
// documents.
//
// As such, it also provides escaping and filter mechanisms closely related to
// the filters and escapers of Go's stdlib package html/template.
//
// # Terminology
//
// A FILTER is a function that is given untrusted data for a specific content
// type.
//
// If the function deems the data safe, it returns it as is.
//
// If the data is deemed unsafe, the filter will replace the data or some of
// its parts with a safe replacement or delete unsafe parts, CHANGING OR FULLY
// REPLACING THE DATA to ensure safety.
//
// An ESCAPER is a function that is given untrusted data for a specific content
// type and replaces unsafe parts with escape sequences, so that the data can
// be used in its content domain without unintended side effects.
package woof

import (
	"reflect"
	"sort"
)

func Ptr[T any](t T) *T {
	return &t
}

func ResolveDefault[T any](val *T, defaultVal T) T {
	if val != nil {
		return *val
	}
	return defaultVal
}

func Ternary[T any](cond bool, ifTrue, ifFalse T) T {
	if cond {
		return ifTrue
	}

	return ifFalse
}

func IsZero[T comparable](t T) bool {
	var zero T
	return t == zero
}

func CanIndex(val any, i any) bool {
	rval := reflect.ValueOf(val)

	switch rval.Kind() { //nolint:exhaustive
	case reflect.Array, reflect.Slice, reflect.String:
		i, ok := i.(int)
		return ok && rval.Len() > i
	case reflect.Map:
		val := rval.MapIndex(reflect.ValueOf(i))
		return val.IsZero()
	default:
		return false
	}
}

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 |
		~string
}

type MapSlice[K Ordered, V any] []MapEntry[K, V]

type MapEntry[K Ordered, V any] struct {
	K K
	V V
}

func (m MapSlice[K, V]) Len() int           { return len(m) }
func (m MapSlice[K, V]) Less(i, j int) bool { return m[i].K < m[j].K }
func (m MapSlice[K, V]) Swap(i, j int)      { m[i], m[j] = m[j], m[i] }

var _ sort.Interface = MapSlice[int, any]{}

func OrderedMap[K Ordered, V any, M ~map[K]V](m M) MapSlice[K, V] {
	if len(m) <= 1 {
		for k, v := range m {
			return MapSlice[K, V]{{k, v}}
		}
		return nil
	}

	vals := make(MapSlice[K, V], 0, len(m))

	for k, v := range m {
		vals = append(vals, MapEntry[K, V]{k, v})
	}

	sort.Sort(vals)
	return vals
}
