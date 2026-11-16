<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.new

```ant signature
func Random.new(seed: int | null = null): Random
```

Random.new returns a generator seeded with seed. Without a seed, or with null, it is seeded from the
host's secure source, as [`Crypto.randomBytes`](../Crypto/randomBytes.md) is, and its values differ
from run to run.

<!-- end generated -->

## Details

- The same seed gives the same values in the same order, on every run and every implementation of
  Ant. Any `int` is a seed, including `0` and negative numbers.
- A generator is not fit for keys or tokens. Use [Crypto.randomBytes](../Crypto/randomBytes.md) for
  those.

## Basic examples

Make a generator from a seed and draw from it:

```ant
let r = Random.new(42)
r.int(0, 100) // 22
r.int(0, 100) // 33
```

Without a seed, the generator starts somewhere new on every run:

<!-- test: check-only prints a different number on every run -->

```ant
let r = Random.new()
println(r.int(1, 7))
```

## Scope

Two generators made from one seed give the same sequence:

```ant
let a = Random.new(7)
let b = Random.new(7)
a.int(0, 1000) // 602
b.int(0, 1000) // 602
```

Zero and negative numbers are seeds like any other:

```ant
Random.new(0).int(0, 100)  // 52
Random.new(-1).int(0, 100) // 8
```

`null` asks for a seed from the host, so an `int | null` passes through. Here the seed comes from
the command line when one is given:

<!-- test: args 42 -->

```ant
let seed = Os.args().get(0)?.toInt()
let r = Random.new(seed)
println(r.int(0, 100))
```

```text
22
```

A generator renders as `Random`, never as its state:

```ant
Random.new(42) // Random
```

## Named arguments

### seed

Name the seed to make its meaning plain at the call:

```ant
let dice = Random.new(seed: 2024)
[dice.int(1, 7), dice.int(1, 7), dice.int(1, 7)] // [6, 2, 4]
```

## Applications

Shuffle a deck so a test can check the order, since the seed fixes the shuffle:

```ant
let deck = ["A", "K", "Q", "J", "10", "9"]
Random.new(1).shuffle(deck)
deck // ["K", "J", "10", "A", "Q", "9"]
```

Make repeatable IDs for tests by passing a seeded generator to the code that makes IDs:

```ant
let alphabet = "0123456789abcdefghjkmnpqrstvwxyz".chars()

func testId(r: Random): string {
  Array.range(0, 6)
    | Array.map({ alphabet[r.int(0, alphabet.length())] })
    | String.join("")
}

let r = Random.new(42)
println([testId(r), testId(r)])
```

```text
["py114r", "j7yxh5"]
```

These IDs are predictable. For keys or tokens, use [Crypto.randomBytes](../Crypto/randomBytes.md).

Estimate pi with a Monte Carlo simulation. The fixed seed makes the estimate the same on every run:

```ant
let r = Random.new(5)
let n = 10000
var inside = 0
for _ in Iter.range(0, n) {
  let x = r.float()
  let y = r.float()
  if x * x + y * y < 1.0 {
    inside += 1
  }
}
println(4.0 * float64(inside) / float64(n))
```

```text
3.1476
```

Draw a fresh seed, print it, and build the generator from it. When a run goes wrong, passing the
printed seed back to `Random.new` gives the same draws again:

<!-- test: check-only prints a different seed on every run -->

```ant
let seed = Math.randomInt(0, 1_000_000)
println("seed ${seed}")
let r = Random.new(seed)
println(r.int(1, 7))
```

## Properties and relations

The same seed gives the same draws:

```ant
let a = Random.new(99)
let b = Random.new(99)
let xs = Array.range(0, 10).map({ a.int(0, 1_000_000) })
let ys = Array.range(0, 10).map({ b.int(0, 1_000_000) })
assert(xs == ys)
```

Generators compare by identity, so two generators with one seed are not equal, and a generator is
equal to itself:

```ant
let a = Random.new(1)
a == Random.new(1) // false
a == a             // true
```

## Possible issues

Making a generator inside a loop with a fixed seed starts it over on every pass, so every pass draws
the same value. Make it once, before the loop:

```ant
let same = Array.range(0, 3).map({ Random.new(42).int(0, 100) })
same // [22, 22, 22]
let r = Random.new(42)
let varied = Array.range(0, 3).map({ r.int(0, 100) })
varied // [22, 33, 33]
```

Goroutines that draw from one generator without a lock race. Give each its own with
[Random.fork](fork.md).

## See also

[Random](index.md) · [Random.fork](fork.md) · [Random.int](int.md) · [Random.float](float.md) ·
[Random.bool](bool.md) · [Random.pick](pick.md) · [Random.shuffle](shuffle.md) ·
[Math.random](../Math/random.md) · [Crypto.randomBytes](../Crypto/randomBytes.md)
