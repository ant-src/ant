package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// The Test namespace of section 11: what a test body needs beyond assert, Panic.catch, and
// defer. Each function acts on the calling goroutine's test and panics outside one.

func init() {
	register("Test", func() value.Value {
		return namespace("Test",
			fn("Test.skip", "reason", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				g.TestSkip(who, stringArg(g, who, a[0]))
				return nil
			}),
			fn("Test.parallel", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				g.TestParallel(who)
				return value.Void
			}),
			fn("Test.tempDir", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return g.TestTempDir(who)
			}),
			fn("Test.context", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return g.TestContext(who)
			}),
			fn("Test.name", "", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return g.TestName(who)
			}),
		)
	})
}
