<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Set.has

```ant signature
func Set.has[T](s: Set[T], x: T): bool
```

Set.has reports whether x is a member of s.

<!-- end generated -->

## Details

- Members compare by [`==`](../../language/operators/equality.md), so a record or an array matches
  by contents.
- To test and insert in one step, use [Set.add](add.md): it returns `true` when the value was not a
  member.

## Basic examples

```ant
var colors = Set.of(["red", "green", "blue"])
Set.has(colors, "green")  // true
Set.has(colors, "purple") // false
```

Call it through `.` on any set:

```ant
var primes = Set.of([2, 3, 5, 7])
primes.has(5) // true
primes.has(9) // false
```

## Scope

On an empty set, `has` is `false` for every value:

```ant
var none: Set[string] = Set.of([])
none.has("") // false
```

Records and arrays compare by value:

```ant
type Point = { x: int, y: int }
var visited = Set.of([Point{ x: 0, y: 0 }, Point{ x: 1, y: 0 }])
var pairs = Set.of([[1, 2], [3, 4]])
visited.has(Point{ x: 1, y: 0 }) // true
visited.has(Point{ x: 0, y: 1 }) // false
pairs.has([3, 4])                // true
pairs.has([4, 3])                // false
```

A tagged union member matches by tag and fields, so equal fields under another tag do not match:

```ant
type Circle = { size: float64 }
type Square = { size: float64 }
type Shape = Circle | Square
var shapes: Set[Shape] = Set.of([Circle{ size: 1.0 }])
shapes.has(Circle{ size: 1.0 }) // true
shapes.has(Square{ size: 1.0 }) // false
```

A set may hold `null` when its element type includes it:

```ant
var answers: Set[bool | null] = Set.of([true, null])
answers.has(null)  // true
answers.has(false) // false
```

Strings compare by code points, without normalizing:

```ant
var words = Set.of(["café"])
words.has("café")        // true
words.has("cafe\u{301}") // false
```

On a set that may itself be `null`, `?.` gives `null`; `??` turns that into a `bool`:

```ant
var tags: Set[string] | null = null
tags?.has("a")          // null
tags?.has("a") ?? false // false
```

In a pipeline, the set is the first argument:

```ant
Set.of(["a", "b"]) | Set.has("b") // true
```

## Applications

Drop the stop words from a sentence:

```ant
var stop = Set.of(["a", "an", "the", "of"])
"the end of a story".split(" ").filter({ !stop.has(x) }) // ["end", "story"]
```

Walk a graph breadth first, visiting each node once even when edges form a cycle:

```ant
var edges = Map{ a: ["b", "c"], b: ["c"], c: ["a", "d"], d: [] }
var seen = Set.of(["a"])
var queue = ["a"]
var i = 0
while i < queue.length() {
  var node = queue[i]
  i += 1
  println(node)
  for next in edges[node] ?? [] {
    if !seen.has(next) {
      seen.add(next)
      queue.push(next)
    }
  }
}
```

```text
a
b
c
d
```

Check a value against a list of allowed ones:

```ant
var methods = Set.of(["GET", "HEAD", "POST"])

func allowed(method: string): bool {
  methods.has(method.toUpper())
}

println(allowed("get"))
println(allowed("DELETE"))
```

```text
true
false
```

## Properties and relations

`Set.has` agrees with `Array.contains` on the members, and with the array the set was built from:

```ant
var xs = [3, 1, 3, 2]
var s = Set.of(xs)
for x in [0, 1, 2, 3, 4] {
  assert(s.has(x) == s.values().contains(x))
  assert(s.has(x) == xs.contains(x))
}
```

`Set.add` puts a member in and `Set.remove` takes it out:

```ant
var s: Set[string] = Set.of([])
s.add("x")
assert(s.has("x"))
s.remove("x")
assert(!s.has("x"))
assert(s.isEmpty())
```

Membership in a union, an intersection, or a difference follows from membership in the two sets:

```ant
var a = Set.of([1, 2, 3])
var b = Set.of([2, 3, 4])
for x in [1, 2, 3, 4, 5] {
  assert(a.union(b).has(x) == (a.has(x) || b.has(x)))
  assert(a.intersect(b).has(x) == (a.has(x) && b.has(x)))
  assert(a.difference(b).has(x) == (a.has(x) && !b.has(x)))
}
```

## Possible issues

The value must have the set's element type:

```ant
var ids = Set.of([1, 2, 3])
println(ids.has("1")) // error: cannot use string where int is expected
```

`NaN` is not `==` to itself, so a `NaN` member is never found. `0.0` and `-0.0` are `==`, so they
find each other:

```ant
Set.of([Math.NaN]).has(Math.NaN) // false
Set.of([0.0]).has(-0.0)          // true
```

A member is hashed when it is added. Mutating an array or record after you add it is a program
error, and `Set.has` may no longer find it. Add a copy, or do not change the member.

## See also

[Set.of](of.md) · [Set.add](add.md) · [Set.remove](remove.md) · [Set.values](values.md) ·
[Set.isSubset](isSubset.md) · [Map.has](../Map/has.md) · [Array.contains](../Array/contains.md)
