package writeutil

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

type FieldChainItem struct {
	Name string
}

func (FieldChainItem) _typeIsSetChainItem() {}

type FuncChainItem struct {
	Name string
	Args []any
}

func (FuncChainItem) _typeIsSetChainItem() {}

// IsSet reports whether base and all chained elements resolve to non-nil
// values of their respective types.
// It also checks if slice/array indexes are in bounds and if map keys exists.
func IsSet(base any, chain ...IsSetChainItem) bool {
	val := reflect.ValueOf(base)

	for _, itm := range chain {
		var ok bool
		val, ok = deref(val)
		if !ok {
			return false
		}

		switch itm := itm.(type) {
		case IndexChainItm:
			indexVal := reflect.ValueOf(itm.Index)

			switch val.Kind() {
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

				if indexNum < 0 || indexNum >= val.Len() {
					return false
				}

				val = val.Index(indexNum)
			case reflect.Map:
				if val.Type().Key() != indexVal.Type() {
					return false
				}

				val = val.MapIndex(indexVal)
				if !val.IsValid() {
					return false
				}
			default:
				return false
			}
		case FieldChainItem:
			if val.Kind() != reflect.Struct {
				return false
			}

			val = val.FieldByName(itm.Name)
			if !val.IsValid() {
				return false
			}
		case FuncChainItem:
			rargs := make([]reflect.Value, len(itm.Args))
			for i, arg := range itm.Args {
				rargs[i] = reflect.ValueOf(arg)
			}

			val = val.MethodByName(itm.Name)
			if !val.IsValid() {
				return false
			}

			rret := val.Call(rargs)
			if len(rret) == 2 {
				err, ok := rret[1].Interface().(error)
				if !ok {
					return false
				}

				if err != nil {
					return false
				}
			}

			val = rret[0]
		}
	}

	return true
}

func deref(val reflect.Value) (reflect.Value, bool) {
	for val.Kind() == reflect.Pointer {
		if val.IsNil() {
			return reflect.Value{}, false
		}

		val = val.Elem()
	}

	return val, true
}
