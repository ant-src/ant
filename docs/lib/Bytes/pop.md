<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.pop

```ant signature
func Bytes.pop(b: Bytes): byte | null
```

Bytes.pop removes and returns the last byte of b, or returns null when b is empty.

<!-- end generated -->

## Details

- Changes `b` in place. Every name that shares it sees the shorter `Bytes`.

## Basic examples

Pop the last byte, and `b` is one byte shorter:

```ant
let b = Bytes.fromArray([1, 2, 3])
Bytes.pop(b) // 3
b            // Bytes(01 02)
```

Empty `Bytes` gives `null`:

```ant
let empty = Bytes.new()
Bytes.pop(empty) // null
empty            // Bytes()
```

Each call removes one more byte, until none is left:

```ant
let b = String.toBytes("hi")
b.pop() // 105
b.pop() // 104
b.pop() // null
```

## Scope

The result is a `byte`, not a one-byte `Bytes`. Slice before popping to keep the last byte as
`Bytes`:

```ant
let b = String.toBytes("ant")
let n = b.length()
b.slice(n - 1, n) // Bytes(74)
b.pop()           // 116
```

Bytes from text are UTF-8, so a character outside ASCII comes off one byte at a time:

```ant
let b = String.toBytes("é")
b.pop() // 169
b       // Bytes(c3)
```

`Bytes` bound by `let` can still lose bytes, and every name that shares it sees the change,
including the caller of a function that pops:

```ant
func dropLast(b: Bytes) {
  b.pop()
}

let a = Bytes.fromArray([1, 2, 3])
let alias = a
dropLast(alias)
println(a)
```

```text
Bytes(01 02)
```

A copy made with `Bytes.clone` or `Bytes.slice` is separate, so popping it leaves the original
alone:

```ant
let original = Bytes.fromArray([1, 2, 3])
let copy = original.clone()
copy.pop()                 // 3
original.slice(0, 2).pop() // 2
original                   // Bytes(01 02 03)
```

A `for` loop follows the current length, so popping inside it stops the loop earlier:

```ant
let b = Bytes.fromArray([1, 2, 3])
for x in b {
  if x == byte(2) {
    b.pop()
  }
  println(x)
}
```

```text
1
2
```

In a pipe, the bytes are the first argument:

```ant
let b = Bytes.fromArray([0xca, 0xfe])
b | Bytes.pop() // 254
b               // Bytes(ca)
```

## Applications

Bind the result and check it for `null` when empty `Bytes` needs its own handling. After the guard,
the result is a `byte`:

```ant
func lastByte(b: Bytes): string {
  let x = b.pop()
  if x == null {
    return "empty"
  }
  "last ${x}"
}

println(lastByte(String.toBytes("ant")))
println(lastByte(Bytes.new()))
```

```text
last 116
empty
```

Drop one trailing newline from a line read as bytes:

```ant
func chomp(line: Bytes) {
  if line.endsWith(Bytes.fromArray([0x0a])) {
    line.pop()
  }
}

let line = String.toBytes("ant\n")
chomp(line)
println(String.fromBytes(line))
```

```text
ant
```

Strip the zero bytes that pad a fixed-size field:

```ant
func trimZeros(field: Bytes) {
  while field.length() > 0 && field[field.length() - 1] == byte(0) {
    field.pop()
  }
}

let field = Bytes.fromArray([0x41, 0x42, 0, 0, 0])
trimZeros(field)
println(field)
```

```text
Bytes(41 42)
```

Use `Bytes` as a stack of bytes: apply backspaces (byte 8) to typed input, where popping empty
output does nothing:

```ant
func applyBackspaces(input: Bytes): Bytes {
  let out = Bytes.new()
  for x in input {
    if x == byte(8) {
      out.pop()
    } else {
      out.push(x)
    }
  }
  out
}

let typed = Bytes.fromArray([8, 0x61, 0x6e, 0x78, 8, 0x74])
println(String.fromBytes(applyBackspaces(typed)))
```

```text
ant
```

Drain `Bytes` from the end with a `while` loop, and stop at `null` with `?? break`:

```ant
let src = String.toBytes("abc")
let reversed = Bytes.new()
while true {
  let x = src.pop() ?? break
  reversed.push(x)
}
println(String.fromBytes(reversed))
println(src)
```

```text
cba
Bytes()
```

## Properties and relations

`Bytes.pop` returns the byte at the last index, and the length drops by one:

```ant
let b = Bytes.fromArray([4, 5, 6])
let n = b.length()
let last = b[n - 1]
assert(b.pop() == last)
assert(b.length() == n - 1)
```

`Bytes.pop` undoes `Bytes.push`:

```ant
let b = Bytes.fromArray([1, 2])
b.push(byte(9))
assert(b.pop() == byte(9))
assert(b == Bytes.fromArray([1, 2]))
```

On non-empty `Bytes`, what remains is what `Bytes.truncate` keeps with one byte fewer, and what
`Bytes.slice` copies without the last byte:

```ant
let a = String.toBytes("anthill")
let b = a.clone()
let rest = a.slice(0, a.length() - 1)
a.pop()
b.truncate(b.length() - 1)
assert(a == b)
assert(a == rest)
```

## Possible issues

The result is `byte | null`, not `byte`, so using it as a `byte` is a compile error:

```ant
let b = Bytes.fromArray([7])
let x: byte = b.pop() // error: cannot use byte | null where byte is expected
```

Give a default with `??` when a byte is needed:

```ant
let b = Bytes.fromArray([7])
b.pop() ?? byte(0) // 7
```

`Bytes` is not an array, so `Array.pop` does not take it:

```ant
let b = Bytes.fromArray([1])
println(Array.pop(b)) // error: cannot use Bytes where Array[T] is expected
```

Popping UTF-8 text removes a byte, not a character. Text whose last character takes more than one
byte is no longer valid UTF-8, and `String.fromBytes` returns `null`:

```ant
let b = String.toBytes("café")
b.pop()             // 169
String.fromBytes(b) // null
```

## See also

[Bytes.push](push.md) · [Bytes.truncate](truncate.md) · [Bytes.length](length.md) ·
[Bytes.endsWith](endsWith.md) · [Bytes.slice](slice.md) · [Bytes.clone](clone.md) ·
[Array.pop](../Array/pop.md) · [Bytes](index.md)
