package lib

import (
	"os"
	"strings"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

func init() {
	register("Os", func() value.Value {
		return namespace("Os", append(processEntries(),
			fn("Os.args", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				args := g.Instance().Args()
				out := make([]value.Value, len(args))
				for i, s := range args {
					out[i] = value.Text(s)
				}
				return value.NewArray(out)
			}),
			fn("Os.env", "name", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				v, ok := g.Instance().LookupEnv(stringArg(g, who, a[0]))
				if !ok {
					return value.Null
				}
				return value.Text(v)
			}),
			fn("Os.environ", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				m := value.NewMap()
				for _, kv := range g.Instance().Environ() {
					k, v, _ := strings.Cut(kv, "=")
					m.Set(value.Text(k), value.Text(v))
				}
				return m
			}),
			fn("Os.cwd", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				dir, err := os.Getwd()
				if err != nil {
					return errResult(err)
				}
				return value.Ok(value.Text(dir))
			}),
			fn("Os.exit", "status", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				g.Instance().Exit(int(intArg(g, who, a[0])))
				return nil
			}),
			// The standard streams are one handle each per run, so that a program reading its
			// input from two places shares one buffer.
			fn("Os.interrupt", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return g.Instance().Interrupt()
			}),
			fn("Os.stdin", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep(who, func() any {
					p := &pump{src: in.Stdin(), done: in.Context().Done()}
					h := newHandle(nil, p, nil)
					h.std, h.in, h.pump, h.childIn = true, in, p, in.ChildStdin()
					return h
				}).(*Handle)
			}),
			fn("Os.stdout", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep(who, func() any {
					h := newHandle(nil, nil, in.Stdout())
					h.std, h.childOut = true, in.ChildStdout()
					return h
				}).(*Handle)
			}),
			fn("Os.stderr", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				in := g.Instance()
				return in.Keep(who, func() any {
					h := newHandle(nil, nil, in.Stderr())
					h.std, h.childOut = true, in.ChildStderr()
					return h
				}).(*Handle)
			}),
		)...)
	})
}
