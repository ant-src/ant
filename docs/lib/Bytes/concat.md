<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.concat

```ant signature
func Bytes.concat(parts: Array[Bytes]): Bytes
```

Bytes.concat returns the byte sequences in parts joined in order.

<!-- end generated -->

## Details

- `Bytes` has no `+`, so `Bytes.concat` joins byte sequences, two or many.

## Basic examples

Join two byte sequences:

```ant
let a = Bytes.fromArray([0x89, 0x50])
let b = Bytes.fromArray([0x4e, 0x47])
Bytes.concat([a, b]) // Bytes(89 50 4e 47)
```

Join the UTF-8 bytes of two strings:

```ant
Bytes.concat(["ant".toBytes(), "hill".toBytes()]) // Bytes(61 6e 74 68 69 6c 6c)
```

## Scope

An empty array gives empty bytes, and empty parts add nothing:

```ant
let none: Array[Bytes] = []
Bytes.concat(none)      // Bytes()
len(Bytes.concat(none)) // 0
let empty = "".toBytes()
Bytes.concat([empty, "x".toBytes(), empty]) // Bytes(78)
```

One part gives bytes equal to it:

```ant
let a = "ant".toBytes()
Bytes.concat([a]) == a // true
```

Pass the parts with a pipe, which puts the array in as the first argument:

```ant
let a = Bytes.fromArray([1, 2])
let b = Bytes.fromArray([3])
[a, b] | Bytes.concat() // Bytes(01 02 03)
```

## Applications

Prefix a payload of up to 255 bytes with its length in one byte:

```ant
func frame(payload: Bytes): Bytes {
  Bytes.concat([Bytes.fromArray([len(payload)]), payload])
}

frame("hi".toBytes()) // Bytes(02 68 69)
```

Collect chunks in an array as a loop goes, then join them once:

```ant
let chunks: Array[Bytes] = []
for word in ["ant", " ", "hill"] {
  chunks.push(word.toBytes())
}
chunks | Bytes.concat() | String.fromBytes() | println()
```

```text
ant hill
```

`Bytes.concat` puts nothing between the parts. To separate records with a newline, put the separator
between them first:

```ant
let records = ["a,1", "b,2", "c,3"]
let newline = "\n".toBytes()
let parts: Array[Bytes] = []
for r in records {
  if len(parts) > 0 {
    parts.push(newline)
  }
  parts.push(r.toBytes())
}
parts | Bytes.concat() | String.fromBytes() | println()
```

```text
a,1
b,2
c,3
```

Put the eight-byte PNG signature in front of image data:

```ant
let signature = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a])
let data = Bytes.concat([signature, "...".toBytes()])
println(Encoding.toHex(data))
```

```text
89504e470d0a1a0a2e2e2e
```

Build a key for a map from two parts. `Bytes` compares by contents, so a key built another way finds
the same entry:

```ant
let m: Map[Bytes, string] = Map{}
m[Bytes.concat(["user:".toBytes(), "42".toBytes()])] = "Ann"
m["user:42".toBytes()] // "Ann"
```

## Properties and relations

The length of the result is the sum of the lengths of the parts:

```ant
let a = "日本".toBytes()
let b = Bytes.fromArray([0, 255])
assert(len(Bytes.concat([a, b])) == len(a) + len(b))
```

The result starts with the first part and ends with the last:

```ant
let head = "GET ".toBytes()
let tail = "/index.html".toBytes()
let line = Bytes.concat([head, tail])
assert(line.startsWith(head))
assert(line.endsWith(tail))
assert(line.slice(len(head), len(line)) == tail)
```

It undoes [Bytes.slice](slice.md): joining the pieces of a cut gives the bytes back:

```ant
let b = "héllo".toBytes()
let cut = 2
assert(Bytes.concat([b.slice(0, cut), b.slice(cut, len(b))]) == b)
```

Joining the bytes of strings gives the bytes of the joined string, as
[String.concat](../String/concat.md) does for text:

```ant
let parts = ["é", "日本", "🐜"]
let joined = Bytes.concat(parts.map({ x.toBytes() }))
assert(joined == String.concat(parts).toBytes())
```

It is the same as joining the arrays of [Bytes.toArray](toArray.md):

```ant
let a = Bytes.fromArray([1, 2])
let b = Bytes.fromArray([3])
assert(Bytes.concat([a, b]).toArray() == Array.concat(a.toArray(), b.toArray()))
```

## Possible issues

`Bytes` has no `+`. Join the parts with `Bytes.concat`:

```ant
let a = "a".toBytes()
let s = a + a
// error: + applies to a single numeric kind, string, or a named primitive type over one, not to Bytes; narrow first
```

The parts go in one array. Bytes passed as separate arguments are a compile error:

```ant
let a = "a".toBytes()
let s = Bytes.concat(a, a) // error: Bytes.concat takes 1 argument, got 2
```

`b.concat()` on bytes is a compile error, since the first parameter is an array:

```ant
let a = "a".toBytes()
let s = a.concat() // error: Bytes.concat is not a function on Bytes: it takes Array[Bytes] first
```

Every part must be `Bytes`. Convert a string with [String.toBytes](../String/toBytes.md) first:

```ant
let s = Bytes.concat(["a", "b"])
// error: cannot use string where Bytes is expected
// error: cannot use string where Bytes is expected
```

A part that may be `null` is a compile error too. Handle the `null` first:

```ant
let x: Bytes | null = null
let s = Bytes.concat(["a".toBytes(), x]) // error: cannot use Bytes | null where Bytes is expected
```

A chunk may end inside a UTF-8 character, so one chunk alone may not decode. Join the chunks first,
then decode:

```ant
let b = "é".toBytes()
let first = b.slice(0, 1)
let second = b.slice(1, 2)
String.fromBytes(first)                         // null
String.fromBytes(Bytes.concat([first, second])) // "é"
```

## See also

[Bytes.slice](slice.md), [Bytes.fromArray](fromArray.md), [Bytes.toArray](toArray.md),
[String.toBytes](../String/toBytes.md), [String.fromBytes](../String/fromBytes.md),
[String.concat](../String/concat.md), [Array.concat](../Array/concat.md)
