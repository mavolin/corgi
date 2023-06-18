package woof

import (
	"reflect"
)

type IsSetChainItem interface {
	_typeIsSetChainItem()
}

type IndexChainItm struct {
	Index any
}

func (IndexChainItm) _typeIsSetChainItem() {}

type FieldMethodChainItem struct {
	Name string
}

func (FieldMethodChainItem) _typeIsSetChainItem() {}

type FuncCallChainItem struct {
	Args []any
}

func (FuncCallChainItem) _typeIsSetChainItem() {}

// IsSet reports whether base and all chained elements resolve to non-nil
// values of their respective types.
// It also checks if slice/array indexes are in bounds and if map keys exists.
func IsSet(base any, chain ...IsSetChainItem) bool {
	rval := reflect.ValueOf(base)

Chain:
	for _, itm := range chain {
		var ok bool
		rval, ok = deref(rval)
		if !ok {
			return false
		}

		switch itm := itm.(type) {
		case IndexChainItm:
			indexVal := reflect.ValueOf(itm.Index)

			switch rval.Kind() {
			case reflect.Array, reflect.Slice:
				var indexNum int

				switch indexVal.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					indexNum = int(indexVal.Int())
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					indexNum = int(indexVal.Uint())
				default:
					return false
				}

				if indexNum < 0 || indexNum >= rval.Len() {
					return false
				}

				rval = rval.Index(indexNum)
			case reflect.Map:
				if rval.Type().Key() != indexVal.Type() {
					return false
				}

				rval = rval.MapIndex(indexVal)
				if !rval.IsValid() {
					return false
				}
			default:
				return false
			}
		case FieldMethodChainItem:
			methVal := rval.MethodByName(itm.Name)
			if methVal.IsValid() {
				rval = methVal
				continue Chain
			}

			if rval.Kind() != reflect.Struct {
				return false
			}

			rval = rval.FieldByName(itm.Name)
			if !rval.IsValid() {
				return false
			}
		case FuncCallChainItem:
			rargs := make([]reflect.Value, len(itm.Args))
			for i, arg := range itm.Args {
				rargs[i] = reflect.ValueOf(arg)
			}

			rret := rval.Call(rargs)
			if len(rret) == 2 {
				err, ok := rret[1].Interface().(error)
				if !ok {
					return false
				}

				if err != nil {
					return false
				}
			}

			rval = rret[0]
		}
	}

	_, ok := deref(rval)
	return ok
}

func deref(rval reflect.Value) (reflect.Value, bool) {
	for rval.Kind() == reflect.Pointer || rval.Kind() == reflect.Interface {
		if rval.IsNil() {
			return reflect.Value{}, false
		}

		rval = rval.Elem()
	}

	if !rval.IsValid() {
		return reflect.Value{}, false
	}

	switch rval.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice:
		if rval.IsNil() {
			return reflect.Value{}, false
		}
	}

	return rval, true
}
