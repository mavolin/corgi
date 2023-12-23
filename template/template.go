// Package template provides utilities used by the generated code from corgi
// compiler.
//
// Normally, there is no reason for a human to use any of the types in this
// package.
// This is especially true for the [Context] type, which should only be called
// by generated code, as otherwise corgi's security guarantees might be
// undermined.
package template

import (
	"reflect"
	"sort"
)

type OptionalArg[T any] struct {
	Val T
	Set bool
}

// SetArg returns an OptionalArg with the given value set.
func SetArg[T any](t T) OptionalArg[T] {
	return OptionalArg[T]{t, true}
}

func Ternary[T any](cond bool, ifTrue, ifFalse T) T {
	if cond {
		return ifTrue
	}

	return ifFalse
}

func IsZero(t any) bool {
	if t == nil {
		return true
	}
	return reflect.ValueOf(t).IsZero()
}

func CanIndex(val any, i any) bool {
	rval := reflect.ValueOf(val)

	switch rval.Kind() { //nolint:exhaustive
	case reflect.Array, reflect.Slice, reflect.String:
		i, ok := i.(int)
		return ok && rval.Len() > i
	case reflect.Map:
		val := rval.MapIndex(reflect.ValueOf(i))
		return val.IsValid()
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
