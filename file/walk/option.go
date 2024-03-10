package walk

import (
	"reflect"

	"github.com/mavolin/corgi/file/ast"
)

// An Option is a functions that limits the scope of [Walk].
//
// If at least one Option returns [Ignore], the item will be ignored, i.e. the
// walk function with that option will not be called.
//
// If at least one Option returns [NoDive], the item will not be dived into,
// even if the walk function returns true.
//
// Note that an Option is called on every item, even if the walk function
// is typed.
type Option func(parents []Context, wctx Context) error

// ChildOf asserts that the visited item must be a child of the passed sequence
// of types.
// Other types may appear in between or in front/after the types.
func ChildOf(types ...ast.ScopeNode) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(parents []Context, wctx Context) error {
		var typI int
		for _, p := range parents {
			pt := reflect.TypeOf(p.Node)
			if pt != rTypes[typI] {
				continue
			}

			typI++
			if typI == len(rTypes) {
				return nil
			}
		}

		return Ignore
	}
}

// ChildOfAny asserts that the visited item must be a child of an item of
// the passed types.
func ChildOfAny(types ...ast.ScopeNode) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(parents []Context, wctx Context) error {
		for _, p := range parents {
			pt := reflect.TypeOf(p.Node)
			for _, rType := range rTypes {
				if pt == rType {
					return nil
				}
			}
		}

		return Ignore
	}
}

// NotChildOf asserts that the visited item must not be a child of exactly the
// passed sequence of types.
// Other types may appear in between or in front/after the types and the
// assertion will still fail.
func NotChildOf(types ...ast.ScopeNode) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(parents []Context, wctx Context) error {
		var typI int
		for _, p := range parents {
			pt := reflect.TypeOf(p.Node)
			if pt != rTypes[typI] {
				continue
			}

			typI++
			if typI == len(rTypes) {
				return IgnoreNoDive
			}
		}

		if typI == len(rTypes)-1 && reflect.TypeOf(wctx.Node) == rTypes[typI] {
			return NoDive
		}

		return nil
	}
}

// NotChildOfAny asserts that the visited item must not be a child of the
// passed types.
func NotChildOfAny(types ...ast.ScopeNode) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(parents []Context, wctx Context) error {
		for _, p := range parents {
			pt := reflect.TypeOf(p.Node)
			for _, rType := range rTypes {
				if pt == rType {
					return IgnoreNoDive
				}
			}
		}

		return nil
	}
}

// DontDiveAny prevents the function from diving if the current item is of
// the passed types.
func DontDiveAny(types ...ast.ScopeNode) Option {
	rTypes := make([]reflect.Type, len(types))
	for i, t := range types {
		rTypes[i] = reflect.TypeOf(t)
	}

	return func(_ []Context, wctx Context) error {
		t := reflect.TypeOf(wctx.Node)
		for _, rType := range rTypes {
			if rType == t {
				return NoDive
			}
		}

		return nil
	}
}

// TopLevel asserts that the visited item must be a top-level item, as defined
// by [IsTopLevel].
func TopLevel() Option {
	return func(parents []Context, current Context) error {
		topLevel, childTopLevel := childIsTopLevel(parents, current)
		if topLevel && childTopLevel {
			return nil
		} else if !childTopLevel {
			return NoDive
		}

		return Ignore
	}
}
