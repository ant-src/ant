<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Array.isEmpty

```ant signature
func Array.isEmpty[T](xs: Array[T]): bool
```

Array.isEmpty reports whether xs has no elements.

<!-- end generated -->

## Details

- Ant has no truthiness, so an array is never a condition by itself; use `Array.isEmpty(xs)`.

## Basic examples

```ant
Array.isEmpty([1, 2, 3]) // false
let none: Array[int] = []
Array.isEmpty(none) // true
```

Call it through `.` on any array:

```ant
let queue = ["a", "b"]
queue.isEmpty() // false
```

## Scope

An array that holds only empty arrays, empty strings, or `null` still has elements:

```ant
[[1], []].isEmpty() // false
[""].isEmpty()      // false
[null].isEmpty()    // false
```

The answer follows the array as it changes:

```ant
let stack = [1]
stack.isEmpty() // false
stack.pop()     // 1
stack.isEmpty() // true
stack.push(2)
stack.isEmpty() // false
```

The result of another function, such as `Array.filter`, is often what you test:

```ant
[1, 2, 3].filter({ x > 5 }).isEmpty() // true
```

In a pipeline, the array is the first argument:

```ant
"a b" | String.split(" ") | Array.isEmpty() // false
```

As a function value, it maps an array of arrays to an array of `bool`s:

```ant
Array.map([[1], [], [2, 3]], Array.isEmpty) // [false, true, false]
```

For an array that may be `null`, `?.` yields `null` instead of calling:

```ant
let tags: Array[string] | null = null
tags?.isEmpty()         // null
tags?.isEmpty() ?? true // true
```

## Named arguments

The parameter is named `xs`:

```ant
Array.isEmpty(xs: [1]) // false
```

## Applications

Report an empty input instead of computing on it:

```ant
func average(xs: Array[float64]): float64 | null {
  if xs.isEmpty() {
    return null
  }
  var total = 0.0
  for x in xs {
    total += x
  }
  total / float64(len(xs))
}

average([1.0, 2.0, 6.0]) // 3.0
average([])              // null
```

Work through a queue until it runs dry:

```ant
let todo = ["b", "a"]
let done: Array[string] = []
while !todo.isEmpty() {
  let item = todo.pop()
  if item != null {
    done.push(item)
  }
}
println(done)
```

```text
["a", "b"]
```

Collect validation problems and succeed only when there are none:

```ant
type User = { name: string, age: int }

func check(u: User): Result[User, Array[string]] {
  let problems: Array[string] = []
  if String.isEmpty(u.name) {
    problems.push("name is empty")
  }
  if u.age < 0 {
    problems.push("age is negative")
  }
  if problems.isEmpty() { Result.ok(u) } else { Result.err(problems) }
}

println(check(User{ name: "Ann", age: 31 }))
println(check(User{ name: "", age: -1 }))
```

```text
Ok{ value: User{ name: "Ann", age: 31 } }
Err{ error: ["name is empty", "age is negative"] }
```

## Properties and relations

`Array.isEmpty` agrees with `len` and `Array.length`, with `==` against an annotated empty array,
since arrays are equal when they have the same length and `==` elements, and, for an array whose
element type excludes `null`, with `Array.first`:

```ant
let none: Array[int] = []
let cases: Array[Array[int]] = [[], [7], [1, 2]]
for xs in cases {
  assert(xs.isEmpty() == (len(xs) == 0))
  assert(xs.isEmpty() == (Array.length(xs) == 0))
  assert(xs.isEmpty() == (xs == none))
  assert(xs.isEmpty() == (xs.first() == null))
}
```

## Possible issues

An array is not a condition. Ask `Array.isEmpty` explicitly:

```ant
let xs = [1]
if xs { // error: cannot use Array[int] where bool is expected
  println("has elements")
}
```

A nullable array must be narrowed or reached through `?.` first:

```ant
let xs: Array[int] | null = null
println(Array.isEmpty(xs)) // error: cannot use Array[int] | null where Array[T] is expected
```

An empty array literal has no element type to infer. Annotate it:

```ant
println([].isEmpty()) // error: cannot infer the element type of an empty array; annotate it
```

An operand of `==` gives no expected type, so `xs == []` cannot infer the element type. Use
`Array.isEmpty`:

```ant
let xs = [1]
println(xs == []) // error: cannot infer the element type of an empty array; annotate it
```

`Array.isEmpty` takes only arrays. Strings, maps, and sets have their own:

```ant
println(Array.isEmpty("abc")) // error: cannot use string where Array[T] is expected
```

`Array.first` is `null` when the array is empty, but also when the first element is `null`. Use
`Array.isEmpty` to test for elements:

```ant
let n: Array[int | null] = [null]
n.first() == null // true
n.isEmpty()       // false
```

## See also

[Array.length](length.md) · [len](../prelude/len.md) · [Array.first](first.md) · [Array.pop](pop.md)
· [String.isEmpty](../String/isEmpty.md) · [Map.isEmpty](../Map/isEmpty.md) ·
[Set.isEmpty](../Set/isEmpty.md)
