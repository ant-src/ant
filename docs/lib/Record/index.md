<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Record

Record holds the functions of record values.

<!-- end generated -->

A record is a value of a record type, such as `Point{ x: 1.0, y: 2.0 }`. Records are reference
values: assigning or passing one shares it. `==` compares records by structure, and
[`Record.same`](same.md) asks whether two names refer to one record.

## Details

- Two records are `==` when they carry the same tag and their fields are `==`. Two record types with
  the same fields are still different types, so their records are never equal.
- A record has no dynamic access: `p["x"]` is a compile error, and `for` does not iterate a record.
  When the keys are data, use a [Map](../Map/index.md).
- A spread, `Point{ ...p }`, makes a new record with the same fields. It copies one level: a field
  that holds a record, an array, a map, or a set still refers to the same value.

## Basic examples

Assigning a record shares it, so a change through `q` shows through `p`:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let q = p
q.x = 5.0
p                 // Point{ x: 5.0, y: 2.0 }
Record.same(p, q) // true
```

A spread makes a copy that is equal and not the same record:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let copy = Point{ ...p }
p == copy            // true
Record.same(p, copy) // false
copy.x = 9.0
p // Point{ x: 1.0, y: 2.0 }
```

## Scope

A function that changes a field of its parameter changes the caller's record:

```ant
type Point = { x: float64, y: float64 }
func shift(pt: Point) {
  pt.x += 10.0
}
let p = Point{ x: 1.0, y: 2.0 }
shift(p)
p // Point{ x: 11.0, y: 2.0 }
```

An array can hold one record twice. A change through one element shows through the other:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let ps = [p, p]
ps[0].y = 0.0
ps                        // [Point{ x: 1.0, y: 0.0 }, Point{ x: 1.0, y: 0.0 }]
Record.same(ps[0], ps[1]) // true
```

A spread with a later field replaces that field, and a spread into another record type with the same
fields converts the record. Later fields win, and the tag comes from the head:

```ant
type Point = { x: float64, y: float64 }
type Vec = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
Point{ ...p, x: 0.0 } // Point{ x: 0.0, y: 2.0 }
Vec{ ...p }           // Vec{ x: 1.0, y: 2.0 }
```

`==` compares the tag as well as the fields. Through a parameter of anonymous record type, a `Point`
and a `Vec` with the same fields compare unequal:

```ant
type Point = { x: float64, y: float64 }
type Vec = { x: float64, y: float64 }
func equal(a: { x: float64 }, b: { x: float64 }): bool {
  a == b
}
let p = Point{ x: 1.0, y: 2.0 }
equal(p, Point{ ...p })         // true
equal(p, Vec{ ...p })           // false
equal(p, Point{ ...p, y: 9.0 }) // false
```

`==` compares every field the record has, including `y`, which the parameter type does not mention.

`==` ends on cyclic records too. Two one-node loops with equal values are equal, though they are two
records:

```ant
type Node = { value: int, next: Node | null }
let a = Node{ value: 1, next: null }
a.next = a
let b = Node{ value: 1, next: null }
b.next = b
a == b            // true
Record.same(a, b) // false
```

## Functions by job

### Identity

- [Record.same](same.md): whether two names refer to one record

### Identity of other reference values

- [Array.same](../Array/same.md), [Map.same](../Map/same.md), [Set.same](../Set/same.md)

## Applications

Keep a history of a settings record. Each entry is a spread copy, so later changes to the live
record do not reach the history:

```ant
type Settings = { volume: int, muted: bool }
let live = Settings{ volume: 5, muted: false }
let history: Array[Settings] = []
history.push(Settings{ ...live })
live.volume = 8
history.push(Settings{ ...live })
live.muted = true
println(history)
println(live)
```

```text
[Settings{ volume: 5, muted: false }, Settings{ volume: 8, muted: false }]
Settings{ volume: 8, muted: true }
```

Without the spread, `history.push(live)` would store the live record itself, and every entry would
show its latest fields.

## Properties and relations

`Record.same` implies `==`: one record is always equal to itself, even after a change through
another name:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let q = p
q.y = 7.0
for b in [p, Point{ ...p }, Point{ ...p, x: 0.0 }] {
  if Record.same(p, b) {
    assert(p == b)
  }
}
```

## Possible issues

A spread copy shares the arrays, maps, sets, and records its fields hold. Changing the copy's `tags`
changes the original's, while setting its `title` does not:

```ant
type Doc = { title: string, tags: Array[string] }
let d = Doc{ title: "draft", tags: ["x"] }
let copy = Doc{ ...d }
copy.title = "final"
copy.tags.push("y")
d // Doc{ title: "draft", tags: ["x", "y"] }
```

Copy the field as well to keep the two apart:

```ant
type Doc = { title: string, tags: Array[string] }
let d = Doc{ title: "draft", tags: ["x"] }
let copy = Doc{ ...d, tags: [...d.tags] }
copy.tags.push("y")
d // Doc{ title: "draft", tags: ["x"] }
```

A record has no dynamic access, and `for` does not iterate one:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let x = p["x"] // error: a record has no dynamic access; use a Map when the keys are data
for f in p {
  // error: for needs an array, a map, a set, a string, Bytes, or a sequence, got Point
  println(f)
}
```

Records of two different types cannot be compared with `==`, even when the fields match:

```ant
type Point = { x: float64, y: float64 }
type Vec = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let same = p == Vec{ ...p } // error: Point and Vec cannot be compared with ==
```

A spread needs a value of a named record type, so it cannot copy a parameter of anonymous record
type:

```ant
type Point = { x: float64, y: float64 }
func copyOf(r: { x: float64, y: float64 }): Point {
  Point{ ...r } // error: a spread needs a value of a named record type, got { x: float64, y: float64 }
}
```

A value of anonymous record type may carry fields its type does not list, so take a `Point`
parameter instead.

`Record.same` takes records only. For arrays, maps, and sets use `Array.same`, `Map.same`, and
`Set.same`:

```ant
Record.same([1], [1]) // error: T would be Array[int], which is not assignable to its bound {}
```

## See also

[Record.same](same.md) · [Array.same](../Array/same.md) · [Map.same](../Map/same.md) ·
[Set.same](../Set/same.md) · [Records](../../guide/records.md) ·
[Record literals](../../language/literals/records.md) ·
[Equality](../../language/operators/equality.md) · [Map](../Map/index.md) ·
[Reflect](../Reflect/index.md)
