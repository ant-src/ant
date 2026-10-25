package lib

import (
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
	"math"
)

// Argument checks: every library function panics when an argument has a kind it does not take
// (section 12, the library row).

// given reports whether an optional argument was passed and is not null.
func given(v value.Value) bool { return v != nil && v != value.Null }

// typedArg takes an argument of the Go type T, which the panic message calls what, as in
// "a Mutex".
func typedArg[T value.Value](g *interp.Goroutine, who, what string, v value.Value) T {
	t, ok := v.(T)
	if !ok {
		g.Panic("%s needs %s, got %s", who, what, value.Describe(v))
	}
	return t
}

func arrayArg(g *interp.Goroutine, who string, v value.Value) *value.Array {
	return typedArg[*value.Array](g, who, "an array", v)
}

func mapArg(g *interp.Goroutine, who string, v value.Value) *value.Map {
	return typedArg[*value.Map](g, who, "a map", v)
}

func setArg(g *interp.Goroutine, who string, v value.Value) *value.Set {
	return typedArg[*value.Set](g, who, "a set", v)
}

func stringArg(g *interp.Goroutine, who string, v value.Value) string {
	return typedArg[string](g, who, "a string", v)
}

// maxCount bounds a count that sizes an allocation: a count above it is an Ant panic, the "a
// count" row of section 12's table, since the runtime could not hold the result and a Go panic
// is not an Ant panic.
const maxCount = math.MaxInt32

// boundCount panics when n, a count of what, is above maxCount.
func boundCount(g *interp.Goroutine, who, what string, n int64) {
	if n > maxCount {
		g.Panic("%s needs a %s of at most %d, got %d", who, what, maxCount, n)
	}
}

// boundSize panics when a string result of size bytes, as a function computes it before building
// the string, is above maxCount: the "a result size" row of section 12's table.
func boundSize(g *interp.Goroutine, who string, size int64) {
	if size > maxCount {
		g.Panic("%s would give a string longer than %d bytes", who, maxCount)
	}
}

// sizeProduct is a * b for non-negative sizes, or maxCount + 1 when that is above maxCount, so
// that a sum of sizes cannot overflow.
func sizeProduct(a, b int64) int64 {
	if a != 0 && b > maxCount/a {
		return maxCount + 1
	}
	return a * b
}

func intArg(g *interp.Goroutine, who string, v value.Value) int64 {
	return typedArg[int64](g, who, "an int", v)
}

func floatArg(g *interp.Goroutine, who string, v value.Value) float64 {
	return typedArg[float64](g, who, "a float64", v)
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

// extremeLess is less with -0.0 below 0.0, as Math.min, Math.max, Array.min and Array.max order
// them; < holds the two zeros equal.
func extremeLess(g *interp.Goroutine, who string, a, b value.Value) bool {
	if za, zb := zeroSign(a), zeroSign(b); za != 0 && zb != 0 {
		return za < zb
	}
	return less(g, who, a, b)
}

// zeroSign is -1 for a float -0.0, 1 for a float 0.0, tagged or not, and 0 for anything else.
func zeroSign(v value.Value) int {
	if t, ok := v.(*value.Tagged); ok {
		v = t.Payload
	}
	var x float64
	switch f := v.(type) {
	case float64:
		x = f
	case float32:
		x = float64(f)
	default:
		return 0
	}
	if x != 0 {
		return 0
	}
	if math.Signbit(x) {
		return -1
	}
	return 1
}

// sortLess is less with NaN after every number, so that a sort has a total order to follow.
func sortLess(g *interp.Goroutine, who string, a, b value.Value) bool {
	an, bn := isNaN(a), isNaN(b)
	if an || bn {
		return !an && bn
	}
	return less(g, who, a, b)
}

// sliceBounds checks a half-open range against a length.
func sliceBounds(g *interp.Goroutine, who string, from, to int64, length int) {
	if from < 0 || to > int64(length) || from > to {
		g.Panic("%s: the range %d..%d is not within 0..%d", who, from, to, length)
	}
}
