<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.makeScalar

```ant signature
func Reflect.makeScalar(t: Reflect.ErasedType, v: Reflect.Scalar): Reflect.Value | null
```

Reflect.makeScalar boxes v at the scalar type t, the inverse of Reflect.scalar: a decoder that holds
only a slot's erased type boxes every leaf through it. It returns null where t is not null, void,
bool, a number, or string, or where v is not of t's kind, so the one call tests that a leaf fits the
slot. A number converts to any number type as the conversion functions do, except that a conversion
that would panic returns null: a fraction, a value out of range, NaN, or an infinity for an integer
type, and a finite value beyond the range of a float32 for float32.

<!-- end generated -->

## Details

- The result is `null`, never a panic, when `t` is not a scalar type or `v` does not fit it. A
  union, a named primitive type such as `type UserId = string`, or a container is not a scalar type.
- `v` must have the kind of `t`: a `bool` for `bool`, a `string` for `string`, `null` for `null`,
  and for a number type any number. A string of digits is not a number.

## Basic examples

Box a value at a type known only at run time:

```ant
let port = Reflect.erase(type(uint16))
Reflect.makeScalar(port, 8080) // Reflect.Value(8080: uint16)
```

A value that does not fit the type gives `null`:

```ant
let small = Reflect.erase(type(int8))
Reflect.makeScalar(small, 300) // null
Reflect.makeScalar(small, "1") // null
```

## Scope

### Numbers

A whole float converts to an integer type, and a fraction does not:

```ant
let count = Reflect.erase(type(int))
Reflect.makeScalar(count, 2.0) // Reflect.Value(2: int)
Reflect.makeScalar(count, 2.5) // null
```

An integer converts to a float type:

```ant
Reflect.makeScalar(Reflect.erase(type(float64)), 3) // Reflect.Value(3.0: float64)
```

A float type rounds to the nearest value it holds:

```ant
Reflect.makeScalar(Reflect.erase(type(float32)), 16777217) // Reflect.Value(16777216.0: float32)
```

A negative number does not fit an unsigned type, and the full `uint64` range fits only `uint64`:

```ant
Reflect.makeScalar(Reflect.erase(type(byte)), -1)               // null
Reflect.makeScalar(Reflect.erase(type(byte)), 255)              // Reflect.Value(255: byte)
Reflect.makeScalar(Reflect.erase(type(uint64)), Math.maxUint64) // Reflect.Value(18446744073709551615: uint64)
Reflect.makeScalar(Reflect.erase(type(int)), Math.maxUint64)    // null
```

`NaN` and the infinities box as themselves at a float type, and give `null` at an integer type:

```ant
Reflect.makeScalar(Reflect.erase(type(float64)), Math.NaN)      // Reflect.Value(NaN: float64)
Reflect.makeScalar(Reflect.erase(type(float32)), Math.Infinity) // Reflect.Value(Infinity: float32)
Reflect.makeScalar(Reflect.erase(type(int)), Math.NaN)          // null
```

A finite value too large for `float32` gives `null` rather than an infinity:

```ant
Reflect.makeScalar(Reflect.erase(type(float32)), 1e300) // null
```

### Strings, booleans, `null`, and `void`

Each takes only a value of its own kind:

```ant
Reflect.makeScalar(Reflect.erase(type(string)), "ant") // Reflect.Value("ant": string)
Reflect.makeScalar(Reflect.erase(type(string)), 1)     // null
Reflect.makeScalar(Reflect.erase(type(bool)), true)    // Reflect.Value(true: bool)
Reflect.makeScalar(Reflect.erase(type(null)), null)    // Reflect.Value(null: null)
Reflect.makeScalar(Reflect.erase(type(void)), null)    // null
```

The `void` value comes from a function that returns nothing:

```ant
func reset() {}
Reflect.makeScalar(Reflect.erase(type(void)), reset()) // Reflect.Value(void: void)
```

### Types that are not scalar

A union, a named primitive type, and a container all give `null`, whatever the value:

```ant
type UserId = string
Reflect.makeScalar(Reflect.erase(type(int | null)), 1)      // null
Reflect.makeScalar(Reflect.erase(type(UserId)), "u-1")      // null
Reflect.makeScalar(Reflect.erase(type(Array[string])), "a") // null
```

### Optional values

`null` is a scalar, so a conversion that may fail passes its result straight in:

```ant
let count = Reflect.erase(type(int))
Reflect.makeScalar(count, "42".toInt())  // Reflect.Value(42: int)
Reflect.makeScalar(count, "4x2".toInt()) // null
```

### Many values

Box every value of an array at one type; the values that do not fit become `null`:

```ant
let level = Reflect.erase(type(byte))
[1, 200, 300].map({ Reflect.makeScalar(level, x) }) // [Reflect.Value(1: byte), Reflect.Value(200: byte), null]
```

## Applications

A JSON decoder handles the records it knows and gives every leaf to `Reflect.makeScalar`, which is
the whole test that the leaf fits its field:

```ant
type Point = { x: int16, y: int16, label: string }

func decode(t: Reflect.ErasedType, j: Json.Value): Reflect.Value | null {
  let s = Reflect.shape(t)
  if s is Reflect.Record {
    if !(j is Map) { return null }
    let fields: Map[string, Reflect.Value] = Map{}
    for f in s.fields {
      let v = j.get(f.name)
      if v == null { return null }
      let d = decode(f.type, v)
      if d == null { return null }
      fields[f.name] = d
    }
    return Reflect.makeRecord(t, fields)
  }
  if j is Array || j is Map { return null }
  Reflect.makeScalar(t, j)
}

let point = Reflect.erase(type(Point))
let inputs = [
  `{"x": 3, "y": -4, "label": "a"}`,
  `{"x": 3.5, "y": 0, "label": "b"}`,
  `{"x": 40000, "y": 0, "label": "c"}`,
  `{"x": 1, "y": 2, "label": 3}`,
]
for text in inputs {
  let d = decode(point, Json.parse(text).unwrap())
  if d == null {
    println("rejected ${text}")
  } else {
    println(Reflect.to(d, type(Point)))
  }
}
```

```text
Point{ x: 3, y: -4, label: "a" }
rejected {"x": 3.5, "y": 0, "label": "b"}
rejected {"x": 40000, "y": 0, "label": "c"}
rejected {"x": 1, "y": 2, "label": 3}
```

Read a setting written as text into a slot of any scalar type. Parse the text by the slot's shape,
then let `Reflect.makeScalar` narrow the number:

```ant
func setting(t: Reflect.ErasedType, text: string): Reflect.Value | null {
  let s = Reflect.shape(t)
  if s is Reflect.Int {
    let n = if s.signed { text.toInt() } else { text.toUint() }
    return Reflect.makeScalar(t, n)
  }
  if s is Reflect.Float { return Reflect.makeScalar(t, text.toFloat()) }
  Reflect.makeScalar(t, text)
}

println(setting(Reflect.erase(type(uint16)), "8080"))
println(setting(Reflect.erase(type(uint16)), "70000"))
println(setting(Reflect.erase(type(float32)), "0.5"))
println(setting(Reflect.erase(type(string)), "localhost"))
```

```text
Reflect.Value(8080: uint16)
null
Reflect.Value(0.5: float32)
Reflect.Value("localhost": string)
```

## Properties and relations

For a value of a scalar type `T`, `Reflect.makeScalar` gives the same box as
[`Reflect.valueOf`](valueOf.md):

```ant
let v = int8(5)
assert(Reflect.makeScalar(Reflect.erase(type(int8)), v) == Reflect.valueOf(type(int8), v))
```

It is the inverse of [`Reflect.scalar`](scalar.md): boxing a leaf at its own type and reading it
back gives the leaf:

```ant
let d = Reflect.valueOf(type(uint32), uint32(7))
let back = Reflect.makeScalar(Reflect.type(d), Reflect.scalar(d))
assert(back == d)
```

The box has the type `t`, not the type of `v`:

```ant
let d = Reflect.makeScalar(Reflect.erase(type(float32)), 2)
assert(d != null && Reflect.type(d) == Reflect.erase(type(float32)))
```

## Possible issues

The result may be `null`, so test it before you pass it on:

```ant
let d = Reflect.makeScalar(Reflect.erase(type(int)), 1)
let n = Reflect.to(d, type(int))
// error: cannot use Reflect.Value | null where Reflect.Value is expected
```

`t` is an erased type. Pass `type(T)` through [`Reflect.erase`](erase.md):

```ant
let d = Reflect.makeScalar(type(uint16), 5)
// error: cannot use Reflect.Type[uint16] where Reflect.ErasedType is expected
```

A named primitive type is not scalar, so its values give `null`. Box the base value, then wrap it
with [`Reflect.makeTagged`](makeTagged.md):

```ant
type Celsius = float64
let t = Reflect.erase(type(Celsius))
println(Reflect.makeScalar(t, 21.5))
let s = Reflect.shape(t)
if s is Reflect.Tagged {
  let base = Reflect.makeScalar(s.base, 21.5)
  if base != null { println(Reflect.makeTagged(t, base)) }
}
```

```text
null
Reflect.Value(Celsius(21.5): Celsius)
```

A union is not scalar either, even when the value is one of its members. Pick the member first, from
the union's shape, and box the value at it:

```ant
let t = Reflect.erase(type(int | null))
Reflect.makeScalar(t, 7) // null
```

Box the value at each member and keep the first box that fits:

```ant
let t = Reflect.erase(type(int | null))
let s = Reflect.shape(t)
if s is Reflect.Union {
  println(s.members | Array.map({ Reflect.makeScalar(x, 7) }) | Array.find({ x != null }))
}
```

```text
Reflect.Value(7: int)
```

## See also

[Reflect.scalar](scalar.md), [Reflect.valueOf](valueOf.md), [Reflect.makeTagged](makeTagged.md),
[Reflect.makeRecord](makeRecord.md), [Reflect.shape](shape.md), [Reflect.erase](erase.md),
[Reflect.Scalar](types/Scalar.md)
