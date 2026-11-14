<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.float

```ant signature
func Random.float(r: Random): float64
```

Random.float returns a uniformly random number in \[0, 1).

<!-- end generated -->

## Details

- Each call draws from `r` and so moves it on: the next call returns the next number in its
  sequence.
- The same seed gives the same numbers in the same order, on every run and every implementation of
  Ant.
- The result is a multiple of 2^-53, and `0.0` is a possible result.
- Two goroutines that draw from one generator without a lock race. Give each its own with
  [Random.fork](fork.md).
- It is not fit for keys or tokens. Use [Crypto.randomBytes](../Crypto/randomBytes.md) for those.

## Basic examples

Draw a number from a seeded generator:

```ant
let r = Random.new(42)
Random.float(r) // 0.08386297105988216
```

Each draw moves the generator on, so the next call gives the next number:

```ant
let r = Random.new(42)
Random.float(r) // 0.08386297105988216
Random.float(r) // 0.3789802506626686
Random.float(r) // 0.6800434110281394
```

## Scope

Call it through `.` or in a pipe:

```ant
let r = Random.new(42)
r.float()          // 0.08386297105988216
r | Random.float() // 0.3789802506626686
```

Call it in a lambda to make an array of draws:

```ant
let r = Random.new(7)
Array.range(0, 3).map({ r.float() }) // [0.7005764821796896, 0.2787512294737843, 0.8396274618764198]
```

A generator from [Random.new](new.md) without a seed gives different numbers on every run:

<!-- test: check-only prints a different number on every run -->

```ant
let r = Random.new()
println(r.float())
```

## Applications

Scale the number to a range of floats, here \[-1, 1):

```ant
func between(r: Random, lo: float64, hi: float64): float64 {
  lo + r.float() * (hi - lo)
}

let r = Random.new(3)
between(r, -1.0, 1.0) // 0.3812765902355759
```

Do something with probability `p`:

```ant
func chance(r: Random, p: float64): bool {
  r.float() < p
}

let r = Random.new(1)
Array.range(0, 8).map({ chance(r, 0.25) }) // [false, false, false, false, false, true, true, false]
```

Pick an item by weight. Silver weighs three times as much as gold, so it wins about 750 of 1000
draws:

```ant
type Prize = { name: string, weight: float64 }

func draw(r: Random, prizes: Array[Prize]): string {
  let total = prizes.map({ x.weight }).sum() ?? 0.0
  var roll = r.float() * total
  for p in prizes {
    if roll < p.weight {
      return p.name
    }
    roll -= p.weight
  }
  prizes[prizes.length() - 1].name
}

let prizes = [Prize{ name: "gold", weight: 1.0 }, Prize{ name: "silver", weight: 3.0 }]
let r = Random.new(11)
let wins = Array.range(0, 1000).map({ draw(r, prizes) })
println(wins.countWhere({ x == "silver" }))
```

```text
735
```

Estimate π from the share of random points inside a quarter circle. The seed makes the estimate the
same on every run:

```ant
let r = Random.new(42)
let n = 100_000
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
3.13676
```

## Properties and relations

Every draw is at least `0.0` and below `1.0`:

```ant
let r = Random.new(1)
for _ in Iter.range(0, 1000) {
  let x = r.float()
  assert(x >= 0.0 && x < 1.0)
}
```

Two generators with the same seed give the same numbers:

```ant
let a = Random.new(2024)
let b = Random.new(2024)
for _ in Iter.range(0, 100) {
  assert(a.float() == b.float())
}
```

Each draw is a whole number of 2^-53 steps, so scaling it by 2^53 gives a whole number:

```ant
let r = Random.new(9)
for _ in Iter.range(0, 1000) {
  let scaled = r.float() * 9007199254740992.0
  assert(scaled == Math.floor(scaled))
}
```

A float takes one draw, as [Random.bool](bool.md) does, so after either the two generators agree
again:

```ant
let a = Random.new(5)
let b = Random.new(5)
a.float()
b.bool()
assert(a.float() == b.float())
```

A [fork](fork.md) draws four numbers from its source, which then goes on after them:

```ant
let a = Random.new(42)
let b = Random.new(42)
b.fork()
for _ in Array.range(0, 4) {
  _ = a.float()
}
assert(a.float() == b.float())
```

[Math.random](../Math/random.md) gives numbers in the same range, but cannot be seeded.

## Possible issues

The result is a `float64`, and an `int` literal does not become one. Write `10.0`, not `10`:

```ant
let r = Random.new(1)
let x = r.float() * 10 // error: * needs both operands of one type, got float64 and int
```

For a whole number, do not scale and round the float. [Math.round](../Math/round.md) makes the two
ends half as likely as the numbers between them. Use [Random.int](int.md), which is uniform:

```ant
let r = Random.new(1)
let die = r.int(1, 7)
die >= 1 && die <= 6 // true
```

`Random.float` takes no range. Scale the result instead:

```ant
let r = Random.new(1)
let x = r.float(10.0) // error: Random.float takes 0 arguments after the receiver, got 1
```

## See also

[Random.new](new.md) · [Random.int](int.md) · [Random.bool](bool.md) · [Random.fork](fork.md) ·
[Math.random](../Math/random.md) · [Random namespace](index.md)
