<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.indexOf

```ant signature
func Bytes.indexOf(b: Bytes, sub: Bytes): int | null
```

Bytes.indexOf returns the index of the first occurrence of sub in b, or null when it does not occur.
It returns 0 for an empty sub.

<!-- end generated -->

## Details

- The index counts bytes, from 0.
- The search compares bytes only. It knows nothing of UTF-8, so a match may start or end inside a
  multibyte character.

## Basic examples

```ant
let b = Bytes.fromArray([1, 2, 3, 2, 3])
Bytes.indexOf(b, Bytes.fromArray([2, 3])) // 1
Bytes.indexOf(b, Bytes.fromArray([9]))    // null
```

Call it through `.` on any `Bytes`:

```ant
let data = String.toBytes("key=value")
data.indexOf(String.toBytes("=")) // 3
```

## Scope

The first occurrence wins, even when a later one overlaps it:

```ant
Bytes.indexOf(Bytes.fromArray([0, 0, 0]), Bytes.fromArray([0, 0])) // 0
```

`sub` may be all of `b`, but not longer:

```ant
let b = Bytes.fromArray([7, 8])
Bytes.indexOf(b, b)                          // 0
Bytes.indexOf(b, Bytes.fromArray([7, 8, 9])) // null
```

Only an empty `sub` occurs in empty bytes, and an empty `sub` gives `0` for any `b`:

```ant
let empty = String.toBytes("")
Bytes.indexOf(empty, Bytes.fromArray([1]))  // null
Bytes.indexOf(String.toBytes("ant"), empty) // 0
Bytes.indexOf(empty, empty)                 // 0
```

On the UTF-8 bytes of text, the index counts bytes. `ï` is two bytes, so `v` is at byte 4:

```ant
Bytes.indexOf(String.toBytes("naïve"), String.toBytes("v")) // 4
```

A single byte matches even when it is the second byte of a character. `é` is `c3 a9`:

```ant
String.toBytes("é")                                         // Bytes(c3 a9)
Bytes.indexOf(String.toBytes("é"), Bytes.fromArray([0xa9])) // 1
```

Call it with named arguments, or in a pipe, where `b` is the first argument:

```ant
let b = Bytes.fromArray([1, 2, 3])
Bytes.indexOf(b: b, sub: Bytes.fromArray([3])) // 2
b | Bytes.indexOf(Bytes.fromArray([2]))        // 1
```

## Applications

Split an HTTP request at the blank line that ends its headers:

```ant
func splitHead(request: Bytes): Array[Bytes] | null {
  let blank = String.toBytes("\r\n\r\n")
  let end = Bytes.indexOf(request, blank) ?? return null
  let head = Bytes.slice(request, 0, end)
  let body = Bytes.slice(request, end + Bytes.length(blank), Bytes.length(request))
  [head, body]
}

let parts = splitHead(String.toBytes("GET / HTTP/1.1\r\nHost: example.com\r\n\r\nhello"))
assert(parts != null)
String.fromBytes(parts[0]) // "GET / HTTP/1.1\r\nHost: example.com"
parts[1]                   // Bytes(68 65 6c 6c 6f)

splitHead(String.toBytes("GET / HTTP/1.1\r\nHost: exa")) // null
```

Read a zero-terminated name out of a fixed-size field, and return early with `?? return` when the
field has no zero byte:

```ant
func cString(field: Bytes): string | null {
  let n = Bytes.indexOf(field, Bytes.fromArray([0])) ?? return null
  String.fromBytes(Bytes.slice(field, 0, n))
}

println(cString(Bytes.fromArray([0x61, 0x6e, 0x74, 0, 0, 0])))
println(cString(Bytes.fromArray([0x61, 0x6e, 0x74])))
```

```text
ant
null
```

Find every occurrence by searching again after each match, and stop with `?? break`:

```ant
let data = Bytes.fromArray([0xff, 0xd8, 1, 2, 0xff, 0xd8, 3])
let marker = Bytes.fromArray([0xff, 0xd8])
var from = 0
let hits: Array[int] = []
while true {
  let i = Bytes.indexOf(Bytes.slice(data, from, Bytes.length(data)), marker) ?? break
  hits.push(from + i)
  from = from + i + Bytes.length(marker)
}
println(hits)
```

```text
[0, 4]
```

Check a file signature with [`Bytes.startsWith`](startsWith.md) when it must be at the start, and
with `Bytes.indexOf` when it may be anywhere:

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
let file = Bytes.fromArray([0, 0, 0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a])
Bytes.startsWith(file, png) // false
Bytes.indexOf(file, png)    // 2
```

## Properties and relations

When `sub` occurs, the slice of `b` at the index is `sub`:

```ant
let b = String.toBytes("mississippi")
let sub = String.toBytes("ssi")
let i = Bytes.indexOf(b, sub)
assert(i != null)
assert(Bytes.slice(b, i, i + Bytes.length(sub)) == sub)
```

The result is `0` exactly when [`Bytes.startsWith`](startsWith.md) is `true`:

```ant
let b = Bytes.fromArray([1, 2, 1, 2])
for xs in [[1], [1, 2], [2], [2, 1, 2], [3]] {
  let sub = Bytes.fromArray(xs)
  assert((Bytes.indexOf(b, sub) == 0) == Bytes.startsWith(b, sub))
}
```

On ASCII text, where every character is one byte, it gives the same index as
[`String.indexOf`](../String/indexOf.md):

```ant
let s = "the cat and the hat"
for t in ["cat", "the", "hat", "dog"] {
  assert(Bytes.indexOf(String.toBytes(s), String.toBytes(t)) == String.indexOf(s, t))
}
```

## Possible issues

A missing match is `null`, not `-1`, and the result is not an `int` until you narrow it, so
arithmetic on it is a compile error:

```ant
let i = Bytes.indexOf(Bytes.fromArray([1, 2, 3]), Bytes.fromArray([2]))
println(i + 1) // error: + needs both operands of one type, got int | null and int
```

Give a default with `??`, or compare against `null`:

```ant
let b = Bytes.fromArray([1, 2, 3])
(Bytes.indexOf(b, Bytes.fromArray([2])) ?? 0) + 1 // 2
Bytes.indexOf(b, Bytes.fromArray([9])) != null    // false
```

`sub` is `Bytes`, not a single byte. Wrap one byte with [`Bytes.fromArray`](fromArray.md):

```ant
let b = Bytes.fromArray([1, 2, 3])
println(Bytes.indexOf(b, 2)) // error: cannot use int where Bytes is expected
```

An empty array literal has no element type, so `Bytes.fromArray([])` does not compile. Write empty
bytes as `String.toBytes("")`, or annotate the array:

```ant
let b = Bytes.fromArray([1, 2, 3])
println(Bytes.indexOf(b, Bytes.fromArray([])))
// error: cannot infer the element type of an empty array; annotate it
```

Annotate it on a `let`:

```ant
let none: Array[int] = []
Bytes.indexOf(Bytes.fromArray([1, 2, 3]), Bytes.fromArray(none)) // 0
```

The index counts bytes, so on text it differs from the code point index that
[`String.indexOf`](../String/indexOf.md) gives once a character takes more than one byte:

```ant
Bytes.indexOf(String.toBytes("café!"), String.toBytes("!")) // 5
String.indexOf("café!", "!")                                // 4
```

Slicing at a byte match inside a character cuts it in two, and the halves are not valid UTF-8:

```ant
let b = String.toBytes("é")
let i = Bytes.indexOf(b, Bytes.fromArray([0xa9])) ?? 0
String.fromBytes(Bytes.slice(b, 0, i)) // null
```

## See also

[Bytes.startsWith](startsWith.md) · [Bytes.endsWith](endsWith.md) · [Bytes.slice](slice.md) ·
[Bytes.length](length.md) · [Bytes.fromArray](fromArray.md) · [String.indexOf](../String/indexOf.md)
· [String.toBytes](../String/toBytes.md) · [Array.indexOf](../Array/indexOf.md)
