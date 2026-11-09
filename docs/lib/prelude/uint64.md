<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# uint64

```ant signature
func uint64[T: Number](x: T): uint64
```

uint64 converts x to uint64 exactly. It panics if x is out of range, has a fraction, or is NaN or infinite; uint64(Math.trunc(x)) truncates. x may also be a named primitive over a number.

<!-- end generated -->

## Details

- An integer literal is always an `int`, so a `uint64` comes from a conversion, such as
  `uint64(42)`, or from a function that returns one, such as `String.toUint`.
- To drop a fraction, round first with `Math.trunc`, `Math.floor`, `Math.ceil`, or `Math.round`.
- Arithmetic on two `uint64` values stays a `uint64` and panics on overflow, including subtraction
  below zero. `Bits.wrapAdd`, `Bits.wrapSub`, and `Bits.wrapMul` wrap modulo 2^64 instead.

## Basic examples

Convert an `int` that is not negative:

```ant
uint64(42) // 42
uint64(0)  // 0
```

Convert a float that holds a whole number:

```ant
uint64(3.0)  // 3
uint64(1e19) // 10000000000000000000
```

Convert a value before mixing it with a `uint64`:

```ant
let total = uint64(100)
total + uint64(byte(7)) // 107
```

## Scope

Every numeric kind converts, as long as the value is in range:

```ant
uint64(int8(127))          // 127
uint64(uint32(4000000000)) // 4000000000
uint64(Math.maxInt)        // 9223372036854775807
uint64(float32(1e10))      // 10000000000
uint64(-0.0)               // 0
```

The range goes past `Math.maxInt`, up to `Math.maxUint64`:

```ant
Math.maxUint64                              // 18446744073709551615
uint64(Math.maxInt) * uint64(2) + uint64(1) // 18446744073709551615
uint64(Math.pow(2.0, 63.0))                 // 9223372036854775808
```

A value of a named primitive type loses its tag:

```ant
type UserId = uint64
let id = UserId(uint64(7))
id         // UserId(7)
uint64(id) // 7
```

`uint64` is a function value, so it can be passed to `Array.map`:

```ant
[1, 2, 3] | Array.map(uint64) // [1, 2, 3]
```

## Applications

Hash a string with 64-bit FNV-1a, which relies on multiplication that wraps:

```ant
func fnv1a(s: string): uint64 {
  var h = String.toUint("14695981039346656037") ?? panic("bad offset")
  let prime = uint64(1099511628211)
  for b in String.toBytes(s) {
    h = Bits.wrapMul(Bits.xor(h, uint64(b)), prime)
  }
  h
}

println(fnv1a(""))
println(fnv1a("a"))
```

```text
14695981039346656037
12638187200555641996
```

Add up file sizes as `uint64`, where a negative size panics instead of slipping into the total:

```ant
let sizes = [1200, 3400, 560]
(sizes | Array.map(uint64) | Array.sum()) ?? uint64(0) // 5160
```

## Properties and relations

Converting a non-negative `int` to `uint64` and back gives the same value:

```ant
for n in [0, 1, Math.maxInt] {
  assert(int(uint64(n)) == n)
}
```

`Math.maxUint64` is the largest `uint64`, and adding one with `Bits.wrapAdd` wraps to zero:

```ant
assert(Bits.wrapAdd(Math.maxUint64, uint64(1)) == uint64(0))
assert(Bits.wrapSub(uint64(0), uint64(1)) == Math.maxUint64)
```

Round first to drop a fraction: `Math.trunc` rounds toward zero, `Math.floor` down:

```ant
uint64(Math.trunc(2.7))  // 2
uint64(Math.floor(9.99)) // 9
```

## Possible issues

A negative value panics, even `-1`:

```ant
uint64(-1)   // panic: -1 out of range for uint64
uint64(-1.0) // panic: -1.0 out of range for uint64
```

A float with a fraction panics:

```ant
uint64(2.5) // panic: 2.5 has a fraction, so it does not convert to uint64
```

`NaN` and the infinities have no integer value:

```ant
uint64(Math.NaN)       // panic: NaN does not convert to uint64
uint64(-Math.Infinity) // panic: -Infinity does not convert to uint64
```

`float64(Math.maxUint64)` rounds up to 2^64, which is one past `Math.maxUint64`, so converting it
back panics:

```ant
uint64(float64(Math.maxUint64)) // panic: 18446744073709552000.0 out of range for uint64
```

Subtraction below zero and addition past `Math.maxUint64` panic. Use `Bits.wrapSub` or
`Bits.wrapAdd` when wrapping is what you want:

```ant
uint64(1) - uint64(2)              // panic: integer overflow
Math.maxUint64 + uint64(1)         // panic: integer overflow
Bits.wrapSub(uint64(1), uint64(2)) // 18446744073709551615
```

A large `uint64` does not fit an `int`:

```ant
int(Math.maxUint64) // panic: 18446744073709551615 out of range for int
```

A `uint64` does not mix with an `int`, not even with a literal:

```ant
let n = uint64(5)
let m = n + 1 // error: + needs both operands of one type, got uint64 and int
```

`uint64` does not parse text. Use [String.toUint](../String/toUint.md):

```ant
let n = uint64("42")
// error: uint64 converts a number or a named primitive type over a number, got string
```

## See also

[uint32](uint32.md) · [int64](int64.md) · [int](int.md) · [byte](byte.md) ·
[String.toUint](../String/toUint.md) · [Bits.wrapAdd](../Bits/wrapAdd.md) ·
[Bits.wrapMul](../Bits/wrapMul.md) · [Math.trunc](../Math/trunc.md)
