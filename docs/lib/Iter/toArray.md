<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.toArray

```ant signature
func Iter.toArray[T](s: Seq[T]): Array[T]
```

Iter.toArray runs s and returns its elements as an array.

<!-- end generated -->

## Details

- Runs `s` when called, unlike the other `Iter` functions, so callbacks passed to `Iter.map` or
  `Iter.filter` run at this point.
- Returns a new array each time.
- Returns only when `s` returns: on an endless sequence, bound it first with [`Iter.take`](take.md).
- A sequence over a channel runs once; a second call returns what is left.

## Basic examples

Collect a range:

```ant
Iter.toArray(Iter.range(0, 5)) // [0, 1, 2, 3, 4]
```

End a pipeline of lazy steps with it:

```ant
Iter.range(0, 30) | Iter.filter({ x % 7 == 0 }) | Iter.toArray() // [0, 7, 14, 21, 28]
```

## Scope

An empty sequence gives an empty array:

```ant
Iter.range(5, 0) | Iter.toArray() // []
let none: Seq[string] = func(yield: func(string): bool) {}
Iter.toArray(none) // []
```

Elements of any type come through as they are, `null` included:

```ant
type Point = { x: int, y: int }
Iter.fromArray([Point{ x: 1, y: 2 }]) | Iter.toArray() // [Point{ x: 1, y: 2 }]
let maybe: Seq[int | null] = Iter.fromArray([1, null, 3])
Iter.toArray(maybe) // [1, null, 3]
```

Collect a sequence you write yourself, as a function that calls `yield`:

```ant
func evens(limit: int): Seq[int] {
  func(yield: func(int): bool) {
    var i = 0
    while i < limit {
      if !yield(i) { return }
      i += 2
    }
  }
}
Iter.toArray(evens(10)) // [0, 2, 4, 6, 8]
```

Call it through `.` on a sequence:

```ant
Iter.range(1, 4).map({ x * x }).toArray() // [1, 4, 9]
```

Collect what a channel delivers until it is closed:

```ant
let squares: Channel[int] = Channel.new(3)
go {
  for i in [1, 2, 3] { Channel.send(squares, i * i) }
  Channel.close(squares)
}
println(Channel.values(squares) | Iter.toArray())
```

```text
[1, 4, 9]
```

## Applications

Take the first few elements of an endless sequence:

```ant
func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var i = 0
    while yield(i) { i += 1 }
  }
}
func isPrime(n: int): bool {
  n >= 2 && Iter.range(2, n).filter({ x * x <= n && n % x == 0 }).toArray().isEmpty()
}
naturals() | Iter.filter(isPrime) | Iter.take(6) | Iter.toArray() // [2, 3, 5, 7, 11, 13]
```

Build a pipeline once and collect it later, when its callbacks should run:

```ant
let labels = Iter.range(0, 3) | Iter.map({
  println("label ${x}")
  "item ${x}"
})
println("built")
println(Iter.toArray(labels))
```

```text
built
label 0
label 1
label 2
["item 0", "item 1", "item 2"]
```

## Properties and relations

Collecting `Iter.fromArray(xs)` gives an array equal to `xs`, but not the same array:

```ant
let xs = [3, 1, 2]
let copy = Iter.toArray(Iter.fromArray(xs))
assert(copy == xs)
Array.push(copy, 4)
assert(xs == [3, 1, 2])
```

Collecting a lazy `Iter.map` gives what `Array.map` gives on the same elements:

```ant
let xs = [1, 2, 3]
assert(Iter.fromArray(xs).map({ x + 1 }).toArray() == Array.map(xs, { x + 1 }))
```

A `for` loop over the sequence visits the same elements, in the same order:

```ant
let seen: Array[int] = []
for n in Iter.range(0, 4) { Array.push(seen, n) }
assert(seen == Iter.toArray(Iter.range(0, 4)))
```

## Possible issues

A sequence over an array reads the array each time it runs, so two calls can differ:

```ant
let xs = [1, 2]
let seq = Iter.fromArray(xs)
Iter.toArray(seq) // [1, 2]
Array.push(xs, 3)
Iter.toArray(seq) // [1, 2, 3]
```

A sequence over a channel gives each value once, so a second run finds nothing:

```ant
let c: Channel[int] = Channel.new(3)
Channel.send(c, 1)
Channel.send(c, 2)
Channel.close(c)
let vs = Channel.values(c)
println(Iter.toArray(vs))
println(Iter.toArray(vs))
```

```text
[1, 2]
[]
```

An array is not a sequence. Convert it with `Iter.fromArray`, or use the `Array` functions directly:

```ant
let xs = [1, 2, 3]
let ys = Iter.toArray(xs) // error: cannot use Array[int] where Seq[T] is expected
```

A panic in the sequence reaches the caller of `Iter.toArray`. `Iter.take` stops the run before the
panic, so the first call returns:

```ant
func countdown(n: int): Seq[int] {
  func(yield: func(int): bool) {
    var i = n
    while true {
      if i < 0 { panic("countdown below zero") }
      if !yield(i) { return }
      i -= 1
    }
  }
}
countdown(2) | Iter.take(3) | Iter.toArray() // [2, 1, 0]
countdown(2) | Iter.toArray()                // panic: countdown below zero
```

## See also

[Iter.fromArray](fromArray.md), [Iter.range](range.md), [Iter.map](map.md),
[Iter.filter](filter.md), [Iter.take](take.md), [Channel.values](../Channel/values.md),
[for](../../language/keywords/for.md)
