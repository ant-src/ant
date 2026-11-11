<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Seq

```ant signature
type Seq[T] = func(func(T): bool): void
```

Seq is a sequence: a function that pushes each element to yield, in order, until yield returns false
or the elements run out. A for loop iterates one directly.

<!-- end generated -->

## Details

- A sequence is lazy: building one runs nothing. A `for` loop, [`Iter.toArray`](../toArray.md), or a
  direct call runs it. The other functions of [`Iter`](../index.md) return a new sequence without
  running it.
- A sequence must return once `yield` returns `false`. Calling `yield` again after that, or after
  the loop has ended, panics.
- A sequence calls `yield` only on the goroutine that runs it. To produce elements on other
  goroutines, send them on a channel and return [`Channel.values`](../../Channel/values.md).
- A sequence reads its source each time it runs: a range gives the same elements again, an array
  gives its current elements, and a channel's values are gone after one run.

## Basic examples

Write a sequence as a function that calls `yield` for each element and stops when `yield` returns
`false`. A `for` loop iterates it:

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

for n in evens(7) {
  println(n)
}
```

```text
0
2
4
6
```

The `Iter` functions return sequences, and `Iter.toArray` collects one into an array:

```ant
let s: Seq[int] = Iter.range(1, 4)
s.toArray()                // [1, 2, 3]
s.map({ x * x }).toArray() // [1, 4, 9]
```

## Scope

A sequence of no elements never calls `yield`:

```ant
let none: Seq[string] = func(yield: func(string): bool) {}
none.toArray() // []
```

Any function of the right type is a sequence. A lambda's `x` here is the `yield` function, and `&&`
stops at the first `false`. The body of a sequence is `void`, so `_ =` discards the `bool`:

```ant
let abc: Seq[string] = {
  _ = x("a") && x("b") && x("c")
}
abc.toArray() // ["a", "b", "c"]
```

A sequence may never run out. The sequence produces only the elements a consumer asks for, so
[`Iter.take`](../take.md) makes it finite:

```ant
func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var i = 0
    while yield(i) { i += 1 }
  }
}

println(naturals().take(4).toArray())
println(naturals().filter({ x % 7 == 0 }).take(3).toArray())
```

```text
[0, 1, 2, 3]
[0, 7, 14]
```

The elements may be records, and a `for` loop destructures them:

```ant
type Point = { x: int, y: int }
let diagonal = Iter.range(0, 3).map({ Point{ x: x, y: x } })
for { x, y } in diagonal {
  println("(${x}, ${y})")
}
```

```text
(0, 0)
(1, 1)
(2, 2)
```

A generic function takes any sequence through `Seq[T]`:

```ant
func count[T](s: Seq[T]): int {
  var n = 0
  for _ in s { n += 1 }
  n
}

println(count(Iter.range(0, 5)))
println(count(Iter.fromArray(["a", "b"])))
```

```text
5
2
```

Call a sequence directly with your own `yield`. Return `true` to ask for the next element and
`false` to stop:

```ant
let words = Iter.fromArray(["ant", "bee", "fly"])
words({
  println(x)
  x != "bee"
})
```

```text
ant
bee
```

Chain the `Iter` functions through `.` or in a pipe:

```ant
Iter.range(0, 30).filter({ x % 7 == 0 }).toArray()                              // [0, 7, 14, 21, 28]
Iter.range(0, 30) | Iter.filter({ x % 7 == 0 }) | Iter.take(2) | Iter.toArray() // [0, 7]
```

Produce the elements on another goroutine by sending them on a channel, and return the channel's
values as the sequence:

```ant
func squares(n: int): Seq[int] {
  let c: Channel[int] = Channel.new()
  go {
    for i in Iter.range(1, n + 1) {
      c.send(i * i)
    }
    c.close()
  }
  c.values()
}

println(squares(4).toArray())
```

```text
[1, 4, 9, 16]
```

## Applications

Walk a tree in depth-first order. A helper passes `yield` down and reports whether to go on; the
sequence discards that report with `_ =`:

```ant
type Tree = { value: int, children: Array[Tree] }

func visit(t: Tree, yield: func(int): bool): bool {
  if !yield(t.value) { return false }
  for c in t.children {
    if !visit(c, yield) { return false }
  }
  true
}

func Tree.values(t: Tree): Seq[int] {
  func(yield: func(int): bool) { _ = visit(t, yield) }
}

let leaf = func(n: int): Tree { Tree{ value: n, children: [] } }
let t = Tree{ value: 1, children: [leaf(2), Tree{ value: 3, children: [leaf(4)] }] }
println(t.values().toArray())
println(t.values().filter({ x % 2 == 0 }).toArray())
```

```text
[1, 2, 3, 4]
[2, 4]
```

Find the first element past a limit. The `return` in the loop body leaves the function and ends the
sequence, so an endless sequence is safe here:

```ant
func firstOver(s: Seq[int], n: int): int | null {
  for x in s {
    if x > n { return x }
  }
  null
}

let powers: Seq[int] = func(yield: func(int): bool) {
  var p = 1
  while yield(p) { p *= 2 }
}
println(firstOver(powers, 1000))
```

```text
1024
```

## Properties and relations

A `for` loop and `Iter.toArray` see the same elements, in the same order:

```ant
let s = Iter.range(0, 5).map({ x * 3 })
let seen: Array[int] = []
for x in s {
  seen.push(x)
}
assert(seen == s.toArray())
```

`Iter.fromArray` and `Iter.toArray` round-trip an array:

```ant
let xs = ["a", "b", "c"]
assert(Iter.fromArray(xs).toArray() == xs)
```

A sequence built from a range gives the same elements every time it runs:

```ant
let s = Iter.range(0, 3)
assert(s.toArray() == s.toArray())
```

## Possible issues

A sequence that ignores what `yield` returns keeps calling it after a `break`, and the loop panics:

```ant
let ab: Seq[string] = func(yield: func(string): bool) {
  for w in ["a", "b"] {
    yield(w)
    // panic: yield called again after it returned false; a sequence must return when yield returns false
  }
}
for s in ab {
  if s == "a" { break }
}
```

Stop at the first `false` with `if !yield(w) { return }`.

A channel's values run once. Running the sequence again finds the channel drained:

```ant
let c: Channel[int] = Channel.new(2)
c.send(1)
c.send(2)
c.close()
let s = c.values()
println(s.toArray())
println(s.toArray())
```

```text
[1, 2]
[]
```

`Iter.fromArray(a)` reads `a` again on each run, so a push between runs shows up in the next run:

```ant
let a = [1]
let s = Iter.fromArray(a)
println(s.toArray())
a.push(2)
println(s.toArray())
```

```text
[1]
[1, 2]
```

A sequence returns nothing. A function that returns a `bool` is not a `Seq`:

```ant
let s: Seq[int] = func(yield: func(int): bool): bool { yield(1) }
// error: cannot use func(func(int): bool): bool where Seq[int] is expected
```

`for` iterates only a function of the sequence's shape:

```ant
let positive: func(int): bool = { x > 0 }
for v in positive { // error: for needs an array, a map, a set, a string, Bytes, or a sequence, got func(int): bool
  println(v)
}
```

## See also

[Iter](../index.md) · [Iter.range](../range.md) · [Iter.fromArray](../fromArray.md) ·
[Iter.map](../map.md) · [Iter.filter](../filter.md) · [Iter.take](../take.md) ·
[Iter.toArray](../toArray.md) · [Channel.values](../../Channel/values.md)
