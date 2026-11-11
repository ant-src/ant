<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Indexed

```ant signature
type Indexed[T] = { index: int, value: T }
```

Indexed is an element with its index, as [`Array.entries`](../entries.md) yields them.

<!-- end generated -->

## Details

- [`Array.entries`](../entries.md) returns one `Indexed[T]` per element: `index` is its position and
  `value` is the element.
- A literal leaves out the type argument, as `Indexed{ index: 0, value: "a" }`; an annotation writes
  it, as `Indexed[int]`.

## Basic examples

Get `Indexed` records from `Array.entries`, and read their fields:

```ant
let es = ["a", "b"].entries()
es[1]       // Indexed{ index: 1, value: "b" }
es[1].index // 1
es[1].value // "b"
```

Write one with a literal. The checker infers `T` from `value`:

```ant
let e = Indexed{ index: 2, value: 3.5 }
e // Indexed{ index: 2, value: 3.5 }
```

Destructure the fields in a loop:

```ant
for { index, value } in ["ant", "bee"].entries() {
  println("${index}: ${value}")
}
```

```text
0: ant
1: bee
```

## Scope

Annotate a variable with the type argument:

```ant
let e: Indexed[string] = Indexed{ index: 0, value: "x" }
e.value // "x"
```

The value may be `null`, a record, an array, or a union member:

```ant
type Point = { x: int, y: int }
let n: Indexed[int | null] = Indexed{ index: 0, value: null }
n                                               // Indexed{ index: 0, value: null }
Indexed{ index: 1, value: Point{ x: 1, y: 2 } } // Indexed{ index: 1, value: Point{ x: 1, y: 2 } }
Indexed{ index: 2, value: [1, 2] }              // Indexed{ index: 2, value: [1, 2] }
let r: Indexed[Result[int, string]] = Indexed{ index: 3, value: Result.err("no") }
r.value // Err{ error: "no" }
```

Assign to a field, as with any record:

```ant
let e = Indexed{ index: 0, value: 8 }
e.value = e.value + 1
e // Indexed{ index: 0, value: 9 }
```

Destructure one record into variables:

```ant
let { index, value } = Indexed{ index: 4, value: "d" }
index // 4
value // "d"
```

Return it from a generic function, where `T` is the function's type parameter:

```ant
func at[T](xs: Array[T], i: int): Indexed[T] {
  Indexed{ index: i, value: xs[i] }
}

println(at(["x", "y", "z"], 1))
println(at([true, false], 0))
```

```text
Indexed{ index: 1, value: "y" }
Indexed{ index: 0, value: true }
```

Test a union member with `is`, written without the type argument:

```ant
func describe(x: Indexed[int] | string): string {
  if x is Indexed { "item ${x.index} is ${x.value}" } else { x }
}

println(describe(Indexed{ index: 1, value: 20 }))
println(describe("none"))
```

```text
item 1 is 20
none
```

## Applications

Sort the entries by value to get the original positions in sorted order:

```ant
let scores = [30, 10, 20]
scores.entries().sortBy({ x.value }).map({ x.index }) // [1, 2, 0]
```

Find the first element a test accepts together with where it is:

```ant
let words = ["ant", "beetle", "cicada"]
words.entries().find({ x.value.length() > 3 }) // Indexed{ index: 1, value: "beetle" }
```

Keep the position of each line that fails a check, so that a report can name it:

```ant
func blankLines(lines: Array[string]): Array[Indexed[string]] {
  lines.entries().filter({ x.value.trim().isEmpty() })
}

for { index } in blankLines(["a", " ", "b", ""]) {
  println("line ${index + 1} is blank")
}
```

```text
line 2 is blank
line 4 is blank
```

## Properties and relations

`Indexed` records compare by their fields, so an entry equals a literal with the same fields:

```ant
assert(["a"].entries()[0] == Indexed{ index: 0, value: "a" })
assert(Indexed{ index: 0, value: [1, 2] } == Indexed{ index: 0, value: [1, 2] })
assert(Indexed{ index: 0, value: 1 } != Indexed{ index: 1, value: 1 })
```

Each entry's `value` is the element at its `index`:

```ant
let xs = [5, 6, 7]
assert(xs.entries().every({ xs[x.index] == x.value }))
```

## Possible issues

A literal gives both fields. Leaving one out is a compile error:

```ant
let e = Indexed{ index: 0 }
// error: Indexed literal is missing the field value
// error: cannot infer the type parameter T of Indexed; annotate the result
```

The type argument goes in an annotation, never in a literal:

```ant
let e = Indexed[int]{ index: 0, value: 1 }
// error: a record literal takes no type arguments; write Indexed{ ... }
```

`Indexed` without a type argument is not a type:

```ant
let e: Indexed = Indexed{ index: 0, value: 1 } // error: Indexed takes 1 type argument, got 0
```

Type arguments are invariant, so an `Indexed[int]` does not fit where an `Indexed[int | string]` is
expected:

```ant
let e = Indexed{ index: 0, value: 1 }
let f: Indexed[int | string] = e
// error: cannot use Indexed[int] where Indexed[int | string] is expected
```

`Indexed` is not `Entry`. A map's entries have `key` and `value`, and an `Indexed` has no `key`:

```ant
for { key, value } in ["a", "b"].entries() { // error: Indexed[string] has no field key
  println("${key}: ${value}")
}
```

`Array.entries` builds a new array of new `Indexed` records, so assigning to an entry's `value`
leaves the array unchanged:

```ant
let xs = [1, 2]
let es = xs.entries()
es[0].value = 9
xs // [1, 2]
```

Write to the array through the index instead:

```ant
let xs = [1, 2]
for e in xs.entries() {
  xs[e.index] = e.value * 10
}
xs // [10, 20]
```

`for` over an array yields its elements, not `Indexed` records, so destructuring `{ index, value }`
needs `.entries()`:

```ant
for { index, value } in ["a", "b"] {
  // error: a record destructuring needs a record, got string
  // error: string is not a record, so it has no field index
  // error: string is not a record, so it has no field value
  println("${index}: ${value}")
}
```

## See also

[Array.entries](../entries.md) · [Entry](../../Map/types/Entry.md) ·
[Array.findIndex](../findIndex.md) · [Array.range](../range.md) ·
[Map.entries](../../Map/entries.md)
