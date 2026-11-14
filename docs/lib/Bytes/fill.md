<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.fill

```ant signature
func Bytes.fill(b: Bytes, x: byte, from: int = 0, to: int | null = null): void
```

Bytes.fill sets the bytes of b from index from up to but excluding index to, the length when null,
to x. It panics as Bytes.slice does.

<!-- end generated -->

## Details

- `Bytes.fill` changes `b` in place. Every name for `b` sees the new bytes.
- It overwrites bytes that are already there and never grows `b`.
- `x` is a `byte`. An integer literal is an `int`, so write `byte(0xff)`.
- It panics when `from` and `to` are not a range of `b`: `from` greater than `to`, or either outside
  `0 ..= len(b)`.

## Basic examples

Set every byte to `0xff`:

```ant
let b = Bytes.new(4)
Bytes.fill(b, byte(0xff))
b // Bytes(ff ff ff ff)
```

Set the bytes from index 1 up to but excluding index 3:

```ant
let b = String.toBytes("abcde")
Bytes.fill(b, byte(0x2d), 1, 3)
String.fromBytes(b) // "a--de"
```

## Scope

Call it through `.`, or in a pipe, where `b` is the first argument:

```ant
let b = Bytes.new(3)
b.fill(byte(1))
b // Bytes(01 01 01)
b | Bytes.fill(byte(2))
b // Bytes(02 02 02)
```

Equal indices fill nothing, at any position up to the length:

```ant
let b = Bytes.fromArray([1, 2, 3])
b.fill(byte(0), 1, 1)
b.fill(byte(0), 3, 3)
b // Bytes(01 02 03)
```

Empty `Bytes` stays empty, since `Bytes.fill` never adds bytes:

```ant
let empty = Bytes.new()
empty.fill(byte(7))
empty // Bytes()
```

`Bytes` refers to one collection, so a fill shows through every name for it. A slice is a copy and
keeps its bytes:

```ant
let b = String.toBytes("ant")
let same = b
let part = b.slice(0, 2)
same.fill(byte(0))
b    // Bytes(00 00 00)
part // Bytes(61 6e)
```

The byte can come from a variable of type `byte`, such as an element of other `Bytes`:

```ant
let pattern = String.toBytes("*")
let line = Bytes.new(5)
line.fill(pattern[0])
String.fromBytes(line) // "*****"
```

## Named arguments

`from` defaults to 0, so give `to` alone to fill a prefix:

```ant
let b = Bytes.new(5)
b.fill(byte(9), to: 2)
b // Bytes(09 09 00 00 00)
```

`to` defaults to `null`, the length, so give `from` alone to fill a suffix:

```ant
let b = Bytes.new(5)
b.fill(byte(9), from: 3)
b // Bytes(00 00 00 09 09)
```

Name both to make a range stand out:

```ant
let b = Bytes.new(5)
Bytes.fill(b, byte(0xaa), from: 1, to: 4)
b // Bytes(00 aa aa aa 00)
```

## Applications

Pad an ASCII name into a fixed-width field of spaces, as a record of a fixed-format file needs. The
field holds ASCII text, because cutting at `width` bytes can split a multi-byte character:

```ant
func field(name: string, width: int): Bytes {
  let out = Bytes.new(width)
  out.fill(byte(0x20))
  let bytes = String.toBytes(name)
  out.copy(0, bytes, to: Math.min(len(bytes), width))
  out
}

println("[${String.fromBytes(field("ant", 6))}]")
println("[${String.fromBytes(field("antelope", 6))}]")
```

```text
[ant   ]
[antelo]
```

Wipe a key with zero bytes once it is used. `defer` runs the fill when `sign` returns, after the
last line has read the key:

```ant
func sign(key: Bytes, message: string): string {
  defer { key.fill(byte(0)) }
  Crypto.hmacSha256(key, String.toBytes(message)) | Encoding.toHex() | String.slice(0, 16)
}

let key = String.toBytes("not a real key")
println(sign(key, "hello"))
println(key.toArray().every({ x == byte(0) }))
```

```text
4fa627e189c672db
true
```

Draw a bar chart into one reused row of bytes, clearing it with a fill before each row:

```ant
let row = Bytes.new(10)
for n in [3, 7, 10, 1] {
  row.fill(byte(0x2e))
  row.fill(byte(0x23), to: n)
  println(String.fromBytes(row))
}
```

```text
###.......
#######...
##########
#.........
```

## Properties and relations

Bytes in the range become `x`, and bytes outside it keep their values:

```ant
let before = Bytes.fromArray([1, 2, 3, 4, 5, 6])
let b = before.clone()
b.fill(byte(0), 2, 5)
for i in Iter.range(0, len(b)) {
  assert(b[i] == (if i >= 2 && i < 5 { byte(0) } else { before[i] }))
}
```

It does what a loop that assigns each index does:

```ant
let a = Bytes.new(8)
let b = Bytes.new(8)
a.fill(byte(0x41), 2, 6)
for i in Iter.range(2, 6) {
  b[i] = byte(0x41)
}
assert(a == b)
```

Filling all of `Bytes.new(n)` gives what [`Bytes.fromArray`](fromArray.md) builds from
[`Array.filled`](../Array/filled.md):

```ant
let b = Bytes.new(4)
b.fill(byte(0x2a))
assert(b == Bytes.fromArray(Array.filled(4, 0x2a)))
```

The length never changes:

```ant
let b = String.toBytes("hello")
b.fill(byte(0), 1)
assert(len(b) == 5)
```

## Possible issues

An integer literal is an `int`, not a `byte`, so it does not compile as `x`:

```ant
let b = Bytes.new(3)
b.fill(0xff) // error: cannot use int where byte is expected
```

Convert it with `byte(...)`, which panics for a value outside `0 ..= 255`:

```ant
Bytes.new(3).fill(byte(256)) // panic: 256 out of range for byte
```

An index past the end panics; `Bytes.fill` does not clamp it and does not grow `b`:

```ant
Bytes.new(3).fill(byte(1), 0, 5) // panic: Bytes.fill: the range 0..5 is not within 0 ..= 3
```

A negative index panics rather than counting from the end:

```ant
Bytes.new(3).fill(byte(1), -1) // panic: Bytes.fill: the range -1..3 is not within 0 ..= 3
```

`from` past the length panics even without `to`, since `to` is then the length:

```ant
Bytes.new(3).fill(byte(1), 4) // panic: Bytes.fill: the range 4..3 is not within 0 ..= 3
```

Filling a slice does not clear part of `b`. [`Bytes.slice`](slice.md) returns a copy, so the fill
goes to the copy and is lost:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
b.slice(1, 3).fill(byte(0))
b // Bytes(01 02 03 04)
```

Give the range to `Bytes.fill` instead, which fills it in place:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
b.fill(byte(0), 1, 3)
b // Bytes(01 00 00 04)
```

## See also

[Bytes.new](new.md) · [Bytes.copy](copy.md) · [Bytes.truncate](truncate.md) ·
[Bytes.slice](slice.md) · [Bytes.clone](clone.md) · [Array.filled](../Array/filled.md) ·
[Array.slice](../Array/slice.md)
