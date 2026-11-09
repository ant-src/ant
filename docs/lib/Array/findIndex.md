<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.findIndex

```ant signature
func Array.findIndex[T](xs: Array[T], f: func(T): bool): int | null
```

Array.findIndex returns the index of the first element of xs that f accepts, or null when there is
none.

<!-- end generated -->

## Details

- Returns `null` when `f` accepts no element, so an empty array gives `null`.
- Calls `f` from the first element and stops at the first element `f` accepts.

## Basic examples

```ant
Array.findIndex([4, 9, 16, 25], { x > 10 }) // 2
Array.findIndex([4, 9, 16, 25], { x > 30 }) // null
```

Call it through `.` on any array:

```ant
var names = ["Ann", "Bob", "Cid"]
names.findIndex({ String.startsWith(x, "B") }) // 1
```

## Scope

The callback runs from the first element and stops at the first match, so a side effect shows which
elements it tested:

```ant
var seen: Array[int] = []
[4, 9, 16, 25].findIndex({ seen.push(x); x > 5 }) // 1
seen                                              // [4, 9]
```

With no match, it tests every element:

```ant
var seen: Array[int] = []
[4, 9, 16].findIndex({ seen.push(x); x > 30 }) // null
seen                                           // [4, 9, 16]
```

An empty array has no index to return. Annotate it so that `T` is known:

```ant
var empty: Array[string] = []
empty.findIndex({ x == "" }) // null
```

Only the first match counts, even when several elements match:

```ant
[3, 8, 5, 8].findIndex({ x == 8 }) // 1
```

The index counts elements, so in an array of graphemes it counts characters as the reader sees them:

```ant
var chars = String.graphemes("ant 🐜 hill")
chars.findIndex({ x == "🐜" }) // 4
```

Test a field of a record:

```ant
type User = { name: string, age: int }
var users = [User{ name: "Ann", age: 15 }, User{ name: "Bob", age: 34 }]
users.findIndex({ x.age >= 18 }) // 1
```

Test the member of a union with `is`:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Circle{ r: 1.0 }, Square{ side: 2.0 }, Square{ side: 3.0 }]
shapes.findIndex({ x is Square }) // 1
```

An array whose element type includes `null` can look for the first `null`:

```ant
var scores: Array[int | null] = [90, null, 75]
scores.findIndex({ x == null }) // 1
```

Pass a named function, or an anonymous `func` with a typed parameter:

```ant
func isNegative(n: int): bool {
  n < 0
}

[3, -1, 4].findIndex(isNegative)                        // 1
[3, -1, 4].findIndex(func(n: int): bool { n % 2 == 0 }) // 2
```

A lambda captures the variables around it:

```ant
var limit = 10
[4, 12, 7].findIndex({ x > limit }) // 1
```

In a pipeline, the array is the first argument:

```ant
"a,b,,c" | String.split(",") | Array.findIndex({ x == "" }) // 2
```

## Applications

Replace the first element that matches, and leave the array alone when none does:

```ant
type Task = { title: string, done: bool }
var tasks = [Task{ title: "write", done: true }, Task{ title: "test", done: false }]
var i = tasks.findIndex({ !x.done })
if i != null {
  tasks[i] = Task{ ...tasks[i], done: true }
}
println(tasks)
```

```text
[Task{ title: "write", done: true }, Task{ title: "test", done: true }]
```

Remove the first element that matches with [`Array.removeAt`](removeAt.md):

```ant
var queue = ["low", "high", "low", "high"]
var i = queue.findIndex({ x == "high" })
if i != null {
  println(queue.removeAt(i))
}
println(queue)
```

```text
high
["low", "low", "high"]
```

Split the lines of a message at the first blank line, into a header and a body:

```ant
var lines = ["From: ann", "To: bob", "", "Hello,", "Bob"]
var at = lines.findIndex({ x == "" }) ?? len(lines)
println(lines.slice(0, at))
println(lines.slice(Math.min(at + 1, len(lines)), len(lines)))
```

```text
["From: ann", "To: bob"]
["Hello,", "Bob"]
```

Stop with a clear message when a required element is missing:

```ant
var columns = ["id", "name", "email"]
var emailAt = columns.findIndex({ x == "email" }) ?? panic("no email column")
println("email is column ${emailAt + 1}")
```

```text
email is column 3
```

## Properties and relations

`Array.findIndex` with an `==` test agrees with [`Array.indexOf`](indexOf.md). It finds an index
exactly when [`Array.some`](some.md) is `true`, and the element at that index is what
[`Array.find`](find.md) returns:

```ant
var xs = [5, 3, 8, 3]
for v in [3, 8, 7] {
  assert(xs.findIndex({ x == v }) == xs.indexOf(v))
  assert((xs.findIndex({ x == v }) != null) == xs.some({ x == v }))
}
var i = xs.findIndex({ x > 4 }) ?? panic("none")
assert(xs[i] == xs.find({ x > 4 }))
```

No element before the index found passes the test, and the first index is never after the last one
that [`Array.findLastIndex`](findLastIndex.md) finds:

```ant
var xs = [1, 4, 6, 9, 12]
var i = xs.findIndex({ x % 3 == 0 }) ?? panic("none")
assert(!xs.slice(0, i).some({ x % 3 == 0 }))
assert(i <= (xs.findLastIndex({ x % 3 == 0 }) ?? panic("none")))
```

The same index comes out of [`Array.entries`](entries.md) and `Array.find`:

```ant
var xs = ["ant", "bee", "cicada"]
var e = xs.entries().find({ String.length(x.value) > 3 })
assert(e?.index == xs.findIndex({ String.length(x) > 3 }))
```

## Possible issues

The result can be `null`, and the type checker rejects it as an index until you narrow it:

```ant
var xs = [1, 2, 3]
println(xs[xs.findIndex({ x == 2 })]) // error: cannot use int | null where int is expected
```

Nor is it a condition. Ant has no truthiness, and index `0` is a match like any other. Compare with
`null` instead:

```ant
if [1, 2, 3].findIndex({ x == 2 }) { // error: cannot use int | null where bool is expected
  println("found")
}
```

```ant
[1, 2, 3].findIndex({ x == 1 }) != null // true
[1, 2, 3].findIndex({ x == 1 })         // 0
```

A miss is `null`, not `-1` as in JavaScript, so `i == -1` is never true. This program passes the
type checker and gives the wrong answer:

```ant
var i = [1, 2, 3].findIndex({ x == 9 })
if i == -1 {
  println("missing")
} else {
  println("found")
}
```

```text
found
```

Compare with `null` instead:

```ant
var i = [1, 2, 3].findIndex({ x == 9 })
if i == null {
  println("missing")
} else {
  println("found")
}
```

```text
missing
```

`f` must return a `bool`. A callback that computes something else is a compile error:

```ant
println([1, 2, 3].findIndex({ x * 2 })) // error: cannot use int where bool is expected
```

`Array.findIndex` takes a test, not a value. To look for a value by `==`, use
[`Array.indexOf`](indexOf.md):

```ant
println([1, 2, 3].findIndex(2)) // error: cannot use int where func(int): bool is expected
```

A lambda needs the expected function type to know what `x` is. Kept in an unannotated `var`, it is a
compile error. Annotate the variable:

```ant
var big = { x > 1 }
// error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
println([1, 2, 3].findIndex(big))
```

```ant
var big: func(int): bool = { x > 1 }
[1, 2, 3].findIndex(big) // 1
```

The call visits the elements `xs` had when it started. Elements the callback pushes are not tested:

```ant
var xs = [1, 2, 3]
xs.findIndex({ xs.push(x * 10); x == 10 }) // null
xs                                         // [1, 2, 3, 10, 20, 30]
```

## See also

[Array.find](find.md) · [Array.findLast](findLast.md) · [Array.findLastIndex](findLastIndex.md) ·
[Array.indexOf](indexOf.md) · [Array.some](some.md) · [Array.entries](entries.md) ·
[Array.removeAt](removeAt.md) · [Array.get](get.md)
