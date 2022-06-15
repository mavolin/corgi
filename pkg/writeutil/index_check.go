package writeutil

import (
	"reflect"
)

type NilCheckChainItm interface {
	_typeNilCheckChainItm()
}

type IndexChainItm struct {
	Index any
}

func (IndexChainItm) _typeNilCheckChainItm() {}

type FieldChainItm struct {
	Name string
}

func (FieldChainItm) _typeNilCheckChainItm() {}

type FuncChainItm struct {
	Name string
	Args []any
}

func (FuncChainItm) _typeNilCheckChainItm() {}

// IsSet reports whether base and any possibly chained elements resolve to
// non-nil values.
// It also checks if slice/array indexes are in bounds and if map keys exists.
func IsSet(base any, chain ...NilCheckChainItm) bool {
	val := reflect.ValueOf(base)

	for _, itm := range chain {
		var ok bool
		val, ok = deref(reflect.ValueOf(base))
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
		case FieldChainItm:
			if val.Kind() != reflect.Struct {
				return false
			}

			val = val.FieldByName(itm.Name)
			if !val.IsValid() {
				return false
			}
		case FuncChainItm:
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
