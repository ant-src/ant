<!-- status: reviewed -->

# Collections and sequences

Ant has three collections and one way to describe a stream of values:

| You need                                       | Use                  |
| ---------------------------------------------- | -------------------- |
| values in order, read by position              | `Array[T]`           |
| a value for each key                           | `Map[K, V]`          |
| each value once, and a fast "is it there?"     | `Set[T]`             |
| values produced one at a time, maybe never all | a sequence, `Seq[T]` |

The three collections hold their values in memory, change in place, and compare by contents. A
sequence holds nothing: it is a function that hands out values when a loop asks for them.

## Arrays, maps, and sets

```ant
let scores = [72, 95, 88]
scores.push(61)

let ages = Map{ ada: 36, linus: 54 }
ages["grace"] = 85

let tags = Set.of(["red", "blue", "red"])
tags.add("green")

println(scores)
println(ages)
println(tags)
```

```text
[72, 95, 88, 61]
Map{ "ada": 36, "linus": 54, "grace": 85 }
Set.of(["red", "blue", "green"])
```

Pick the collection by the question your code asks most. "What is the third score?" is an array.
"How old is Ada?" is a map. "Have I seen this tag?" is a set: `Array.contains` answers it too, but
it searches the array element by element, while `Set.has` goes straight to the member.

`Set.add` returns whether it changed the set, so `if seen.add(line) { println(line) }` keeps the
first of each value. A set also combines with another: `a.union(b)` has the members of `a`, then
those of `b` that `a` lacks; `a.intersect(b)` and `a.difference(b)` keep the members of `a` that `b`
has or lacks; `a.isSubset(b)` asks whether `b` has every member of `a`. Each returns a new set:

```ant
let a = Set.of([1, 2, 3])
let b = Set.of([3, 4])
a.union(b)              // Set.of([1, 2, 3, 4])
a.intersect(b)          // Set.of([3])
a.difference(b)         // Set.of([1, 2])
Set.of([3]).isSubset(b) // true
```

Maps and sets keep insertion order. A `for` loop, `keys`, `values`, and printing all visit entries
in the order their keys were first added, so output does not shuffle between runs. Replacing a value
keeps its entry where it was; a key removed and added again goes to the end.

A set has no literal: `Set.of` builds one from an array and keeps the first of each value. An empty
collection needs its type written out, since nothing in it says what it holds:

```ant
let names: Array[string] = []
let counts: Map[string, int] = Map{}
let seen: Set[int] = Set.of([])
```

### Reading what may not be there

An index out of range panics, whether you read or write it. A map read never panics: a missing key
gives `null`, so the type of `ages["alan"]` is `int | null`.

```ant
let xs = [10, 20, 30]
let ages = Map{ ada: 36 }
xs.get(5)         // null
ages["alan"] ?? 0 // 0
ages.has("ada")   // true
xs[5]             // panic: index 5 out of range for an array of length 3
```

The `?? 0` default is how you count or add up per key, the job of `defaultdict` in Python or
`m[k]++` in Go:

```ant
let words = ["a", "b", "a"]
let counts: Map[string, int] = Map{}
for w in words {
  counts[w] = (counts[w] ?? 0) + 1
}
println(counts)
```

```text
Map{ "a": 2, "b": 1 }
```

When the counts come from an array, `Array.countBy` does this in one call: `words.countBy({ x })`.
To collect the values themselves, `Array.groupBy` builds a map of arrays.

Use `xs[i]` when an index outside the array is a bug, and `xs.get(i)` when it is an ordinary case.
`Map.has` tells a missing key apart from a key whose value is `null`, when the value type allows
`null`.

## Changing a collection

A collection is a reference. `let` fixes which collection a name holds, not what is inside it, so
`scores.push(61)` above works on a `let`. A function that receives a collection receives the same
one, and its changes show through every name:

```ant
func addDefaults(tags: Set[string]) {
  tags.add("new")
}

let tags = Set.of(["red"])
addDefaults(tags)
println(tags)
```

```text
Set.of(["red", "new"])
```

Few functions change a collection in place: `push`, `pop`, `insert`, `removeAt`, and `extend` on
arrays, `set` and `remove` on maps, `add` and `remove` on sets, and `Random.shuffle`, plus the
assignments `xs[i] = v` and `m[k] = v`. Everything else returns a new collection and leaves its
argument alone. That includes `sort` and `reverse`, which change the array in place in JavaScript
and Python:

```ant
let xs = [3, 1, 2]
xs.sort()    // [1, 2, 3]
xs           // [3, 1, 2]
xs.reverse() // [2, 1, 3]
xs           // [3, 1, 2]
```

Keep the result: `let sorted = xs.sort()`.

The variable of a `for` loop is a fresh `let` for each element, so assigning it is a compile error
rather than a silent no-op. To change the elements, build a new array with `map`, or write by index:

```ant
let xs = [1, 2, 3]
for x in xs { x = x * 2 }
// error: cannot assign to x, a for variable; declare a var to hold a value that changes
```

```ant
let xs = [1, 2, 3]
let doubled = xs.map({ x * 2 })
for i in Array.range(0, xs.length()) {
  xs[i] *= 2
}
println(doubled)
println(xs)
```

```text
[2, 4, 6]
[2, 4, 6]
```

Keep the `Array.range` loop for writing by index. To read each element with its index, loop over
`xs.entries()`, which pairs them as `Indexed` records:

```ant
let names = ["ada", "linus"]
for { index, value } in names.entries() {
  println("${index}: ${value}")
}
```

```text
0: ada
1: linus
```

The same goes for maps. Each `Entry` a loop over a map visits is a new record, so assigning to its
`value` field changes the entry, not the map. Write through the map instead:

```ant
let stock = Map{ apples: 3, pears: 0 }
for { key, value } in stock {
  if value == 0 { stock[key] = 10 }
}
println(stock)
```

```text
Map{ "apples": 3, "pears": 10 }
```

A loop over an array reads the length on each step, so elements pushed during the loop are visited
too. A loop over a map or a set visits entries added during the loop as well.

## Equality and identity

`==` compares contents, for every collection. Two arrays are equal when they have the same elements
in the same order; two maps when they have the same keys with equal values, in any order; two sets
when they have the same members, in any order:

```ant
[1, 2, 3] == [1, 2, 3]                   // true
Map{ a: 1, b: 2 } == Map{ b: 2, a: 1 }   // true
Set.of(["x", "y"]) == Set.of(["y", "x"]) // true
[[1, 2], [3]] == [[1, 2], [3]]           // true
```

This is not JavaScript, where `==` on two arrays asks whether they are one object, and not Go, where
`==` on slices and maps does not compile. Ant has no `===`. When you need to know whether two names
hold the same collection, ask with `same`:

```ant
let xs = [1, 2, 3]
let alias = xs
let copy = [...xs]
xs == copy            // true
Array.same(xs, alias) // true
Array.same(xs, copy)  // false
```

The spread `[...xs]` makes a new array with the same elements. It copies one level: an array of
records holds the same records. `Map.same`, `Set.same`, and `Record.same` ask the same question for
maps, sets, and records.

### Keys and members

Maps and sets find keys with `==` too, so any value can be a key, including a record or an array.
Two keys that are `==` are the same key:

```ant
type Point = { x: int, y: int }
let visited: Set[Point] = Set.of([])
visited.add(Point{ x: 1, y: 2 }) // true
visited.has(Point{ x: 1, y: 2 }) // true
visited.add(Point{ x: 1, y: 2 }) // false
visited.size()                   // 1
```

A map hashes a key when you insert it. Do not change a record or an array after you use it as a key
or a member: the map does not notice, and may no longer find the entry. Insert a copy, such as
`Point{ ...p }`, when you still need to change the original.

## Lazy sequences with Iter

`Array.map` and `Array.filter` are eager: each one runs over the whole array and builds a new one.
For most work that is what you want. When the source is large or endless, or you need only the first
few results, use the `Iter` functions instead. They build a sequence that does nothing until a loop
or `Iter.toArray` runs it, and they stop the source as soon as the consumer is done:

```ant
var calls = 0
func square(n: int): int {
  calls += 1
  n * n
}

let eager = Array.range(0, 1000)
  | Array.map(square)
  | Array.filter({ x % 3 == 0 })
println("eager: ${eager.slice(0, 3)} after ${calls} calls")

calls = 0
let lazy = Iter.range(0, 1000)
  | Iter.map(square)
  | Iter.filter({ x % 3 == 0 })
  | Iter.take(3)
  | Iter.toArray()
println("lazy: ${lazy} after ${calls} calls")
```

```text
eager: [0, 9, 36] after 1000 calls
lazy: [0, 9, 36] after 7 calls
```

`Iter.fromArray(xs)` turns an array into a sequence, and `Iter.range(from, to)` counts without
building an array. A sequence runs again, from the start, each time you loop over it, when its
source allows: `Iter.range` and `Iter.fromArray` do. A pipeline of `Iter` calls is a description,
not a result, so name its `toArray` when you need the values more than once.

`Iter` has only these few functions. A pipeline that does not need to stop early reads better as
`Array` calls, which have the full set of `find`, `some`, `sortBy`, `groupBy`, and the rest.

## Writing a sequence with yield

A `Seq[T]` is a function that takes a callback, `yield`, and calls it once per element. A `for` loop
over a sequence passes the loop body as `yield`. When the body ends with `break`, `yield` returns
`false`, and the sequence must return:

```ant
type Tree = { name: string, children: Array[Tree] }

func names(t: Tree): Seq[string] {
  func(yield: func(string): bool) {
    if !yield(t.name) { return }
    for child in t.children {
      for name in names(child) {
        if !yield(name) { return }
      }
    }
  }
}

let root = Tree{
  name: "src",
  children: [
    Tree{ name: "lib", children: [Tree{ name: "io", children: [] }] },
    Tree{ name: "main", children: [] },
  ],
}

for name in names(root) {
  println(name)
  if name == "io" { break }
}
```

```text
src
lib
io
```

The tree is never flattened into an array: the loop stops at `io`, and `main` is never visited. The
`return` in the inner loop leaves the sequence function, which is what `yield` returning `false`
asks for. Because a sequence is an ordinary function, the `Iter` functions work on it too:
`names(root) | Iter.take(2) | Iter.toArray()` is `["src", "lib"]`.

Write a sequence when the values are expensive, endless, or come from a recursive walk. Return an
array when the caller needs them all anyway: an array can be indexed, measured, and looped over many
times.

A sequence that calls `yield` after it returned `false` panics, so a missing
`if !yield(x) { return }` shows up the first time a loop breaks early:

```ant
func upTo(n: int): Seq[int] {
  func(yield: func(int): bool) {
    var i = 0
    while i < n {
      yield(i)
      // panic: yield called again after it returned false; a sequence must return when yield returns false
      i += 1
    }
  }
}

for i in upTo(3) {
  if i == 1 { break }
}
```

### A sequence stays on one goroutine

`yield` runs the loop body, and the body belongs to the goroutine running the loop: a `return` or a
`try` in it leaves the function that loop is in. So a sequence calls `yield` only on the loop's
goroutine. A call from a `go` block panics. A sequence whose values come from goroutines of its own
has them send the values on a channel and yields what `Channel.values` receives:

```ant
func squares(ns: Array[int]): Seq[int] {
  func(yield: func(int): bool) {
    let results: Channel[int] = Channel.new()
    go {
      defer { results.close() }
      for n in ns { results.send(n * n) }
    }
    for r in Channel.values(results) {
      if !yield(r) { return }
    }
  }
}

for r in squares([1, 2, 3]) {
  println(r)
}
```

```text
1
4
9
```

The goroutine computes, the loop's goroutine yields. When the loop breaks early, the values not yet
received stay in the channel; give the goroutine a context to cancel when it would otherwise wait
forever on a send. The [concurrency guide](concurrency.md) covers channels and contexts.

## See also

- [Arrays, maps, and sets](../tour/05-collections.md) and [Sequences](../tour/09-sequences.md) in
  the tour
- [Array](../lib/Array/index.md), [Map](../lib/Map/index.md), [Set](../lib/Set/index.md), and
  [Iter](../lib/Iter/index.md) in the standard library
- [Seq](../lib/Iter/types/Seq.md), [for](../language/keywords/for.md), and
  [equality](../language/operators/equality.md)
- [Array.same](../lib/Array/same.md) and [Record.same](../lib/Record/same.md)
