<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.pick

```ant signature
func Random.pick[T](r: Random, xs: Array[T]): T | null
```

Random.pick returns a uniformly random element of xs, or null if xs is empty, in which case it draws
nothing.

<!-- end generated -->

## Details

- Each call on a non-empty array draws from `r` and so changes it.
- `xs` is not changed.
- Not for keys or tokens: use [Crypto.randomBytes](../Crypto/randomBytes.md).

## Basic examples

```ant
Random.pick(Random.new(42), ["red", "green", "blue"]) // "blue"
Random.pick(Random.new(42), [10, 20, 30, 40])         // 30
```

Call it through `.` on a generator. Each call draws again, so later picks differ:

```ant
let r = Random.new(42)
let xs = ["a", "b", "c", "d"]
Array.range(0, 6).map({ r.pick(xs) }) // ["c", "c", "b", "b", "a", "a"]
```

An empty array gives `null`. Annotate it so that `T` is known:

```ant
let empty: Array[string] = []
Random.new(42).pick(empty) // null
```

## Scope

An array with one element always gives that element:

```ant
Random.new(42).pick([7]) // 7
Random.new(99).pick([7]) // 7
```

Records come back as they are stored, and `?.` reads a field of the result:

```ant
type Card = { rank: string, suit: string }
let deck = [
  Card{ rank: "A", suit: "spades" },
  Card{ rank: "7", suit: "hearts" },
  Card{ rank: "Q", suit: "clubs" },
]
Random.new(42).pick(deck)       // Card{ rank: "Q", suit: "clubs" }
Random.new(42).pick(deck)?.suit // "clubs"
```

When the element type includes `null`, a stored `null` can be picked:

```ant
let slots: Array[string | null] = [null, "b", null]
Random.new(3).pick(slots) // null
```

To pick a character a reader sees, pick from the graphemes of a string:

```ant
Random.new(42).pick("🐜🐝🐞".graphemes()) // "🐞"
```

A map or a set is not an array; pick from its keys or values:

```ant
Random.new(42).pick(Map{ "a": 1, "b": 2, "c": 3 }.keys()) // "c"
Random.new(42).pick(Set.of([3, 1, 2]).values())           // 2
```

## Applications

Roll a die ten times:

```ant
let dice = Random.new(1)
let faces = [1, 2, 3, 4, 5, 6]
let rolls = Array.range(0, 10).map({ dice.pick(faces) ?? 0 })
println(rolls)
```

```text
[6, 3, 5, 4, 3, 6, 2, 1, 2, 2]
```

Greet each name with a random greeting:

```ant
let r = Random.new(5)
let greetings = ["Hello", "Hi", "Hey"]
for name in ["Ann", "Bob", "Cy"] {
  let greeting = r.pick(greetings) ?? "Hello"
  println("${greeting}, ${name}")
}
```

```text
Hi, Ann
Hello, Bob
Hey, Cy
```

Weight the choices by repeating an element. Here `"common"` comes up three times as often as
`"rare"`:

```ant
let r = Random.new(9)
let loot = ["common", "common", "common", "rare"]
let counts = Array.range(0, 1000)
  | Array.map({ r.pick(loot) ?? "" })
  | Array.groupBy({ x })
  | Map.map({ x.value.length() })
println(counts)
```

```text
Map{ "common": 734, "rare": 266 }
```

Calling `pick` again can return the same element. To draw several distinct elements, shuffle a copy
and take the front:

```ant
let r = Random.new(42)
let names = ["Ann", "Bob", "Cy", "Di", "Ed"]
let hand = [...names]
r.shuffle(hand)
println(hand.slice(0, 2))
```

```text
["Di", "Cy"]
```

## Properties and relations

`Random.pick` is indexing at `Random.int(r, 0, len(xs))`. Two generators with one seed agree, and
stay in step afterwards:

```ant
let xs = ["a", "b", "c", "d"]
for seed in [1, 2, 3] {
  let a = Random.new(seed)
  let b = Random.new(seed)
  assert(a.pick(xs) == xs[b.int(0, xs.length())])
  assert(a.float() == b.float())
}
```

A pick from an empty array draws nothing, so the generator goes on as if the call had not happened:

```ant
let empty: Array[int] = []
let a = Random.new(5)
let b = Random.new(5)
assert(a.pick(empty) == null)
assert(a.int(0, 100) == b.int(0, 100))
```

The result is always an element of a non-empty array, and the array is not changed:

```ant
let r = Random.new(8)
let xs = [3, 1, 4, 1, 5]
for _ in Array.range(0, 100) {
  let x = r.pick(xs)
  assert(x != null && xs.contains(x))
}
assert(xs == [3, 1, 4, 1, 5])
```

Each element is picked about equally often:

```ant
let r = Random.new(2)
let faces = [1, 2, 3, 4, 5, 6]
let tally = Array.range(0, 6000)
  | Array.map({ r.pick(faces) ?? 0 })
  | Array.groupBy({ x })
  | Map.map({ x.value.length() })
assert(tally.values().every({ x > 900 && x < 1100 }))
```

## Possible issues

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
let r = Random.new(42)
let color: string = r.pick(["red", "blue"]) // error: cannot use string | null where string is expected
```

Give a default with `??`, or bind the result and test it for `null`.

A new generator with the same seed for each call picks the same element every time. Make the
generator once and draw from it:

```ant
let xs = ["a", "b", "c", "d"]
Array.range(0, 3).map({ Random.new(42).pick(xs) }) // ["c", "c", "c"]
let r = Random.new(42)
Array.range(0, 3).map({ r.pick(xs) }) // ["c", "c", "b"]
```

The generator comes first, so piping the array into `Random.pick` puts it in the wrong place:

```ant
let r = Random.new(42)
println([10, 20, 30] | Random.pick(r))
// error: cannot use Array[int] where Random is expected
```

Only arrays are accepted. Split a string into graphemes first:

```ant
let r = Random.new(42)
println(r.pick("abc")) // error: cannot use string where Array[T] is expected
```

When the element type includes `null`, a `null` result does not tell an empty array from a picked
`null`; ask the array with [Array.isEmpty](../Array/isEmpty.md).

## See also

[Random.int](int.md) · [Random.shuffle](shuffle.md) · [Random.new](new.md) · [Random.fork](fork.md)
· [Random](index.md) · [Math.randomInt](../Math/randomInt.md) · [Array.get](../Array/get.md) ·
[Crypto.randomBytes](../Crypto/randomBytes.md)
