<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.max

```ant signature
func Math.max[T: Ordered](a: T, b: T): T
```

Math.max returns the larger of a and b.

<!-- end generated -->

## Details

- On floats, a `NaN` argument gives `NaN`, whichever side it is on.
- `Math.max` holds `-0.0` less than `0.0`, so `Math.max(-0.0, 0.0)` and `Math.max(0.0, -0.0)` are
  both `0.0`, though `-0.0 < 0.0` is `false`.
- For the largest element of an array, use [Array.max](../Array/max.md).

## Basic examples

```ant
Math.max(3, 7)              // 7
Math.max(-2.5, -1.0)        // -1.0
Math.max("apple", "banana") // "banana"
```

## Scope

Every integer and float type works, and the result keeps its type:

```ant
Math.max(uint8(200), uint8(17))       // 200
Math.max(int8(-5), int8(-100))        // -5
Math.max(float32(1.5), float32(2.25)) // 2.25
Math.max(Math.maxInt, Math.minInt)    // 9223372036854775807
```

Strings, the empty string included:

```ant
Math.max("", "a")      // "a"
Math.max("ant", "ant") // "ant"
```

Special float values:

```ant
Math.max(-Math.Infinity, 2.0)     // 2.0
Math.max(Math.NaN, 1.0)           // NaN
Math.max(1.0, Math.NaN)           // NaN
Math.max(Math.Infinity, Math.NaN) // NaN
```

Signed zeros, told apart by dividing by them:

```ant
Math.max(0.0, -0.0)       // 0.0
1.0 / Math.max(-0.0, 0.0) // Infinity
1.0 / Math.min(-0.0, 0.0) // -Infinity
```

In a pipe, the value is `a`:

```ant
5 | Math.max(8) // 8
```

For three or more values, nest the calls, or use [Array.max](../Array/max.md) on an array:

```ant
Math.max(3, Math.max(9, 4)) // 9
```

Pass `Math.max` as a function value:

```ant
[3, 9, 2].reduce(Math.minInt, Math.max) // 9
```

## Applications

Keep a value from going below a floor:

```ant
func withdraw(balance: int, amount: int): int {
  Math.max(balance - amount, 0)
}

withdraw(100, 30)  // 70
withdraw(100, 130) // 0
```

The longest name in a list, kept as a running maximum:

```ant
type User = { name: string }

let users = [User{ name: "Ann" }, User{ name: "Bartholomew" }, User{ name: "Cy" }]
var width = 0
for u in users {
  width = Math.max(width, u.name.length())
}
width // 11
```

The overlap of two ranges, empty when they do not meet:

```ant
type Span = { start: int, end: int }

func overlap(a: Span, b: Span): int {
  Math.max(0, Math.min(a.end, b.end) - Math.max(a.start, b.start))
}

overlap(Span{ start: 0, end: 10 }, Span{ start: 5, end: 20 })  // 5
overlap(Span{ start: 0, end: 10 }, Span{ start: 15, end: 20 }) // 0
```

## Properties and relations

`Math.max` is symmetric, gives one of its arguments, and is at least as large as both:

```ant
for [a, b] in [[3, 7], [-1, -4], [5, 5], [Math.minInt, 0]] {
  let m = Math.max(a, b)
  assert(m == Math.max(b, a))
  assert(m == a || m == b)
  assert(m >= a && m >= b)
}
```

It agrees with [Array.max](../Array/max.md) on two elements, and with [Math.min](min.md) and
[Math.clamp](clamp.md):

```ant
let a = 4
let b = 11
assert(Math.max(a, b) == [a, b].max())
assert(Math.max(a, b) + Math.min(a, b) == a + b)
assert(Math.clamp(20, a, b) == Math.min(Math.max(20, a), b))
```

## Possible issues

Both arguments must have the same type. An `int` and a `float64` are a compile error; convert one:

```ant
let x = Math.max(3, 2.5) // error: cannot use float64 where int is expected
```

```ant
Math.max(float64(3), 2.5) // 3.0
```

A value typed `Number`, or a nullable value, must be narrowed first:

```ant
let a: Number = 1
let b: Number = 2
let m = Math.max(a, b)
// error: T would be Number, which is not a single member of its bound Ordered; narrow first
```

```ant
let x: int | null = 3
let m = Math.max(x ?? 0, 1)
m // 3
```

A named primitive type is not `Ordered`. Unwrap it, then tag the result again:

```ant
type Score = int
let m = Math.max(Score(3), Score(5))
// error: T would be Score, which is not a single member of its bound Ordered; narrow first
```

```ant
type Score = int
let best = Score(Math.max(int(Score(3)), int(Score(5))))
best // Score(5)
```

`bool` is not `Ordered`:

```ant
let m = Math.max(true, false)
// error: T would be bool, which is not a single member of its bound Ordered; narrow first
```

A number has no namespace, so a call through `.` is a compile error. Write `Math.max(n, 4)`, or use
a pipe:

```ant
let n = 3
let m = n.max(4) // error: int has no field or function max
```

## See also

[Math.min](min.md) · [Math.clamp](clamp.md) · [Array.max](../Array/max.md) ·
[Array.maxBy](../Array/maxBy.md) · [Math namespace](index.md)
