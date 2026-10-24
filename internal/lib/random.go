package lib

import (
	"math/bits"

	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/value"
)

// Random is the seeded generator of section 11: xoshiro256** over four words, seeded by
// SplitMix64. Every value it gives is fixed by the spec, so no step here may change.
type Random struct{ s [4]uint64 }

// SyncName names the generator for rendering.
func (r *Random) SyncName() string { return "Random" }

func newRandom(seed int64) *Random {
	x := uint64(seed)
	r := &Random{}
	for i := range r.s {
		x += 0x9E3779B97F4A7C15
		z := x
		z = (z ^ z>>30) * 0xBF58476D1CE4E5B9
		z = (z ^ z>>27) * 0x94D049BB133111EB
		r.s[i] = z ^ z>>31
	}
	return r
}

// next is one draw.
func (r *Random) next() uint64 {
	s := &r.s
	out := bits.RotateLeft64(s[1]*5, 7) * 9
	t := s[1] << 17
	s[2] ^= s[0]
	s[3] ^= s[1]
	s[1] ^= s[2]
	s[0] ^= s[3]
	s[2] ^= t
	s[3] = bits.RotateLeft64(s[3], 45)
	return out
}

// below draws uniformly in [0, n), n >= 1, by masked rejection: one draw per try, n = 1 included.
func (r *Random) below(n uint64) uint64 {
	mask := ^uint64(0) >> bits.LeadingZeros64(n-1)
	for {
		if v := r.next() & mask; v < n {
			return v
		}
	}
}

var randomJump = [4]uint64{0x180ec6d33cfd0aba, 0xd5a61266f0c9392c, 0xa9582618e03fc9aa, 0x39abdc4529b1661c}

// fork returns a generator with r's state and moves r 2^128 draws ahead.
func (r *Random) fork() *Random {
	child := &Random{r.s}
	var t [4]uint64
	for _, j := range randomJump {
		for b := 0; b < 64; b++ {
			if j&(1<<b) != 0 {
				for k := range t {
					t[k] ^= r.s[k]
				}
			}
			r.next()
		}
	}
	r.s = t
	return child
}

func randomArg(g *interp.Goroutine, who string, v value.Value) *Random {
	return typedArg[*Random](g, who, "a Random", v)
}

func init() {
	register("Random", func() value.Value {
		return syncTag("Random",
			fn("Random.new", "seed", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return newRandom(intArg(g, who, a[0]))
			}),
			fn("Random.int", "r, from, to", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				r, from, to := randomArg(g, who, a[0]), intArg(g, who, a[1]), intArg(g, who, a[2])
				if to <= from {
					g.Panic("Random.int needs from < to, got %d and %d", from, to)
				}
				// The span may exceed the int range, so it is taken as a uint64.
				return int64(uint64(from) + r.below(uint64(to)-uint64(from)))
			}),
			fn("Random.float", "r", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return float64(randomArg(g, who, a[0]).next()>>11) * 0x1p-53
			}),
			fn("Random.bool", "r", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return randomArg(g, who, a[0]).next()>>63 == 1
			}),
			fn("Random.pick", "r, xs", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				r, xs := randomArg(g, who, a[0]), arrayArg(g, who, a[1])
				vals := xs.Values()
				if len(vals) == 0 {
					return value.Null
				}
				return vals[r.below(uint64(len(vals)))]
			}),
			fn("Random.shuffle", "r, xs", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				r, xs := randomArg(g, who, a[0]), arrayArg(g, who, a[1])
				vals := xs.Values()
				for i := len(vals) - 1; i > 0; i-- {
					j := r.below(uint64(i + 1))
					vals[i], vals[j] = vals[j], vals[i]
				}
				for i, v := range vals {
					xs.Set(i, v)
				}
				return value.Void
			}),
			fn("Random.fork", "r", func(g *interp.Goroutine, who string, a []value.Value) value.Value {
				return randomArg(g, who, a[0]).fork()
			}),
		)
	})
}
