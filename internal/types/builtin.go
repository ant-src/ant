package types

import "github.com/ant-src/ant/value"

// The built-in type names, which no declaration binds and which the runtime resolves by name,
// as is does (section 12, Expressions). The resolver, the checker, and the runtime resolver of
// type expressions share these tables, so a name is a built-in for all three or for none.

// BuiltinKinds maps the primitive type names, int and byte included, to their kinds (section 4,
// Primitives).
var BuiltinKinds = map[string]value.Kind{
	"int8": value.KindInt8, "int16": value.KindInt16, "int32": value.KindInt32, "int64": value.KindInt64,
	"uint8": value.KindUint8, "uint16": value.KindUint16, "uint32": value.KindUint32, "uint64": value.KindUint64,
	"float32": value.KindFloat32, "float64": value.KindFloat64, "int": value.KindInt64, "byte": value.KindUint8,
	"bool": value.KindBool, "string": value.KindString,
}

// builtinArity gives the number of type arguments each built-in name other than a primitive
// takes.
var builtinArity = map[string]int{
	"null": 0, "void": 0, "never": 0, "Array": 1, "Map": 2, "Set": 1, "Bytes": 0, "Html": 0,
}

// BuiltinArity reports whether name is a built-in type name, a primitive, null, void, never,
// Array, Map, Set, Bytes, or Html, and how many type arguments it takes.
func BuiltinArity(name string) (int, bool) {
	if _, ok := BuiltinKinds[name]; ok {
		return 0, true
	}
	n, ok := builtinArity[name]
	return n, ok
}

// BuiltinName reports whether name is a built-in type name.
func BuiltinName(name string) bool {
	_, ok := BuiltinArity(name)
	return ok
}

// Builtin makes the type a built-in name denotes from its resolved type arguments, as many as
// BuiltinArity gives. opaque gives the declaration of Bytes or Html, which each resolver keeps.
func Builtin(name string, args []Type, opaque func(name string) *Decl) Type {
	if k, ok := BuiltinKinds[name]; ok {
		return Prim{Kind: k}
	}
	switch name {
	case "null":
		return Null{}
	case "void":
		return Void{}
	case "never":
		return Never{}
	case "Array":
		return Array{Elem: args[0]}
	case "Map":
		return Map{Key: args[0], Val: args[1]}
	case "Set":
		return Set{Elem: args[0]}
	case "Bytes", "Html":
		return Named{Decl: opaque(name)}
	}
	panic("types.Builtin: " + name + " is not a built-in type name")
}
