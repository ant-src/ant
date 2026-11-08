<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Indexed

```ant signature
type Indexed[T] = { index: int, value: T }
```

Indexed is an element with its index, as [`Array.entries`](../entries.md) yields them.

<!-- end generated -->

## Details

- `Indexed[T]` is a generic record type that the prelude declares in every module, so it needs no
  import ([Prelude](https://github.com/ant-src/spec/blob/master/ant-lang.md#prelude)).
- [`Array.entries`](../entries.md) returns one `Indexed[T]` per element, where `index` is the
  element's position and `value` is the element
  ([Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)).
- A literal names the type without type arguments, as `Indexed{ index: 0, value: "a" }`; the checker
  infers `T` from `value` or from the expected type. Both fields have no default, so a literal gives
  both ([Records](https://github.com/ant-src/spec/blob/master/ant-lang.md#records),
  [Generics](https://github.com/ant-src/spec/blob/master/ant-lang.md#generics)).
- An annotation writes the type argument, as `Indexed[int]`. `Indexed` without one is a compile
  error ([Generics](https://github.com/ant-src/spec/blob/master/ant-lang.md#generics),
  [Types as values](https://github.com/ant-src/spec/blob/master/ant-lang.md#types-as-values)).
- Type arguments are invariant: `Indexed[int]` is not an `Indexed[int | string]`
  ([Assignability](https://github.com/ant-src/spec/blob/master/ant-lang.md#assignability)).
- The fields are mutable, like the fields of every record
  ([Records](https://github.com/ant-src/spec/blob/master/ant-lang.md#records)).

## Basic examples

Get `Indexed` records from `Array.entries`, and read their fields:

```ant
var es = ["a", "b"].entries()
es[1]       // Indexed{ index: 1, value: "b" }
es[1].index // 1
es[1].value // "b"
```

Write one with a literal. The checker infers `T` from `value`:

```ant
var e = Indexed{ index: 2, value: 3.5 }
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
var e: Indexed[string] = Indexed{ index: 0, value: "x" }
e.value // "x"
```

The value may be `null`, a record, an array, or a union member:

```ant
type Point = { x: int, y: int }
var n: Indexed[int | null] = Indexed{ index: 0, value: null }
n                                               // Indexed{ index: 0, value: null }
Indexed{ index: 1, value: Point{ x: 1, y: 2 } } // Indexed{ index: 1, value: Point{ x: 1, y: 2 } }
Indexed{ index: 2, value: [1, 2] }              // Indexed{ index: 2, value: [1, 2] }
var r: Indexed[Result[int, string]] = Indexed{ index: 3, value: Result.err("no") }
r.value // Err{ error: "no" }
```

Assign to a field, as with any record:

```ant
var e = Indexed{ index: 0, value: 8 }
e.value = e.value + 1
e // Indexed{ index: 0, value: 9 }
```

Destructure one record into variables:

```ant
var { index, value } = Indexed{ index: 4, value: "d" }
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
var scores = [30, 10, 20]
scores.entries().sortBy({ x.value }).map({ x.index }) // [1, 2, 0]
```

Find the first element a test accepts together with where it is:

```ant
var words = ["ant", "beetle", "cicada"]
words.entries().find({ String.length(x.value) > 3 }) // Indexed{ index: 1, value: "beetle" }
```

Keep the position of each line that fails a check, so that a report can name it:

```ant
func blankLines(lines: Array[string]): Array[Indexed[string]] {
  lines.entries().filter({ String.trim(x.value) == "" })
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
var xs = [5, 6, 7]
assert(xs.entries().every({ xs[x.index] == x.value }))
```

## Possible issues

A literal gives both fields. Leaving one out is a compile error:

<!-- test: error Indexed literal is missing the field value -->

```ant
var e = Indexed{ index: 0 }
```

The type argument goes in an annotation, never in a literal:

<!-- test: error expected a newline or ';' before '{' -->

```ant
var e = Indexed[int]{ index: 0, value: 1 }
```

`Indexed` without a type argument is not a type:

<!-- test: error Indexed takes 1 type argument, got 0 -->

```ant
var e: Indexed = Indexed{ index: 0, value: 1 }
```

Type arguments are invariant, so an `Indexed[int]` does not fit where an `Indexed[int | string]` is
expected:

<!-- test: error cannot use Indexed[int] where Indexed[int | string] is expected -->

```ant
var e = Indexed{ index: 0, value: 1 }
var f: Indexed[int | string] = e
```

`Indexed` is not `Entry`. A map's entries have `key` and `value`, and an `Indexed` has no `key`:

<!-- test: error Indexed[string] has no field key -->

```ant
for { key, value } in ["a", "b"].entries() {
  println("${key}: ${value}")
}
```

`Array.entries` builds a new array of new `Indexed` records, so assigning to an entry's `value`
leaves the array unchanged ([Array](https://github.com/ant-src/spec/blob/master/ant-lang.md#array)):

```ant
var xs = [1, 2]
var es = xs.entries()
es[0].value = 9
xs // [1, 2]
```

Write to the array through the index instead:

```ant
var xs = [1, 2]
for e in xs.entries() {
  xs[e.index] = e.value * 10
}
xs // [10, 20]
```

`for` over an array yields its elements, not `Indexed` records, so destructuring `{ index, value }`
needs `.entries()` ([Loops](https://github.com/ant-src/spec/blob/master/ant-lang.md#loops)):

<!-- test: error a record destructuring needs a record, got string -->

```ant
for { index, value } in ["a", "b"] {
  println("${index}: ${value}")
}
```

## See also

[Array.entries](../entries.md) · [Entry](../../Map/types/Entry.md) ·
[Array.findIndex](../findIndex.md) · [Array.range](../range.md) ·
[Map.entries](../../Map/entries.md)
