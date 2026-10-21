// Package lib implements the standard library (section 11 and stdlib.md) as builtins, and
// builds the prelude table that binds every prelude name.
package lib

import (
	"strings"
	"sync"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// fn declares a builtin. params is its parameter list, comma separated.
func fn(name, params string, impl func(g *interp.Goroutine, args []value.Value) value.Value) *interp.Builtin {
	var ps []string
	if params != "" {
		for _, p := range strings.Split(params, ",") {
			ps = append(ps, strings.TrimSpace(p))
		}
	}
	return &interp.Builtin{Name: name, Params: ps, Fn: impl}
}

// namespace builds a record of functions under a tag named name.
func namespace(name string, funcs []*interp.Builtin) *value.Record {
	fields := make([]string, len(funcs))
	vals := make([]value.Value, len(funcs))
	for i, f := range funcs {
		fields[i] = f.Name[strings.LastIndex(f.Name, ".")+1:]
		vals[i] = f
	}
	tag := value.NewTag(name, value.RecordTag, fields)
	return value.NewRecord(tag, vals)
}

// withFuncs attaches functions to a tag, as functions declared on its type.
func withFuncs(tag *value.Tag, funcs []*interp.Builtin) *value.Tag {
	for _, f := range funcs {
		tag.SetFunc(f.Name[strings.LastIndex(f.Name, ".")+1:], f)
	}
	return tag
}

var (
	once  sync.Once
	table map[string]value.Value
)

// Prelude returns the table that binds every prelude name. It is built once and shared by
// every run; nothing in it changes after construction.
func Prelude() map[string]value.Value {
	once.Do(func() {
		table = map[string]value.Value{}
		table["print"] = fn("print", "value", func(g *interp.Goroutine, args []value.Value) value.Value {
			g.Instance().Print(value.Render(args[0]))
			return value.Void
		})
		table["panic"] = fn("panic", "message", func(g *interp.Goroutine, args []value.Value) value.Value {
			msg, ok := args[0].(string)
			if !ok {
				g.Panic("panic needs a string message, got %s", value.Describe(args[0]))
			}
			g.Panic("%s", msg)
			return nil
		})
		for _, name := range []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "int", "byte", "string", "bool"} {
			kind, _ := value.KindByName(name)
			table[name] = fn(name, "x", func(g *interp.Goroutine, args []value.Value) value.Value {
				v, err := value.Convert(kind, args[0])
				if err != nil {
					g.Panic("%s", err)
				}
				return v
			})
		}
		table["Ok"] = value.OkTag
		table["Err"] = value.ErrTag
		table["Error"] = value.ErrorTag
		table["Entry"] = value.EntryTag
		table["Indexed"] = value.IndexedTag
		for _, name := range []string{"never", "Integer", "Number", "Ordered", "Seq"} {
			table[name] = value.NewTag(name, value.AliasTag, nil)
		}
		table["Result"] = withFuncs(value.NewTag("Result", value.AliasTag, nil), resultFuncs())
		for name, build := range namespaces {
			table[name] = build()
		}
	})
	return table
}

// namespaces lists the namespace builders; each library file registers its own in init.
var namespaces = map[string]func() value.Value{}

func register(name string, build func() value.Value) { namespaces[name] = build }
