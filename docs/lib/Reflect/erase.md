<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.erase

```ant signature
func Reflect.erase[T](t: Reflect.Type[T]): Reflect.ErasedType
```

Reflect.erase returns t with its static type forgotten, for code that walks types. The conversion is
always written: a Reflect.Type\[T\] is not assignable to a Reflect.ErasedType.

<!-- end generated -->

## Details

- The erased value denotes the same type as `t` and renders the same way. Only the static type is
  gone.
- The functions that take a type apart or build values from one, such as `Reflect.shape`,
  `Reflect.assignable`, and `Reflect.makeRecord`, take a `Reflect.ErasedType`. Pass them
  `Reflect.erase(type(T))`, never `type(T)` itself.
- The children `Reflect.shape` returns are already erased, so a recursive walk calls `Reflect.erase`
  once, at its entry.

## Basic examples

```ant
type Config = { name: string, port: int = 8080 }
Reflect.erase(type(Array[int])) // Array[int]
Reflect.erase(type(Config))     // Config
```

Give the result to a function that walks types:

```ant
type Config = { name: string, port: int = 8080 }
let t = Reflect.erase(type(Config))
Reflect.shape(t) is Reflect.Record // true
```

## Scope

Any type a `type(...)` expression can hold erases. The result renders in annotation syntax, with an
alias by its name and `int64` and `uint8` as `int` and `byte`:

```ant
type Ids = Array[int]
Reflect.erase(type(int | null))         // int | null
Reflect.erase(type(Ids))                // Ids
Reflect.erase(type({ x: int, y: int })) // { x: int, y: int }
Reflect.erase(type(func(int): string))  // func(int): string
Reflect.erase(type(Result[int, Error])) // Result[int, Error]
Reflect.erase(type(Json.Value))         // Json.Value
Reflect.erase(type(Bytes))              // Bytes
Reflect.erase(type(uint8))              // byte
```

The builders keep the static type, so erase what they return:

```ant
type Point = { x: float64, y: float64 }
Reflect.erase(Reflect.arrayOf(type(Point)))           // Array[Point]
Reflect.erase(Reflect.mapOf(type(string), type(int))) // Map[string, int]
```

Call it through `.` on a type value, or end a pipe with it:

```ant
let t = type(Map[string, int])
t.erase()                   // Map[string, int]
type(int) | Reflect.erase() // int
```

## Applications

Choose a type at run time. `Reflect.Type[int64]` and `Reflect.Type[int32]` are two types that one
`if` cannot return, so erase each branch. `int64` renders as `int`:

```ant
func counterType(wide: bool): Reflect.ErasedType {
  if wide { Reflect.erase(type(int64)) } else { Reflect.erase(type(int32)) }
}

println(counterType(true))
println(counterType(false))
```

```text
int
int32
```

Walk a record type and print its fields. The function erases its argument once and recurses on the
erased children:

```ant
type Point = { x: float64, y: float64 }
type Line = { from: Point, to: Point, label: string | null, tags: Array[string] }

func describe(t: Reflect.ErasedType, depth: int): void {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    for f in s.fields {
      println("${"  ".repeat(depth)}${f.name}: ${f.type}")
      describe(f.type, depth + 1)
    }
  }
}

describe(Reflect.erase(type(Line)), 0)
```

```text
from: Point
  x: float64
  y: float64
to: Point
  x: float64
  y: float64
label: string | null
tags: Array[string]
```

Keep the typed form at the two ends only. `zero` takes a `Reflect.Type[T]` and returns a `T`; in
between, `zeroOf` works on erased types and boxed values, and `Reflect.to` brings the static type
back:

```ant
type Point = { x: float64, y: float64 }
type Shape = { name: string, visible: bool, points: Array[Point], origin: Point }

func zeroOf(t: Reflect.ErasedType): Reflect.Value {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    let fields = s.fields.map({ Entry{ key: x.name, value: zeroOf(x.type) } })
    return Reflect.makeRecord(t, Map.fromEntries(fields))
  }
  if s is Reflect.Array {
    return Reflect.makeArray(t, [])
  }
  let scalar: Reflect.Scalar = if s is Reflect.String { "" } else if s is Reflect.Bool { false } else { 0 }
  let v = Reflect.makeScalar(t, scalar)
  if v == null {
    panic("no zero value for ${t}")
  }
  v
}

func zero[T](t: Reflect.Type[T]): T {
  Reflect.to(zeroOf(Reflect.erase(t)), t)
}

let shape = zero(type(Shape))
println(shape)
println(shape.origin.x + 1.5)
```

```text
Shape{ name: "", visible: false, points: [], origin: Point{ x: 0.0, y: 0.0 } }
1.5
```

Key a table by erased types, so one map holds an entry for each type. Equal types find the same
entry:

```ant
let names: Map[Reflect.ErasedType, string] = Map{}
names[Reflect.erase(type(int))] = "integer"
names[Reflect.erase(type(string))] = "text"
println(names[Reflect.erase(type(int64))])
println(names.get(Reflect.erase(type(bool))))
```

```text
integer
null
```

Handle an opaque type specially by comparing erased type values, not names:

```ant
func encoding(t: Reflect.ErasedType): string {
  if t == Reflect.erase(type(Bytes)) { "base64" } else { "json" }
}

println(encoding(Reflect.erase(type(Bytes))))
println(encoding(Reflect.erase(type(Array[byte]))))
```

```text
base64
json
```

## Properties and relations

Two erased values are `==` when they denote the same type: aliases expand and union members compare
as a set, whatever the rendering says:

```ant
type Ids = Array[int]
let a = Reflect.erase(type(Ids))
let b = Reflect.erase(type(Array[int]))
assert(a == b)
assert(String.from(a) != String.from(b))
assert(Reflect.erase(type(int | null)) == Reflect.erase(type(null | int)))
assert(Reflect.erase(type(int)) == Reflect.erase(type(int64)))
```

Erasing changes nothing but the static type, so the rendering matches:

```ant
type Point = { x: float64, y: float64 }
let t = type(Array[Point])
assert(String.from(Reflect.erase(t)) == String.from(t))
```

`Reflect.type` of a boxed value is the erased type the value was boxed at, with aliases expanded and
a union resolved to the value's member:

```ant
type Ids = Array[int]
assert(String.from(Reflect.type(Reflect.valueOf(type(Ids), [1]))) == "Array[int]")
assert(Reflect.type(Reflect.valueOf(type(string), "a")) == Reflect.erase(type(string)))
assert(Reflect.type(Reflect.valueOf(type(int | null), 42)) == Reflect.erase(type(int)))
```

## Possible issues

A `Reflect.Type[T]` is not a `Reflect.ErasedType`, so passing `type(T)` where an erased type is
expected is a compile error:

```ant
println(Reflect.shape(type(int))) // error: cannot use Reflect.Type[int] where Reflect.ErasedType is expected
```

For the same reason, a type value and an erased one cannot be compared. Erase both sides:

```ant
println(type(int) == Reflect.erase(type(int)))
// error: Reflect.Type[int] and Reflect.ErasedType cannot be compared with ==: neither is assignable to the other
```

Two different `Reflect.Type` instances cannot meet in one array or one `if`, since no union holds
two of them. Erase each one:

```ant
let ts = [type(int), type(string)]
// error: the elements do not form a valid type
```

```ant
let ts = [Reflect.erase(type(int)), Reflect.erase(type(string))]
println(ts)
```

```text
[int, string]
```

A declaration with a primitive body is a new type, not an alias, so its erased type is not `==` to
its base:

```ant
type UserId = int
Reflect.erase(type(UserId)) == Reflect.erase(type(int)) // false
```

## See also

[Reflect](index.md) · [Reflect.Type](types/Type.md) · [Reflect.ErasedType](types/ErasedType.md) ·
[Reflect.shape](shape.md) · [Reflect.to](to.md) · [Reflect.type](type.md) ·
[Reflect.arrayOf](arrayOf.md) · [Reflect.assignable](assignable.md)
