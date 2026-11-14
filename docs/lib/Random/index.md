<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random

```ant signature
type Random
```

Random is a seeded generator of pseudo-random numbers. The same seed gives the same values in the
same order on every implementation of Ant, so a simulation, a test, or a game can be run again
exactly. Every function but [`Random.new`](new.md) draws from the generator and changes it, so
goroutines that share one without a lock race; [`Random.fork`](fork.md) gives each its own.
Generators compare by identity. None is fit for keys or tokens, which take
[`Crypto.randomBytes`](../Crypto/randomBytes.md).

<!-- end generated -->

The `Random` namespace holds one function that makes a generator, four that draw a value from it,
one that shuffles an array with it, and one that splits it into two generators whose values never
overlap.

## Details

- Without a seed, [Random.new](new.md) starts from the host's secure source, and the values differ
  from run to run.
- A generator is a reference. Every function but `Random.new` draws from it and moves it on, and a
  second name for it draws from the same sequence.
- Give each goroutine its own generator with [Random.fork](fork.md), before the goroutines start.
- For a one-off value that need not repeat, [Math.random](../Math/random.md) and
  [Math.randomInt](../Math/randomInt.md) need no generator.

## Basic examples

Make a generator with `Random.new` and draw an `int`, a `float64`, and a `bool` from it:

```ant
let r = Random.new(42)
Random.int(r, 1, 7) // 2
Random.float(r)     // 0.9246929453253876
Random.bool(r)      // true
```

Pick an element with `Random.pick`, and reorder an array in place with `Random.shuffle`:

```ant
let r = Random.new(42)
Random.pick(r, ["rock", "paper", "scissors"]) // "scissors"
let deck = ["A", "K", "Q", "J"]
Random.shuffle(r, deck)
deck // ["A", "J", "K", "Q"]
```

Leave out the seed when the values need not repeat. They differ from run to run:

```ant
let r = Random.new()
r.int(1, 7) >= 1 // true
```

## Scope

A draw through `.` reads from the generator:

```ant
let r = Random.new(3)
Array.range(0, 10) | Array.map({ r.int(1, 7) }) // [1, 2, 3, 3, 4, 2, 5, 5, 2, 1]
```

A generator renders as `Random`, never as its state, and compares by identity:

```ant
let r = Random.new(1)
r                  // Random
r == Random.new(1) // false
r == r             // true
```

`Random.pick` returns `null` for an empty array, and draws nothing:

```ant
let r = Random.new(42)
let none: Array[string] = []
r.pick(none)             // null
r.pick(["a", "b"]) ?? "" // "a"
```

A generator fits in a record field, so a game keeps its randomness with the rest of its state:

```ant
type Game = { rng: Random, rolls: Array[int] }

func Game.roll(g: Game): int {
  let n = g.rng.int(1, 7)
  g.rolls.push(n)
  n
}

let game = Game{ rng: Random.new(2024), rolls: [] }
game.roll()
game.roll()
game.roll()
println(game.rolls)
```

```text
[6, 2, 4]
```

## Functions by job

### Making a generator

- [Random.new](new.md): a generator from a seed, or from the host's secure source without one
- [Random.fork](fork.md): a second generator that never meets the first, for another goroutine or
  another part of a program

### Drawing values

- [Random.int](int.md): an `int` from `from`, inclusive, to `to`, exclusive
- [Random.float](float.md): a `float64` in [0, 1)
- [Random.bool](bool.md): `true` or `false`, each half the time
- [Random.pick](pick.md): an element of an array, or `null` when it is empty

### Reordering

- [Random.shuffle](shuffle.md): a random permutation of an array, in place

## Applications

Take a sample without repeats: shuffle a copy and keep the front of it. The original array stays as
it was:

```ant
let names = ["Ann", "Bob", "Cid", "Dee", "Eve", "Fay"]
let r = Random.new(2026)
let pool = [...names]
r.shuffle(pool)
println(pool.slice(0, 3))
println(names)
```

```text
["Bob", "Eve", "Dee"]
["Ann", "Bob", "Cid", "Dee", "Eve", "Fay"]
```

Choose with weights: draw a number below the total weight and walk the table until it falls inside
an entry:

```ant
type Loot = { name: string, weight: int }

func choose(r: Random, table: Array[Loot]): string {
  var roll = r.int(0, table.map({ x.weight }).sum() ?? 0)
  for item in table {
    if roll < item.weight { return item.name }
    roll -= item.weight
  }
  ""
}

let table = [
  Loot{ name: "coin", weight: 70 },
  Loot{ name: "potion", weight: 25 },
  Loot{ name: "sword", weight: 5 },
]
let r = Random.new(9)
let drops = Array.range(0, 1000) | Array.map({ choose(r, table) })
println("coins ${drops.countWhere({ x == "coin" })}, swords ${drops.countWhere({ x == "sword" })}")
```

```text
coins 707, swords 49
```

Generate test inputs from a seed. When a check fails, its message names the seed, which replays the
same inputs:

```ant
func randomWord(r: Random): string {
  let letters = "abcdefghijklmnopqrstuvwxyz".chars()
  Array.range(0, r.int(1, 8))
    | Array.map({ r.pick(letters) ?? "" })
    | String.join("")
}

let seed = 17
let r = Random.new(seed)
let words = Array.range(0, 100) | Array.map({ randomWord(r) })
for word in words {
  assert(word.toUpper().toLower() == word, "seed ${seed}: ${word}")
}
println(words.slice(0, 3))
```

```text
["bzh", "y", "qepqlui"]
```

## Properties and relations

The same seed gives the same values:

```ant
let a = Random.new(99)
let b = Random.new(99)
let xs = Array.range(0, 10) | Array.map({ a.int(0, 1_000_000) })
let ys = Array.range(0, 10) | Array.map({ b.int(0, 1_000_000) })
assert(xs == ys)
```

`Random.pick` is an index drawn with `Random.int`:

```ant
let xs = ["a", "b", "c", "d", "e"]
let a = Random.new(5)
let b = Random.new(5)
for _ in Array.range(0, 20) {
  assert(a.pick(xs) == xs[b.int(0, xs.length())])
}
```

A shuffle keeps every element and only changes the order:

```ant
let xs = Array.range(0, 20)
let shuffled = [...xs]
Random.new(8).shuffle(shuffled)
assert(shuffled.sort() == xs)
```

`Random.float` stays in [0, 1):

```ant
let r = Random.new(1)
assert(Array.range(0, 1000).every({ let f = r.float(); f >= 0.0 && f < 1.0 }))
```

## Possible issues

A generator made anew for each draw starts over each time, so every draw gives the same value. Make
the generator once and pass it along:

```ant
Array.range(0, 3) | Array.map({ Random.new(42).int(0, 100) }) // [22, 22, 22]
let r = Random.new(42)
Array.range(0, 3) | Array.map({ r.int(0, 100) }) // [22, 33, 33]
```

A second name for a generator is not a second generator. Both names draw from one sequence:

```ant
let r = Random.new(42)
let other = r
other.int(0, 100) // 22
r.int(0, 100)     // 33
```

`Random.int` excludes its upper bound, so `r.int(5, 5)` has nothing to draw from and panics:

```ant
let r = Random.new(1)
r.int(5, 5) // panic: Random.int needs from < to, got 5 and 5
```

Goroutines that draw from one shared generator race. Fork one generator per goroutine before they
start:

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

A record that holds a generator has no JSON form, and `Json.stringify` panics on it. Store the seed
instead, and make the generator again from it:

```ant
type Game = { rng: Random, rolls: Array[int] }

let game = Game{ rng: Random.new(2024), rolls: [] }
Json.stringify(game) // panic: Json.stringify: Random has no JSON form
```

The values follow from the seed, so anyone who learns it can predict them. Make keys and tokens with
[Crypto.randomBytes](../Crypto/randomBytes.md) instead.

## See also

[Random.new](new.md) · [Random.int](int.md) · [Random.float](float.md) · [Random.bool](bool.md) ·
[Random.pick](pick.md) · [Random.shuffle](shuffle.md) · [Random.fork](fork.md) ·
[Math.random](../Math/random.md) · [Math.randomInt](../Math/randomInt.md) ·
[Crypto.randomBytes](../Crypto/randomBytes.md)
