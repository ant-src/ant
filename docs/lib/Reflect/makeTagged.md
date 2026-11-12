<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeTagged

```ant signature
func Reflect.makeTagged(t: Reflect.ErasedType, payload: Reflect.Value): Reflect.Value
```

Reflect.makeTagged returns a new value of the named primitive type t over payload. It panics if t is
not a named primitive type, or payload's type is not its base.

<!-- end generated -->

## Details

- `Reflect.makeTagged` does not convert the payload: a number of another numeric type panics. Box it
  at the base first, for example with `Reflect.makeScalar`.
- `Reflect.makeTagged` builds the value without calling any function declared on the type, so a
  check such a function makes does not run.

## Basic examples

Box a payload at the base type, then wrap it in the named type:

```ant
type UserId = string
let payload = Reflect.valueOf(type(string), "u-1")
Reflect.makeTagged(type(UserId).erase(), payload) // Reflect.Value(UserId("u-1"): UserId)
```

Take the typed value out with `Reflect.to`. It is equal to the one the constructor makes:

```ant
type Celsius = float64
let t = type(Celsius).erase()
let d = Reflect.makeTagged(t, Reflect.valueOf(type(float64), 19.5))
d.to(type(Celsius))                  // Celsius(19.5)
d.to(type(Celsius)) == Celsius(19.5) // true
```

## Scope

Any primitive base works: a string, a number of any size, or `bool`. Box a sized number with a
conversion:

```ant
type Sku = string
type Port = uint16
type Ratio = float32
type Flag = bool
let sku = Reflect.valueOf(type(string), "A-100")
let port = Reflect.valueOf(type(uint16), uint16(8080))
let ratio = Reflect.valueOf(type(float32), float32(0.5))
let flag = Reflect.valueOf(type(bool), true)
Reflect.makeTagged(type(Sku).erase(), sku)     // Reflect.Value(Sku("A-100"): Sku)
Reflect.makeTagged(type(Port).erase(), port)   // Reflect.Value(Port(8080): Port)
Reflect.makeTagged(type(Ratio).erase(), ratio) // Reflect.Value(Ratio(0.5): Ratio)
Reflect.makeTagged(type(Flag).erase(), flag)   // Reflect.Value(Flag(true): Flag)
```

For a generic named primitive, `Reflect.makeTagged` builds the value at the type arguments that `t`
names. `Id[User]` and `Id[Account]` share a tag and are still different types:

```ant
type Id[T] = string
type User = { name: string }
type Account = { number: int }
let d = Reflect.makeTagged(type(Id[User]).erase(), Reflect.valueOf(type(string), "u-1"))
d.type()                                                // Id[User]
Reflect.assignable(d.type(), type(Id[Account]).erase()) // false
```

The type often comes from a shape rather than from `type(...)`. Here it is the type of a record
field:

```ant
type OrderId = string
type Order = { id: OrderId, total: int }
let s = Reflect.shape(type(Order).erase())
let id = if s is Reflect.Record { s.fields[0].type } else { panic("not a record") }
Reflect.makeTagged(id, Reflect.valueOf(type(string), "o-7")) // Reflect.Value(OrderId("o-7"): OrderId)
```

Name the arguments, in any order:

```ant
type UserId = string
let payload = Reflect.valueOf(type(string), "u-1")
Reflect.makeTagged(payload: payload, t: type(UserId).erase()) // Reflect.Value(UserId("u-1"): UserId)
```

## Applications

Fill a named primitive field of a record built from parts. The field's value is a tagged value, not
its base:

```ant
type Celsius = float64
type Reading = { sensor: string, value: Celsius }
let d = Reflect.makeRecord(type(Reading).erase(), Map{
  sensor: Reflect.valueOf(type(string), "a1"),
  value: Reflect.makeTagged(type(Celsius).erase(), Reflect.valueOf(type(float64), 21.5)),
})
d.to(type(Reading)).value // Celsius(21.5)
```

A decoder that holds only a slot's erased type builds a named primitive from its base. It decodes
the base, here with `Reflect.makeScalar`, and wraps the result:

```ant
type OrderId = string
type Cents = int64

func decode(t: Reflect.ErasedType, raw: Reflect.Scalar): Reflect.Value | null {
  let s = Reflect.shape(t)
  if s is Reflect.Tagged {
    let base = decode(s.base, raw)
    if base == null { return null }
    return Reflect.makeTagged(t, base)
  }
  Reflect.makeScalar(t, raw)
}

println(decode(type(OrderId).erase(), "o-7"))
println(decode(type(Cents).erase(), 1250))
println(decode(type(Cents).erase(), "1250"))
```

```text
Reflect.Value(OrderId("o-7"): OrderId)
Reflect.Value(Cents(1250): Cents)
null
```

A transformation written once for every named primitive over `string` reads the payload, changes it,
and wraps it in the same type again:

```ant
type Email = string
type Name = string

func trimmed(d: Reflect.Value): Reflect.Value {
  let s = Reflect.shape(d.type())
  if !(s is Reflect.Tagged) || s.base != type(string).erase() { return d }
  let text = d.payload().to(type(string)).trim()
  Reflect.makeTagged(d.type(), Reflect.valueOf(type(string), text))
}

println(trimmed(Reflect.valueOf(type(Email), Email(" ann@example.com "))))
println(trimmed(Reflect.valueOf(type(Name), Name("Ann  "))))
println(trimmed(Reflect.valueOf(type(int), 7)))
```

```text
Reflect.Value(Email("ann@example.com"): Email)
Reflect.Value(Name("Ann"): Name)
Reflect.Value(7: int)
```

## Properties and relations

`Reflect.payload` undoes `Reflect.makeTagged`, and the new value has exactly the type `t`:

```ant
type UserId = string
let t = type(UserId).erase()
let p = Reflect.valueOf(type(string), "u-1")
let d = Reflect.makeTagged(t, p)
assert(d.payload() == p)
assert(d.type() == t)
```

Wrapping the payload of a tagged value again gives an equal value:

```ant
type Celsius = float64
let d = Reflect.valueOf(type(Celsius), Celsius(21.5))
assert(Reflect.makeTagged(d.type(), d.payload()) == d)
```

## Possible issues

A built-in type is not a named primitive type, and neither is a record or a union:

```ant
type User = { name: string }
let payload = Reflect.valueOf(type(string), "u-1")
Reflect.makeTagged(type(string).erase(), payload) // panic: Reflect.makeTagged needs a named primitive type, got string
Reflect.makeTagged(type(User).erase(), payload)   // panic: Reflect.makeTagged needs a named primitive type, got User
```

A payload that is already tagged is not its base. Unwrap it with `Reflect.payload` first:

```ant
type UserId = string
let t = type(UserId).erase()
let d = Reflect.valueOf(type(UserId), UserId("u-1"))
Reflect.makeTagged(t, d)
// panic: Reflect.makeTagged: a payload of type UserId is not assignable to the base type string of UserId
Reflect.makeTagged(t, d.payload()) // Reflect.Value(UserId("u-1"): UserId)
```

`Reflect.makeTagged` does not convert a number of another kind. Box it at the base with
`Reflect.makeScalar`, which converts it:

```ant
type Celsius = float64
let t = type(Celsius).erase()
let n = Reflect.valueOf(type(int), 21)
Reflect.makeTagged(t, n)
// panic: Reflect.makeTagged: a payload of type int is not assignable to the base type float64 of Celsius
let f = Reflect.makeScalar(type(float64).erase(), 21) ?? panic("21 is not a float64")
Reflect.makeTagged(t, f) // Reflect.Value(Celsius(21.0): Celsius)
```

A function declared on the type does not run, so a value it would reject can be built:

```ant
type Percent = int

func Percent.of(n: int): Percent | null {
  if n < 0 || n > 100 { return null }
  Percent(n)
}

println(Percent.of(150))
let d = Reflect.makeTagged(type(Percent).erase(), Reflect.valueOf(type(int), 150))
println(d.to(type(Percent)))
```

```text
null
Percent(150)
```

## See also

[Reflect.payload](payload.md), [Reflect.makeScalar](makeScalar.md),
[Reflect.makeRecord](makeRecord.md), [Reflect.shape](shape.md), [Reflect.valueOf](valueOf.md),
[Reflect.to](to.md), [Reflect.Shape](types/Shape.md)
