<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.slice

```ant signature
func Bytes.slice(b: Bytes, from: int, to: int): Bytes
```

Bytes.slice returns the bytes of b from index from up to but excluding index to. It panics if from
\> to, or if either is outside 0 ..= Bytes.length(b).

<!-- end generated -->

## Details

- Indices count bytes. `from` is inclusive and `to` is exclusive, so `from == to` gives empty
  `Bytes`.

## Basic examples

Take the bytes from index 1 up to but excluding index 4:

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a])
Bytes.slice(png, 1, 4) // Bytes(50 4e 47)
```

Take a prefix from index 0, or a suffix up to the length:

```ant
let b = Bytes.fromArray([1, 2, 3, 4, 5])
Bytes.slice(b, 0, 2)      // Bytes(01 02)
Bytes.slice(b, 3, len(b)) // Bytes(04 05)
```

## Scope

Equal indices give empty `Bytes`, at any position up to the length, empty `Bytes` included:

```ant
let b = Bytes.fromArray([1, 2, 3])
Bytes.slice(b, 1, 1)                  // Bytes()
Bytes.slice(b, 3, 3)                  // Bytes()
Bytes.slice(String.toBytes(""), 0, 0) // Bytes()
```

Indices count bytes, not code points. In UTF-8, `é` takes the two bytes at indices 1 and 2:

```ant
let utf8 = String.toBytes("héllo")
Bytes.slice(utf8, 1, 3)                   // Bytes(c3 a9)
String.fromBytes(Bytes.slice(utf8, 1, 3)) // "é"
```

The slice is `Bytes` like any other: it compares by contents, indexes to a `byte`, and works as a
map key:

```ant
let b = Bytes.fromArray([0xca, 0xfe, 0xba, 0xbe])
Bytes.slice(b, 0, 2) == Bytes.fromArray([0xca, 0xfe]) // true
Bytes.slice(b, 2, 4)[0]                               // 186
let seen: Map[Bytes, string] = Map{}
seen[Bytes.slice(b, 0, 2)] = "cafe"
seen[Bytes.fromArray([0xca, 0xfe])] // "cafe"
```

Call it through `.`, in a pipe, or with named arguments:

```ant
let b = String.toBytes("hello")
b.slice(1, 4)                  // Bytes(65 6c 6c)
b | Bytes.slice(1, 4)          // Bytes(65 6c 6c)
Bytes.slice(b, from: 1, to: 3) // Bytes(65 6c)
```

## Applications

Read a big-endian unsigned integer from a range of a binary header:

```ant
func bigEndian(b: Bytes, from: int, to: int): int {
  var n = 0
  for x in Bytes.slice(b, from, to) {
    n = n * 256 + int(x)
  }
  n
}

let header = Bytes.fromArray([0x01, 0xbb, 0x00, 0x00, 0x00, 0x50])
println(bigEndian(header, 0, 2))
println(bigEndian(header, 2, 6))
```

```text
443
80
```

Split at the first newline with `Bytes.indexOf`, into a record of the two parts:

```ant
type Frame = { head: Bytes, body: Bytes }

func frame(b: Bytes): Frame | null {
  let i = Bytes.indexOf(b, Bytes.fromArray([0x0a])) ?? return null
  Frame{ head: Bytes.slice(b, 0, i), body: Bytes.slice(b, i + 1, len(b)) }
}

let f = frame(String.toBytes("GET /\nHost: ant"))
if f != null {
  println(String.fromBytes(f.head))
  println(String.fromBytes(f.body))
}
println(frame(String.toBytes("no newline")))
```

```text
GET /
Host: ant
null
```

Cut data into blocks of a fixed size, the last one shorter:

```ant
func blocks(b: Bytes, size: int): Array[Bytes] {
  let out: Array[Bytes] = []
  var i = 0
  while i < len(b) {
    Array.push(out, Bytes.slice(b, i, Math.min(i + size, len(b))))
    i += size
  }
  out
}

for block in blocks(String.toBytes("abcdefg"), 3) {
  println(Encoding.toHex(block))
}
```

```text
616263
646566
67
```

## Properties and relations

The result holds `to - from` bytes, and byte `i` of it is byte `from + i` of `b`:

```ant
let b = Bytes.fromArray([5, 6, 7, 8, 9])
let part = Bytes.slice(b, 1, 4)
assert(len(part) == 3)
assert(Array.every(Array.range(0, len(part)), { part[x] == b[1 + x] }))
```

Two slices that meet at the same index join back to `b` with `Bytes.concat`:

```ant
let b = String.toBytes("añt🐜")
let n = len(b)
for i in Iter.range(0, n + 1) {
  assert(Bytes.concat([Bytes.slice(b, 0, i), Bytes.slice(b, i, n)]) == b)
}
```

It takes the same bytes that `Array.slice` takes from `Bytes.toArray`:

```ant
let b = Bytes.fromArray([10, 20, 30, 40, 50])
assert(Bytes.slice(b, 1, 4) == Bytes.fromArray(Array.slice(Bytes.toArray(b), 1, 4)))
```

Slicing a slice is one slice with the offsets added:

```ant
let b = Bytes.fromArray([1, 2, 3, 4, 5, 6])
assert(Bytes.slice(Bytes.slice(b, 1, 5), 1, 3) == Bytes.slice(b, 2, 4))
```

## Possible issues

An index past the end panics; `Bytes.slice` does not clamp it:

```ant
Bytes.slice(Bytes.fromArray([1, 2, 3]), 0, 5) // panic: Bytes.slice: the range 0..5 is not within 0..3
```

A negative index panics rather than counting from the end:

```ant
Bytes.slice(Bytes.fromArray([1, 2, 3]), -1, 3) // panic: Bytes.slice: the range -1..3 is not within 0..3
```

`from` greater than `to` panics rather than giving empty `Bytes`:

```ant
Bytes.slice(Bytes.fromArray([1, 2, 3]), 2, 1) // panic: Bytes.slice: the range 2..1 is not within 0..3
```

All three arguments are required; there is no form that runs to the end:

```ant
println(Bytes.slice(Bytes.fromArray([1, 2, 3]), 1)) // error: Bytes.slice needs an argument for to
```

A byte range can cut a UTF-8 character in two. `String.fromBytes` then returns `null`, and
`String.slice` counts code points instead:

```ant
let utf8 = String.toBytes("héllo")
String.fromBytes(Bytes.slice(utf8, 0, 2)) // null
String.slice("héllo", 0, 2)               // "hé"
```

## See also

[Bytes.length](length.md) · [Bytes.indexOf](indexOf.md) · [Bytes.concat](concat.md) ·
[Bytes.startsWith](startsWith.md) · [Bytes.toArray](toArray.md) · [Array.slice](../Array/slice.md) ·
[String.slice](../String/slice.md) · [String.fromBytes](../String/fromBytes.md)
