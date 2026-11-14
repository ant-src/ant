<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.clone

```ant signature
func Bytes.clone(b: Bytes): Bytes
```

Bytes.clone returns a copy of b.

<!-- end generated -->

## Details

- The copy is a new collection: writing, growing, or truncating it leaves `b` as it is, and the
  other way around.
- Assignment, arguments, containers, and channel sends share `Bytes` rather than copy them. Clone
  before you hand on bytes that you go on changing.

## Basic examples

Copy bytes, then change the copy:

```ant
let b = Bytes.fromArray([1, 2, 3])
let c = Bytes.clone(b)
c[0] = byte(9)
c // Bytes(09 02 03)
b // Bytes(01 02 03)
```

The copy is equal to the original but is another collection:

```ant
let b = "ant".toBytes()
Bytes.clone(b) == b           // true
Bytes.same(Bytes.clone(b), b) // false
```

## Scope

Empty bytes give empty bytes, again a new collection:

```ant
let empty = Bytes.new()
empty.clone()                    // Bytes()
Bytes.same(empty.clone(), empty) // false
```

Growth of the copy does not reach the original:

```ant
let b = Bytes.fromArray([0xca, 0xfe])
let c = b.clone()
c.extend(Bytes.fromArray([0xba, 0xbe]))
c      // Bytes(ca fe ba be)
len(b) // 2
```

Neither does truncation of the original reach the copy:

```ant
let b = "hello".toBytes()
let c = b.clone()
b.truncate(0)
b // Bytes()
c // Bytes(68 65 6c 6c 6f)
```

Call it through `.` or in a pipe:

```ant
let b = Bytes.fromArray([1, 2])
b.clone()         // Bytes(01 02)
b | Bytes.clone() // Bytes(01 02)
```

## Applications

Reuse one buffer to build several messages, and send a copy of each, so the next message does not
overwrite the one in the channel:

```ant
let out: Channel[Bytes] = Channel.new(2)
let buf = Bytes.new()
for word in ["ant", "hill"] {
  buf.truncate(0)
  buf.extend(word.toBytes())
  out.send(buf.clone())
}
out.close()
for m in out.values() {
  println(String.fromBytes(m))
}
```

```text
ant
hill
```

Keep a snapshot of a buffer as it was before a change:

```ant
let header = Bytes.fromArray([0x01, 0x00, 0x00, 0x00])
let before = header.clone()
header[0] = byte(0x02)
println(Encoding.toHex(before))
println(Encoding.toHex(header))
```

```text
01000000
02000000
```

Copy bytes a record keeps, so a caller that changes its buffer later does not change the record:

```ant
type Packet = { id: int, payload: Bytes }

func packet(id: int, payload: Bytes): Packet {
  Packet{ id, payload: payload.clone() }
}

let buf = "ping".toBytes()
let p = packet(1, buf)
buf[1] = byte(0x6f)
println(String.fromBytes(buf))
println(String.fromBytes(p.payload))
```

```text
pong
ping
```

Use a copy as a map key, since changing bytes while they are a key is a program error:

```ant
let seen: Map[Bytes, int] = Map{}
let buf = "a".toBytes()
seen[buf.clone()] = 1
buf.push(byte(0x62))
seen["a".toBytes()] // 1
```

## Properties and relations

The copy has the same length and the same bytes, and is not the same collection:

```ant
let b = "日本🐜".toBytes()
let c = b.clone()
assert(len(c) == len(b))
assert(c == b)
assert(!Bytes.same(c, b))
```

It gives the same bytes as a slice of the whole, or a concatenation of one part, which are copies
too:

```ant
let b = Bytes.fromArray([5, 6, 7])
assert(b.clone() == b.slice(0, len(b)))
assert(b.clone() == Bytes.concat([b]))
```

A round trip through an array gives a copy as well:

```ant
let b = Bytes.fromArray([0, 128, 255])
let c = Bytes.fromArray(b.toArray())
assert(c == b.clone())
assert(!Bytes.same(c, b))
```

## Possible issues

Assignment does not copy. Both names see the change:

```ant
let b = Bytes.fromArray([1, 2, 3])
let alias = b
alias[0] = byte(9)
b // Bytes(09 02 03)
```

A channel send shares the bytes too. Writing the buffer after sending it changes what the receiver
gets, so both messages read `hill` here:

```ant
let out: Channel[Bytes] = Channel.new(2)
let buf = Bytes.new()
for word in ["ant", "hill"] {
  buf.truncate(0)
  buf.extend(word.toBytes())
  out.send(buf)
}
out.close()
for m in out.values() {
  println(String.fromBytes(m))
}
```

```text
hill
hill
```

A clone is needed only for bytes you go on changing after you hand them on. A library function
returns new `Bytes` that nothing else holds, so `"ant".toBytes().clone()` copies twice for nothing.
A library function that takes `Bytes` only reads them, so `Encoding.toHex(b)` needs no clone either:

```ant
let b = "ant".toBytes()
b[0] = byte(0x41)
Encoding.toHex(b)   // "416e74"
String.fromBytes(b) // "Ant"
```

`Bytes.clone` takes only `Bytes`. A value that may be `null` is a compile error; handle the `null`
first:

```ant
let b: Bytes | null = null
let c = Bytes.clone(b) // error: cannot use Bytes | null where Bytes is expected
```

It does not copy an array or a string:

```ant
let a = Bytes.clone([1, 2]) // error: cannot use Array[int] where Bytes is expected
let s = Bytes.clone("ab")   // error: cannot use string where Bytes is expected
```

## See also

[Bytes.same](same.md), [Bytes.slice](slice.md), [Bytes.concat](concat.md), [Bytes.new](new.md),
[Bytes.extend](extend.md), [Bytes.copy](copy.md)
