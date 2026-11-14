<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.same

```ant signature
func Bytes.same(a: Bytes, b: Bytes): bool
```

Bytes.same reports whether a and b are one collection.

<!-- end generated -->

## Details

- Returns `false` for two different collections, even when they hold the same bytes.
- `==` compares the bytes, so use `Bytes.same` to tell a shared collection from a copy. A write
  through one name shows through every name that is the same collection.

## Basic examples

A collection is the same as itself, and as any other name for it:

```ant
let a = Bytes.fromArray([1, 2])
let b = a
Bytes.same(a, a) // true
Bytes.same(a, b) // true
```

A copy holds equal bytes, and is not the same collection:

```ant
let a = Bytes.fromArray([1, 2])
let c = a.clone()
a == c           // true
Bytes.same(a, c) // false
```

## Scope

Call it through `.` on any `Bytes`:

```ant
let a = String.toBytes("ant")
let b = a
a.same(b) // true
```

Two empty collections are equal, and still two collections:

```ant
let a = Bytes.new()
let b = Bytes.new()
a == b           // true
Bytes.same(a, b) // false
```

Every library function that returns `Bytes` returns a new collection, even when it holds the same
bytes as its argument:

```ant
let a = Bytes.fromArray([7, 8, 9])
Bytes.same(a, a.slice(0, len(a)))           // false
Bytes.same(a, Bytes.concat([a]))            // false
Bytes.same(a, Bytes.fromArray(a.toArray())) // false
```

`String.toBytes` gives a new collection on every call:

```ant
let s = "hi"
Bytes.same(s.toBytes(), s.toBytes()) // false
```

A spread copies each field, and a `Bytes` field holds a reference, so both records share one
collection:

```ant
type Packet = { id: int, body: Bytes }
let p = Packet{ id: 1, body: String.toBytes("ping") }
let q = Packet{ ...p, id: 2 }
Bytes.same(p.body, q.body) // true
```

An array may hold one collection in several slots:

```ant
let a = Bytes.fromArray([0])
let slots = [a, a, a.clone()]
Bytes.same(slots[0], slots[1]) // true
Bytes.same(slots[0], slots[2]) // false
```

In a pipeline, the left side is `a`:

```ant
let a = Bytes.fromArray([1])
let b = a
a | Bytes.same(b) // true
```

Name the arguments, in any order:

```ant
let a = Bytes.fromArray([1])
Bytes.same(a: a, b: a)         // true
Bytes.same(b: a, a: a.clone()) // false
```

## Applications

Replace the contents of `dst` with those of `src`. Truncating first would erase `src` when the
caller passes one collection as both, so the guard returns early:

```ant
func assign(dst: Bytes, src: Bytes) {
  if Bytes.same(dst, src) {
    return
  }
  dst.truncate(0)
  dst.extend(src)
}

let buf = String.toBytes("old")
let msg = String.toBytes("new")
assign(buf, msg)
println(String.fromBytes(buf))
assign(buf, buf)
println(String.fromBytes(buf))
```

```text
new
new
```

Keep a frame before reusing the buffer it was read into. Without `clone`, every saved frame is the
one buffer, and shows its last contents:

```ant
let buf = Bytes.new()
let shared: Array[Bytes] = []
let kept: Array[Bytes] = []
for word in ["ping", "pong"] {
  buf.truncate(0)
  buf.extend(word.toBytes())
  shared.push(buf)
  kept.push(buf.clone())
}
println(Bytes.same(shared[0], shared[1]))
println(shared.map({ String.fromBytes(x) ?? "" }))
println(kept.map({ String.fromBytes(x) ?? "" }))
```

```text
true
["pong", "pong"]
["ping", "pong"]
```

## Properties and relations

`Bytes.same` is reflexive and symmetric:

```ant
let a = Bytes.fromArray([1, 2])
let b = a
assert(Bytes.same(a, a))
assert(Bytes.same(a, b) == Bytes.same(b, a))
```

The same collection is always `==` to itself, so `Bytes.same` implies `==`, but not the reverse:

```ant
let a = Bytes.fromArray([1, 2])
let b = a
let c = a.clone()
assert(Bytes.same(a, b) && a == b)
assert(a == c && !Bytes.same(a, c))
```

A write through one name shows through the other when they are the same collection, and not through
a copy:

```ant
let a = Bytes.fromArray([1])
let b = a
let c = a.clone()
a.push(byte(2))
assert(Bytes.same(a, b) && b == Bytes.fromArray([1, 2]))
assert(!Bytes.same(a, c) && c == Bytes.fromArray([1]))
```

`Reflect.same` reports the same for two reflected `Bytes`:

```ant
let a = Bytes.fromArray([1])
let t = type(Bytes)
assert(Reflect.same(Reflect.valueOf(t, a), Reflect.valueOf(t, a)))
assert(!Reflect.same(Reflect.valueOf(t, a), Reflect.valueOf(t, a.clone())))
```

## Possible issues

`==` compares the bytes, so it cannot tell whether two names share a collection. Use `Bytes.same`
when that is the question:

```ant
let a = Bytes.fromArray([1])
let copy = a.clone()
a == copy           // true
Bytes.same(a, copy) // false
```

An `Array[byte]` is not `Bytes`, so passing one is a compile error:

```ant
let a = Bytes.fromArray([1])
println(Bytes.same(a, a.toArray())) // error: cannot use Array[byte] where Bytes is expected
```

A nullable value, such as one read from a map, must be narrowed first:

```ant
let a = Bytes.fromArray([1])
let m: Map[string, Bytes] = Map{ k: a }
println(Bytes.same(m["k"], a)) // error: cannot use Bytes | null where Bytes is expected
```

```ant
let a = Bytes.fromArray([1])
let m: Map[string, Bytes] = Map{ k: a }
let found = m["k"]
found != null && Bytes.same(found, a) // true
```

## See also

[Bytes.clone](clone.md) · [Bytes.slice](slice.md) · [Bytes.concat](concat.md) ·
[Array.same](../Array/same.md) · [Map.same](../Map/same.md) · [Set.same](../Set/same.md) ·
[Record.same](../Record/same.md) · [Reflect.same](../Reflect/same.md)
