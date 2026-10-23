package lib

import (
	"context"
	"errors"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/temporal"
	"github.com/ant-src/ant/value"
)

// The Context sync type of section 9 (Waiting), and the helpers every blocking call uses to
// give up when a context is done.

func contextArg(g *interp.Goroutine, who string, v value.Value) *value.Context {
	return typedArg[*value.Context](g, who, "a Context", v)
}

// waitContext reads a blocking call's trailing ctx argument as the Go context the call selects
// on: the given context, which descends from the run's, or the run's own when the argument is
// omitted or null.
func waitContext(g *interp.Goroutine, who string, v value.Value) context.Context {
	if !given(v) {
		return g.Context()
	}
	return contextArg(g, who, v).Ctx
}

// gaveUp is the Err a blocking call returns when its context ended it: the reason as the
// message.
func gaveUp(ctx context.Context) value.Value {
	return value.Err(value.ErrorValue(value.ReasonOf(ctx)))
}

// failed is the Err of a blocking call that failed: the context's reason when the context
// ended it, the Go error otherwise.
func failed(ctx context.Context, err error) value.Value {
	if ctx.Err() != nil {
		return gaveUp(ctx)
	}
	return errResult(err)
}

// goDuration converts a Time.Duration to a Go duration by the rules of stdlib.md: hours and
// smaller units are exact and a day is 24 hours; years, months or weeks, which have no fixed
// length, and a zero or negative duration are errors; a duration beyond Go's range is taken as
// that range.
func goDuration(d temporal.Duration) (time.Duration, error) {
	exact, ok := d.Exact()
	if !ok {
		return 0, errors.New("the timeout has years, months or weeks, which have no fixed length")
	}
	if d.Sign() <= 0 {
		return 0, errors.New("the timeout is zero or negative")
	}
	return exact, nil
}

// spanArg reads the Time.Duration a sleep or a ticker takes, by goDuration's rules except that
// years, months or weeks are a panic, since Time is computation, and a zero or negative span is
// returned for the caller to judge.
func spanArg(g *interp.Goroutine, who string, v value.Value) time.Duration {
	d, ok := durationArg(g, who, v).Exact()
	if !ok {
		g.Panic("%s needs a duration without years, months or weeks, which have no fixed length", who)
	}
	return d
}

func init() {
	register("Context", func() value.Value {
		return syncTag("Context",
			fn("Context.new", "timeout?, parent?", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				var timeout time.Duration
				if given(a[0]) {
					d, err := goDuration(durationArg(g, who, a[0]))
					if err != nil {
						g.Panic("%s: %s", who, err)
					}
					timeout = d
				}
				parent := g.Context()
				if given(a[1]) {
					parent = contextArg(g, who, a[1]).Ctx
				}
				return value.NewContext(parent, timeout)
			}),
			fn("Context.cancel", "ctx", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				contextArg(g, who, a[0]).Cancel(value.ErrCancelled)
				return value.Void
			}),
			fn("Context.isDone", "ctx", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return contextArg(g, who, a[0]).Done()
			}),
			fn("Context.reason", "ctx", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				reason := contextArg(g, who, a[0]).Reason()
				if reason == "" {
					return value.Null
				}
				return reason
			}),
			fn("Context.deadline", "ctx", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				t, ok := contextArg(g, who, a[0]).Ctx.Deadline()
				if !ok {
					return value.Null
				}
				deadline, err := temporal.NewInstant(t.Unix(), int64(t.Nanosecond()))
				must(g, who, err)
				return deadline
			}),
			fn("Context.wait", "ctx", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				c := contextArg(g, who, a[0])
				g.Instance().Flush()
				<-c.Ctx.Done()
				g.AfterWait()
				return value.Void
			}),
		)
	})
}
