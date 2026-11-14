<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.copy

```ant signature
func Bytes.copy(dst: Bytes, at: int, src: Bytes, from: int = 0, to: int | null = null): void
```

Bytes.copy overwrites the bytes of dst from index at with those of src from index from up to but
excluding index to, the length of src when null, as if src were read first. It never grows dst, and
panics before writing unless 0 \<= at \<= Bytes.length(dst) and the range fits.

<!-- end generated -->

## Details

- `Bytes.copy` changes `dst` in place and returns nothing.
- It never grows `dst`: a range that does not fit between `at` and the end of `dst`, or is not
  within `src`, panics before any byte is written. To add bytes at the end, use
  [Bytes.extend](extend.md).
- `src` and `dst` may be the same `Bytes`, even with overlapping ranges: the source range is read
  before any byte is written.

## Basic examples

Copy all of `src` into `dst` from index 1:

```ant
let dst = Bytes.new(6)
Bytes.copy(dst, 1, Bytes.fromArray([1, 2, 3]))
dst // Bytes(00 01 02 03 00 00)
```

Copy only the bytes of `src` from index 1 up to but excluding index 4:

```ant
let dst = Bytes.fromArray([0, 0, 0, 0])
Bytes.copy(dst, 0, "hello".toBytes(), 1, 4)
dst // Bytes(65 6c 6c 00)
```

## Scope

The bytes of `dst` outside the range keep their values, and its length stays the same:

```ant
let dst = "abcdef".toBytes()
dst.copy(2, "XY".toBytes())
String.fromBytes(dst) // "abXYef"
len(dst)              // 6
```

An empty source range writes nothing. `at` may then be the length of `dst`, and `dst` may be empty:

```ant
let dst = Bytes.fromArray([7, 8])
dst.copy(2, Bytes.new())
dst // Bytes(07 08)
let empty = Bytes.new()
empty.copy(0, "abc".toBytes(), 1, 1)
empty // Bytes()
```

`dst` and `src` may be the same `Bytes`. Copying to a higher index shifts bytes right, and to a
lower index shifts them left, even where the ranges overlap:

```ant
let right = Bytes.fromArray([1, 2, 3, 4, 5])
right.copy(1, right, 0, 3)
right // Bytes(01 01 02 03 05)
let left = Bytes.fromArray([1, 2, 3, 4, 5])
left.copy(0, left, 2)
left // Bytes(03 04 05 04 05)
```

A copy shares nothing with `src`: writing to `src` afterwards leaves `dst` as it is:

```ant
let src = Bytes.fromArray([1, 2])
let dst = Bytes.new(2)
dst.copy(0, src)
src[0] = byte(9)
dst // Bytes(01 02)
```

`dst` is shared like any `Bytes`, so the change shows through every name for it:

```ant
let dst = Bytes.new(3)
let alias = dst
dst.copy(0, Bytes.fromArray([5, 6, 7]))
alias // Bytes(05 06 07)
```

## Named arguments

`from` starts the source range at an index other than 0:

```ant
let dst = Bytes.new(3)
dst.copy(0, "hello".toBytes(), from: 2)
String.fromBytes(dst) // "llo"
```

`to` ends it before the end of `src`. With both left out, `Bytes.copy` copies all of `src`:

```ant
let dst = "-----".toBytes()
dst.copy(0, "hello".toBytes(), to: 2)
String.fromBytes(dst) // "he---"
dst.copy(1, "hello".toBytes(), from: 1, to: 3)
String.fromBytes(dst) // "hel--"
```

## Applications

Build a fixed-size record: allocate it once with [Bytes.new](new.md), then copy each field to its
offset:

```ant
func record(name: string, tag: string): Bytes {
  let bytes = name.toBytes()
  let out = Bytes.new(12)
  out.copy(0, "ANT1".toBytes())
  out.copy(4, bytes, to: Math.min(len(bytes), 6))
  out.copy(10, tag.toBytes(), to: 2)
  out
}

println(Encoding.toHex(record("ann", "ok")))
println(Encoding.toHex(record("bartholomew", "ok")))
```

```text
414e5431616e6e0000006f6b
414e543162617274686f6f6b
```

Drop the bytes a parser has consumed from the front of a buffer, in place: shift the rest left, then
cut the end with [Bytes.truncate](truncate.md):

```ant
func consume(buf: Bytes, n: int) {
  buf.copy(0, buf, n)
  buf.truncate(len(buf) - n)
}

let buf = "GET /\nHost".toBytes()
consume(buf, 6)
println(String.fromBytes(buf))
```

```text
Host
```

Insert bytes in the middle: grow the buffer at the end, shift the tail right, then copy the new
bytes into the gap:

```ant
func insert(b: Bytes, at: int, part: Bytes) {
  let n = len(b)
  b.extend(Bytes.new(len(part)))
  b.copy(at + len(part), b, at, n)
  b.copy(at, part)
}

let b = "ant hill".toBytes()
insert(b, 4, "big ".toBytes())
println(String.fromBytes(b))
```

```text
ant big hill
```

## Properties and relations

After the copy, the range of `dst` from `at` equals the source range, and the length of `dst` is
unchanged:

```ant
let src = "héllo".toBytes()
let dst = Bytes.new(8)
dst.copy(2, src, 1, 4)
assert(dst.slice(2, 5) == src.slice(1, 4))
assert(len(dst) == 8)
```

When `dst` and `src` are different `Bytes`, it writes the same bytes as assigning them one index at
a time. When they are the same `Bytes` with overlapping ranges, the two differ, as the overlap
example under Scope shows:

```ant
let src = Bytes.fromArray([10, 20, 30, 40])
let a = Bytes.new(5)
let b = Bytes.new(5)
a.copy(1, src, 1, 3)
for i in Iter.range(1, 3) {
  b[i] = src[i]
}
assert(a == b)
```

Copying all of `src` to index 0 of a new `Bytes` of the same length gives what
[Bytes.clone](clone.md) gives:

```ant
let src = "ant".toBytes()
let dst = Bytes.new(len(src))
dst.copy(0, src)
assert(dst == src.clone())
assert(!Bytes.same(dst, src))
```

## Possible issues

`Bytes.copy` never grows `dst`. A range that runs past its end panics, and the message names the
range of `dst` it would write:

```ant
Bytes.new(3).copy(1, "abc".toBytes()) // panic: Bytes.copy: the range 1..4 is not within 0 ..= 3
```

The panic comes before any byte is written, so `dst` keeps its old bytes:

```ant
let dst = "abc".toBytes()
Panic.catch({ dst.copy(1, "xyz".toBytes()) }).isOk() // false
String.fromBytes(dst)                                // "abc"
```

`at` past the end panics even when nothing would be written, and a negative `at` does not count from
the end:

```ant
Bytes.new(3).copy(4, Bytes.new())  // panic: Bytes.copy: the range 4..4 is not within 0 ..= 3
Bytes.new(3).copy(-1, Bytes.new()) // panic: Bytes.copy: the range -1..-1 is not within 0 ..= 3
```

A source range that is not within `src` panics too, as [Bytes.slice](slice.md) does:

```ant
Bytes.new(3).copy(0, "abc".toBytes(), 2, 1) // panic: Bytes.copy: the range 2..1 is not within 0 ..= 3
Bytes.new(8).copy(0, "abc".toBytes(), 0, 5) // panic: Bytes.copy: the range 0..5 is not within 0 ..= 3
```

The destination comes first, then where to write, then the source. Passing the source first is a
compile error:

```ant
let src = "abc".toBytes()
let dst = Bytes.new(3)
Bytes.copy(src, dst, 0) // error: cannot use Bytes where int is expected
```

## See also

[Bytes.slice](slice.md) · [Bytes.extend](extend.md) · [Bytes.fill](fill.md) · [Bytes.new](new.md) ·
[Bytes.clone](clone.md) · [Bytes.truncate](truncate.md) · [Bytes.write](write.md)
