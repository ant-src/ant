<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.add

```ant signature
func Set.add[T](s: Set[T], x: T): bool
```

Set.add adds x to s and reports whether it was not already a member. It mutates s. Adding an
existing member changes nothing, so it keeps its position.

<!-- end generated -->

## Details

- Every name that shares the set sees the new member. Membership is by `==`.
- A new member goes to the end of the insertion order, so a member removed and added again moves to
  the end.
- A `for` loop over `s` visits members added during the loop.
- A member is hashed when it is added. Mutating an array or a record after adding it is a program
  error: the set may no longer find it.

## Basic examples

Add a member, and learn whether it was new:

```ant
var s = Set.of([1, 2])
Set.add(s, 3) // true
Set.add(s, 2) // false
s             // Set.of([1, 2, 3])
```

Start from an empty set, whose element type comes from the annotation:

```ant
var tags: Set[string] = Set.of([])
Set.add(tags, "red")  // true
Set.add(tags, "blue") // true
Set.add(tags, "red")  // false
tags                  // Set.of(["red", "blue"])
```

## Scope

Records and arrays are members by value, so an equal one is not added again:

```ant
type Point = { x: int, y: int }
var points: Set[Point] = Set.of([])
Set.add(points, Point{ x: 1, y: 2 }) // true
Set.add(points, Point{ x: 1, y: 2 }) // false
var rows = Set.of([[1, 2]])
Set.add(rows, [1, 2]) // false
Set.add(rows, [2, 1]) // true
```

`null` is a member like any other in a set that allows it:

```ant
var maybe: Set[int | null] = Set.of([1])
Set.add(maybe, null) // true
Set.add(maybe, null) // false
maybe                // Set.of([1, null])
```

Floats compare as `==` compares them, so `-0.0` is the member `0.0`:

```ant
var zeros = Set.of([0.0])
Set.add(zeros, -0.0) // false
```

A `let` set can be added to, since the binding stays the same:

```ant
let s = Set.of(["a"])
Set.add(s, "b") // true
s               // Set.of(["a", "b"])
```

## Applications

Keep the first of each line, with no separate membership test:

```ant
var seen: Set[string] = Set.of([])
for line in ["b", "a", "b", "c", "a"] {
  if Set.add(seen, line) {
    println(line)
  }
}
```

```text
b
a
c
```

Find the values that repeat, each reported once:

```ant
func repeated(words: Array[string]): Array[string] {
  var seen: Set[string] = Set.of([])
  var twice: Set[string] = Set.of([])
  for w in words {
    if !Set.add(seen, w) {
      Set.add(twice, w)
    }
  }
  Set.values(twice)
}

repeated(["to", "be", "or", "not", "to", "be", "to"]) // ["to", "be"]
```

Collect every node reachable from a start. The `for` loop visits the members added while it runs. A
node already in the set is not added again, so a cycle ends. `edges[node] ?? []` is an empty array
for a node with no edges:

```ant
func reachable(edges: Map[string, Array[string]], start: string): Set[string] {
  var found = Set.of([start])
  for node in found {
    for next in edges[node] ?? [] {
      Set.add(found, next)
    }
  }
  found
}

var edges = Map{ a: ["b", "c"], b: ["d"], c: ["a"], d: ["b"], e: ["a"] }
reachable(edges, "a") // Set.of(["a", "b", "c", "d"])
reachable(edges, "e") // Set.of(["e", "a", "b", "c", "d"])
```

Count the distinct values of a field in a list of records:

```ant
type Order = { customer: string, total: int }
var orders = [
  Order{ customer: "ada", total: 30 },
  Order{ customer: "linus", total: 12 },
  Order{ customer: "ada", total: 7 },
]
var customers: Set[string] = Set.of([])
for o in orders {
  Set.add(customers, o.customer)
}
Set.size(customers) // 2
```

## Properties and relations

The result is the opposite of [Set.has](has.md) before the call, and `x` is a member after it:

```ant
var s = Set.of([1, 2])
var had = Set.has(s, 5)
assert(Set.add(s, 5) == !had)
assert(Set.has(s, 5))
```

The size grows by one exactly when the result is `true`:

```ant
var s = Set.of(["a"])
var before = Set.size(s)
var added = Set.add(s, "b")
assert(Set.size(s) == before + (if added { 1 } else { 0 }))
assert(!Set.add(s, "b"))
assert(Set.size(s) == 2)
```

A new member goes last in [Set.values](values.md):

```ant
var s = Set.of([3, 1])
Set.add(s, 2)
assert(Array.last(Set.values(s)) == 2)
```

It does in place what [Set.union](union.md) with a one-member set does in a new set:

```ant
var s = Set.of([1, 2])
var u = Set.union(s, Set.of([3]))
var added = Set.add(s, 3)
assert(added)
assert(s == u)
assert(Set.values(s) == Set.values(u))
```

[Set.remove](remove.md) undoes the add of a new member:

```ant
var s = Set.of([1, 2])
var copy = Set.of(Set.values(s))
Set.add(s, 3)
assert(Set.remove(s, 3))
assert(s == copy)
```

Every other name for the set sees the add:

```ant
var s = Set.of([1])
var alias = s
Set.add(s, 2) // true
assert(Set.same(alias, s))
alias // Set.of([1, 2])
```

## Possible issues

`Set.add` returns a `bool`, not the set, so its result cannot be assigned back:

```ant
var s = Set.of([1])
s = Set.add(s, 2) // error: cannot use bool where Set[int] is expected
```

To get a new set and keep the old one, use [Set.union](union.md):

```ant
var s = Set.of([1])
var t = Set.union(s, Set.of([2]))
s // Set.of([1])
t // Set.of([1, 2])
```

`Set.of([])` has no element type to infer, so annotate the variable:

```ant
var s = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
Set.add(s, 1)
```

`x` must have the element type:

```ant
var s = Set.of([1])
Set.add(s, "2") // error: cannot use string where int is expected
```

A set has no indexing, so add with `Set.add`, not an assignment as for a map:

```ant
var s: Set[int] = Set.of([])
s[1] = true // error: Set[int] cannot be indexed
```

NaN is not `==` to itself, so each add of it adds a new member:

```ant
var s: Set[float64] = Set.of([])
let nan = 0.0 / 0.0
Set.add(s, nan) // true
Set.add(s, nan) // true
Set.size(s)     // 2
```

Adding an existing member does not move it. To move a member to the end, remove it first:

```ant
var s = Set.of(["a", "b", "c"])
Set.add(s, "a")    // false
s                  // Set.of(["a", "b", "c"])
Set.remove(s, "a") // true
Set.add(s, "a")    // true
s                  // Set.of(["b", "c", "a"])
```

Adding to one set from several goroutines without a lock is a data race. Guard the set with a
`Mutex`:

```ant
var seen: Set[int] = Set.of([])
var lock = Mutex.new()
var wg = WaitGroup.new()
for n in [1, 2, 3, 2, 1] {
  WaitGroup.add(wg, 1)
  go {
    defer { WaitGroup.done(wg) }
    Mutex.withLock(lock, { Set.add(seen, n) })
  }
}
WaitGroup.wait(wg)
println(Set.size(seen))
```

```text
3
```

## See also

[Set.has](has.md), [Set.remove](remove.md), [Set.of](of.md), [Set.union](union.md),
[Set.values](values.md), [Map.set](../Map/set.md), [Array.push](../Array/push.md)
