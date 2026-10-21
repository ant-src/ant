package value

import (
	"reflect"
	"unsafe"
)

func tagID(t *Tag) unsafe.Pointer { return unsafe.Pointer(t) }

// identity returns a number that identifies a pointer-shaped value.
func identity(v Value) uint64 {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		return uint64(rv.Pointer())
	}
	return 0
}
