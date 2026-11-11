<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Math.random

```ant signature
func Math.random(): float64
```

Math.random returns a uniformly random number in \[0, 1). The sequence cannot be seeded; a
[`Random`](../Random/index.md) generator is the seeded form.

<!-- end generated -->

## Details

- The values differ from run to run.
- It is not fit for keys or tokens. Use [Crypto.randomBytes](../Crypto/randomBytes.md) for those.

## Basic examples

Draw a number:

<!-- test: check-only prints a different number on every run -->

```ant
println(Math.random())
```

The number is in \[0, 1):

```ant
let x = Math.random()
x >= 0.0 && x < 1.0 // true
```

## Scope

`Math.random` takes no arguments. As a function value its type is `func(): float64`:

```ant
let draw: func(): float64 = Math.random
draw() < 1.0 // true
```

Call it in a lambda to make an array of draws:

```ant
let xs = Array.range(0, 5).map({ Math.random() })
xs.length()                       // 5
xs.every({ x >= 0.0 && x < 1.0 }) // true
```

## Applications

Scale the number to a range of floats, here \[-1, 1):

```ant
func between(lo: float64, hi: float64): float64 {
  lo + Math.random() * (hi - lo)
}

let v = between(-1.0, 1.0)
v >= -1.0 && v < 1.0 // true
```

Do something with probability `p`. A `p` of `0.0` never happens, and `1.0` always does:

```ant
func chance(p: float64): bool {
  Math.random() < p
}

chance(0.0) // false
chance(1.0) // true
```

Spread out retries, so that clients that failed together do not retry together:

```ant
func backoff(attempt: int): float64 {
  let base = 100.0 * Math.pow(2.0, float64(attempt))
  base + Math.random() * base
}

let delay = backoff(3)
delay >= 800.0 && delay < 1600.0 // true
```

Pick an item by weight:

```ant
type Prize = { name: string, weight: float64 }

func draw(prizes: Array[Prize]): string {
  var total = 0.0
  for p in prizes {
    total += p.weight
  }
  var roll = Math.random() * total
  for p in prizes {
    if roll < p.weight {
      return p.name
    }
    roll -= p.weight
  }
  prizes[prizes.length() - 1].name
}

let prizes = [Prize{ name: "gold", weight: 1.0 }, Prize{ name: "silver", weight: 3.0 }]
["gold", "silver"].contains(draw(prizes)) // true
```

Estimate π from the share of random points that fall inside a quarter circle:

<!-- test: check-only the estimate changes from run to run -->

```ant
let n = 100_000
var inside = 0
for _ in Iter.range(0, n) {
  let x = Math.random()
  let y = Math.random()
  if x * x + y * y < 1.0 {
    inside += 1
  }
}
println(4.0 * float64(inside) / float64(n))
```

## Properties and relations

Every draw is at least `0.0` and below `1.0`:

```ant
for _ in Iter.range(0, 1000) {
  let x = Math.random()
  assert(x >= 0.0 && x < 1.0)
}
```

Scaling by `n` and taking [Math.floor](floor.md) gives a whole number below `n`; for an `int`,
[Math.randomInt](randomInt.md) does this directly:

```ant
for _ in Iter.range(0, 1000) {
  let i = int(Math.floor(Math.random() * 6.0))
  assert(i >= 0 && i < 6)
  let j = Math.randomInt(0, 6)
  assert(j >= 0 && j < 6)
}
```

[Random.float](../Random/float.md) gives numbers in the same range, and the same ones for the same
seed:

```ant
let a = Random.new(42)
let b = Random.new(42)
a.float() // 0.08386297105988216
b.float() // 0.08386297105988216
```

## Possible issues

`Math.random` cannot be seeded, so a seed is a compile error. Use [Random.new](../Random/new.md) for
a seeded generator:

```ant
let x = Math.random(42) // error: Math.random takes 0 arguments, got 1
```

The result is a `float64`, and an `int` literal does not become one. Write `10.0`, not `10`:

```ant
let x = Math.random() * 10 // error: * needs both operands of one type, got float64 and int
```

```ant
let x = Math.random() * 10.0
x < 10.0 // true
```

[Math.round](round.md) of a scaled draw makes the two ends half as likely as the numbers between
them. For a uniform `int`, use [Math.randomInt](randomInt.md):

```ant
let die = Math.randomInt(1, 7)
die >= 1 && die <= 6 // true
```

## See also

[Math.randomInt](randomInt.md) · [Random](../Random/index.md) · [Random.float](../Random/float.md) ·
[Crypto.randomBytes](../Crypto/randomBytes.md) · [Math namespace](index.md)
