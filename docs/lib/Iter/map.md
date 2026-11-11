<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.map

```ant signature
func Iter.map[T, U](s: Seq[T], f: func(T): U): Seq[U]
```

Iter.map returns a sequence of f applied to each element of s. Like every Iter function but
[`Iter.toArray`](toArray.md), it runs nothing until the result is run.

<!-- end generated -->

## Details

- Lazy: `Iter.map` calls `f` only when the result runs, through `Iter.toArray` or `for`, and calls
  it once for each element the run takes from `s`, in the order of `s`, so a `break` or
  [`Iter.take`](take.md) stops the calls.
- Each run of the result runs `s` again and calls `f` again; nothing is cached. A second run sees
  only what `s` gives the second time, so a sequence from [`Channel.values`](../Channel/values.md)
  gives nothing more once the channel is drained.

## Basic examples

Square each number of a range, then collect the result:

```ant
Iter.map(Iter.range(1, 4), { x * x }) | Iter.toArray() // [1, 4, 9]
```

Loop over the mapped sequence with `for`:

```ant
for word in Iter.map(Iter.fromArray(["ant", "bee"]), String.toUpper) {
  println(word)
}
```

```text
ANT
BEE
```

## Scope

An empty sequence gives an empty sequence:

```ant
Iter.range(0, 0) | Iter.map({ x * 2 }) | Iter.toArray() // []
```

The element type may change:

```ant
Iter.range(1, 4) | Iter.map({ "#${x}" }) | Iter.toArray() // ["#1", "#2", "#3"]
```

Read a field of each record:

```ant
type User = { name: string, age: int }
let users = [User{ name: "Ann", age: 31 }, User{ name: "Bob", age: 27 }]
Iter.fromArray(users) | Iter.map({ x.name }) | Iter.toArray() // ["Ann", "Bob"]
```

Replace the `null` elements with `??`:

```ant
let scores: Array[int | null] = [90, null, 75]
Iter.fromArray(scores) | Iter.map({ x ?? 0 }) | Iter.toArray() // [90, 0, 75]
```

Pass a named function or a conversion instead of a lambda:

```ant
func cube(n: int): int { n * n * n }
Iter.range(1, 4) | Iter.map(cube) | Iter.toArray()    // [1, 8, 27]
Iter.range(1, 3) | Iter.map(float64) | Iter.toArray() // [1.0, 2.0]
```

Map a sequence you write yourself. A sequence is a function that calls `yield` for each element:

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
evens(7) | Iter.map({ x / 2 }) | Iter.toArray() // [0, 1, 2, 3]
```

Call it through `.` or in a pipe:

```ant
Iter.range(0, 3).map({ x + 1 }).toArray()            // [1, 2, 3]
Iter.range(0, 3) | Iter.map({ -x }) | Iter.toArray() // [0, -1, -2]
```

Name the arguments to pass them in any order:

```ant
Iter.map(f: { x * 10 }, s: Iter.range(1, 3)) | Iter.toArray() // [10, 20]
```

## Applications

Take the first few elements of a sequence that never ends:

```ant
func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var n = 0
    while yield(n) {
      n += 1
    }
  }
}
naturals() | Iter.map({ x * x }) | Iter.take(5) | Iter.toArray() // [0, 1, 4, 9, 16]
```

Transform the values a goroutine sends on a channel as they arrive:

```ant
let c: Channel[int] = Channel.new()
go {
  for i in Iter.range(1, 4) {
    c.send(i)
  }
  c.close()
}
for line in Channel.values(c) | Iter.map({ "got ${x}" }) {
  println(line)
}
```

```text
got 1
got 2
got 3
```

Chain `map` with `filter` and collect once at the end:

```ant
Iter.range(1, 10) | Iter.map({ x * 3 }) | Iter.filter({ x % 2 == 0 }) | Iter.toArray() // [6, 12, 18, 24]
```

## Properties and relations

On an array, collecting the mapped sequence gives what `Array.map` gives:

```ant
let xs = [1, 2, 3]
assert((Iter.fromArray(xs) | Iter.map({ x * 2 }) | Iter.toArray()) == Array.map(xs, { x * 2 }))
```

Building the sequence calls nothing. Each run calls `f` once per element, and running it again calls
`f` again:

```ant
var calls = 0
let doubled = Iter.range(0, 3) | Iter.map({ calls += 1; x * 2 })
calls             // 0
doubled.toArray() // [0, 2, 4]
calls             // 3
doubled.toArray() // [0, 2, 4]
calls             // 6
```

Followed by `Iter.take(n)`, the map calls `f` only `n` times, however long `s` is:

```ant
var mapped = 0
let s = Iter.range(0, 1000000) | Iter.map({ mapped += 1; x })
s | Iter.take(5) | Iter.toArray() // [0, 1, 2, 3, 4]
mapped                            // 5
```

A `break` in a `for` loop stops the calls too. `f` never sees the elements after the break:

```ant
let seen: Array[int] = []
for n in Iter.range(1, 100) | Iter.map({ seen.push(x); x * x }) {
  if n > 10 { break }
}
println(seen)
```

```text
[1, 2, 3, 4]
```

When `f` and `g` have no side effects, two maps in a row are one map with the composed function:

```ant
let f: func(int): int = { x + 1 }
let g: func(int): int = { x * 10 }
let s = Iter.range(0, 5)
assert((s | Iter.map(f) | Iter.map(g) | Iter.toArray()) == (s | Iter.map({ g(f(x)) }) | Iter.toArray()))
```

## Possible issues

`Iter.map` takes a sequence, not an array. Pass an array through `Iter.fromArray` first:

```ant
let xs = [1, 2, 3]
let doubled = Iter.map(xs, { x * 2 }) // error: cannot use Array[int] where Seq[T] is expected
```

The result is a sequence, not an array. Collect it with `Iter.toArray` where an array is expected:

```ant
let squares: Array[int] = Iter.range(0, 3) | Iter.map({ x * x })
// error: cannot use Seq[int] where Array[int] is expected
```

A panic in `f` comes from the run, not from the line that builds the sequence:

```ant
let s = Iter.range(0, 3) | Iter.map({ 10 / (x - 1) })
println("built")
let results = s.toArray() // panic: division by zero
```

```text
built
```

On an array, `.map` is [`Array.map`](../Array/map.md), which is eager and gives an array. Wrap the
array with [`Iter.fromArray`](fromArray.md) to get a lazy sequence:

```ant
[1, 2, 3].map({ x * 2 })                                  // [2, 4, 6]
Iter.fromArray([1, 2, 3]).map({ x * 2 }) | Iter.toArray() // [2, 4, 6]
```

The sequence reads the array when it runs, not when you build it:

```ant
let a = [1, 2]
let s = Iter.fromArray(a) | Iter.map({ x * 10 })
a.push(3)
s.toArray() // [10, 20, 30]
```

`f` takes one argument, the element. A function of two parameters is a compile error:

```ant
let s = Iter.range(0, 2) | Iter.map(func(n: int, i: int): int { n + i })
// error: cannot use func(int, int): int where func(int): U is expected
```

## See also

[Iter.filter](filter.md), [Iter.take](take.md), [Iter.toArray](toArray.md),
[Iter.fromArray](fromArray.md), [Iter.range](range.md), [Array.map](../Array/map.md),
[Channel.values](../Channel/values.md)
