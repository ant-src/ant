<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.take

```ant signature
func Iter.take[T](s: Seq[T], n: int): Seq[T]
```

Iter.take returns a sequence of at most the first n elements of s. It panics if n \< 0.

<!-- end generated -->

## Details

- `Iter.take` is lazy: it returns a new sequence without running `s`. The source runs when a `for`
  loop or [Iter.toArray](toArray.md) runs the result.
- When `s` has fewer than `n` elements, the result has all of them. `n` of `0` gives an empty
  sequence.
- The result is finite even when `s` never ends, so `Iter.take` is how you bound an endless
  sequence.

## Basic examples

Keep the first three numbers of a range:

```ant
Iter.range(0, 10) | Iter.take(3) | Iter.toArray() // [0, 1, 2]
```

Loop over the first two elements of an array:

```ant
for name in Iter.take(Iter.fromArray(["Ann", "Bob", "Cy"]), 2) {
  println(name)
}
```

```text
Ann
Bob
```

## Scope

A count past the end gives every element; a count of `0` gives none:

```ant
Iter.range(0, 2) | Iter.take(5) | Iter.toArray() // [0, 1]
Iter.range(0, 2) | Iter.take(0) | Iter.toArray() // []
Iter.range(5, 5) | Iter.take(3) | Iter.toArray() // []
```

Records and `null` elements pass through unchanged:

```ant
type Task = { title: string, done: bool }
var tasks = [Task{ title: "write", done: true }, Task{ title: "test", done: false }]
Iter.fromArray(tasks) | Iter.take(1) | Iter.toArray() // [Task{ title: "write", done: true }]

Iter.fromArray([null, 2, null]) | Iter.take(2) | Iter.toArray() // [null, 2]
```

Call it through `.`:

```ant
Iter.fromArray(["a", "b", "c"]).take(2).toArray() // ["a", "b"]
```

## Applications

Find the first five multiples of 7 without deciding how far to search. `Iter.take` stops the endless
source once it has five:

```ant
func from(start: int): Seq[int] {
  func(yield: func(int): bool) {
    var i = start
    while true {
      if !yield(i) { return }
      i += 1
    }
  }
}

println(from(1) | Iter.filter({ x % 7 == 0 }) | Iter.take(5) | Iter.toArray())
```

```text
[7, 14, 21, 28, 35]
```

Print the first six Fibonacci numbers, one per line:

```ant
func fibonacci(): Seq[int] {
  func(yield: func(int): bool) {
    var a = 0
    var b = 1
    while true {
      if !yield(a) { return }
      let next = a + b
      a = b
      b = next
    }
  }
}

for n in fibonacci() | Iter.take(6) {
  println(n)
}
```

```text
0
1
1
2
3
5
```

Read only the first values sent on a channel with [Channel.values](../Channel/values.md):

```ant
var c: Channel[string] = Channel.new(capacity: 4)
for word in ["one", "two", "three", "four"] {
  Channel.send(c, word)
}
Channel.close(c)
for word in Channel.values(c) | Iter.take(2) {
  println(word)
}
```

```text
one
two
```

## Properties and relations

Taking `n` elements gives the same array as slicing the first `n`, when there are at least `n`:

```ant
var xs = [4, 8, 15, 16, 23, 42]
assert((Iter.fromArray(xs) | Iter.take(3) | Iter.toArray()) == Array.slice(xs, 0, 3))
```

The result never has more than `n` elements, nor more than `s` has:

```ant
for n in Iter.range(0, 6) {
  let got = Iter.range(0, 3) | Iter.take(n) | Iter.toArray()
  assert(len(got) == Math.min(n, 3))
}
```

Taking twice keeps the smaller count:

```ant
let s = Iter.range(0, 10)
assert((s | Iter.take(5) | Iter.take(2) | Iter.toArray()) == (s | Iter.take(2) | Iter.toArray()))
```

## Possible issues

A negative count panics:

```ant
Iter.take(Iter.range(0, 5), -1) // panic: Iter.take needs a count of at least 0, got -1
```

`Iter.take` returns a sequence, not an array. Run it with [Iter.toArray](toArray.md) to get one:

```ant
var first: Array[int] = Iter.range(0, 5) | Iter.take(2)
// error: cannot use Seq[int] where Array[int] is expected
```

## See also

[Iter.range](range.md), [Iter.fromArray](fromArray.md), [Iter.filter](filter.md),
[Iter.map](map.md), [Iter.toArray](toArray.md), [Array.slice](../Array/slice.md),
[Channel.values](../Channel/values.md)
