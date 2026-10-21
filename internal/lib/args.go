package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Argument checks: every library function panics when an argument has a kind it does not take
// (section 12, the library row).

func arrayArg(g *interp.Goroutine, who string, v value.Value) *value.Array {
	a, ok := v.(*value.Array)
	if !ok {
		g.Panic("%s needs an array, got %s", who, value.Describe(v))
	}
	return a
}

func mapArg(g *interp.Goroutine, who string, v value.Value) *value.Map {
	m, ok := v.(*value.Map)
	if !ok {
		g.Panic("%s needs a map, got %s", who, value.Describe(v))
	}
	return m
}

func stringArg(g *interp.Goroutine, who string, v value.Value) string {
	s, ok := v.(string)
	if !ok {
		g.Panic("%s needs a string, got %s", who, value.Describe(v))
	}
	return s
}

func intArg(g *interp.Goroutine, who string, v value.Value) int64 {
	i, ok := v.(int64)
	if !ok {
		g.Panic("%s needs an int, got %s", who, value.Describe(v))
	}
	return i
}

func floatArg(g *interp.Goroutine, who string, v value.Value) float64 {
	f, ok := v.(float64)
	if !ok {
		g.Panic("%s needs a float64, got %s", who, value.Describe(v))
	}
	return f
}

func funcArg(g *interp.Goroutine, who string, v value.Value) value.Value {
	if _, ok := v.(value.Callable); !ok {
		g.Panic("%s needs a function, got %s", who, value.Describe(v))
	}
	return v
}

// boolResult requires a callback's result to be a bool.
func boolResult(g *interp.Goroutine, who string, v value.Value) bool {
	b, ok := v.(bool)
	if !ok {
		g.Panic("the function given to %s must return a bool, got %s", who, value.Describe(v))
	}
	return b
}

func intResult(g *interp.Goroutine, who string, v value.Value) int64 {
	i, ok := v.(int64)
	if !ok {
		g.Panic("the function given to %s must return an int, got %s", who, value.Describe(v))
	}
	return i
}

// strings converts an array that must hold only strings.
func stringsArg(g *interp.Goroutine, who string, v value.Value) []string {
	vals := arrayArg(g, who, v).Values()
	out := make([]string, len(vals))
	for i, x := range vals {
		s, ok := x.(string)
		if !ok {
			g.Panic("%s needs an array of strings, got an element of kind %s", who, value.Describe(x))
		}
		out[i] = s
	}
	return out
}

// less orders two Ordered values, or panics.
func less(g *interp.Goroutine, who string, a, b value.Value) bool {
	r, err := value.Compare("<", a, b)
	if err != nil {
		g.Panic("%s: %s", who, err)
	}
	return r
}

// sliceBounds checks a half-open range against a length.
func sliceBounds(g *interp.Goroutine, who string, from, to int64, length int) {
	if from < 0 || to > int64(length) || from > to {
		g.Panic("%s: the range %d..%d is not within 0..%d", who, from, to, length)
	}
}
