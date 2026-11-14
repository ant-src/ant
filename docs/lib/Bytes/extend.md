<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.extend

```ant signature
func Bytes.extend(b: Bytes, src: Bytes): void
```

Bytes.extend appends the bytes src holds at the call to b.

<!-- end generated -->

## Details

- To keep `b` as it is and get new `Bytes`, use [Bytes.concat](concat.md).
- Appends the bytes `src` holds at the call, so `b.extend(b)` doubles `b`, and a later write to
  `src` does not reach `b`.

## Basic examples

Append the bytes of one collection to another:

```ant
let b = Bytes.fromArray([0xca, 0xfe])
Bytes.extend(b, Bytes.fromArray([0xba, 0xbe]))
b // Bytes(ca fe ba be)
```

Append the UTF-8 bytes of a string:

```ant
let b = "ant".toBytes()
Bytes.extend(b, "hill".toBytes())
String.fromBytes(b) // "anthill"
```

## Scope

Empty `src` leaves `b` as it was, and empty `b` ends up with the bytes of `src`:

```ant
let b = Bytes.fromArray([1, 2])
b.extend(Bytes.new())
b // Bytes(01 02)
let empty = Bytes.new()
empty.extend(Bytes.fromArray([3, 4]))
empty // Bytes(03 04)
```

`Bytes` extended with itself doubles, because the call reads `src` as it was at the start:

```ant
let b = Bytes.fromArray([1, 2])
b.extend(b)
b // Bytes(01 02 01 02)
```

The bytes are copied in. Writing to `src` afterwards leaves `b` as it is:

```ant
let b = Bytes.fromArray([1])
let src = Bytes.fromArray([2])
b.extend(src)
src[0] = byte(9)
b   // Bytes(01 02)
src // Bytes(09)
```

## Applications

Build a message in one buffer: a type byte, a two-byte length, then the payload:

```ant
func message(kind: byte, payload: Bytes): Bytes {
  let out = Bytes.new()
  out.push(kind)
  out.append(type(uint16), uint16(len(payload)), Bytes.BigEndian{})
  out.extend(payload)
  out
}

println(message(byte(1), "hi".toBytes()))
```

```text
Bytes(01 00 02 68 69)
```

Append only part of another `Bytes`. [Bytes.slice](slice.md) copies the range first, so `src` can be
any range of any `Bytes`, `b` included:

```ant
let input = "id=42;".toBytes()
let out = "value: ".toBytes()
out.extend(input.slice(3, 5))
println(String.fromBytes(out) ?? "")
out.extend(out.slice(0, 5))
println(String.fromBytes(out) ?? "")
```

```text
value: 42
value: 42value
```

Read a file in chunks and collect them in one buffer:

<!-- test: file data.bin
hello, chunks
-->

```ant
func readFile(path: string): Result[Bytes, Error] {
  let f = try File.open(path)
  defer { f.close() }
  let out = Bytes.new()
  while true {
    let chunk = try f.readChunk(4) ?? break
    out.extend(chunk)
  }
  Result.ok(out)
}

let data = try readFile("data.bin")
println(len(data))
print(String.fromBytes(data) ?? "")
```

```text
14
hello, chunks
```

Reuse one buffer for each record: [Bytes.truncate](truncate.md) empties it, and `Bytes.extend` fills
it again:

```ant
let line = Bytes.new()
for word in ["ant", "hill"] {
  line.truncate(0)
  line.extend(word.toBytes())
  line.extend("\n".toBytes())
  print(String.fromBytes(line) ?? "")
}
```

```text
ant
hill
```

## Properties and relations

The length grows by the length of `src`:

```ant
let b = "日本".toBytes()
let src = Bytes.fromArray([0, 255])
let before = len(b)
b.extend(src)
assert(len(b) == before + len(src))
```

It gives in place what [Bytes.concat](concat.md) returns as new `Bytes`:

```ant
let b = Bytes.fromArray([1, 2])
let src = Bytes.fromArray([3])
let joined = Bytes.concat([b, src])
b.extend(src)
assert(b == joined)
```

It is [Bytes.push](push.md) of each byte of `src`, in order:

```ant
let a = Bytes.fromArray([1])
let c = Bytes.fromArray([1])
let src = Bytes.fromArray([2, 3])
a.extend(src)
for x in src {
  c.push(x)
}
assert(a == c)
```

`b` stays the same collection, so every other name for it sees the new bytes:

```ant
let b = Bytes.fromArray([1])
let alias = b
b.extend(Bytes.fromArray([2]))
assert(Bytes.same(alias, b))
alias // Bytes(01 02)
```

## Possible issues

`Bytes.extend` returns `void`, not the extended bytes. To keep `b` and get new `Bytes`, use
[Bytes.concat](concat.md):

```ant
let b = Bytes.fromArray([1])
let src = Bytes.fromArray([2])
Bytes.concat([b, src]) // Bytes(01 02)
b                      // Bytes(01)
```

`src` must be `Bytes`. A string or an array of integers is a compile error; convert it with
[String.toBytes](../String/toBytes.md) or [Bytes.fromArray](fromArray.md):

```ant
let b = Bytes.new()
b.extend("hi")   // error: cannot use string where Bytes is expected
b.extend([1, 2]) // error: cannot use Array[int] where Bytes is expected
```

To append one byte, use [Bytes.push](push.md). An integer literal is an `int`, so convert it:

```ant
let b = Bytes.new()
b.push(byte(10))
b // Bytes(0a)
```

A `for` loop over `b` follows its current length, so it also visits the bytes that `Bytes.extend`
appends during the loop. Loop over a copy from [Bytes.clone](clone.md) to visit only the bytes that
were there at the start:

```ant
let b = Bytes.fromArray([1, 2])
for x in b.clone() {
  b.extend(Bytes.fromArray([x * byte(10)]))
}
b // Bytes(01 02 0a 14)
```

## See also

[Bytes.push](push.md), [Bytes.concat](concat.md), [Bytes.append](append.md),
[Bytes.truncate](truncate.md), [Bytes.clone](clone.md), [Bytes.same](same.md),
[Array.extend](../Array/extend.md)
