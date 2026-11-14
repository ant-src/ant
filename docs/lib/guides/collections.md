<!-- status: reviewed -->

# Collections

Ant has three built-in collections and one lazy sequence type. An array holds elements in order, a
map holds a value under each key, a set holds each value once, and a sequence produces its elements
on demand. The functions for each live in a namespace: [Array](../Array/index.md),
[Map](../Map/index.md), and [Set](../Set/index.md) for the collections, and [Iter](../Iter/index.md)
for sequences.

```ant
let scores = [72, 95, 88]
let ages = Map{ ada: 36, linus: 54 }
let tags = Set.of(["red", "blue", "red"])
let sevens = Iter.range(0, 1000000) | Iter.filter({ x % 7 == 0 })
scores[1]                              // 95
ages["ada"]                            // 36
tags                                   // Set.of(["red", "blue"])
tags.has("red")                        // true
sevens | Iter.take(3) | Iter.toArray() // [0, 7, 14]
```

## Choosing a collection

| You need                                         | Use                                      |
| ------------------------------------------------ | ---------------------------------------- |
| Elements in order, read by position              | `Array[T]`, built with `[...]`           |
| A value for each key                             | `Map[K, V]`, built with `Map{ ... }`     |
| Each value once, and membership tests            | `Set[T]`, built with `Set.of([...])`     |
| Elements computed one at a time, maybe endlessly | `Seq[T]`, built and consumed with `Iter` |

Keys and members compare by `==`, so a record or an array can be a key or a member. Maps and sets
keep insertion order: a `for` loop, `Map.keys`, and `Set.values` visit keys and members in the order
they were inserted.

## How they behave

A function that takes a collection takes it first, so a call goes through `.` or down a pipe:

```ant
let prices = Map{ tea: 3, cake: 5, soup: 7 }
prices.filter({ x.value > 4 }).keys() // ["cake", "soup"]
prices | Map.values() | Array.sum()   // 15
```

`Array.sum`, `Array.min`, and `Array.max` return `null` when the array is empty, so `??` supplies a
default:

<!-- test: before
let prices = Map{ tea: 3, cake: 5, soup: 7 }
-->

```ant
prices.filter({ x.value > 10 }).values().sum() ?? 0 // 0
```

Reading a missing key gives `null`, and so does `Array.get` out of range, but indexing an array out
of range panics. Use `??` for a default:

<!-- test: before
let scores = [72, 95, 88]
let ages = Map{ ada: 36, linus: 54 }
-->

```ant
ages["bob"]      // null
ages["bob"] ?? 0 // 0
scores.get(5)    // null
scores[5]        // panic: index 5 out of range for an array of length 3
```

A `for` loop over a map yields each [Entry](../Map/types/Entry.md), over a set each member, and
[Array.entries](../Array/entries.md) pairs each element of an array with its index:

```ant
let ages = Map{ ada: 36, linus: 54 }
for { key, value } in ages {
  println("${key} ${value}")
}
for tag in Set.of(["red", "blue"]) {
  println(tag)
}
for e in ["a", "b"].entries() {
  println("${e.index}: ${e.value}")
}
```

```text
ada 36
linus 54
red
blue
0: a
1: b
```

Indexed assignment (`xs[i] = v`, `m[k] = v`) and the functions under
[Changing a collection in place](#changing-a-collection-in-place) change a collection. Every other
function leaves it unchanged:

```ant
let xs = [3, 1, 2]
xs.sort() // [1, 2, 3]
xs        // [3, 1, 2]
xs.push(4)
xs // [3, 1, 2, 4]
```

A collection is a reference, so copying a variable shares it. `==` compares contents; `same` asks
whether two names hold the same collection. Two maps or two sets are equal in any order, two arrays
only in the same order:

```ant
let a = Map{ x: 1, y: 2 }
let b = Map{ y: 2, x: 1 }
a == b                           // true
Map.same(a, b)                   // false
Set.of([1, 2]) == Set.of([2, 1]) // true
[1, 2] == [2, 1]                 // false
```

A sequence runs nothing until a `for` loop or `Iter.toArray` consumes it, and it stops once the
consumer has what it needs. On an array `.map` is the eager `Array.map`; on a sequence it is the
lazy `Iter.map`:

```ant
var calls = 0
let squares = Iter.range(0, 1000000).map({ calls += 1; x * x })
calls                                   // 0
squares | Iter.take(3) | Iter.toArray() // [0, 1, 4]
calls                                   // 3
```

## Functions by job

### Building

- [Array literals](../../language/literals/arrays.md): `[1, 2, 3]`, with `...` to spread an array
- [Map literals](../../language/literals/maps.md): `Map{ key: value }`
- [Array.range](../Array/range.md): the `int`s from `from` up to `to`
- [Array.filled](../Array/filled.md): `n` copies of one value
- [Set.of](../Set/of.md): a set of the first of each value of an array, in order
- [Map.fromEntries](../Map/fromEntries.md): a map from an array of entries; later entries win
- [Iter.range](../Iter/range.md): a sequence of the `int`s from `from` up to `to`
- [Iter.fromArray](../Iter/fromArray.md): a sequence of the elements of an array

### Size and membership

- [len](../prelude/len.md): the length of an array or the size of a map or set
- [Array.length](../Array/length.md), [Map.size](../Map/size.md), [Set.size](../Set/size.md): the
  number of elements, entries, or members
- [Array.isEmpty](../Array/isEmpty.md), [Map.isEmpty](../Map/isEmpty.md),
  [Set.isEmpty](../Set/isEmpty.md): whether there are none
- [Array.contains](../Array/contains.md): whether a value is an element
- [Map.has](../Map/has.md): whether there is an entry for a key
- [Set.has](../Set/has.md): whether a value is a member

### Reading

- [Array.get](../Array/get.md): element `i`, or `null` when out of range
- [Array.first](../Array/first.md), [Array.last](../Array/last.md): the first or last element, or
  `null`
- [Map.get](../Map/get.md): the value for a key, or `null`; the same as `m[k]`

### Changing a collection in place

- [Array.push](../Array/push.md): append a value
- [Array.extend](../Array/extend.md): append the elements of another array
- [Array.pop](../Array/pop.md): remove and return the last element
- [Array.insert](../Array/insert.md): put a value at index `i`
- [Array.removeAt](../Array/removeAt.md): remove and return element `i`
- [Map.set](../Map/set.md): store a value under a key; the same as `m[k] = v`
- [Map.remove](../Map/remove.md): remove a key and return its value, or `null`
- [Set.add](../Set/add.md): add a member; `true` when it was not already there
- [Set.remove](../Set/remove.md): remove a member; `true` when it was there

### Transforming

- [Array.map](../Array/map.md), [Map.map](../Map/map.md): compute a new element or value from each
- [Array.flatMap](../Array/flatMap.md): map to arrays and concatenate them
- [Array.filter](../Array/filter.md), [Map.filter](../Map/filter.md): keep what a predicate accepts
- [Array.reverse](../Array/reverse.md): the elements in reverse order
- [Array.slice](../Array/slice.md): the elements from `from` up to `to`
- [Array.concat](../Array/concat.md): two arrays, one after the other
- [Array.flatten](../Array/flatten.md): an array of arrays, concatenated
- [Array.unique](../Array/unique.md), [Array.uniqueBy](../Array/uniqueBy.md): the first of each
  equal value or key
- [Map.merge](../Map/merge.md): the entries of two maps; the second map's entries win

### Searching and testing

- [Array.find](../Array/find.md), [Array.findLast](../Array/findLast.md): the first or last element
  a predicate accepts
- [Array.findIndex](../Array/findIndex.md), [Array.findLastIndex](../Array/findLastIndex.md): its
  index
- [Array.indexOf](../Array/indexOf.md), [Array.lastIndexOf](../Array/lastIndexOf.md): the first or
  last index of a value
- [Array.some](../Array/some.md): whether a predicate accepts any element
- [Array.every](../Array/every.md): whether it accepts every element
- [Array.countWhere](../Array/countWhere.md): how many elements it accepts

### Ordering

- [Array.sort](../Array/sort.md): ascending, stable
- [Array.sortBy](../Array/sortBy.md): ascending by a key
- [Array.sortWith](../Array/sortWith.md): by a comparison function
- [Array.min](../Array/min.md), [Array.max](../Array/max.md): the least or greatest element
- [Array.minBy](../Array/minBy.md), [Array.maxBy](../Array/maxBy.md): the element with the least or
  greatest key

### Folding and grouping

- [Array.reduce](../Array/reduce.md): fold the elements into one value
- [Array.sum](../Array/sum.md): the sum of numbers, or `null` when empty
- [Array.chunk](../Array/chunk.md): split into arrays of `n` elements
- [Array.groupBy](../Array/groupBy.md): a map from each key to its elements

### Converting to arrays

- [Map.keys](../Map/keys.md), [Map.values](../Map/values.md): the keys or the values, in insertion
  order
- [Map.entries](../Map/entries.md): the entries, as [Entry](../Map/types/Entry.md) records
- [Set.values](../Set/values.md): the members, in insertion order
- [Array.entries](../Array/entries.md): each element with its index, as
  [Indexed](../Array/types/Indexed.md) records
- [Iter.toArray](../Iter/toArray.md): run a sequence and collect its elements

### Combining sets

- [Set.union](../Set/union.md): the members of `a`, then those of `b` that `a` lacks
- [Set.intersect](../Set/intersect.md): the members of `a` that `b` has
- [Set.difference](../Set/difference.md): the members of `a` that `b` lacks
- [Set.isSubset](../Set/isSubset.md): whether every member of `a` is a member of `b`

### Lazy sequences

- [Seq](../Iter/types/Seq.md): a sequence, `func(func(T): bool): void`
- [Iter.map](../Iter/map.md): apply a function to each element, as the sequence runs
- [Iter.filter](../Iter/filter.md): the elements a predicate accepts
- [Iter.take](../Iter/take.md): at most the first `n` elements, then stop the source

### Identity

- [Array.same](../Array/same.md), [Map.same](../Map/same.md), [Set.same](../Set/same.md): whether
  two names hold the same collection, not only an equal one

## Applications

An empty `[]`, `Map{}`, or `Set.of([])` takes its element type from the declared type, so annotate
the name:

```ant
let seen = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
```

Count words with a map. `??` supplies `0` for a word not seen yet:

```ant
let text = "the ant and the bee and the wasp"
let counts: Map[string, int] = Map{}
for word in text.split(" ") {
  counts[word] = (counts[word] ?? 0) + 1
}
counts // Map{ "the": 3, "ant": 1, "and": 2, "bee": 1, "wasp": 1 }
```

Rank the counts. `Map.entries` turns the map into an array, which sorts:

<!-- test: before
let counts = Map{ the: 3, ant: 1, and: 2, bee: 1, wasp: 1 }
-->

```ant
counts
  | Map.entries()
  | Array.sortBy({ -x.value })
  | Array.map({ "${x.key} ${x.value}" }) // ["the 3", "and 2", "ant 1", "bee 1", "wasp 1"]
```

Print each line the first time it appears. `Set.add` says whether the line was new:

```ant
let lines = ["b", "a", "b", "c", "a"]
let seen: Set[string] = Set.of([])
for line in lines {
  if seen.add(line) {
    println(line)
  }
}
```

```text
b
a
c
```

Find the tags two posts share, and the ones only the first has:

```ant
let first = Set.of(["go", "rust", "ant"])
let second = Set.of(["ant", "zig", "go"])
first.intersect(second)  // Set.of(["go", "ant"])
first.difference(second) // Set.of(["rust"])
```

Group records by a field, then summarize each group. `Array.groupBy` returns a map, so `.map` on it
is `Map.map`:

```ant
type Order = { customer: string, total: int }
let orders = [
  Order{ customer: "Ann", total: 30 },
  Order{ customer: "Bob", total: 12 },
  Order{ customer: "Ann", total: 8 },
]
let byCustomer = orders.groupBy({ x.customer })
byCustomer.map({ x.value.length() }) // Map{ "Ann": 2, "Bob": 1 }
byCustomer.keys()                    // ["Ann", "Bob"]
```

## See also

[Array](../Array/index.md), [Map](../Map/index.md), [Set](../Set/index.md),
[Iter](../Iter/index.md), [Collections and sequences](../../guide/collections.md),
[Array type](../../language/types/Array.md), [Map type](../../language/types/Map.md),
[Set type](../../language/types/Set.md), [for](../../language/keywords/for.md), [Text](text.md)
