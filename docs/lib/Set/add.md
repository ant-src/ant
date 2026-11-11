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
let s = Set.of([1, 2])
Set.add(s, 3) // true
Set.add(s, 2) // false
s             // Set.of([1, 2, 3])
```

Start from an empty set, whose element type comes from the annotation:

```ant
let tags: Set[string] = Set.of([])
Set.add(tags, "red")  // true
Set.add(tags, "blue") // true
Set.add(tags, "red")  // false
tags                  // Set.of(["red", "blue"])
```

## Scope

Records and arrays are members by value, so an equal one is not added again:

```ant
type Point = { x: int, y: int }
let points: Set[Point] = Set.of([])
points.add(Point{ x: 1, y: 2 }) // true
points.add(Point{ x: 1, y: 2 }) // false
let rows = Set.of([[1, 2]])
rows.add([1, 2]) // false
rows.add([2, 1]) // true
```

`null` is a member like any other in a set that allows it:

```ant
let maybe: Set[int | null] = Set.of([1])
maybe.add(null) // true
maybe.add(null) // false
maybe           // Set.of([1, null])
```

Floats compare as `==` compares them, so `-0.0` is the member `0.0`:

```ant
let zeros = Set.of([0.0])
zeros.add(-0.0) // false
```

A `let` set can be added to, since the binding stays the same:

```ant
let s = Set.of(["a"])
s.add("b") // true
s          // Set.of(["a", "b"])
```

## Applications

Keep the first of each line, with no separate membership test:

```ant
let seen: Set[string] = Set.of([])
for line in ["b", "a", "b", "c", "a"] {
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

Find the values that repeat, each reported once:

```ant
func repeated(words: Array[string]): Array[string] {
  let seen: Set[string] = Set.of([])
  let twice: Set[string] = Set.of([])
  for w in words {
    if !seen.add(w) {
      twice.add(w)
    }
  }
  twice.values()
}

repeated(["to", "be", "or", "not", "to", "be", "to"]) // ["to", "be"]
```

Collect every node reachable from a start. The `for` loop visits the members added while it runs. A
node already in the set is not added again, so a cycle ends. `edges[node] ?? []` is an empty array
for a node with no edges:

```ant
func reachable(edges: Map[string, Array[string]], start: string): Set[string] {
  let found = Set.of([start])
  for node in found {
    for next in edges[node] ?? [] {
      found.add(next)
    }
  }
  found
}

let edges = Map{ a: ["b", "c"], b: ["d"], c: ["a"], d: ["b"], e: ["a"] }
reachable(edges, "a") // Set.of(["a", "b", "c", "d"])
reachable(edges, "e") // Set.of(["e", "a", "b", "c", "d"])
```

Count the distinct values of a field in a list of records:

```ant
type Order = { customer: string, total: int }
let orders = [
  Order{ customer: "ada", total: 30 },
  Order{ customer: "linus", total: 12 },
  Order{ customer: "ada", total: 7 },
]
let customers: Set[string] = Set.of([])
for o in orders {
  customers.add(o.customer)
}
customers.size() // 2
```

## Properties and relations

The result is the opposite of [Set.has](has.md) before the call, and `x` is a member after it:

```ant
let s = Set.of([1, 2])
let had = s.has(5)
assert(s.add(5) == !had)
assert(s.has(5))
```

The size grows by one exactly when the result is `true`:

```ant
let s = Set.of(["a"])
let before = s.size()
let added = s.add("b")
assert(s.size() == before + (if added { 1 } else { 0 }))
assert(!s.add("b"))
assert(s.size() == 2)
```

A new member goes last in [Set.values](values.md):

```ant
let s = Set.of([3, 1])
s.add(2)
assert(s.values().last() == 2)
```

It does in place what [Set.union](union.md) with a one-member set does in a new set:

```ant
let s = Set.of([1, 2])
let u = s.union(Set.of([3]))
let added = s.add(3)
assert(added)
assert(s == u)
assert(s.values() == u.values())
```

[Set.remove](remove.md) undoes the add of a new member:

```ant
let s = Set.of([1, 2])
let copy = Set.of(s.values())
s.add(3)
assert(s.remove(3))
assert(s == copy)
```

Every other name for the set sees the add:

```ant
let s = Set.of([1])
let alias = s
s.add(2) // true
assert(Set.same(alias, s))
alias // Set.of([1, 2])
```

## Possible issues

`Set.add` returns a `bool`, not the set, so its result cannot be assigned back:

```ant
var s = Set.of([1])
s = s.add(2) // error: cannot use bool where Set[int] is expected
```

To get a new set and keep the old one, use [Set.union](union.md):

```ant
let s = Set.of([1])
let t = s.union(Set.of([2]))
s // Set.of([1])
t // Set.of([1, 2])
```

`Set.of([])` has no element type to infer, so annotate the variable:

```ant
let s = Set.of([]) // error: cannot infer the element type of an empty array; annotate it
s.add(1)
```

`x` must have the element type:

```ant
let s = Set.of([1])
s.add("2") // error: cannot use string where int is expected
```

A set has no indexing, so add with `Set.add`, not an assignment as for a map:

```ant
let s: Set[int] = Set.of([])
s[1] = true // error: Set[int] cannot be indexed
```

NaN is not `==` to itself, so each add of it adds a new member:

```ant
let s: Set[float64] = Set.of([])
let nan = 0.0 / 0.0
s.add(nan) // true
s.add(nan) // true
s.size()   // 2
```

Adding an existing member does not move it. To move a member to the end, remove it first:

```ant
let s = Set.of(["a", "b", "c"])
s.add("a")    // false
s             // Set.of(["a", "b", "c"])
s.remove("a") // true
s.add("a")    // true
s             // Set.of(["b", "c", "a"])
```

Adding to one set from several goroutines without a lock is a data race. Guard the set with a
`Mutex`:

```ant
let seen: Set[int] = Set.of([])
let lock = Mutex.new()
let wg = WaitGroup.new()
for n in [1, 2, 3, 2, 1] {
  wg.add(1)
  go {
    defer { wg.done() }
    lock.withLock({ seen.add(n) })
  }
}
wg.wait()
println(seen.size())
```

```text
3
```

## See also

[Set.has](has.md), [Set.remove](remove.md), [Set.of](of.md), [Set.union](union.md),
[Set.values](values.md), [Map.set](../Map/set.md), [Array.push](../Array/push.md)
