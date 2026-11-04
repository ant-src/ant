package types

import "github.com/ant-src/ant/value"

// BuiltinKinds maps the primitive type names, int and byte included, to their kinds (section 4,
// Primitives). The checker and the runtime resolver of type expressions share it.
var BuiltinKinds = map[string]value.Kind{
	"int8": value.KindInt8, "int16": value.KindInt16, "int32": value.KindInt32, "int64": value.KindInt64,
	"uint8": value.KindUint8, "uint16": value.KindUint16, "uint32": value.KindUint32, "uint64": value.KindUint64,
	"float32": value.KindFloat32, "float64": value.KindFloat64, "int": value.KindInt64, "byte": value.KindUint8,
	"bool": value.KindBool, "string": value.KindString,
}

// BuiltinName reports whether name is a type name that no declaration binds and that the runtime
// resolves by name, as is does (section 12, Expressions): a primitive, null, void, never, Array,
// Map, Set, Bytes, or Html.
func BuiltinName(name string) bool {
	if _, ok := BuiltinKinds[name]; ok {
		return true
	}
	switch name {
	case "null", "void", "never", "Array", "Map", "Set", "Bytes", "Html":
		return true
	}
	return false
}
