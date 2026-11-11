<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.randomInt

```ant signature
func Math.randomInt(from: int, to: int): int
```

Math.randomInt returns a uniformly random int from from, inclusive, to to, exclusive. It draws from
the same unseeded sequence as [`Math.random`](random.md). It panics if to \<= from.

<!-- end generated -->

## Details

- `from` is a possible result and `to` is not: `Math.randomInt(1, 7)` gives one of 1 through 6.
- The values differ from run to run, and nothing can seed them. For a sequence you can repeat, use
  [Random.new](../Random/new.md) with a seed and [Random.int](../Random/int.md).
- Not fit for keys, tokens, or passwords. Use [Crypto.randomBytes](../Crypto/randomBytes.md) for
  those.

## Basic examples

Roll a six-sided die. The result is one of 1 through 6:

```ant
let roll = Math.randomInt(1, 7)
roll >= 1 && roll <= 6 // true
```

Pick a random element of an array by drawing an index from `0` up to its length:

```ant
let colors = ["red", "green", "blue"]
let color = colors[Math.randomInt(0, len(colors))]
colors.contains(color) // true
```

## Scope

A range one wide has only one possible result:

```ant
Math.randomInt(5, 6)   // 5
Math.randomInt(-3, -2) // -3
```

Negative bounds work like any others:

```ant
let t = Math.randomInt(-10, 0)
t >= -10 && t < 0 // true
```

The range may span every `int` but `Math.maxInt`:

```ant
let n = Math.randomInt(Math.minInt, Math.maxInt)
n < Math.maxInt // true
```

The parameters are named `from` and `to`:

```ant
Math.randomInt(from: 0, to: 1) // 0
```

In a pipe, the value on the left becomes `from`:

```ant
6 | Math.randomInt(7) // 6
```

## Applications

Flip a coin:

```ant
let side = if Math.randomInt(0, 2) == 0 { "heads" } else { "tails" }
side == "heads" || side == "tails" // true
```

Roll two dice and add them, which gives 2 through 12, with 7 the most likely:

```ant
func twoDice(): int {
  Math.randomInt(1, 7) + Math.randomInt(1, 7)
}

for _ in Array.range(0, 1000) {
  let sum = twoDice()
  assert(sum >= 2 && sum <= 12)
}
```

Add jitter to a retry delay, so that many clients do not retry at the same moment:

```ant
func retryDelayMs(attempt: int): int {
  let base = 100 * (attempt + 1)
  base + Math.randomInt(-base / 10, base / 10 + 1)
}

let d = retryDelayMs(4)
d >= 450 && d <= 550 // true
```

Count how often each face of a die comes up:

```ant
let counts = Array.filled(6, 0)
for _ in Array.range(0, 6000) {
  let face = Math.randomInt(1, 7)
  counts[face - 1] += 1
}
assert(counts.every({ x > 0 }))
assert(counts.sum() == 6000)
```

## Properties and relations

Every result is at least `from` and less than `to`:

```ant
for _ in Array.range(0, 1000) {
  let x = Math.randomInt(10, 20)
  assert(x >= 10 && x < 20)
}
```

Adding `from` to a draw from zero gives the same range:

```ant
for _ in Array.range(0, 1000) {
  let x = 10 + Math.randomInt(0, 10)
  assert(x >= 10 && x < 20)
}
```

A float from [Math.random](random.md), scaled and cut down to an `int`, gives the same range:

```ant
for _ in Array.range(0, 1000) {
  let x = int(Math.floor(Math.random() * 6.0)) + 1
  assert(x >= 1 && x <= 6)
}
```

## Possible issues

`to` is excluded, so a die rolled with `Math.randomInt(1, 6)` never shows a 6:

```ant
for _ in Array.range(0, 1000) {
  assert(Math.randomInt(1, 6) != 6)
}
```

An empty range panics, and so does a range with its bounds swapped:

```ant
Math.randomInt(0, 0) // panic: Math.randomInt needs from < to, got 0 and 0
Math.randomInt(3, 1) // panic: Math.randomInt needs from < to, got 3 and 1
```

Drawing an index for an empty array is the same empty range, so check the length first:

```ant
let empty: Array[string] = []
empty[Math.randomInt(0, len(empty))] // panic: Math.randomInt needs from < to, got 0 and 0
```

The bounds are `int`. A `float64` or another integer type is a compile error:

```ant
let x = Math.randomInt(1.0, 7) // error: cannot use float64 where int is expected
```

The results cannot be repeated. For a test or a replayable game, seed a [Random](../Random/index.md)
instead, which gives the same values on every run:

```ant
let r = Random.new(42)
r.int(0, 100) // 22
```

## See also

[Math.random](random.md) · [Random.int](../Random/int.md) · [Random.pick](../Random/pick.md) ·
[Random.shuffle](../Random/shuffle.md) · [Crypto.randomBytes](../Crypto/randomBytes.md)
