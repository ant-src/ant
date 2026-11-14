<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.int

```ant signature
func Random.int(r: Random, from: int, to: int): int
```

Random.int returns a uniformly random int from from, inclusive, to to, exclusive. It panics if to
\<= from.

<!-- end generated -->

## Details

- `from` is a possible result and `to` is not: `Random.int(r, 1, 7)` gives one of 1 through 6.
- Each call draws from `r` and moves it on, so the next call gives the next value of the sequence. A
  range one wide still draws.
- The same seed gives the same values in the same order, on every run and every implementation of
  Ant.

## Basic examples

Seed a generator and draw a number from 0 up to 100:

```ant
let r = Random.new(42)
Random.int(r, 0, 100) // 22
```

Each call gives the next value of the sequence. Roll a six-sided die ten times:

```ant
let dice = Random.new(7)
Array.range(0, 10) | Array.map({ Random.int(dice, 1, 7) }) // [3, 3, 1, 1, 2, 5, 5, 1, 4, 1]
```

## Scope

Call it through `.` on the generator:

```ant
let r = Random.new(42)
r.int(0, 100) // 22
r.int(0, 100) // 33
```

Negative bounds work like any others:

```ant
let r = Random.new(1)
r.int(-10, 0) // -5
```

The range may span every `int` but `Math.maxInt`:

```ant
let r = Random.new(1)
r.int(Math.minInt, Math.maxInt) // 3743247123249303749
```

A range one wide has only one possible result, and still uses up a draw, so the next call gives the
second value of the sequence:

```ant
let r = Random.new(42)
r.int(0, 1)   // 0
r.int(0, 100) // 33
```

In a pipe, the generator on the left becomes `r`:

```ant
let r = Random.new(7)
r | Random.int(1, 7) // 3
```

## Named arguments

The bounds are named `from` and `to`:

```ant
let r = Random.new(42)
Random.int(r, from: 0, to: 100) // 22
```

## Applications

Slip a card into a deck at a random place. The deck has one more place than it has cards, the end
included, so the range runs to `length() + 1`:

```ant
let deck = ["A", "B", "C", "D"]
let r = Random.new(42)
let at = r.int(0, deck.length() + 1)
deck.insert(at, "Joker")
println(deck)
```

```text
["A", "Joker", "B", "C", "D"]
```

Roll several dice and add them, the same way on every run:

```ant
func roll(r: Random, dice: int): int {
  (Array.range(0, dice) | Array.map({ r.int(1, 7) }) | Array.sum()) ?? 0
}

let r = Random.new(2024)
println(Array.range(0, 5) | Array.map({ roll(r, 2) }))
```

```text
[8, 10, 6, 11, 5]
```

Choose a prize by weight: a ticket from 0 up to the total weight falls inside one prize's share:

```ant
type Prize = { name: string, weight: int }

func draw(r: Random, prizes: Array[Prize]): string {
  let total = (prizes | Array.map({ x.weight }) | Array.sum()) ?? 0
  var ticket = r.int(0, total)
  for p in prizes {
    if ticket < p.weight { return p.name }
    ticket -= p.weight
  }
  panic("unreachable")
}

let prizes = [
  Prize{ name: "gold", weight: 1 },
  Prize{ name: "silver", weight: 3 },
  Prize{ name: "bronze", weight: 6 },
]
let r = Random.new(42)
println(Array.range(0, 4) | Array.map({ draw(r, prizes) }))
```

```text
["bronze", "silver", "silver", "bronze"]
```

Walk a step left or right at random, and replay the same walk from the same seed:

```ant
func walk(seed: int, steps: int): int {
  let r = Random.new(seed)
  var pos = 0
  for _ in Array.range(0, steps) {
    pos += if r.int(0, 2) == 0 { -1 } else { 1 }
  }
  pos
}

assert(walk(5, 100) == walk(5, 100))
println(walk(5, 10))
```

```text
2
```

## Properties and relations

Every result is at least `from` and less than `to`:

```ant
let r = Random.new(1)
for _ in Array.range(0, 1000) {
  let x = r.int(10, 20)
  assert(x >= 10 && x < 20)
}
```

Two generators with the same seed give the same values:

```ant
let a = Random.new(3)
let b = Random.new(3)
for _ in Array.range(0, 100) {
  assert(a.int(0, 1000) == b.int(0, 1000))
}
```

Only the width of the range decides the draw, so shifting the bounds shifts the result:

```ant
let a = Random.new(9)
let b = Random.new(9)
for _ in Array.range(0, 100) {
  assert(a.int(50, 60) == 50 + b.int(0, 10))
}
```

[Random.pick](pick.md) draws an index the same way:

```ant
let xs = ["a", "b", "c", "d"]
let a = Random.new(5)
let b = Random.new(5)
for _ in Array.range(0, 100) {
  assert(a.pick(xs) == xs[b.int(0, xs.length())])
}
```

## Possible issues

`to` is excluded, so a die rolled with `r.int(1, 6)` never shows a 6:

```ant
let r = Random.new(1)
for _ in Array.range(0, 1000) {
  assert(r.int(1, 6) != 6)
}
```

An empty range panics, and so does a range with its bounds swapped:

```ant
let r = Random.new(1)
r.int(0, 0) // panic: Random.int needs from < to, got 0 and 0
r.int(3, 1) // panic: Random.int needs from < to, got 3 and 1
```

Drawing an index for an empty array is the same empty range. Use [Random.pick](pick.md), which
returns `null` for an empty array, or check the length first:

```ant
let r = Random.new(1)
let empty: Array[string] = []
empty[r.int(0, empty.length())] // panic: Random.int needs from < to, got 0 and 0
```

A generator made anew for each call starts the sequence over, so every call gives the same value.
Make the generator once and pass it along:

```ant
func roll(): int {
  Random.new(42).int(1, 7)
}

[roll(), roll(), roll()] // [2, 2, 2]
```

The generator comes first. Calling it with only the bounds is a compile error:

```ant
let x = Random.int(1, 7) // error: Random.int needs an argument for r
```

The bounds are `int`. A `float64` is a compile error:

```ant
let r = Random.new(1)
let x = r.int(1.0, 7) // error: cannot use float64 where int is expected
```

Each draw changes `r`, so two goroutines that call `r.int` on one generator without a lock race.
Give each goroutine its own with [Random.fork](fork.md), forked before the goroutines start:

```ant
let root = Random.new(7)
let gens = Array.range(0, 3) | Array.map({ root.fork() })
let rolls = Parallel.map(gens, func(g: Random, ctx: Context): int {
  g.int(1, 7)
}, limit: 3) ?? []
println(rolls)
```

```text
[3, 5, 2]
```

The values are predictable from the seed, so they are not fit for keys, tokens, or passwords. Use
[Crypto.randomBytes](../Crypto/randomBytes.md) for those.

## See also

[Random.new](new.md) · [Random.float](float.md) · [Random.pick](pick.md) ·
[Random.shuffle](shuffle.md) · [Random.fork](fork.md) · [Math.randomInt](../Math/randomInt.md)
