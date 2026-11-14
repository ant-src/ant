<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.sizeOf

```ant signature
func Bytes.sizeOf[T: Number](t: Reflect.Type[T]): int
```

Bytes.sizeOf returns the number of bytes a number of type t takes: 1, 2, 4 or 8.

<!-- end generated -->

## Details

- It is the number of bytes [Bytes.read](read.md), [Bytes.write](write.md), and
  [Bytes.append](append.md) take or add for that type.
- The type value must name one number type. A union, `Number` itself, or a type that is not a number
  is a compile error.

## Basic examples

Ask for the width of a number type, written as a type value with `type(...)`:

```ant
Bytes.sizeOf(type(int8))    // 1
Bytes.sizeOf(type(uint16))  // 2
Bytes.sizeOf(type(float32)) // 4
Bytes.sizeOf(type(float64)) // 8
```

## Scope

Every number type, signed and unsigned, has the width its name says:

```ant
Bytes.sizeOf(type(int8))   // 1
Bytes.sizeOf(type(uint8))  // 1
Bytes.sizeOf(type(int16))  // 2
Bytes.sizeOf(type(uint16)) // 2
Bytes.sizeOf(type(int32))  // 4
Bytes.sizeOf(type(uint32)) // 4
Bytes.sizeOf(type(int64))  // 8
Bytes.sizeOf(type(uint64)) // 8
```

The aliases `int` and `byte` are `int64` and `uint8`:

```ant
Bytes.sizeOf(type(int))  // 8
Bytes.sizeOf(type(byte)) // 1
```

A type value stored in a variable works the same:

```ant
let word = type(uint32)
Bytes.sizeOf(word) // 4
```

Inside a generic function, pass on the `Reflect.Type[T]` parameter. This one reads every value of
type `T` packed in `b`:

```ant
func readAll[T: Number](b: Bytes, t: Reflect.Type[T], order: Bytes.Order): Array[T] {
  let size = Bytes.sizeOf(t)
  Array.range(0, b.length() / size).map({ b.read(t, x * size, order) })
}

let b = Bytes.fromArray([0, 1, 0, 2, 0, 3, 0, 4])
readAll(b, type(uint16), Bytes.BigEndian{}) // [1, 2, 3, 4]
readAll(b, type(uint32), Bytes.BigEndian{}) // [65538, 196612]
```

## Applications

Compute the size of a fixed binary header, a `uint32` magic number, a `uint16` version, and a
`uint32` length, instead of writing `10` by hand:

```ant
let headerSize = Bytes.sizeOf(type(uint32)) + Bytes.sizeOf(type(uint16)) + Bytes.sizeOf(type(uint32))
headerSize // 10
```

Allocate room for a known number of values, then write each at its offset:

```ant
let size = Bytes.sizeOf(type(uint16))
let samples = [440, 880, 1760]
let b = Bytes.new(samples.length() * size)
for { index, value } in samples.entries() {
  b.write(type(uint16), index * size, uint16(value), Bytes.LittleEndian{})
}
println(b)
```

```text
Bytes(b8 01 70 03 e0 06)
```

Check that a payload holds whole values before reading it:

```ant
func wholeFloats(b: Bytes): bool {
  b.length() % Bytes.sizeOf(type(float64)) == 0
}

wholeFloats(Bytes.new(16)) // true
wholeFloats(Bytes.new(12)) // false
```

## Properties and relations

`Bytes.append` adds exactly `Bytes.sizeOf(t)` bytes:

```ant
let b = Bytes.new()
b.append(type(uint32), uint32(7), Bytes.BigEndian{})
assert(b.length() == Bytes.sizeOf(type(uint32)))
b.append(type(float64), 1.5, Bytes.BigEndian{})
assert(b.length() == Bytes.sizeOf(type(uint32)) + Bytes.sizeOf(type(float64)))
```

A signed type and its unsigned counterpart have the same width, and so do an alias and its type:

```ant
assert(Bytes.sizeOf(type(int32)) == Bytes.sizeOf(type(uint32)))
assert(Bytes.sizeOf(type(int)) == Bytes.sizeOf(type(int64)))
assert(Bytes.sizeOf(type(byte)) == Bytes.sizeOf(type(uint8)))
```

`Bytes.read` needs `Bytes.sizeOf(t)` bytes from its offset, so reading at the last whole position
works and one byte later panics:

```ant
let b = Bytes.new(6)
let last = b.length() - Bytes.sizeOf(type(uint32))
b.read(type(uint32), last, Bytes.BigEndian{})     // 0
b.read(type(uint32), last + 1, Bytes.BigEndian{}) // panic: Bytes.read: the range 3..7 is not within 0 ..= 6
```

## Possible issues

A type name alone is not a value. Write `type(int32)`:

```ant
println(Bytes.sizeOf(int32)) // error: int32 is a type name; write type(int32) for its value
```

The type must be one number type, not a union of them and not `Number`:

```ant
println(Bytes.sizeOf(type(uint16 | uint32)))
// error: T would be uint16 | uint32, which is not a single member of its bound Number; narrow first
println(Bytes.sizeOf(type(bool))) // error: T would be bool, which is not a member of its bound Number
```

The first parameter is a `Reflect.Type`, not `Bytes`, so a call through `.` on the type value is a
compile error. Use the namespace form or a pipe:

```ant
let t = type(uint32)
println(t.sizeOf()) // error: Reflect.Type[uint32] has no field or function sizeOf
```

Type values of different types do not share an array, so mapping over several widths at once is a
compile error. Ask for each one by itself:

```ant
println([type(int8), type(int16)].map(Bytes.sizeOf))
// error: the elements do not form a valid type: a union may hold at most one type with no discriminator, and Reflect.Type[int8] and Reflect.Type[int16] both have none; give the array an expected type
```

An erased type value, from `Reflect.erase`, no longer says which number it is, so it is a compile
error too:

```ant
println(Bytes.sizeOf(Reflect.erase(type(int8))))
// error: cannot use Reflect.ErasedType where Reflect.Type[T] is expected
```

## See also

[Bytes.read](read.md) · [Bytes.write](write.md) · [Bytes.append](append.md) · [Bytes.new](new.md) ·
[Bytes.length](length.md) · [Reflect.erase](../Reflect/erase.md) · [Bytes](index.md)
