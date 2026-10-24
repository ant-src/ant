// Package lib implements the standard library (section 11 and stdlib.md) as builtins, and
// builds the prelude table that binds every prelude name.
package lib

import (
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// fn declares a builtin. params is its parameter list, comma separated; a name ending in ? is
// optional, which only trailing parameters can be, and the builtin receives nil for it when
// the call omits it. The prelude declares the same parameters with their defaults, and a test
// keeps the two in step. impl receives the builtin's name as who, for its panic messages.
func fn(name, params string, impl func(g *interp.Goroutine, who string, args []value.Value) value.Value) *interp.Builtin {
	var ps []string
	optional := 0
	if params != "" {
		for _, p := range strings.Split(params, ",") {
			p = strings.TrimSpace(p)
			if strings.HasSuffix(p, "?") {
				p = strings.TrimSuffix(p, "?")
				optional++
			} else if optional > 0 {
				panic(name + ": an optional parameter precedes a required one")
			}
			ps = append(ps, p)
		}
	}
	call := func(g *interp.Goroutine, args []value.Value) value.Value { return impl(g, name, args) }
	return &interp.Builtin{Name: name, Params: ps, Optional: optional, Fn: call}
}

// entry is one field of a namespace: a function or a constant.
type entry struct {
	name string
	v    value.Value
}

// constant declares a namespace constant such as Math.PI.
func constant(name string, v value.Value) entry { return entry{name: name, v: v} }

// namespace builds a record of functions and constants under a tag named name. A function's
// field is the part of its name after the last dot.
func namespace(name string, entries ...any) *value.Record {
	var fields []string
	var vals []value.Value
	for _, e := range entries {
		switch x := e.(type) {
		case *interp.Builtin:
			fields = append(fields, x.Name[strings.LastIndex(x.Name, ".")+1:])
			vals = append(vals, x)
		case entry:
			fields = append(fields, x.name)
			vals = append(vals, x.v)
		}
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
		table["print"] = fn("print", "value", func(g *interp.Goroutine, _ string, args []value.Value) value.Value {
			g.Instance().Print(value.Render(args[0]))
			return value.Void
		})
		table["println"] = fn("println", "value?", func(g *interp.Goroutine, _ string, args []value.Value) value.Value {
			if args[0] == nil { // println() with no argument, not println(null)
				g.Instance().Print("\n")
				return value.Void
			}
			g.Instance().Print(value.Render(args[0]) + "\n")
			return value.Void
		})
		table["panic"] = fn("panic", "message", func(g *interp.Goroutine, _ string, args []value.Value) value.Value {
			msg, ok := args[0].(string)
			if !ok {
				g.Panic("panic needs a string message, got %s", value.Describe(args[0]))
			}
			g.Panic("%s", msg)
			return nil
		})
		table["len"] = fn("len", "x", func(g *interp.Goroutine, _ string, args []value.Value) value.Value {
			v := args[0]
			if t, ok := v.(*value.Tagged); ok {
				v = t.Payload
			}
			switch x := v.(type) {
			case *value.Array:
				return int64(x.Len())
			case *value.Map:
				return int64(x.Len())
			case *value.Set:
				return int64(x.Len())
			case string:
				return int64(utf8.RuneCountInString(x))
			case *value.Chan:
				return int64(x.Len())
			}
			g.Panic("len needs an array, a map, a set, a string, or a channel, got %s", value.Describe(args[0]))
			return nil
		})
		for _, name := range []string{"int8", "int16", "int32", "int64", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "int", "byte", "string", "bool"} {
			kind, _ := value.KindByName(name)
			table[name] = fn(name, "x", func(g *interp.Goroutine, _ string, args []value.Value) value.Value {
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
		for _, name := range []string{"Integer", "Number", "Ordered", "Seq"} {
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
