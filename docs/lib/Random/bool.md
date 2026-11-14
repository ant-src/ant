<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.bool

```ant signature
func Random.bool(r: Random): bool
```

Random.bool returns true or false, each half the time.

<!-- end generated -->

## Details

- Each call takes one draw from `r` and so changes it; the next call can give a different value.
- A generator made with the same seed gives the same values in the same order, on every run.
- It is not fit for secrets. Use [Crypto.randomBytes](../Crypto/randomBytes.md) for those.

## Basic examples

Make a generator with [Random.new](new.md) and flip a coin with it:

```ant
let r = Random.new(42)
Random.bool(r) // false
Random.bool(r) // false
Random.bool(r) // true
```

Call it through `.` on the generator:

```ant
let r = Random.new(7)
r.bool() // true
r.bool() // false
```

## Scope

A generator without a seed gives values that differ from run to run:

<!-- test: check-only prints a different value on every run -->

```ant
let r = Random.new()
println(r.bool())
```

Call it in a lambda to make an array of flips:

```ant
let r = Random.new(42)
Array.range(0, 8) | Array.map({ r.bool() }) // [false, false, true, true, true, true, true, true]
```

`Random.bool` is a function value of type `func(Random): bool`:

```ant
let flip: func(Random): bool = Random.bool
flip(Random.new(42)) // false
```

## Applications

Name the sides of a coin with an `if` expression:

```ant
let coin = Random.new(2024)
let sides = Array.range(0, 6) | Array.map({ if coin.bool() { "heads" } else { "tails" } })
sides | String.join(" ") // "tails heads tails tails heads tails"
```

Take a random walk, one step left or right at a time:

```ant
let r = Random.new(5)
var position = 0
for _ in Iter.range(0, 10) {
  position += if r.bool() { 1 } else { -1 }
}
println(position)
```

```text
4
```

Give each number a random sign:

```ant
let r = Random.new(9)
Array.range(0, 5) | Array.map({ if r.bool() { x } else { -x } }) // [0, -1, -2, 3, 4]
```

Upper-case each letter on a coin flip:

```ant
let r = Random.new(11)
let mixed = "ant colony".chars()
  | Array.map({ if r.bool() { x.toUpper() } else { x } })
  | String.join("")
println(mixed)
```

```text
ant coLONy
```

## Properties and relations

The same seed gives the same flips:

```ant
let a = Random.new(7)
let b = Random.new(7)
for _ in Iter.range(0, 100) {
  assert(a.bool() == b.bool())
}
```

`Random.bool` takes exactly one draw, as [Random.float](float.md) does, so two generators with one
seed stay in step after either call:

```ant
let a = Random.new(42)
let b = Random.new(42)
a.bool()
b.float()
assert(a.int(0, 1000) == b.int(0, 1000))
```

`true` comes up about half the time:

```ant
let r = Random.new(1)
let heads = Array.range(0, 10_000) | Array.countWhere({ r.bool() })
assert(heads > 4800 && heads < 5200)
```

## Possible issues

A new generator with the same seed for every flip gives the same flip every time. Make the generator
once and draw from it:

```ant
Random.bool(Random.new(42)) // false
Random.bool(Random.new(42)) // false
```

`Random.bool` takes no probability, so a second argument is a compile error:

```ant
let rare = Random.bool(Random.new(1), 0.3) // error: Random.bool takes 1 argument, got 2
```

For odds other than one half, compare [Random.float](float.md) with the probability:

```ant
let r = Random.new(1)
let rare = Array.range(0, 10_000) | Array.countWhere({ r.float() < 0.3 })
assert(rare > 2800 && rare < 3200)
```

A call without a generator is a compile error too. For an unseeded coin, make one with
`Random.new()`:

```ant
let heads = Random.bool() // error: Random.bool needs an argument for r
```

## See also

[Random.new](new.md) · [Random.float](float.md) · [Random.int](int.md) · [Random.pick](pick.md) ·
[Math.random](../Math/random.md) · [Random namespace](index.md)
