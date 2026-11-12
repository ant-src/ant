<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Reflect.Scalar

```ant signature
type Reflect.Scalar = null | void | bool | int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64 | float32 | float64 | string
```

Reflect.Scalar is what [`Reflect.scalar`](../scalar.md) returns: a leaf value in its own kind.

<!-- end generated -->

## Details

- `int` and `byte` are in the union as `int64` and `uint8`, the types they alias.
- [`Reflect.scalar`](../scalar.md) returns the leaf in its own kind, so `is` tells the ten number
  types apart.

## Basic examples

Read a leaf out of a box; the result is a `Reflect.Scalar`:

```ant
Reflect.scalar(Reflect.valueOf(type(int), 42))       // 42
Reflect.scalar(Reflect.valueOf(type(string), "ant")) // "ant"
```

Pass a `Reflect.Scalar` to [`Reflect.makeScalar`](../makeScalar.md) to box it at a type known at run
time:

```ant
let leaf: Reflect.Scalar = 8080
Reflect.makeScalar(Reflect.erase(type(uint16)), leaf) // Reflect.Value(8080: uint16)
```

## Scope

A value of any leaf kind fits the type:

```ant
let leaves: Array[Reflect.Scalar] = [null, true, 3, 2.5, float32(0.5), byte(7), "x"]
leaves.length() // 7
```

The leaf keeps its own kind, so `is` tests it exactly. A `uint8` is not an `int`:

```ant
let s = Reflect.scalar(Reflect.valueOf(type(byte), byte(200)))
s is uint8 // true
s is int   // false
```

The `void` value comes from a function that returns nothing:

```ant
func reset() {}
let s = Reflect.scalar(Reflect.valueOf(type(void), reset()))
s is void // true
```

`null` is a member, so a parse that may fail fits the type as it is:

```ant
let parsed: Reflect.Scalar = "4x2".toInt()
parsed // null
```

## Applications

An encoder renders any leaf with `String.from`, and tests only the kinds it writes differently:

```ant
func encode(v: Reflect.Scalar): string {
  if v is string { return "\"${v}\"" }
  if v == null { return "null" }
  String.from(v)
}

println(encode(Reflect.scalar(Reflect.valueOf(type(float32), float32(0.5)))))
println(encode(Reflect.scalar(Reflect.valueOf(type(int64), 9))))
println(encode(Reflect.scalar(Reflect.valueOf(type(string), "hi"))))
println(encode(Reflect.scalar(Reflect.valueOf(type(int | null), null))))
```

```text
0.5
9
"hi"
null
```

A decoder parses text into the kind its slot needs, as a `Reflect.Scalar`, and leaves the range and
the exact type to `Reflect.makeScalar`. A parse that fails gives `null`, which `Reflect.makeScalar`
rejects:

```ant
func leaf(t: Reflect.ErasedType, text: string): Reflect.Value | null {
  let s = Reflect.shape(t)
  let v: Reflect.Scalar = if s is Reflect.Int {
    text.toInt()
  } else if s is Reflect.Float {
    text.toFloat()
  } else if s is Reflect.Bool {
    if text == "true" { true } else if text == "false" { false } else { null }
  } else {
    text
  }
  Reflect.makeScalar(t, v)
}

let port = Reflect.erase(type(uint16))
println(leaf(port, "8080"))
println(leaf(port, "-1"))
println(leaf(port, "http"))
println(leaf(Reflect.erase(type(bool)), "true"))
```

```text
Reflect.Value(8080: uint16)
null
null
Reflect.Value(true: bool)
```

## Properties and relations

Boxing a leaf at its box's type gives the box back:

```ant
let d = Reflect.valueOf(type(int16), int16(-3))
let s: Reflect.Scalar = Reflect.scalar(d)
assert(Reflect.makeScalar(Reflect.type(d), s) == d)
```

## Possible issues

A named primitive value is not a `Reflect.Scalar`. Use its base value, with a conversion such as
`string(id)`:

```ant
type UserId = string
let id = UserId("u-1")
let leaf: Reflect.Scalar = id // error: cannot use UserId where Reflect.Scalar is expected
```

A `Reflect.Scalar` is not a number, so arithmetic on it is a compile error. Narrow it with `is`
first, or take the value out with [`Reflect.to`](../to.md) when you know its type:

```ant
let s = Reflect.scalar(Reflect.valueOf(type(int), 2))
let n = s + 1 // error: + needs both operands of one type, got Reflect.Scalar and int
```

```ant
let d = Reflect.valueOf(type(int), 2)
let s = Reflect.scalar(d)
if s is int { s + 1 } else { 0 } // 3
Reflect.to(d, type(int)) + 1     // 3
```

A box of a named primitive type holds no `Reflect.Scalar`, so `Reflect.scalar` panics on it. Take
the base value out with [`Reflect.payload`](../payload.md) first:

```ant
type UserId = string
let d = Reflect.valueOf(type(UserId), UserId("u-1"))
Reflect.scalar(d)                  // panic: Reflect.scalar needs a value of a scalar type, got one of type UserId
Reflect.scalar(Reflect.payload(d)) // "u-1"
```

## See also

[Reflect.scalar](../scalar.md), [Reflect.makeScalar](../makeScalar.md), [Reflect.Value](Value.md),
[Reflect.Shape](Shape.md), [Reflect.payload](../payload.md)
