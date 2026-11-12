<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.payload

```ant signature
func Reflect.payload(d: Reflect.Value): Reflect.Value
```

Reflect.payload returns the base value of the named primitive in d, boxed at the base type. It
panics if d's type is not a named primitive type.

<!-- end generated -->

## Details

- The box has the base type, `string`, `bool`, or a numeric kind, never the named type. It is the
  `base` of the type's [`Reflect.Tagged`](types/Shape.md) shape.

## Basic examples

Unwrap a boxed `UserId` to its string:

```ant
type UserId = string
let d = Reflect.valueOf(type(UserId), UserId("u-1"))
d                  // Reflect.Value(UserId("u-1"): UserId)
Reflect.payload(d) // Reflect.Value("u-1": string)
```

Take the payload out with [`Reflect.to`](to.md), at the base type:

```ant
type Meters = float64
let d = Reflect.valueOf(type(Meters), Meters(2.5))
Reflect.to(Reflect.payload(d), type(float64)) // 2.5
```

## Scope

The payload is boxed at the exact base type, here `uint16`. The constructor takes a `uint16`, so the
literal is converted first:

```ant
type Port = uint16
let d = Reflect.valueOf(type(Port), Port(uint16(8080)))
d.payload()        // Reflect.Value(8080: uint16)
d.payload().type() // uint16
```

A named primitive over `bool`:

```ant
type Enabled = bool
Reflect.valueOf(type(Enabled), Enabled(true)).payload() // Reflect.Value(true: bool)
```

A generic named primitive keeps its type arguments on the tagged value, and drops them on the
payload:

```ant
type User = { name: string }
type Id[T] = string
let id: Id[User] = Id("ann")
let d = Reflect.valueOf(type(Id[User]), id)
d           // Reflect.Value(Id("ann"): Id[User])
d.payload() // Reflect.Value("ann": string)
```

A box made at a union type holds the member, so the payload is there when the value is tagged:

```ant
type UserId = string
let d = Reflect.valueOf(type(UserId | null), UserId("u-1"))
Reflect.type(d) // UserId
d.payload()     // Reflect.Value("u-1": string)
```

A record field of a named primitive type comes out of [`Reflect.field`](field.md) tagged, ready for
`payload`:

```ant
type UserId = string
type Session = { user: UserId, ttl: int }
let d = Reflect.valueOf(type(Session), Session{ user: UserId("u-7"), ttl: 60 })
d.field("user").payload() // Reflect.Value("u-7": string)
```

Call it in a pipe, and end with [`Reflect.scalar`](scalar.md) for the bare value:

```ant
type Meters = float64
let d = Reflect.valueOf(type(Meters), Meters(3.0))
d | Reflect.payload() | Reflect.scalar() // 3.0
```

## Applications

Render any record as `key=value` pairs, unwrapping tagged fields to their payload:

```ant
type UserId = string
type Port = uint16
type Conn = { user: UserId, port: Port, secure: bool }

func plain(d: Reflect.Value): Reflect.Scalar {
  let tagged = d.type().shape() is Reflect.Tagged
  if tagged { d.payload().scalar() } else { d.scalar() }
}

func render(d: Reflect.Value): string {
  let s = d.type().shape()
  if !(s is Reflect.Record) {
    return String.from(plain(d))
  }
  s.fields
    | Array.map({ "${x.name}=${plain(d.field(x.name))}" })
    | String.join(" ")
}

let conn = Conn{ user: UserId("ann"), port: Port(uint16(443)), secure: true }
println(render(Reflect.valueOf(type(Conn), conn)))
```

```text
user=ann port=443 secure=true
```

Decode text into a named primitive over `string` or an integer kind, the inverse of `payload`: box
the payload at the base with [`Reflect.makeScalar`](makeScalar.md), then tag it with
[`Reflect.makeTagged`](makeTagged.md):

```ant
type Port = uint16
type Name = string

func decode(t: Reflect.ErasedType, text: string): Reflect.Value | null {
  let s = Reflect.shape(t)
  if !(s is Reflect.Tagged) {
    return null
  }
  let base = Reflect.shape(s.base)
  let raw: Reflect.Scalar = if base is Reflect.String { text } else { text.toInt() ?? return null }
  let payload = Reflect.makeScalar(s.base, raw) ?? return null
  Reflect.makeTagged(t, payload)
}

println(decode(Reflect.erase(type(Port)), "8080"))
println(decode(Reflect.erase(type(Port)), "70000"))
println(decode(Reflect.erase(type(Name)), "ant"))
```

```text
Reflect.Value(Port(8080): Port)
null
Reflect.Value(Name("ant"): Name)
```

## Properties and relations

The payload's type is the `base` of the type's shape:

```ant
type Meters = float64
let d = Reflect.valueOf(type(Meters), Meters(1.5))
let s = d.type().shape()
assert(s is Reflect.Tagged)
if s is Reflect.Tagged {
  assert(d.payload().type() == s.base)
}
```

It unwraps as converting to the base type does:

```ant
type UserId = string
let id = UserId("u-1")
let d = Reflect.valueOf(type(UserId), id)
assert(Reflect.to(d.payload(), type(string)) == string(id))
```

[`Reflect.makeTagged`](makeTagged.md) puts the tag back, giving an equal value:

```ant
type User = { name: string }
type Id[T] = string
let id: Id[User] = Id("ann")
let d = Reflect.valueOf(type(Id[User]), id)
assert(Reflect.makeTagged(d.type(), d.payload()) == d)
```

## Possible issues

A box of the base type itself has no tag, and panics:

```ant
Reflect.valueOf(type(string), "u-1").payload()
// panic: Reflect.payload needs a value of a named primitive type, got one of type string
```

So does a record, even one with a single field:

```ant
type Wrapped = { value: string }
let d = Reflect.valueOf(type(Wrapped), Wrapped{ value: "u-1" })
d.payload() // panic: Reflect.payload needs a value of a named primitive type, got one of type Wrapped
```

A box made at a nullable type can hold `null`, as an optional field does, and `payload` panics on
it:

```ant
type UserId = string
Reflect.valueOf(type(UserId | null), null).payload()
// panic: Reflect.payload needs a value of a named primitive type, got one of type null
```

Check the shape of the value's type first:

```ant
type UserId = string

func userId(d: Reflect.Value): Reflect.Value | null {
  if d.type().shape() is Reflect.Tagged { d.payload() } else { null }
}

println(userId(Reflect.valueOf(type(UserId | null), null)))
println(userId(Reflect.valueOf(type(UserId | null), UserId("u-1"))))
```

```text
null
Reflect.Value("u-1": string)
```

A named primitive type is not assignable to its base, so `Reflect.to` at the base panics on the
tagged box. Unwrap it with `payload` first:

```ant
type UserId = string
let d = Reflect.valueOf(type(UserId), UserId("u-1"))
Reflect.to(d, type(string))           // panic: Reflect.to: UserId is not assignable to string
Reflect.to(d.payload(), type(string)) // "u-1"
```

## See also

[Reflect.makeTagged](makeTagged.md) · [Reflect.shape](shape.md) · [Reflect.field](field.md) ·
[Reflect.scalar](scalar.md) · [Reflect.to](to.md) · [Reflect.valueOf](valueOf.md)
