<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Integer

```ant signature
type Integer = int8 | int16 | int32 | int64 | uint8 | uint16 | uint32 | uint64
```

Integer is any of the eight sized integer types; int and byte are among them.

<!-- end generated -->

## Details

- `Integer` exists to bound a type parameter. `[T: Integer]` accepts any one of the eight integer
  types, and inside the body the arithmetic and comparison operators apply to `T`.
- The solution for `T` is a single member. A value typed `Integer`, a float, and a named primitive
  type over an integer are all rejected.
- A variable typed `Integer` holds any integer, but no arithmetic or comparison operator applies to
  it until `is` narrows it to one member. `==` works, and values of different members are never
  equal. `is Integer` itself is a compile error, since `Integer` is an alias.
- Arithmetic in a generic body panics on overflow at the width of `T`, as it does outside one.

## Basic examples

Write a function once for every integer type:

```ant
func double[T: Integer](x: T): T { x + x }
double(21)        // 42
double(byte(100)) // 200
double(int8(-3))  // -6
```

The `Bits` functions take any integer type and return the same type:

```ant
Bits.and(byte(0b1100), byte(0b1010)) // 8
Bits.not(uint16(0))                  // 65535
Bits.not(int16(0))                   // -1
```

## Scope

Every member works, signed and unsigned, from 8 to 64 bits. `int` is `int64` and `byte` is `uint8`:

```ant
func isZero[T: Integer](x: T): bool { x == x - x }
isZero(int8(0))   // true
isZero(int16(4))  // false
isZero(int32(0))  // true
isZero(0)         // true
isZero(byte(0))   // true
isZero(uint16(1)) // false
isZero(uint32(0)) // true
isZero(uint64(0)) // true
```

An array of `T` takes one member for all its elements, and `<` and `>` compare them:

```ant
func largest[T: Integer](xs: Array[T]): T {
  var best = xs[0]
  for x in xs {
    if x > best {
      best = x
    }
  }
  best
}
largest([3, 9, 4])                           // 9
largest([int8(-5), int8(-2)])                // -2
largest(Bytes.toArray(String.toBytes("hi"))) // 105
```

Annotate a variable or an array element as `Integer` to hold integers of different types side by
side:

```ant
let mixed: Array[Integer] = [1, byte(2), int8(-3)]
mixed // [1, 2, -3]
```

Narrow a value typed `Integer` with `is` on a member before using an operator on it:

```ant
func describe(x: Integer): string {
  if x is int {
    "int ${x + 1}"
  } else if x is byte {
    "byte ${x}"
  } else {
    "another integer"
  }
}
describe(7)        // "int 8"
describe(byte(3))  // "byte 3"
describe(int16(3)) // "another integer"
```

## Applications

Format any integer as a fixed-width binary string:

```ant
func binary[T: Integer](x: T, width: int): string {
  let digits = String.fromInt(x, base: 2)
  String.repeat("0", width - String.length(digits)) + digits
}
binary(byte(5), 8)      // "00000101"
binary(uint16(300), 16) // "0000000100101100"
```

Compute the parity bit of a word of any width:

```ant
func parity[T: Integer](x: T): int { Bits.popCount(x) % 2 }
parity(byte(0b0110)) // 0
parity(uint32(7))    // 1
parity(uint64(0xFF)) // 0
```

Library functions bounded by `Integer` accept an array of any one member:

```ant
Bytes.fromArray([72, 105])             // Bytes(48 69)
Bytes.fromArray([byte(72), byte(105)]) // Bytes(48 69)
```

## Properties and relations

`Bits.not` undoes itself on every member:

```ant
assert(Bits.not(Bits.not(byte(5))) == byte(5))
assert(Bits.not(Bits.not(int32(-9))) == int32(-9))
assert(Bits.not(Bits.not(uint64(1))) == uint64(1))
```

On any member, `Bits.wrapAdd` agrees with `+` when the sum fits:

```ant
assert(Bits.wrapAdd(byte(100), byte(27)) == byte(100) + byte(27))
assert(Bits.wrapAdd(int8(-100), int8(27)) == int8(-100) + int8(27))
```

## Possible issues

A value typed `Integer` is a union, so it neither solves `T` nor takes an operator:

```ant
func double[T: Integer](x: T): T { x + x }
let n: Integer = 5
let a = n + 1 // error: + needs both operands of one type, got Integer and int
let b = double(n)
// error: T would be Integer, which is not a single member of its bound Integer; narrow first
```

`==` on two `Integer` values compares the member too: an `int` and a `byte` with the same value are
not equal:

```ant
let a: Integer = 5
let b: Integer = byte(5)
a == b // false
```

An `Array[int]` is not an `Array[Integer]`, because type arguments are invariant:

```ant
func total(xs: Array[Integer]): int { 0 }
let xs = [1, 2]
total(xs) // error: cannot use Array[int] where Array[Integer] is expected
```

Take `[T: Integer](xs: Array[T])` instead, as `largest` in Scope does.

`is Integer` does not compile, because an alias has no run-time identity. Test the members:

```ant
func isWhole(x: Number): bool { x is Integer }
// error: Integer is an alias and has no run-time identity; test its members instead
```

A float is not an `Integer`, even with no fractional part:

```ant
func double[T: Integer](x: T): T { x + x }
let c = double(2.0)
// error: T would be float64, which is not a single member of its bound Integer; narrow first
```

A named primitive type over an integer is not a member. Unwrap it, then tag the result again:

```ant
func double[T: Integer](x: T): T { x + x }
type Count = int
let d = double(Count(2))
// error: T would be Count, which is not a single member of its bound Integer; narrow first
```

```ant
func double[T: Integer](x: T): T { x + x }
type Count = int
Count(double(int(Count(2)))) // Count(4)
```

Overflow panics at the width of the member `T` stands for:

```ant
func double[T: Integer](x: T): T { x + x }
double(100)       // 200
double(int8(100)) // panic: integer overflow
```

## See also

[Number](Number.md) · [Ordered](Ordered.md) · [int](../../../language/types/int.md) ·
[byte](../../../language/types/byte.md) · [Bits](../../Bits/index.md) ·
[Bits.wrapAdd](../../Bits/wrapAdd.md) · [String.fromInt](../../String/fromInt.md)
