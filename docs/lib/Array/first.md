<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.first

```ant signature
func Array.first[T](xs: Array[T]): T | null
```

Array.first returns the first element of xs, or null when xs is empty.

<!-- end generated -->

## Details

- Returns `null` when `xs` is empty; `xs` is unchanged.
- When `T` includes `null`, a `null` result does not tell an empty array from a `null` first
  element; use [Array.isEmpty](isEmpty.md).

## Basic examples

```ant
Array.first([10, 20, 30]) // 10
Array.first(["a", "b"])   // "a"
```

An empty array gives `null`. Annotate it so that `T` is known:

```ant
var empty: Array[int] = []
Array.first(empty) // null
```

Call it through `.` on any array:

```ant
var queue = ["build", "test", "deploy"]
queue.first() // "build"
```

## Scope

An array of records gives the first record. Read a field of it with `?.`, which yields `null` for an
empty array:

```ant
type User = { name: string, age: int }
var users = [User{ name: "Ann", age: 31 }, User{ name: "Bob", age: 27 }]
var nobody: Array[User] = []
users.first()                      // User{ name: "Ann", age: 31 }
users.first()?.name                // "Ann"
nobody.first()?.name               // null
nobody.first()?.name ?? "no users" // "no users"
```

A nested array gives its first inner array, and `?.[i]` indexes into it:

```ant
var grid = [[1, 2], [3]]
grid.first()      // [1, 2]
grid.first()?.[1] // 2
```

Elements of a union keep their tag, so `is` tells which member came first:

```ant
type Circle = { r: float64 }
type Square = { side: float64 }
type Shape = Circle | Square
var shapes: Array[Shape] = [Square{ side: 2.0 }, Circle{ r: 1.0 }]
shapes.first()           // Square{ side: 2.0 }
shapes.first() is Square // true
```

An array of functions gives a function, which `?.(…)` calls when it is there:

```ant
var steps: Array[func(int): int] = [{ x + 1 }, { x * 2 }]
var none: Array[func(int): int] = []
steps.first()?.(10) // 11
none.first()?.(10)  // null
```

For the first grapheme of a string, what a reader sees as one character, split it with
`String.graphemes` first. `String.chars` splits into code points instead, and the two differ when a
character is written with a combining mark:

```ant
String.graphemes("🐜🐝🐞").first()       // "🐜"
String.graphemes("e\u{301}x").first() // "é"
String.chars("e\u{301}x").first()     // "e"
```

The keys, values, and entries of a map, and the values of a set, are arrays in insertion order, so
`Array.first` gives the earliest one:

```ant
Map.keys(Map{ b: 2, a: 1 }).first() // "b"
Set.of([3, 1, 2]).values().first()  // 3
```

In a pipe, the array is the first argument:

```ant
"the quick brown fox" | String.split(" ") | Array.first() // "the"
```

## Applications

Take the first command-line argument, or a default when there is none:

<!-- test: args Ann -->

```ant
var name = Os.args().first() ?? "world"
println("Hello, ${name}")
```

```text
Hello, Ann
```

Bind the result to a local and return early when the array is empty. After the test the local is a
`string`:

```ant
func greet(names: Array[string]): string {
  var head = names.first()
  if head == null {
    return "hello, nobody"
  }
  "hello, ${head.toUpper()}"
}

println(greet(["ann", "bob"]))
println(greet([]))
```

```text
hello, ANN
hello, nobody
```

Keep the first word for each initial letter:

```ant
var words = ["apple", "avocado", "banana", "blueberry", "cherry"]
var byLetter = words.groupBy({ String.graphemes(x)[0] })
byLetter.map({ x.value.first() ?? "" }) // Map{ "a": "apple", "b": "banana", "c": "cherry" }
```

Stop with a panic when an array that must not be empty is:

```ant
var config: Array[string] = []
config.first() ?? panic("no config file found") // panic: no config file found
```

## Properties and relations

`Array.first` agrees with `Array.get` at index 0, with indexing on a non-empty array, with
`Array.find` given a function that accepts everything, and with `Array.last` of the reversed array:

```ant
for xs in [[4, 5, 6], [7], []] {
  assert(xs.first() == xs.get(0))
  assert(xs.first() == xs.find({ true }))
  assert(xs.first() == xs.reverse().last())
  if !xs.isEmpty() {
    assert(xs.first() == xs[0])
  }
}
```

When the element type does not include `null`, the result is `null` exactly when the array is empty:

```ant
for xs in [["a"], []] {
  assert((xs.first() == null) == xs.isEmpty())
}
```

`Array.first` does not change the array:

```ant
var xs = [1, 2, 3]
var head = xs.first()
assert(head == 1)
assert(xs == [1, 2, 3])
```

## Possible issues

The result is `T | null`, not `T`, so using it as a `T` is a compile error:

```ant
var xs = [3, 1, 2]
println(xs.first() + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`, or narrow a local:

```ant
var xs = [3, 1, 2]
(xs.first() ?? 0) + 1 // 4
```

Testing the call itself does not narrow a second call, because only locals, parameters, and
top-level `let`s narrow:

```ant
var xs = [3, 1, 2]
if xs.first() != null {
  println(xs.first() * 2) // error: * needs both operands of one type, got int | null and int
}
```

Bind the result to a local and test that:

```ant
var xs = [3, 1, 2]
var head = xs.first()
if head != null {
  println(head * 2)
}
```

```text
6
```

When the element type includes `null`, a `null` result does not tell an empty array from one whose
first element is `null`. Ask the array with [`Array.isEmpty`](isEmpty.md) instead:

```ant
var slots: Array[string | null] = [null, "b"]
slots.first()   // null
slots.isEmpty() // false
```

Indexing with `[0]` panics on an empty array, where `Array.first` gives `null`:

```ant
var empty: Array[int] = []
empty.first() // null
empty[0]      // panic: index 0 out of range for an array of length 0
```

## See also

[Array.last](last.md) · [Array.get](get.md) · [Array.find](find.md) · [Array.pop](pop.md) ·
[Array.isEmpty](isEmpty.md) · [Array.slice](slice.md) ·
[?? operator](../../language/operators/null-coalescing.md) ·
[?. operator](../../language/operators/optional-chaining.md) · [Null guide](../../guide/null.md)
