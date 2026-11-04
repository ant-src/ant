package types

import "testing"

// Every name BuiltinArity admits has a type Builtin makes from that many arguments.
func TestBuiltinTable(t *testing.T) {
	names := []string{"int", "byte", "string", "null", "void", "never", "Array", "Map", "Set", "Bytes", "Html"}
	for name := range BuiltinKinds {
		names = append(names, name)
	}
	for name := range builtinArity {
		names = append(names, name)
	}
	opaque := func(name string) *Decl { return &Decl{Name: name, Kind: KindOpaqueDecl} }
	for _, name := range names {
		n, ok := BuiltinArity(name)
		if !ok {
			t.Errorf("%s is not a built-in name", name)
			continue
		}
		args := make([]Type, n)
		for i := range args {
			args[i] = Prim{Kind: BuiltinKinds["int"]}
		}
		if Builtin(name, args, opaque) == nil {
			t.Errorf("Builtin(%s) gave no type", name)
		}
	}
	if BuiltinName("Config") || BuiltinName("Reflect") {
		t.Error("a declared name counts as built-in")
	}
}
