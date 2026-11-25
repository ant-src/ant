<!-- status: reviewed -->

# Records

A record holds a fixed set of named fields under a type name. Use one when you know the fields as
you write the code: a point, a server's settings, a row of a report. This chapter covers declaring
and building records, field defaults, taking records apart with destructuring, and what `==`
compares.

## Declaring and constructing

```ant
type Point = { x: int, y: int }

let p = Point{ x: 1, y: 2 }
let x = 5
let q = Point{ y: 0, x }
println(p)
println(q)
println(q.x + p.y)
```

```text
Point{ x: 1, y: 2 }
Point{ x: 5, y: 0 }
7
```

A `type` declaration lists the fields and their types. A literal names the type and gives the fields
in any order; a record prints its fields in the order the type declares them. A bare name in a
literal, `x`, is short for `x: x`.

Every literal names its type. Ant has no anonymous record literals like JavaScript's `{ x: 1 }`. The
type checker rejects a literal that leaves out a field without a default (see
[Field defaults](#field-defaults) below) or names one the type does not list, which catches typos:

```ant
type Point = { x: int, y: int }

let a = Point{ x: 1 }       // error: Point literal is missing the field y
let b = Point{ x: 1, z: 2 } // error: Point has no field z
```

### Each declaration is its own type

Two declarations with the same fields are two types. A `Vec` does not go where a `Point` is
expected, and `==` does not compare them:

```ant
type Point = { x: int, y: int }
type Vec = { x: int, y: int }

let p = Point{ x: 1, y: 2 }
let v: Vec = p // error: cannot use Point where Vec is expected
println(p == Vec{ x: 1, y: 2 })
// error: Point and Vec cannot be compared with ==: neither is assignable to the other
```

To convert, spread one into a literal of the other: `Vec{ ...p }` copies every field of `p` into a
new `Vec`.

### Accepting any record with some fields

An anonymous record type in an annotation, such as `{ name: string }`, accepts any record that has
at least those fields, whatever its type name. Use it for a function that needs only a few fields:

```ant
type User = { name: string, email: string }
type Product = { name: string, price: int }

func label(item: { name: string }): string { "[${item.name}]" }

println(label(User{ name: "ann", email: "ann@example.com" }))
println(label(Product{ name: "tea", price: 4 }))
```

```text
[ann]
[tea]
```

A named type does not work this way. With `type Named = { name: string }`, a parameter of type
`Named` accepts only `Named` values, not a `User`.

To give a record type its own operations, such as `p.norm()`, declare
[functions on the type](functions.md#functions-on-types).

## Changing records

Fields are mutable. `let` stops you from reassigning the name, not from changing the record it
holds:

```ant
type Counter = { name: string, hits: int }
let home = Counter{ name: "home", hits: 0 }
home.hits += 1
home // Counter{ name: "home", hits: 1 }
```

A record is shared, not copied. Assigning a record to another name or passing it to a function
shares the one value, so a change through any name shows through all of them. If you come from Go,
where assigning a struct copies it, this is the difference to keep in mind:

```ant
type Account = { owner: string, balance: int }

func deposit(account: Account, amount: int) {
  account.balance += amount
}

let checking = Account{ owner: "ann", balance: 10 }
let alias = checking
deposit(alias, 5)
println(checking.balance)
```

```text
15
```

To change a copy and leave the original alone, build a new record with a spread. `...` copies every
field, and a later entry wins:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let moved = Point{ ...p, x: 10 }
moved // Point{ x: 10, y: 2 }
p     // Point{ x: 1, y: 2 }
```

The copy is shallow. A spread copies each field's value, so an array, map, or record in a field is
shared between the two records:

```ant
type Team = { name: string, members: Array[string] }
let core = Team{ name: "core", members: ["ann"] }
let docs = Team{ ...core, name: "docs" }
docs.members.push("bob")
core.members // ["ann", "bob"]
```

Give the copy its own array when the two must change apart:
`Team{ ...core, name: "docs", members: [...core.members] }`.

## Field defaults

A field in a `type` declaration may give a default after `=`. A literal may leave that field out,
and it gets the default:

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

A default runs each time a literal uses it, so every `Server` gets its own empty `tags` array:

```ant
type Server = { addr: string, tags: Array[string] = [] }
let a = Server{ addr: ":1" }
let b = Server{ addr: ":2" }
a.tags.push("prod")
b.tags // []
```

A default belongs to literals only. It does not change the type: every `Server` has every field, so
`s.workers` is a plain `int`, with no `null` to check. Ant has no zero values either: a field
without a default must be given in every literal, unlike Go, which fills in `0` or `""`.

A default runs in the scope where the type is declared, so it sees top-level names and imports but
not the record's other fields:

```ant
type Window = { start: int, end: int = start + 10 } // error: start is not defined
```

When one field depends on another, write a function that builds the record.

A field that a spread or an entry gives never runs its default. So a spread can turn a record into a
type with more fields, when the extra fields have defaults:

```ant
type Point = { x: float64, y: float64 }
type Point3 = { x: float64, y: float64, z: float64 = 0.0 }
let p = Point{ x: 1.0, y: 2.0 }
Point3{ ...p } // Point3{ x: 1.0, y: 2.0, z: 0.0 }
```

Choose defaults with care:

- Give a default to a field callers set independently of the others: a timeout, a flag, a list that
  starts empty.
- Keep fields that must agree, such as `min` and `max`, required, or add a function that builds the
  record and checks them together. A caller who sets only one could otherwise break the rule.
- Leave out the default on a field every caller should decide. When you add such a field, every
  literal fails to check until someone gives it.

Only the record body of a `type` declaration takes defaults. An anonymous record type cannot, since
no literal names it:

```ant
func connect(options: { host: string, port: int = 80 }): string { options.host }
// error: a field default belongs in the record body of a type declaration
```

## Destructuring

`let`, `var`, and `for` take a destructuring in place of a name. `{ name }` binds the field `name`,
`{ email: address }` binds the field `email` as `address`, and fields you do not mention are
ignored:

```ant
type Address = { city: string, zip: string }
type User = { name: string, email: string, address: Address }

let user = User{
  name: "ann",
  email: "ann@example.com",
  address: Address{ city: "Oslo", zip: "0150" },
}
let { name, email: contact } = user
let { address: { city } } = user
println("${name} <${contact}> in ${city}")
```

```text
ann <ann@example.com> in Oslo
```

A destructuring in a `for` loop takes each record apart as the loop reaches it:

```ant
type Score = { player: string, points: int }

let scores = [Score{ player: "ann", points: 12 }, Score{ player: "bob", points: 9 }]
for { player, points } in scores {
  println("${player}: ${points}")
}
```

```text
ann: 12
bob: 9
```

A record is also the way to return several values from a function. Name the result type, then
destructure the call. Here `bounds` returns `null` for an empty array, and the caller handles that
case before it destructures:

```ant
type Range = { low: int, high: int }

func bounds(xs: Array[int]): Range | null {
  let low = xs.min() ?? return null
  let high = xs.max() ?? return null
  Range{ low, high }
}

let { low, high } = bounds([3, 1, 4, 1, 5]) ?? panic("empty")
println("${low}..${high}")
```

```text
1..5
```

A record destructuring always declares new names. There is no record destructuring assignment like
JavaScript's `({ x, y } = p)`, because a `{` at the start of a statement is a lambda; write
`let { x, y } = p` instead. Array destructuring can assign, as in `[a, b] = [b, a]`.

Destructure a nullable record only after you narrow it. The checker reports a field read on a value
that may be `null`:

```ant
type User = { name: string }

func greet(user: User | null): string {
  let { name } = user // error: User | null has no field name: narrow the value first
  "hi ${name}"
}
```

Put a guard first, `if user == null { return "hi" }`, and the destructuring checks.

## Equality and identity

`==` compares records by structure: two records are equal when they have the same type and equal
fields. `Record.same` asks whether two names hold one record:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let q = Point{ x: 1, y: 2 }
let r = p
p == q            // true
Record.same(p, q) // false
Record.same(p, r) // true
```

If you come from JavaScript or Java, where `===` or the default `equals` compares identity, note
that Ant has no `===`, and `==` never compares identity for a record. The comparison goes into every
field, so records that hold arrays or other records compare by their contents:

```ant
type Team = { name: string, members: Array[string] }
Team{ name: "core", members: ["ann"] } == Team{ name: "core", members: ["ann"] } // true
Team{ name: "core", members: ["ann"] } == Team{ name: "core", members: ["bob"] } // false
```

Because a map and a set compare keys with `==`, a record works as a map key or a set member, and an
equal record finds it:

```ant
type Cell = { row: int, col: int }
let board = Map{ (Cell{ row: 0, col: 0 }): "x" }
let visited = Set.of([Cell{ row: 1, col: 2 }])
board[Cell{ row: 0, col: 0 }]       // "x"
visited.has(Cell{ row: 1, col: 2 }) // true
```

Do not change a record after you use it as a key or add it to a set. The map hashes a key when it is
inserted, so a changed key may no longer be found.

## Records in the standard library

The standard library uses records too, so everything above applies to them. `Error` is a record with
a `message` field, `Result` is a union of the records `Ok` and `Err`, and a caught `Panic` is a
record with a `message` and a `trace`:

```ant
let parsed: Result[int, Error] = Result.ok(1)
parsed == Ok{ value: 1 }                                // true
Error{ message: "disk full" }.message                   // "disk full"
Error.wrap(Error{ message: "disk full" }, "save notes") // Error{ message: "save notes: disk full", kind: null }
```

Destructuring works on them as on your own records:

```ant
let caught = Panic.catch(func(): int { [1, 2][5] })
if caught is Err {
  let { message } = caught.error
  println(message)
}
```

```text
index 5 out of range for an array of length 2
```

## Records, maps, or unions

| You have                                        | Use                             |
| ----------------------------------------------- | ------------------------------- |
| fields you know as you write the code           | a record                        |
| keys that are data, read or counted at run time | a `Map`                         |
| a value that is one of several shapes           | a [union](unions.md) of records |

A record has no dynamic access. You cannot index it by a field name or loop over its fields:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let field = "x"
println(p[field]) // error: a record has no dynamic access; use a Map when the keys are data
```

When the names of the keys come from input, such as headers or word counts, use a `Map`.

## See also

- [Record literals](../language/literals/records.md): every form of a literal
- [`type`](../language/keywords/type.md): declaring record types
- [`let`](../language/keywords/let.md) and [`for`](../language/keywords/for.md): destructuring
- [`==`](../language/operators/equality.md) and [Record.same](../lib/Record/same.md)
- [Functions on types](functions.md#functions-on-types), for `func Point.norm(p: Point)` and calling
  it as `p.norm()`
- [Unions](unions.md), for a value that is one of several records
- [Values](values.md), for how sharing works for every kind of value
- [Errors](errors.md), for `Result`, `Error`, and `Panic`
