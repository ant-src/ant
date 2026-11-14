<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.truncate

```ant signature
func Bytes.truncate(b: Bytes, n: int): void
```

Bytes.truncate keeps the first n bytes of b. It panics if n is outside 0 ..= Bytes.length(b).

<!-- end generated -->

## Details

- It changes `b` in place: every name that refers to the same `Bytes` sees the shorter length.
- It only shortens; to add bytes, use `Bytes.push` or `Bytes.extend`.
- `b.truncate(0)` empties `b` so it can be filled again. The dropped bytes never come back: bytes
  added later are the new ones.

## Basic examples

Keep the first three bytes:

```ant
let b = Bytes.fromArray([1, 2, 3, 4, 5])
Bytes.truncate(b, 3)
b // Bytes(01 02 03)
```

Truncate to `0` to empty `Bytes`:

```ant
let b = String.toBytes("hello")
Bytes.truncate(b, 0)
b      // Bytes()
len(b) // 0
```

## Scope

`n` equal to the length changes nothing, and empty `Bytes` truncates to `0`:

```ant
let b = Bytes.fromArray([7, 8])
b.truncate(len(b))
b // Bytes(07 08)
let empty = Bytes.new()
empty.truncate(0)
empty // Bytes()
```

`b` is shared by reference, so every name for it, a record field included, sees the change:

```ant
type Packet = { data: Bytes }
let b = String.toBytes("ping!")
let p = Packet{ data: b }
b.truncate(4)
p.data                   // Bytes(70 69 6e 67)
String.fromBytes(p.data) // "ping"
Bytes.same(b, p.data)    // true
```

Call it through `.`, in a pipe, or with a named argument:

```ant
let b = String.toBytes("anthill")
b.truncate(6)
b | Bytes.truncate(5)
Bytes.truncate(b, n: 4)
b // Bytes(61 6e 74 68)
```

A `for` loop follows the current length, so truncating inside the loop ends it early:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
for x in b {
  println(x)
  if x == byte(2) {
    b.truncate(2)
  }
}
```

```text
1
2
```

## Applications

Remove a trailing line break, `\n` or `\r\n`, in place:

```ant
func chomp(b: Bytes) {
  if !b.endsWith(String.toBytes("\n")) {
    return
  }
  b.truncate(len(b) - 1)
  if b.endsWith(String.toBytes("\r")) {
    b.truncate(len(b) - 1)
  }
}

let line = String.toBytes("GET / HTTP/1.1\r\n")
chomp(line)
println(String.fromBytes(line))
```

```text
GET / HTTP/1.1
```

Strip the zero bytes that pad a fixed-size field:

```ant
let field = Bytes.fromArray([0x61, 0x6e, 0x74, 0, 0, 0, 0, 0])
var n = len(field)
while n > 0 && field[n - 1] == byte(0) {
  n -= 1
}
field.truncate(n)
println(String.fromBytes(field))
```

```text
ant
```

Reuse one buffer for each record instead of making a new one every time:

```ant
let buf = Bytes.new()
for name in ["ant", "bee", "wasp"] {
  buf.truncate(0)
  buf.extend(String.toBytes(name))
  buf.push(byte(0x0a))
  println(Encoding.toHex(buf))
}
```

```text
616e740a
6265650a
776173700a
```

Undo a partial write: remember the length, append, and truncate back when the result is too long:

```ant
func addField(b: Bytes, value: string, limit: int): bool {
  let mark = len(b)
  b.extend(String.toBytes(value))
  b.push(byte(0x3b))
  if len(b) > limit {
    b.truncate(mark)
    return false
  }
  true
}

let out = Bytes.new()
for v in ["id=7", "name=ant", "colony=formicidae"] {
  if !addField(out, v, 16) {
    println("skipped ${v}")
  }
}
println(String.fromBytes(out))
```

```text
skipped colony=formicidae
id=7;name=ant;
```

## Properties and relations

Afterwards the length is `n`, and the bytes are those `Bytes.slice` takes from `0` to `n`:

```ant
let b = String.toBytes("añt🐜")
for n in Iter.range(0, len(b) + 1) {
  let c = b.clone()
  c.truncate(n)
  assert(len(c) == n)
  assert(c == b.slice(0, n))
}
```

It undoes `Bytes.extend`: truncating back to the old length restores the old contents:

```ant
let b = Bytes.fromArray([1, 2, 3])
let before = b.clone()
b.extend(Bytes.fromArray([4, 5]))
b.truncate(len(before))
assert(b == before)
```

Bytes added after a truncation are the new ones, not the bytes it dropped:

```ant
let b = Bytes.fromArray([1, 2, 3])
b.truncate(1)
b.push(byte(9))
assert(b == Bytes.fromArray([1, 9]))
```

## Possible issues

`Bytes.truncate` does not grow `b`. A length above the current one panics; use `Bytes.extend` or
`Bytes.push` to add bytes:

```ant
let b = Bytes.fromArray([1, 2, 3])
b.truncate(4) // panic: Bytes.truncate needs a length within 0 ..= 3, got 4
```

A negative length panics rather than counting from the end. To drop the last `k` bytes, pass
`len(b) - k`:

```ant
let b = Bytes.fromArray([1, 2, 3])
b.truncate(-1) // panic: Bytes.truncate needs a length within 0 ..= 3, got -1
```

It changes the `Bytes` every alias refers to. For a shorter copy that leaves `b` as it is, use
`Bytes.slice`:

```ant
let original = String.toBytes("anthill")
let alias = original
alias.truncate(3)
original // Bytes(61 6e 74)
let copy = String.toBytes("anthill")
copy.slice(0, 3) // Bytes(61 6e 74)
copy             // Bytes(61 6e 74 68 69 6c 6c)
```

`n` counts bytes, not characters. Text from `String.toBytes` cut inside a multi-byte character is
invalid UTF-8: `String.fromBytes` returns `null`, and `String.fromBytesLossy` puts U+FFFD in place
of the cut character:

```ant
let b = String.toBytes("añt")
b.truncate(2)
b                        // Bytes(61 c3)
String.fromBytes(b)      // null
String.fromBytesLossy(b) // "a�"
```

An index that was valid before the truncation panics after it:

```ant
let b = Bytes.fromArray([1, 2, 3])
b.truncate(1)
b[0] // 1
b[2] // panic: index 2 out of range for Bytes of length 1
```

## See also

[Bytes.pop](pop.md) · [Bytes.slice](slice.md) · [Bytes.extend](extend.md) · [Bytes.push](push.md) ·
[Bytes.clone](clone.md) · [Bytes.new](new.md) · [Bytes.length](length.md) · [Bytes](index.md)
