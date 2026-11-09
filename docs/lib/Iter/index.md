<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter

Iter holds the functions of lazy sequences, Seq\[T\]. Each returns a new sequence without running
its source; [`Iter.toArray`](toArray.md) and for run one.

<!-- end generated -->

The `Iter` namespace holds the functions on sequences. A [Seq](types/Seq.md) is a function that
pushes its elements to a callback one at a time, so a sequence can be endless and still be used: the
consumer asks for only the elements it needs. `s.map(f)` and `s | Iter.map(f)` call
`Iter.map(s, f)`. `Iter.range` and `Iter.fromArray` build a sequence from something else, so you
call them as `Iter.fromArray(xs)` or `xs | Iter.fromArray()`, not through `.`.

## Details

- Building a sequence runs nothing. A `for` loop or [Iter.toArray](toArray.md) runs it, and the
  callbacks of [Iter.map](map.md) and [Iter.filter](filter.md) run then, element by element.
- Each run runs the source again, and the callbacks with it. A source that cannot repeat, such as
  [Channel.values](../Channel/values.md), gives only what is left.
- A chain stops as soon as its consumer does: after a `break`, or once [Iter.take](take.md) has its
  elements, the source produces nothing more.
- On an array, `.map` and `.filter` are the eager `Array` functions. On a sequence they are the lazy
  `Iter` ones. The type of the receiver decides.

## Basic examples

Chain the functions in a pipe, and collect the result with `Iter.toArray`:

```ant
Iter.range(0, 1000) | Iter.filter({ x % 7 == 0 }) | Iter.take(5) | Iter.toArray() // [0, 7, 14, 21, 28]
```

A `for` loop runs a sequence directly:

```ant
for i in Iter.range(1, 4) {
  println("step ${i}")
}
```

```text
step 1
step 2
step 3
```

The same chain reads left to right through `.`:

```ant
Iter.fromArray(["ant", "bee", "fly"]).map({ String.toUpper(x) }).toArray() // ["ANT", "BEE", "FLY"]
```

## Scope

An empty range is an empty sequence, and every function passes it through:

```ant
Iter.range(3, 3).toArray()                // []
Iter.range(5, 0).map({ x * 2 }).toArray() // []
```

A sequence you write yourself works with every function here. This one never runs out, and
`Iter.take` ends it:

```ant
func powersOfTwo(): Seq[int] {
  func(yield: func(int): bool) {
    var p = 1
    while yield(p) { p *= 2 }
  }
}
powersOfTwo() | Iter.take(6) | Iter.toArray() // [1, 2, 4, 8, 16, 32]
```

The elements may be records, and a `for` loop destructures them:

```ant
type Item = { name: string, price: int }
var items = [Item{ name: "tea", price: 3 }, Item{ name: "cake", price: 5 }]
for { name, price } in Iter.fromArray(items).filter({ x.price > 4 }) {
  println("${name}: ${price}")
}
```

```text
cake: 5
```

## Functions by job

### Building sequences

- [Iter.range](range.md): the `int`s from `from` up to `to`
- [Iter.fromArray](fromArray.md): the elements of an array, in order

### Transforming

- [Iter.map](map.md): apply a function to each element
- [Iter.filter](filter.md): the elements a predicate accepts
- [Iter.take](take.md): at most the first `n` elements

### Running

- [Iter.toArray](toArray.md): run a sequence and collect its elements

### Types

- [Seq](types/Seq.md): a sequence, `func(func(T): bool): void`

Other namespaces return sequences too, such as [Channel.values](../Channel/values.md).

## Applications

Find the first square above 500 without choosing how many to compute:

```ant
func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var n = 1
    while yield(n) { n += 1 }
  }
}

for sq in naturals() | Iter.map({ x * x }) {
  if sq > 500 {
    println(sq)
    break
  }
}
```

```text
529
```

Generate the Collatz steps from a number, as a sequence that ends on its own:

```ant
func collatz(start: int): Seq[int] {
  func(yield: func(int): bool) {
    var n = start
    while n != 1 {
      if !yield(n) { return }
      n = if n % 2 == 0 { n / 2 } else { 3 * n + 1 }
    }
    _ = yield(1)
  }
}

println(collatz(6).toArray())
println(len(collatz(27).toArray()))
```

```text
[6, 3, 10, 5, 16, 8, 4, 2, 1]
112
```

## Properties and relations

On an array, the `Iter` chain gives what the `Array` functions give:

```ant
var xs = [4, 9, 2, 7]
var lazy = Iter.fromArray(xs).filter({ x > 3 }).map({ x * 10 }).toArray()
assert(lazy == xs.filter({ x > 3 }).map({ x * 10 }))
```

`Iter.range` and `Array.range` hold the same integers:

```ant
assert(Iter.range(2, 6).toArray() == Array.range(2, 6))
```

The callbacks run only when the sequence runs, and only as far as the consumer asks:

```ant
var seen: Array[int] = []
var s = Iter.range(0, 100).map({ Array.push(seen, x); x })
seen                // []
s.take(3).toArray() // [0, 1, 2]
seen                // [0, 1, 2]
```

## Possible issues

Each run calls the callbacks again. Collect a sequence once with `Iter.toArray` when they are costly
or have effects:

```ant
var calls = 0
var doubled = Iter.range(0, 3).map({ calls += 1; x * 2 })
doubled.toArray() // [0, 2, 4]
doubled.toArray() // [0, 2, 4]
calls             // 6
```

The `Iter` functions take a sequence, not an array. Wrap an array with `Iter.fromArray`:

```ant
var doubled = Iter.map([1, 2], { x * 2 }) // error: cannot use Array[int] where Seq[T] is expected
```

`Iter.take` panics on a negative count:

```ant
Iter.range(0, 5).take(-1) // panic: Iter.take needs a count of at least 0, got -1
```

`Iter.fromArray` reads the array when the sequence runs, not when you build it:

```ant
var a = [1, 2]
var s = Iter.fromArray(a)
Array.push(a, 3)
s.toArray() // [1, 2, 3]
```

Build it from a copy, `Iter.fromArray([...a])`, to keep the elements as they are now.

`Iter.toArray` runs the whole sequence, so on an endless one it never returns. Put `Iter.take` in
front of it, or consume the sequence with a `for` loop that breaks.

## See also

[Seq](types/Seq.md), [Array](../Array/index.md), [Channel.values](../Channel/values.md),
[for](../../language/keywords/for.md), [Sequences](../../tour/09-sequences.md),
[Collections](../guides/collections.md)
