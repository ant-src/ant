<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Record.same

```ant signature
func Record.same[T: {}](a: T, b: T): bool
```

Record.same reports whether a and b are the same record, not merely equal ones. The bound \{\}
admits any record type.

<!-- end generated -->

## Details

- Returns `false` for two different records, even when they are `==`.
- A nullable record, such as a value read from a map, is a compile error; narrow it first.

## Basic examples

A record is the same as itself, and as any other name for it:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let q = p
Record.same(p, p) // true
Record.same(p, q) // true
```

A copy made with a spread is equal, and not the same record:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
p == Point{ ...p }            // true
Record.same(p, Point{ ...p }) // false
```

## Scope

Two records of an empty type are equal, and still two records:

```ant
type Empty = {}
let a = Empty{}
let b = Empty{}
a == b            // true
Record.same(a, b) // false
```

A spread copies the fields, and a field that holds a record holds a reference, so the copy shares
it:

```ant
type Point = { x: float64, y: float64 }
type Line = { from: Point, to: Point }
let l = Line{ from: Point{ x: 0.0, y: 0.0 }, to: Point{ x: 1.0, y: 1.0 } }
let m = Line{ ...l, to: Point{ x: 2.0, y: 2.0 } }
Record.same(l, m)           // false
Record.same(l.from, m.from) // true
Record.same(l.to, m.to)     // false
```

A record read from an array, or passed to a function, is the same record:

```ant
type User = { name: string }
func first(us: Array[User]): User {
  us[0]
}
let ann = User{ name: "Ann" }
let users = [ann]
Record.same(users[0], ann)     // true
Record.same(first(users), ann) // true
```

Any record type works, including prelude records such as `Error`, and a union of records such as
`Result`:

```ant
let e = Error{ message: "x" }
Record.same(e, e)                     // true
Record.same(e, Error.wrap(e, "load")) // false
let r: Result[int, Error] = Result.ok(1)
Record.same(r, r) // true
```

A caught panic raised again and caught again is a new record, equal to the first:

```ant
let p = Panic.catch({ panic("boom") })
if p is Err {
  let again = Panic.catch({ Panic.raise(p.error) })
  if again is Err {
    println(again.error == p.error)
    println(Record.same(again.error, p.error))
  }
}
```

```text
true
false
```

In a pipeline, the left side is `a`:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let q = p
p | Record.same(q) // true
```

Name the arguments, in any order:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
Record.same(a: p, b: p)             // true
Record.same(b: p, a: Point{ ...p }) // false
```

## Applications

Skip a save when an update returns its input unchanged. `rename` returns the same record when the
name does not change, and a new one when it does:

```ant
type User = { id: int, name: string }

func rename(u: User, name: string): User {
  if u.name == name { u } else { User{ ...u, name } }
}

let ann = User{ id: 1, name: "Ann" }
for name in ["Ann", "Anna"] {
  let next = rename(ann, name)
  if Record.same(next, ann) {
    println("${name}: unchanged")
  } else {
    println("${name}: save ${next}")
  }
}
```

```text
Ann: unchanged
Anna: save User{ id: 1, name: "Anna" }
```

Find a loop in a linked list. Two nodes may be equal and still be different nodes, so the walk keeps
the nodes it has seen and compares by identity:

```ant
type Node = { value: int, next: Node | null }

func hasLoop(head: Node): bool {
  let seen: Array[Node] = []
  var node: Node | null = head
  while node != null {
    let n = node
    if seen.findIndex({ Record.same(x, n) }) != null {
      return true
    }
    seen.push(n)
    node = n.next
  }
  false
}

let c = Node{ value: 1, next: null }
let b = Node{ value: 1, next: c }
let a = Node{ value: 1, next: b }
println(hasLoop(a))
c.next = a
println(hasLoop(a))
```

```text
false
true
```

## Properties and relations

`Record.same` is reflexive and symmetric:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let q = p
assert(Record.same(p, p))
assert(Record.same(p, q) == Record.same(q, p))
```

The same record is `==` to itself unless it holds a `NaN`, in a field or anywhere below one, since
floats compare with IEEE 754 rules:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: Math.NaN, y: 0.0 }
p == p            // false
Record.same(p, p) // true
```

When two names refer to the same record, a change through one shows through the other. For a copy it
does not:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let q = p
let r = Point{ ...p }
q.x = 5.0
assert(Record.same(p, q) && p.x == 5.0)
assert(!Record.same(p, r) && r.x == 1.0)
```

`Reflect.same` reports the same for two reflected records:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let t = type(Point)
assert(Reflect.same(Reflect.valueOf(t, p), Reflect.valueOf(t, p)))
assert(!Reflect.same(Reflect.valueOf(t, p), Reflect.valueOf(t, Point{ ...p })))
```

## Possible issues

`Record.same` is not reached through `.` on a record, since a record's functions are the ones
declared on its type:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
println(p.same(p)) // error: Point has no field or function same
```

Both arguments have the one type `T`, fixed by the first argument. Two record types with the same
fields are still different types:

```ant
type Point = { x: float64, y: float64 }
type Vec = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
println(Record.same(p, Vec{ ...p })) // error: cannot use Vec where Point is expected
```

A nullable record, such as a value read from a map, is not a record. Narrow it first:

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let m = Map{ "a": p }
println(Record.same(m["a"], p))
// error: T would be Point | null, which is not assignable to its bound {}
```

```ant
type Point = { x: float64, y: float64 }
let p = Point{ x: 1.0, y: 2.0 }
let m = Map{ "a": p }
let found = m["a"]
found != null && Record.same(found, p) // true
```

Arrays, maps, and sets have their own identity functions:

```ant
println(Record.same([1], [1]))
// error: T would be Array[int], which is not assignable to its bound {}
```

## See also

[Array.same](../Array/same.md) · [Map.same](../Map/same.md) · [Set.same](../Set/same.md) ·
[Reflect.same](../Reflect/same.md) · [Error.wrap](../Error/wrap.md) ·
[== operator](../../language/operators/equality.md)
