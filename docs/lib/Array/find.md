<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.find

```ant signature
func Array.find[T](xs: Array[T], f: func(T): bool): T | null
```

Array.find returns the first element of xs that f accepts, or null when there is none.

<!-- end generated -->

## Details

- Returns `null` when `f` accepts no element, so an empty array gives `null`.
- Calls `f` from the first element and stops at the first element `f` accepts.
- When `T` includes `null`, a `null` result does not tell a missing element from a found `null`; use
  [Array.findIndex](findIndex.md).
- The result is the element in `xs`, not a copy.

## Basic examples

Find the first number above 2:

```ant
Array.find([1, 2, 3, 4], { x > 2 }) // 3
```

When no element matches, the result is `null`:

```ant
Array.find([1, 2, 3], { x > 5 }) // null
```

Find a record by a field, and read a field of the result with `?.`:

```ant
type User = { id: int, name: string }
var users = [User{ id: 1, name: "Ann" }, User{ id: 7, name: "Bob" }]
users.find({ x.id == 7 })?.name // "Bob"
users.find({ x.id == 9 })?.name // null
```

## Scope

An empty array has no element to find. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
empty.find({ true }) // null
```

Only the first match counts, even when later elements match too:

```ant
["fig", "plum", "pear", "lime"].find({ String.length(x) == 4 }) // "plum"
```

`f` runs from the first element and stops at the first match. A callback that records each call
shows it never sees `12`:

```ant
var seen: Array[int] = []
var found = [3, 8, 5, 12].find({
  seen.push(x)
  x > 4
})
println(found)
println(seen)
```

```text
8
[3, 8]
```

Pass a named function instead of a lambda:

```ant
func isEven(n: int): bool {
  n % 2 == 0
}

[3, 5, 8, 10].find(isEven) // 8
```

A lambda can use the local variables around it:

```ant
var limit = 10
var prices = [4, 12, 7, 30]
prices.find({ x > limit }) // 12
```

Search the graphemes of a string by splitting it with `String.graphemes` first:

```ant
String.graphemes("🐜🐝🐞") | Array.find({ x != "🐜" }) // "🐝"
```

Find a member of a union by its type with `is`. The result type stays `Shape | null`, not
`Square | null`; test it with `is` again to read `side`:

```ant
type Circle = { radius: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ radius: 1.0 }, Square{ side: 2.0 }, Square{ side: 3.0 }]
shapes.find({ x is Square }) // Square{ side: 2.0 }
var s = shapes.find({ x is Square })
if s is Square { s.side } // 2.0
```

Find in a nested array:

```ant
[[1, 2], [], [3]].find({ Array.isEmpty(x) }) // []
```

Call it through `.` or in a pipe:

```ant
[5, 10, 15].find({ x % 2 == 0 })         // 10
[5, 10, 15] | Array.find({ x % 3 == 0 }) // 15
```

## Applications

Look up a record and fall back to a default with `??`:

```ant
type User = { id: int, name: string }
var users = [User{ id: 1, name: "Ann" }, User{ id: 7, name: "Bob" }]

func nameOf(id: int): string {
  users.find({ x.id == id })?.name ?? "anonymous"
}

println(nameOf(7))
println(nameOf(9))
```

```text
Bob
anonymous
```

Narrow the result with `!= null` to work with the whole record:

```ant
type Item = { name: string, stock: int }
var items = [Item{ name: "nut", stock: 0 }, Item{ name: "bolt", stock: 4 }]

var item = items.find({ x.stock > 0 })
if item != null {
  println("${item.name}: ${item.stock} in stock")
}
```

```text
bolt: 4 in stock
```

Find the first line that is not blank:

```ant
var text = "\n  \nTitle\nBody"
String.split(text, "\n").find({ String.trim(x) != "" }) // "Title"
```

Return early from a function when nothing matches, with `?? return`:

```ant
type Route = { path: string, handler: string }

func handle(routes: Array[Route], path: string): string {
  var route = routes.find({ x.path == path }) ?? return "404 not found"
  "200 ${route.handler}"
}

var routes = [Route{ path: "/", handler: "home" }, Route{ path: "/about", handler: "about" }]
println(handle(routes, "/about"))
println(handle(routes, "/missing"))
```

```text
200 about
404 not found
```

## Properties and relations

`Array.find` gives the first element of `Array.filter`, and the element at `Array.findIndex`:

```ant
var xs = [4, 9, 16, 25]
var f: func(int): bool = { x > 5 }
assert(xs.find(f) == xs.filter(f).first())
var i = xs.findIndex(f)
assert(i != null && xs.find(f) == xs[i])
```

It finds something exactly when `Array.some` holds, for an element type without `null`:

```ant
var xs = [1, 3, 5]
assert((xs.find({ x > 4 }) != null) == xs.some({ x > 4 }))
assert((xs.find({ x > 9 }) != null) == xs.some({ x > 9 }))
```

On a reversed array, it gives what `Array.findLast` gives on the original:

```ant
var xs = [2, 7, 4, 9, 6]
assert(xs.reverse().find({ x > 5 }) == xs.findLast({ x > 5 }))
```

The element it returns is the one in the array, so a change to it shows in the array:

```ant
type Item = { name: string, stock: int }
var items = [Item{ name: "nut", stock: 0 }]
var item = items.find({ x.name == "nut" })
if item != null {
  item.stock = 10
}
assert(items[0].stock == 10)
```

## Possible issues

The result may be `null`, so reading a field of it without narrowing is a compile error:

<!-- test: error has no field name -->

```ant
type User = { id: int, name: string }
var users = [User{ id: 1, name: "Ann" }]
println(users.find({ x.id == 1 }).name)
```

Use `?.`, `??`, or a test against `null` first, as in the examples above.

When the element type includes `null`, a found `null` and no match look the same. Use
[`Array.findIndex`](findIndex.md) to tell them apart:

```ant
var scores: Array[int | null] = [90, null, 75]
scores.find({ x == null })      // null
scores.find({ x == 100 })       // null
scores.findIndex({ x == null }) // 1
scores.findIndex({ x == 100 })  // null
```

`f` must return a `bool`. Ant has no truthy values, so a number is not a test:

<!-- test: error cannot use int where bool is expected -->

```ant
var xs = [1, 2, 3]
println(xs.find({ x - 2 }))
```

`Array.find` takes a predicate, not a value. To look for a value by `==`, use
[`Array.contains`](contains.md) or [`Array.indexOf`](indexOf.md), or compare in the lambda:

```ant
[1, 2, 3].find({ x == 2 }) // 2
[1, 2, 3].contains(2)      // true
[1, 2, 3].indexOf(2)       // 1
```

## See also

[Array.findIndex](findIndex.md) · [Array.findLast](findLast.md) ·
[Array.findLastIndex](findLastIndex.md) · [Array.filter](filter.md) · [Array.some](some.md) ·
[Array.contains](contains.md) · [Array.indexOf](indexOf.md) · [Array.first](first.md) ·
[?. operator](../../language/operators/optional-chaining.md) ·
[?? operator](../../language/operators/null-coalescing.md)
