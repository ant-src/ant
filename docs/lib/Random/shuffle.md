<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.shuffle

```ant signature
func Random.shuffle[T](r: Random, xs: Array[T]): void
```

Random.shuffle reorders xs in place into a uniformly random permutation.

<!-- end generated -->

## Details

- `Random.shuffle` changes `xs` and returns nothing. To keep the original order, shuffle a copy such
  as `[...xs]`.
- The same seed gives the same order on every run and every implementation of Ant.

## Basic examples

Shuffle the numbers 1 to 10 with a seeded generator:

```ant
let xs = Array.range(1, 11)
Random.shuffle(Random.new(42), xs)
println(xs)
```

```text
[6, 8, 10, 4, 3, 1, 5, 9, 2, 7]
```

Call it through `.` on the generator:

```ant
let names = ["Ann", "Bob", "Cy", "Dee"]
let rng = Random.new(7)
rng.shuffle(names)
println(names)
```

```text
["Bob", "Ann", "Dee", "Cy"]
```

## Scope

Any element type works, records included:

```ant
type Card = { rank: string, suit: string }

let cards = [
  Card{ rank: "A", suit: "S" },
  Card{ rank: "K", suit: "H" },
  Card{ rank: "Q", suit: "D" },
]
Random.new(1).shuffle(cards)
println(cards)
```

```text
[Card{ rank: "Q", suit: "D" }, Card{ rank: "A", suit: "S" }, Card{ rank: "K", suit: "H" }]
```

Elements of a union, `null` among them, move like any other:

```ant
let slots: Array[int | null] = [1, null, 2, null]
Random.new(1).shuffle(slots)
println(slots)
```

```text
[null, 1, 2, null]
```

An empty array and an array of one element stay as they are and draw nothing, so the generator goes
on as if the calls were not there:

```ant
let a = Random.new(1)
let b = Random.new(1)
let empty: Array[int] = []
a.shuffle(empty)
a.shuffle([5])
assert(empty.isEmpty())
assert(a.int(0, 1000) == b.int(0, 1000))
```

Without a seed, the order changes from run to run:

<!-- test: check-only the order depends on the host's secure source -->

```ant
let xs = [1, 2, 3, 4, 5]
Random.new().shuffle(xs)
println(xs)
```

## Applications

Deal two hands of five from a shuffled deck:

```ant
let suits = ["S", "H", "D", "C"]
let ranks = ["A", "K", "Q", "J", "10"]
let deck = suits.flatMap(func(suit: string): Array[string] { ranks.map({ "${x}${suit}" }) })
Random.new(2026).shuffle(deck)
for hand in deck.chunk(5).slice(0, 2) {
  println(hand | String.join(" "))
}
```

```text
10H 10C AC 10S JD
KH AS 10D JS JC
```

Draw a sample of `k` elements without repeats, leaving the input in its order:

```ant
func sample[T](rng: Random, xs: Array[T], k: int): Array[T] {
  let copy = [...xs]
  rng.shuffle(copy)
  copy.slice(0, k)
}

let people = ["Ann", "Bob", "Cy", "Dee", "Eve", "Fay"]
println(sample(Random.new(5), people, 3))
println(people)
```

```text
["Ann", "Dee", "Fay"]
["Ann", "Bob", "Cy", "Dee", "Eve", "Fay"]
```

Pair people at random:

```ant
let order = ["Ann", "Bob", "Cy", "Dee", "Eve", "Fay"]
Random.new(11).shuffle(order)
for pair in order.chunk(2) {
  println("${pair[0]} + ${pair[1]}")
}
```

```text
Eve + Dee
Fay + Cy
Ann + Bob
```

## Properties and relations

A shuffle keeps every element, so the sorted result equals the sorted input:

```ant
let xs = [5, 3, 3, 9, 1, 7]
let ys = [...xs]
Random.new(8).shuffle(ys)
assert(ys.length() == xs.length())
assert(ys.sort() == xs.sort())
```

The same seed gives the same permutation:

```ant
let xs = Array.range(0, 20)
let ys = Array.range(0, 20)
Random.new(99).shuffle(xs)
Random.new(99).shuffle(ys)
assert(xs == ys)
```

`Random.shuffle` is Fisher-Yates from the end: it swaps each position, from the last down to the
second, with one drawn by [`Random.int`](int.md) at or before it:

```ant
let xs = Array.range(0, 8)
let ys = Array.range(0, 8)
Random.new(3).shuffle(xs)

let rng = Random.new(3)
var i = ys.length() - 1
while i >= 1 {
  let j = rng.int(0, i + 1)
  let t = ys[i]
  ys[i] = ys[j]
  ys[j] = t
  i -= 1
}
assert(xs == ys)
```

## Possible issues

`Random.shuffle` returns `void`, not the shuffled array, so its result cannot stand for an array:

```ant
let rng = Random.new(42)
let xs = [3, 1, 2]
let ys: Array[int] = rng.shuffle(xs) // error: cannot use void where Array[int] is expected
```

It does not end a pipe either: the generator is the first parameter, so `xs | Random.shuffle(rng)`
passes the array where the generator goes. Shuffle in a statement of its own, such as
`rng.shuffle(xs)`.

The array is changed in place, so every name for it sees the new order:

```ant
let original = [1, 2, 3, 4]
let alias = original
Random.new(42).shuffle(alias)
println(original)
```

```text
[1, 2, 4, 3]
```

Each shuffle draws from the generator and changes it, so goroutines that share one generator without
a lock race. Give each goroutine its own with `rng.fork()`, made before the `go`:

```ant
let rng = Random.new(42)
let decks = [Array.range(0, 6), Array.range(0, 6)]
let wg = WaitGroup.new()
for deck in decks {
  let own = rng.fork()
  wg.add(1)
  go {
    defer { wg.done() }
    own.shuffle(deck)
  }
}
wg.wait()
println(decks)
```

```text
[[1, 0, 2, 4, 3, 5], [5, 4, 2, 3, 0, 1]]
```

## See also

[Random.pick](pick.md) · [Random.int](int.md) · [Random.new](new.md) · [Random.fork](fork.md) ·
[Array.sort](../Array/sort.md) · [Array.reverse](../Array/reverse.md)
