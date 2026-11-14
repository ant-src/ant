<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.new

```ant signature
func Bytes.new(size: int = 0): Bytes
```

Bytes.new returns size zero bytes. It panics if size \< 0 or is above 2^31 − 1.

<!-- end generated -->

## Details

- Each call returns a new collection that nothing else holds.
- The bytes are part of the length, not spare room: `push` and `extend` append after them.

## Basic examples

Make four zero bytes:

```ant
Bytes.new(4) // Bytes(00 00 00 00)
```

Without a size, `Bytes.new` gives empty bytes to grow:

```ant
let b = Bytes.new()
b.push(byte(0x41))
b.push(byte(0x6e))
b // Bytes(41 6e)
```

## Scope

A size of 0 is the same as no size:

```ant
Bytes.new(0)                // Bytes()
Bytes.new(0) == Bytes.new() // true
```

The bytes start at zero and take writes by index:

```ant
let b = Bytes.new(3)
b[1] = byte(0xff)
b // Bytes(00 ff 00)
```

Two calls give equal contents but two collections:

```ant
let a = Bytes.new(2)
let b = Bytes.new(2)
a == b           // true
Bytes.same(a, b) // false
```

## Named arguments

`size:` names the number of bytes:

```ant
Bytes.new(size: 3) // Bytes(00 00 00)
```

## Applications

When the size is known, make it once and assign each byte by index:

```ant
let n = 5
let b = Bytes.new(n)
for i in Iter.range(0, n) {
  b[i] = byte(i * 2)
}
println(b)
```

```text
Bytes(00 02 04 06 08)
```

Lay out a fixed-size header and write numbers into it:

```ant
let header = Bytes.new(6)
header.write(type(uint16), 0, uint16(443), Bytes.BigEndian{})
header.write(type(uint32), 2, uint32(80), Bytes.BigEndian{})
header // Bytes(01 bb 00 00 00 50)
```

Pad data with zero bytes to a multiple of a block size:

```ant
func pad(b: Bytes, block: int): Bytes {
  let out = b.clone()
  let rest = len(b) % block
  if rest == 0 {
    return out
  }
  out.extend(Bytes.new(block - rest))
  out
}

println(pad("abcde".toBytes(), 4))
println(pad("abcd".toBytes(), 4))
```

```text
Bytes(61 62 63 64 65 00 00 00)
Bytes(61 62 63 64)
```

Start from zeros and fill them with another byte:

```ant
let b = Bytes.new(4)
b.fill(byte(0xff))
b // Bytes(ff ff ff ff)
```

## Properties and relations

The length is `size`, and every byte is zero:

```ant
let b = Bytes.new(8)
assert(len(b) == 8)
assert(b.toArray().every({ x == byte(0) }))
```

It holds the same bytes as `Bytes.fromArray` of `size` zeros:

```ant
assert(Bytes.new(3) == Bytes.fromArray(Array.filled(3, 0)))
```

## Possible issues

`size` is a count, not a capacity. Pushing appends after the zeros:

```ant
let b = Bytes.new(2)
b.push(byte(1))
b // Bytes(00 00 01)
```

Empty bytes have no index to write; grow them with `push` instead:

```ant
let b = Bytes.new()
b[0] = byte(1) // panic: index 0 out of range for Bytes of length 0
```

A negative size panics, and so does one above 2^31 − 1:

```ant
Bytes.new(-1)         // panic: Bytes.new needs a size of at least 0, got -1
Bytes.new(2147483648) // panic: Bytes.new needs a size of at most 2147483647, got 2147483648
```

`size` is an `int`. Convert a length read as another integer type:

```ant
let prefix = Bytes.fromArray([0, 0, 0, 3])
let n = prefix.read(type(uint32), 0, Bytes.BigEndian{})
let body = Bytes.new(n) // error: cannot use uint32 where int is expected
```

Convert it with `int`:

```ant
let prefix = Bytes.fromArray([0, 0, 0, 3])
let n = prefix.read(type(uint32), 0, Bytes.BigEndian{})
Bytes.new(int(n)) // Bytes(00 00 00)
```

`Array.filled` repeats one value, so every element is the same collection. Call `Bytes.new` once per
element:

```ant
let shared = Array.filled(2, Bytes.new(2))
shared[0][0] = byte(7)
shared // [Bytes(07 00), Bytes(07 00)]
let own = Array.range(0, 2).map({ Bytes.new(2) })
own[0][0] = byte(7)
own // [Bytes(07 00), Bytes(00 00)]
```

## See also

[Bytes](index.md) · [Bytes.fromArray](fromArray.md) · [Bytes.push](push.md) ·
[Bytes.extend](extend.md) · [Bytes.fill](fill.md) · [Bytes.write](write.md) ·
[Bytes.clone](clone.md) · [Array.filled](../Array/filled.md)
