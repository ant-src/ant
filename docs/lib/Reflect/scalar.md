<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.scalar

```ant signature
func Reflect.scalar(d: Reflect.Value): Reflect.Scalar
```

Reflect.scalar returns the value in d in its own kind. It panics if d's type is not null, void,
bool, a number, or string.

<!-- end generated -->

## Details

- The result has the value's own kind, not a common one: a box of type `uint8` gives a `uint8`, and
  a box of type `float32` gives a `float32`. Test the kind with `is` before using it as a number.
- A box made at a union holds the member its value belongs to, so a box of type `int | null` gives
  the `int` or the `null` inside it.
- It panics on any other type: a named primitive type such as `type Celsius = float64`, a record, a
  container, or an opaque type such as `Bytes`. Unwrap a named primitive with
  [`Reflect.payload`](payload.md) first.
- `int` and `byte` are aliases of `int64` and `uint8`, so a box of type `int` gives a value that is
  both `int` and `int64`.

## Basic examples

Take a leaf out of a box:

```ant
Reflect.scalar(Reflect.valueOf(type(int), 42))       // 42
Reflect.scalar(Reflect.valueOf(type(string), "ant")) // "ant"
```

The value keeps its own number type:

```ant
let v = Reflect.scalar(Reflect.valueOf(type(uint8), uint8(7)))
v is uint8 // true
v is int   // false
```

## Scope

### Every scalar kind

Booleans, sized integers, floats, `null`, and `void` all come out as themselves:

```ant
func reset() {}
Reflect.scalar(Reflect.valueOf(type(bool), true))                // true
Reflect.scalar(Reflect.valueOf(type(float32), float32(1.5)))     // 1.5
Reflect.scalar(Reflect.valueOf(type(float64), 2.5))              // 2.5
Reflect.scalar(Reflect.valueOf(type(int64), int64(-3))) is int64 // true
Reflect.scalar(Reflect.valueOf(type(int), 1)) is int64           // true
Reflect.scalar(Reflect.valueOf(type(null), null))                // null
Reflect.scalar(Reflect.valueOf(type(void), reset())) is void     // true
```

### Boxes made at a union

A box made at `int | null` holds whichever member the value is:

```ant
let some: int | null = 3
let none: int | null = null
Reflect.scalar(Reflect.valueOf(type(int | null), some)) // 3
Reflect.scalar(Reflect.valueOf(type(int | null), none)) // null
```

### Through `.` and pipes

`Reflect.scalar` takes the box first, so it ends a chain of readers:

```ant
type Point = { x: int, y: int }
let d = Reflect.valueOf(type(Point), Point{ x: 1, y: 2 })
d.field("y").scalar()                     // 2
d | Reflect.field("x") | Reflect.scalar() // 1
```

### Named primitive types

A named primitive type is not a scalar type. Read its payload, which has the base type:

```ant
type Celsius = float64
let d = Reflect.valueOf(type(Celsius), Celsius(21.5))
d.payload().scalar() // 21.5
```

## Applications

Write any record as a CSV row. Each field's leaf goes through `String.from`, whatever its number
type:

```ant
type Reading = { sensor: string, value: float32, count: uint16, ok: bool }

func csvRow(d: Reflect.Value): string {
  let s = Reflect.shape(Reflect.type(d))
  if !(s is Reflect.Record) {
    return String.from(d.scalar())
  }
  s.fields | Array.map({ String.from(d.field(x.name).scalar()) }) | String.join(",")
}

let r = Reading{ sensor: "a1", value: float32(21.5), count: uint16(3), ok: true }
println(csvRow(Reflect.valueOf(type(Reading), r)))
```

```text
a1,21.5,3,true
```

Render a leaf as JSON text, testing only the kinds that need their own form:

```ant
func jsonLeaf(d: Reflect.Value): string {
  let v = d.scalar()
  if v is string { return Json.stringify(v) }
  if v == null || v is void { return "null" }
  String.from(v)
}

func reset() {}
let leaves = [
  Reflect.valueOf(type(string), "say \"hi\""),
  Reflect.valueOf(type(uint64), Math.maxUint64),
  Reflect.valueOf(type(float32), float32(0.1)),
  Reflect.valueOf(type(bool), false),
  Reflect.valueOf(type(void), reset()),
]
for d in leaves {
  println(jsonLeaf(d))
}
```

```text
"say \"hi\""
18446744073709551615
0.1
false
null
```

## Properties and relations

For a scalar type, `Reflect.scalar` gives the same value as [`Reflect.to`](to.md) at that type:

```ant
let d = Reflect.valueOf(type(int16), int16(-5))
assert(Reflect.scalar(d) == Reflect.to(d, type(int16)))
```

[`Reflect.makeScalar`](makeScalar.md) is its inverse: boxing the leaf at the box's type gives an
equal box:

```ant
let d = Reflect.valueOf(type(float64), 2.5)
assert(Reflect.makeScalar(Reflect.type(d), d.scalar()) == d)
```

## Possible issues

The result is a `Reflect.Scalar`, a union, so it is not a number until you test its kind:

```ant
let d = Reflect.valueOf(type(int), 1)
let n = Reflect.scalar(d) + 1
// error: + needs both operands of one type, got Reflect.Scalar and int
```

```ant
let d = Reflect.valueOf(type(int), 1)
let v = d.scalar()
if v is int { v + 1 } else { 0 } // 2
```

A `uint8` is not an `int`, so a test for `int` alone misses the other number types:

```ant
let v = Reflect.valueOf(type(uint8), uint8(200)).scalar()
if v is int { "int" } else { "something else" } // "something else"
```

A named primitive type panics, even when its base is a scalar:

```ant
type Celsius = float64
Reflect.scalar(Reflect.valueOf(type(Celsius), Celsius(21.5)))
// panic: Reflect.scalar needs a value of a scalar type, got one of type Celsius
```

So do records and containers. Test the shape first when the type is not known:

```ant
Reflect.scalar(Reflect.valueOf(type(Array[int]), [1]))
// panic: Reflect.scalar needs a value of a scalar type, got one of type Array[int]
```

```ant
let d = Reflect.valueOf(type(Array[int]), [1])
let s = Reflect.shape(Reflect.type(d))
let leaf = s is Reflect.Bool || s is Reflect.Int || s is Reflect.Float || s is Reflect.String
if leaf { String.from(d.scalar()) } else { "not a leaf" } // "not a leaf"
```

## See also

[Reflect.makeScalar](makeScalar.md) · [Reflect.to](to.md) · [Reflect.payload](payload.md) ·
[Reflect.field](field.md) · [Reflect.items](items.md) · [Reflect.shape](shape.md) ·
[Reflect.valueOf](valueOf.md) · [Reflect.Scalar](types/Scalar.md)
