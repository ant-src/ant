package lib

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/temporal"
	"github.com/ant-src/ant/value"
)

// The Context sync type of section 9 (Waiting), and the helpers every blocking call uses to
// give up when a context is done.

func contextArg(g *interp.Goroutine, who string, v value.Value) *value.Context {
	c, ok := v.(*value.Context)
	if !ok {
		g.Panic("%s needs a Context, got %s", who, value.Describe(v))
	}
	return c
}

// optionalContext reads a trailing ctx argument: nil when omitted or null.
func optionalContext(g *interp.Goroutine, who string, v value.Value) *value.Context {
	if !given(v) {
		return nil
	}
	return contextArg(g, who, v)
}

// waitContext is the Go context a blocking call selects on: the given context, which descends
// from the run's, or the run's own when none was given.
func waitContext(g *interp.Goroutine, c *value.Context) context.Context {
	if c == nil {
		return g.Context()
	}
	return c.Ctx
}

// gaveUp is the Err a blocking call returns when its context ended it: the reason as the
// message.
func gaveUp(ctx context.Context) value.Value {
	return value.Err(value.ErrorValue(value.ReasonOf(ctx)))
}

// goDuration converts a Time.Duration to a Go duration by the rules of stdlib.md: hours and
// smaller units are exact and a day is 24 hours; years, months or weeks, which have no fixed
// length, and a zero or negative duration are errors; a duration beyond Go's range is taken as
// that range.
func goDuration(d temporal.Duration) (time.Duration, error) {
	if d.Years != 0 || d.Months != 0 || d.Weeks != 0 {
		return 0, errors.New("the timeout has years, months or weeks, which have no fixed length")
	}
	if d.Sign() <= 0 {
		return 0, errors.New("the timeout is zero or negative")
	}
	seconds := float64(d.Days)*86400 + float64(d.Hours)*3600 + float64(d.Minutes)*60 + float64(d.Seconds) +
		float64(d.Milliseconds)/1e3 + float64(d.Microseconds)/1e6 + float64(d.Nanoseconds)/1e9
	ns := seconds * 1e9
	if ns >= math.MaxInt64 {
		return time.Duration(math.MaxInt64), nil
	}
	return time.Duration(ns), nil
}

func init() {
	register("Context", func() value.Value {
		return syncTag("Context",
			fn("Context.new", "timeout?, parent?", func(g *interp.Goroutine, a []value.Value) value.Value {
				const who = "Context.new"
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
			fn("Context.cancel", "ctx", func(g *interp.Goroutine, a []value.Value) value.Value {
				contextArg(g, "Context.cancel", a[0]).Cancel(value.ErrCancelled)
				return value.Void
			}),
			fn("Context.isDone", "ctx", func(g *interp.Goroutine, a []value.Value) value.Value {
				return contextArg(g, "Context.isDone", a[0]).Done()
			}),
			fn("Context.reason", "ctx", func(g *interp.Goroutine, a []value.Value) value.Value {
				reason := contextArg(g, "Context.reason", a[0]).Reason()
				if reason == "" {
					return value.Null
				}
				return reason
			}),
			fn("Context.deadline", "ctx", func(g *interp.Goroutine, a []value.Value) value.Value {
				t, ok := contextArg(g, "Context.deadline", a[0]).Ctx.Deadline()
				if !ok {
					return value.Null
				}
				return temporal.FromEpochNanos(t.UnixNano())
			}),
			fn("Context.wait", "ctx", func(g *interp.Goroutine, a []value.Value) value.Value {
				c := contextArg(g, "Context.wait", a[0])
				g.Instance().Flush()
				<-c.Ctx.Done()
				return value.Void
			}),
		)
	})
}
