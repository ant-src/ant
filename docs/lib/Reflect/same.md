<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.same

```ant signature
func Reflect.same(a: Reflect.Value, b: Reflect.Value): bool
```

Reflect.same reports whether a and b hold one record, array, map, set, or Bytes, as Record.same,
Array.same, Map.same, Set.same, and Bytes.same do, and whether they are == for any other kind. A
walker tests each child against the boxes on its path with it to find a cycle, which the readers
never detect.

<!-- end generated -->

## Details

- For a record, an array, a map, a set, or `Bytes`, `Reflect.same` reports identity: two boxes of
  one value are the same even when their types differ, and two equal copies are not.
- For any other kind it is `==` on the boxes, so the types must be `==` too: `1` as an `int` and `1`
  as a named type over `int` are not the same.

## Basic examples

Two boxes of one array are the same; a box of an equal copy is not:

```ant
let xs = [1, 2, 3]
let a = Reflect.valueOf(type(Array[int]), xs)
let b = Reflect.valueOf(type(Array[int]), xs)
let c = Reflect.valueOf(type(Array[int]), [1, 2, 3])
Reflect.same(a, b) // true
Reflect.same(a, c) // false
```

Boxes of scalars are the same when they are `==`:

```ant
Reflect.same(Reflect.valueOf(type(int), 1), Reflect.valueOf(type(int), 1))           // true
Reflect.same(Reflect.valueOf(type(string), "a"), Reflect.valueOf(type(string), "b")) // false
```

## Scope

A record is the same through any type that boxes it, here its named type and an anonymous record
type:

```ant
type User = { name: string, age: int }
let ann = User{ name: "Ann", age: 31 }
let named = Reflect.valueOf(type(User), ann)
let partial = Reflect.valueOf(type({ name: string }), ann)
Reflect.same(named, partial)                                     // true
Reflect.same(named, Reflect.valueOf(type(User), User{ ...ann })) // false
```

Maps and sets compare by identity too:

```ant
let m = Map{ "a": 1 }
let s = Set.of([1, 2])
let md = Reflect.valueOf(type(Map[string, int]), m)
let sd = Reflect.valueOf(type(Set[int]), s)
md.same(Reflect.valueOf(type(Map[string, int]), m))             // true
md.same(Reflect.valueOf(type(Map[string, int]), Map{ "a": 1 })) // false
sd.same(Reflect.valueOf(type(Set[int]), s))                     // true
```

A box of a union type holds the member, so a nullable array and a plain one are the same array:

```ant
let xs = [1]
let plain = Reflect.valueOf(type(Array[int]), xs)
let nullable = Reflect.valueOf(type(Array[int] | null), xs)
Reflect.same(plain, nullable) // true
```

Scalars of different types are never the same, even when their values print alike:

```ant
type UserId = int
let n = Reflect.valueOf(type(int), 7)
Reflect.same(n, Reflect.valueOf(type(UserId), UserId(7))) // false
Reflect.same(n, Reflect.valueOf(type(int8), int8(7)))     // false
```

`null` boxed through a nullable type is boxed at `null`, so it is the same as any other `null` box:

```ant
let none = Reflect.valueOf(type(int | null), null)
Reflect.same(none, Reflect.valueOf(type(null), null)) // true
```

`Bytes` are mutable, so two boxes of equal bytes are the same only when they hold one collection.
Any other kind compares with `==`, a function by identity:

```ant
let hi = "hi".toBytes()
let a = Reflect.valueOf(type(Bytes), hi)
a.same(Reflect.valueOf(type(Bytes), hi))             // true
a.same(Reflect.valueOf(type(Bytes), "hi".toBytes())) // false
let double = func(x: int): int { x * 2 }
let f = Reflect.valueOf(type(func(int): int), double)
f.same(Reflect.valueOf(type(func(int): int), double)) // true
```

Children taken out with a reader are boxes of the values in the parent, so a record shared by two
fields is the same through both:

```ant
type Point = { x: int, y: int }
type Line = { from: Point, to: Point }
let p = Point{ x: 0, y: 0 }
let d = Reflect.valueOf(type(Line), Line{ from: p, to: p })
Reflect.same(d.field("from"), d.field("to")) // true
```

## Applications

Render any value of records and scalars, and print `...` where it refers back to a record on the
path, instead of recursing forever:

```ant
type Node = { name: string, next: Node | null }

func show(d: Reflect.Value, path: Array[Reflect.Value]): string {
  if path.some({ Reflect.same(x, d) }) {
    return "..."
  }
  let s = Reflect.shape(Reflect.type(d))
  if !(s is Reflect.Record) {
    return String.from(Reflect.scalar(d))
  }
  let inner = [...path, d]
  let body = s.fields
    | Array.map({ "${x.name}: ${show(d.field(x.name), inner)}" })
    | String.join(", ")
  "{${body}}"
}

let a = Node{ name: "a", next: null }
let b = Node{ name: "b", next: a }
a.next = b
println(show(Reflect.valueOf(type(Node), a), []))
```

```text
{name: a, next: {name: b, next: ...}}
```

Find the fields of a record that hold one array, so that a change through one field shows through
the others:

```ant
type Config = { allow: Array[string], deny: Array[string], admins: Array[string] }

func shared[T](t: Reflect.Type[T], v: T): Array[string] {
  let d = Reflect.valueOf(t, v)
  let s = Reflect.shape(Reflect.erase(t))
  if !(s is Reflect.Record) {
    return []
  }
  let names = s.fields.map({ x.name })
  names.filter(func(name: string): bool {
    names.some({ x != name && Reflect.same(d.field(x), d.field(name)) })
  })
}

let staff = ["ann", "bob"]
let config = Config{ allow: staff, deny: [], admins: staff }
println(shared(type(Config), config))
```

```text
["allow", "admins"]
```

## Properties and relations

For a record, `Reflect.same` on two boxes agrees with [Record.same](../Record/same.md) on the
values, and so for arrays, maps, and sets with their own `same`:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let q = Point{ ...p }
func box(v: Point): Reflect.Value {
  Reflect.valueOf(type(Point), v)
}
assert(Reflect.same(box(p), box(p)) == Record.same(p, p))
assert(Reflect.same(box(p), box(q)) == Record.same(p, q))
```

`==` on boxes compares the types and the contents, so equal copies are `==` and not the same, and
two boxes of one record at different types are the same and not `==`:

```ant
type User = { name: string, age: int }
let ann = User{ name: "Ann", age: 31 }
let named = Reflect.valueOf(type(User), ann)
let partial = Reflect.valueOf(type({ name: string }), ann)
assert(Reflect.same(named, partial) && named != partial)
let a = Reflect.valueOf(type(Array[int]), [1, 2])
let b = Reflect.valueOf(type(Array[int]), [1, 2])
assert(a == b)
assert(!Reflect.same(a, b))
assert(Reflect.same(a, a) && a == a)
```

## Possible issues

A float box holding `NaN` is not the same as itself, because `NaN` is not `==` to itself:

```ant
let d = Reflect.valueOf(type(float64), Math.NaN)
Reflect.same(d, d) // false
```

A box keeps the value it was given, not a variable. After a variable is assigned a new array, the
old box holds the old one:

```ant
var xs = [1]
let before = Reflect.valueOf(type(Array[int]), xs)
xs = [1]
Reflect.same(before, Reflect.valueOf(type(Array[int]), xs)) // false
```

## See also

[Record.same](../Record/same.md) · [Array.same](../Array/same.md) · [Map.same](../Map/same.md) ·
[Set.same](../Set/same.md) · [Reflect.valueOf](valueOf.md) · [Reflect.field](field.md) ·
[Reflect.items](items.md) · [Reflect.entries](entries.md)
