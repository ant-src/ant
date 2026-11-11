<!-- status: reviewed -->

# Record literals

A record literal names a record type and lists its fields in braces: `Point{ x: 1, y: 2 }`. A bare
name is short for a field of the same name, and `...` copies the fields of another record.

## Details

- Every record literal names its type. There are no anonymous record literals.
- A literal gives every field without a default. A missing field or a field the type does not list
  is a compile error.
- Entries and spreads evaluate first, left to right, and a later entry wins. Then each omitted field
  takes its default, evaluated afresh for each literal.
- Each evaluation builds a new record. Fields are mutable, and assigning or passing a record shares
  it rather than copying it.

## Basic examples

```ant
type Point = { x: int, y: int }
Point{ x: 1, y: 2 }   // Point{ x: 1, y: 2 }
Point{ y: 2, x: 1 }   // Point{ x: 1, y: 2 }
Point{ x: 1, y: 2 }.x // 1
```

A bare name is a shorthand entry: `Point{ x, y }` means `Point{ x: x, y: y }`:

```ant
type Point = { x: int, y: int }
let x = 3
let y = 4
Point{ x, y } // Point{ x: 3, y: 4 }
```

A spread copies every field of a record, and a later entry replaces one:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
Point{ ...p, x: 10 } // Point{ x: 10, y: 2 }
p                    // Point{ x: 1, y: 2 }
```

## Scope

### Field defaults

A literal may leave out a field that declares a default after `=`:

```ant
type Server = {
  addr: string,
  workers: int = 4,
  verbose: bool = false,
  tags: Array[string] = [],
}
Server{ addr: ":8080" }                // Server{ addr: ":8080", workers: 4, verbose: false, tags: [] }
Server{ addr: ":9090", verbose: true } // Server{ addr: ":9090", workers: 4, verbose: true, tags: [] }
```

A default runs again for each literal, so every `Server` above gets its own `tags` array:

```ant
type Server = { addr: string, tags: Array[string] = [] }
let a = Server{ addr: ":1" }
let b = Server{ addr: ":2" }
a.tags.push("prod")
b.tags // []
```

### Spreads that convert

The tag comes from the head, not from the spread, so a spread converts between record types with the
same fields. Extra fields of the new type take their defaults:

```ant
type Point = { x: int, y: int }
type Vec = { x: int, y: int }
type Point3 = { x: int, y: int, z: int = 0 }
let p = Point{ x: 1, y: 2 }
Vec{ ...p }          // Vec{ x: 1, y: 2 }
Point3{ ...p }       // Point3{ x: 1, y: 2, z: 0 }
Point3{ ...p, z: 5 } // Point3{ x: 1, y: 2, z: 5 }
```

### Field values

A field takes any expression: another record literal, a collection, a union value, `null`, or a
lambda, which gets its type from the field:

```ant
type Point = { x: int, y: int }
type Shape = {
  name: string,
  corners: Array[Point],
  parent: Shape | null = null,
  area: func(): int,
}
let square = Shape{
  name: "square",
  corners: [Point{ x: 0, y: 0 }, Point{ x: 2, y: 2 }],
  area: { 4 },
}
square.corners[1].x // 2
square.parent       // null
square.area()       // 4
```

### Generic records

The type arguments of a generic record come from its fields. They are never written in the head:

```ant
type Pair[A, B] = { first: A, second: B }
let pair = Pair{ first: "ant", second: 6 }
pair.second + 1 // 7
pair            // Pair{ first: "ant", second: 6 }
```

### Types from a namespace

A record type reached through an imported namespace is written with its path in the head:

```ant [geometry.ant]
export type Point = { x: float64, y: float64 }
```

```ant
import "./geometry.ant" Geometry

let origin = Geometry.Point{ x: 0.0, y: 0.0 }
println(origin)
```

```text
Point{ x: 0.0, y: 0.0 }
```

### Evaluation order

Entries evaluate left to right, in the order the literal writes them:

```ant
type Point = { x: int, y: int }
func note(name: string, n: int): int {
  println("evaluating ${name}")
  n
}
let p = Point{ y: note("y", 2), x: note("x", 1) }
println(p)
```

```text
evaluating y
evaluating x
Point{ x: 1, y: 2 }
```

## Applications

Build options with defaults, and give only what differs:

```ant
type Retry = { attempts: int = 3, delayMs: int = 100, jitter: bool = true }
func describe(r: Retry): string {
  "${r.attempts} attempts, ${r.delayMs} ms apart"
}
describe(Retry{})              // "3 attempts, 100 ms apart"
describe(Retry{ attempts: 5 }) // "5 attempts, 100 ms apart"
```

Update a record without changing the original, with a spread:

```ant
type Account = { owner: string, balance: int }
func deposit(a: Account, amount: int): Account {
  Account{ ...a, balance: a.balance + amount }
}
let before = Account{ owner: "Ann", balance: 100 }
let after = deposit(before, 50)
after.balance  // 150
before.balance // 100
```

Build records from data with shorthand entries:

```ant
type User = { name: string, age: int }
let ages = Map{ "Ann": 31, "Bob": 27 }
let users: Array[User] = []
for { key: name, value: age } in ages {
  users.push(User{ name, age })
}
println(users)
```

```text
[User{ name: "Ann", age: 31 }, User{ name: "Bob", age: 27 }]
```

## Properties and relations

Two records are `==` when they have the same type and equal fields:

```ant
type Point = { x: int, y: int }
type Vec = { x: int, y: int }
assert(Point{ x: 1, y: 2 } == Point{ x: 1, y: 2 })
assert(Point{ x: 1, y: 2 } != Point{ x: 2, y: 1 })
assert(Vec{ ...Point{ x: 1, y: 2 } } == Vec{ x: 1, y: 2 })
```

A spread makes a new record, equal but not the same. A second name shares one record:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let copy = Point{ ...p }
let same = p
assert(copy == p && !Record.same(copy, p))
same.x = 9
assert(p.x == 9 && copy.x == 1)
```

An omitted field with a default equals the default written out:

```ant
type Server = { addr: string, workers: int = 4 }
assert(Server{ addr: ":80" } == Server{ addr: ":80", workers: 4 })
```

## Possible issues

### A missing or unknown field

A literal must give every field without a default:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1 } // error: Point literal is missing the field y
```

A field the type does not list is a compile error, which catches typos:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2, z: 3 } // error: Point has no field z
```

### No anonymous record literals

Braces without a type name are not a record:

```ant
let p = { x: 1, y: 2 } // error: a record or map literal names its type
```

### Type arguments in the head

A generic record takes its type arguments from its fields or the expected type, not from the head:

```ant
type Box[T] = { value: T }
let b = Box[int]{ value: 1 } // error: a record literal takes no type arguments
```

### A generic record with all fields omitted

When every field takes its default, nothing fixes the type arguments. Give an expected type:

```ant
type Bag[T] = { items: Array[T] = [] }
let bag = Bag{} // error: cannot infer the type parameter T of Bag; annotate the result
```

```ant
type Bag[T] = { items: Array[T] = [] }
let bag: Bag[int] = Bag{}
bag.items.push(1)
bag // Bag{ items: [1] }
```

### A default that reads a variable shares its value

A default that names a variable reads it each time, so every record shares the one array it holds:

```ant
let shared: Array[string] = []
type Job = { name: string, tags: Array[string] = shared }
let a = Job{ name: "a" }
let b = Job{ name: "b" }
a.tags.push("urgent")
b.tags // ["urgent"]
```

### A spread needs a named record type

A value of an anonymous record type may hold fields its type does not list, so it cannot be spread:

```ant
type Named = { name: string }
func rename(n: { name: string }): Named {
  Named{ ...n } // error: a spread needs a value of a named record type, got { name: string }
}
```

### A spread copies one level

A spread copies the fields, not the arrays, maps, or records they hold, so both records share them:

```ant
type Server = { addr: string, tags: Array[string] = [] }
let a = Server{ addr: ":1" }
let b = Server{ ...a, addr: ":2" }
b.tags.push("prod")
a.tags // ["prod"]
```

### Same fields, different types

Two types with the same fields are different types. A `Point` does not go where a `Vec` is expected;
convert with a spread:

```ant
type Point = { x: int, y: int }
type Vec = { x: int, y: int }
let v: Vec = Point{ x: 1, y: 2 } // error: cannot use Point where Vec is expected
```

## See also

[type](../keywords/type.md) · [Record](../../lib/Record/index.md) ·
[Record.same](../../lib/Record/same.md) · [Equality](../operators/equality.md) ·
[Array literals](arrays.md) · [Map literals](maps.md) · [Records tour](../../tour/10-records.md)
