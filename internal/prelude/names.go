// Package prelude lists the names the prelude binds in every module (sections 10 and 11).
package prelude

// Names are the prelude's bindings: the functions, the conversion names, the types, and the
// namespaces. No declaration in any scope may shadow one (section 10).
var Names = []string{
	"print", "panic",
	"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64",
	"float32", "float64", "int", "byte", "string", "bool",
	"never", "Integer", "Number", "Ordered", "Seq", "Entry", "Indexed",
	"Error", "Ok", "Err", "Result",
	"Array", "Map", "Set", "String", "Regex", "Iter", "Math", "Bits", "Record",
	"Mutex", "RwMutex", "WaitGroup", "Once", "Cond", "Atomic", "Context", "Channel",
	"File", "Http", "Net", "Path", "Time", "Os", "Json", "Csv", "Encoding", "Url", "Crypto",
}

var names = func() map[string]bool {
	m := make(map[string]bool, len(Names))
	for _, n := range Names {
		m[n] = true
	}
	return m
}()

// IsName reports whether name is a prelude name.
func IsName(name string) bool { return names[name] }
