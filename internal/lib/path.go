package lib

import (
	"fmt"
	"path/filepath"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The Path namespace (stdlib.md): pure functions over paths in the host's convention, from
// path/filepath.

func init() {
	register("Path", func() value.Value {
		str := func(name string, f func(string) string) *interp.Builtin {
			return fn("Path."+name, "p", func(g *interp.Goroutine, _ string, a []value.Value) value.Value {
				return f(stringArg(g, "Path."+name, a[0]))
			})
		}
		return namespace("Path",
			fn("Path.join", "parts", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var parts []string
				for _, v := range arrayArg(g, who, a[0]).Values() {
					parts = append(parts, stringArg(g, who, v))
				}
				return filepath.Join(parts...)
			}),
			str("base", filepath.Base),
			str("dir", filepath.Dir),
			str("ext", filepath.Ext),
			str("clean", filepath.Clean),
			fn("Path.isAbs", "p", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return filepath.IsAbs(stringArg(g, who, a[0]))
			}),
			fn("Path.abs", "p", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				p, err := filepath.Abs(stringArg(g, who, a[0]))
				if err != nil {
					return errResult(err)
				}
				return value.Ok(p)
			}),
			fn("Path.rel", "base, target", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				base, target := stringArg(g, who, a[0]), stringArg(g, who, a[1])
				p, err := filepath.Rel(base, target)
				if err != nil {
					// Go's message names its own function; this one names the paths.
					return errResult(fmt.Errorf("no relative path leads from %s to %s", base, target))
				}
				return value.Ok(p)
			}),
		)
	})
}
