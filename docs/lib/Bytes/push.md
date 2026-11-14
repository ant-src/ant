<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.push

```ant signature
func Bytes.push(b: Bytes, x: byte): void
```

Bytes.push appends x to b.

<!-- end generated -->

## Details

- `Bytes.push` changes `b` in place. Every name for the same `Bytes` sees the new byte.
- An integer literal is an `int`, so write `byte(65)`.
- A `for` loop over `b` visits bytes pushed during the loop.

## Basic examples

Append bytes to empty `Bytes`:

```ant
let b = Bytes.new()
Bytes.push(b, byte(0x41))
Bytes.push(b, byte(0xff))
b      // Bytes(41 ff)
len(b) // 2
```

Append a byte after the UTF-8 bytes of a string:

```ant
let line = "hi".toBytes()
Bytes.push(line, byte(0x0a))
line // Bytes(68 69 0a)
```

## Scope

Push onto `Bytes` that already holds bytes, from any source:

```ant
let magic = Bytes.fromArray([0x89, 0x50])
magic.push(byte(0x4e))
magic // Bytes(89 50 4e)
```

A `byte` read from other `Bytes`, or from a `byte` variable, goes in as it is:

```ant
let src = "Ant".toBytes()
let dst = Bytes.new()
dst.push(src[0])
let zero = byte(0)
dst.push(zero)
dst // Bytes(41 00)
```

`Bytes` is shared by reference, so a push through one name shows through every other:

```ant
let b = Bytes.new()
let alias = b
alias.push(byte(7))
b                    // Bytes(07)
Bytes.same(alias, b) // true
```

A function that pushes onto its parameter changes the caller's `Bytes`:

```ant
func terminate(b: Bytes): void {
  b.push(byte(0))
}

let name = "ant".toBytes()
terminate(name)
println(name)
```

```text
Bytes(61 6e 74 00)
```

## Applications

Encode an unsigned integer as a variable-length integer, seven bits per byte with the high bit set
on every byte but the last:

```ant
func varint(n: int): Bytes {
  let out = Bytes.new()
  var v = n
  while v >= 128 {
    out.push(byte(v % 128 + 128))
    v /= 128
  }
  out.push(byte(v))
  out
}

println(varint(1))
println(varint(300))
println(varint(16384))
```

```text
Bytes(01)
Bytes(ac 02)
Bytes(80 80 01)
```

Append a one-byte checksum, the XOR of every byte before it:

```ant
func withChecksum(payload: Bytes): Bytes {
  let out = payload.clone()
  var sum = byte(0)
  for x in payload {
    sum = Bits.xor(sum, x)
  }
  out.push(sum)
  out
}

let frame = withChecksum(Bytes.fromArray([0x01, 0x02, 0x04]))
println(Encoding.toHex(frame))
```

```text
01020407
```

Keep only the printable ASCII bytes of some input:

```ant
let input = Bytes.fromArray([0x41, 0x00, 0x6e, 0x07, 0x74])
let clean = Bytes.new()
for x in input {
  if x >= byte(0x20) && x < byte(0x7f) {
    clean.push(x)
  }
}
println(String.fromBytes(clean))
```

```text
Ant
```

## Properties and relations

`Bytes.push` adds one byte at the end and leaves the others as they were:

```ant
let b = Bytes.fromArray([1, 2, 3])
let before = b.clone()
b.push(byte(9))
assert(len(b) == len(before) + 1)
assert(b[len(b) - 1] == byte(9))
assert(b.slice(0, len(before)) == before)
```

[Bytes.pop](pop.md) undoes it:

```ant
let b = "ant".toBytes()
let before = b.clone()
b.push(byte(0x21))
assert(b.pop() == byte(0x21))
assert(b == before)
```

It is [Bytes.extend](extend.md) with one byte:

```ant
let a = Bytes.fromArray([1, 2])
let b = a.clone()
a.push(byte(3))
b.extend(Bytes.fromArray([3]))
assert(a == b)
```

When the size is known, [Bytes.new](new.md) and assignment by index build the same bytes:

```ant
let pushed = Bytes.new()
let sized = Bytes.new(3)
for i in Iter.range(0, 3) {
  pushed.push(byte(i * 10))
  sized[i] = byte(i * 10)
}
assert(pushed == sized)
```

## Possible issues

An integer literal is an `int`, never a `byte`, so pushing one is a compile error. Convert it with
`byte(...)`:

```ant
let b = Bytes.new()
b.push(65) // error: cannot use int where byte is expected
```

An `int` variable needs the same conversion, which panics when the value does not fit in a byte:

```ant
let b = Bytes.new()
let n = 300
b.push(byte(n)) // panic: 300 out of range for byte
```

`Bytes.push` returns `void`, not the `Bytes`, so it does not chain. Use the `Bytes` itself after the
call:

```ant
let b = Bytes.new()
let grown: Bytes = b.push(byte(1)) // error: cannot use void where Bytes is expected
let n = b.push(byte(2)).length()   // error: void has no field or function length
```

`Bytes.push` takes one `byte`. Append many with [Bytes.extend](extend.md), and a fixed-size number
with [Bytes.append](append.md):

```ant
let b = Bytes.new()
b.push("hi".toBytes()) // error: cannot use Bytes where byte is expected
```

A `for` loop over `b` reaches the bytes pushed during the loop, so `for x in b { b.push(x) }` never
ends. Double the bytes with [Bytes.extend](extend.md), which appends `b` as it was at the call, or
loop over the indexes that exist before the first push:

```ant
let b = Bytes.fromArray([1, 2])
b.extend(b)
b // Bytes(01 02 01 02)
```

```ant
let b = Bytes.fromArray([1, 2])
for i in Iter.range(0, len(b)) {
  b.push(b[i])
}
println(b)
```

```text
Bytes(01 02 01 02)
```

## See also

[Bytes.extend](extend.md) · [Bytes.append](append.md) · [Bytes.pop](pop.md) ·
[Bytes.truncate](truncate.md) · [Bytes.new](new.md) · [Array.push](../Array/push.md)
