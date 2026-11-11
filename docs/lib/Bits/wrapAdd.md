<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bits.wrapAdd

```ant signature
func Bits.wrapAdd[T: Integer](a: T, b: T): T
```

Bits.wrapAdd returns a + b, wrapping around in two's complement instead of panicking on overflow.

<!-- end generated -->

## Details

- `Bits.wrapAdd` never panics. A sum that does not fit `T` wraps around at the width of `T`: past
  the maximum it continues from the minimum, and below the minimum it continues from the maximum.

## Basic examples

Add one to the largest `byte` and get `0` instead of a panic:

```ant
Bits.wrapAdd(byte(255), byte(1))  // 0
Bits.wrapAdd(byte(250), byte(10)) // 4
```

A signed sum that overflows wraps to the negative side:

```ant
Bits.wrapAdd(int8(127), int8(1)) // -128
```

A sum that fits is ordinary addition:

```ant
Bits.wrapAdd(3, 4) // 7
```

## Scope

Every integer width works. On `int`, the sum wraps at 64 bits:

```ant
Bits.wrapAdd(Math.maxInt, 1) == Math.minInt         // true
Bits.wrapAdd(Math.maxUint64, uint64(2))             // 1
Bits.wrapAdd(uint32(4000000000), uint32(500000000)) // 205032704
```

Negative operands wrap below the minimum back to the maximum:

```ant
Bits.wrapAdd(int16(-32768), int16(-1)) // 32767
Bits.wrapAdd(int8(-100), int8(-100))   // 56
```

Call it in a pipe, where the value on the left becomes `a`:

```ant
uint8(200) | Bits.wrapAdd(uint8(100)) // 44
```

Pass it as a function value, such as the step of `reduce`:

```ant
let bytes = [10, 20, 250].map(byte)
bytes.reduce(byte(0), Bits.wrapAdd) // 24
```

## Applications

Compute an 8-bit checksum, the sum of every byte modulo 256:

```ant
func checksum(data: Bytes): byte {
  var sum = byte(0)
  for b in data {
    sum = Bits.wrapAdd(sum, b)
  }
  sum
}
checksum("hello".toBytes()) // 20
```

Step a 16-bit sequence number that starts again from 0 after 65535:

```ant
func next(seq: uint16): uint16 { Bits.wrapAdd(seq, uint16(1)) }
next(uint16(41))    // 42
next(uint16(65535)) // 0
```

Hash a string with djb2, which multiplies and adds in 32 bits and lets both wrap:

```ant
func djb2(s: string): uint32 {
  var h = uint32(5381)
  for b in s.toBytes() {
    h = Bits.wrapAdd(Bits.wrapMul(h, uint32(33)), uint32(b))
  }
  h
}
djb2("ant") // 193486376
```

## Properties and relations

The order of the operands does not matter:

```ant
let a = int8(100)
let b = int8(90)
assert(Bits.wrapAdd(a, b) == Bits.wrapAdd(b, a))
```

For a `byte`, the result is the exact sum modulo 256:

```ant
for pair in [[200, 100], [255, 1], [3, 4]] {
  let sum = Bits.wrapAdd(byte(pair[0]), byte(pair[1]))
  assert(int(sum) == (pair[0] + pair[1]) % 256)
}
```

`Bits.wrapSub` undoes it, even when the sum wrapped:

```ant
let x = uint8(250)
let y = uint8(10)
assert(Bits.wrapSub(Bits.wrapAdd(x, y), y) == x)
```

## Possible issues

`+` panics where `Bits.wrapAdd` wraps:

```ant
let full = uint8(255)
full + uint8(1) // panic: integer overflow
```

An integer literal is always an `int`, so it does not match a `byte` operand. Convert it:

```ant
let b = byte(255)
let next = Bits.wrapAdd(b, 1) // error: cannot use int where byte is expected
```

A number has no namespace, so a dot call on it finds no function. Call `Bits.wrapAdd(b, byte(1))`
instead:

```ant
let b = byte(255)
let next = b.wrapAdd(byte(1)) // error: byte has no field or function wrapAdd
```

A named primitive type is not a member of `Integer`. Unwrap it, add, and tag the result:

```ant
type Ticks = uint16
let t = Ticks(uint16(65535))
let later = Bits.wrapAdd(t, Ticks(uint16(1)))
// error: T would be Ticks, which is not a member of its bound Integer; unwrap it with uint16(...)
```

```ant
type Ticks = uint16
let t = Ticks(uint16(65535))
Ticks(Bits.wrapAdd(uint16(t), uint16(1))) // Ticks(0)
```

## See also

- [Bits.wrapSub](wrapSub.md) and [Bits.wrapMul](wrapMul.md): the other wrapping operations.
- [Bits](index.md): the other bitwise operations.
