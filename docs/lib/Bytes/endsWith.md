<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Bytes.endsWith

```ant signature
func Bytes.endsWith(b: Bytes, suffix: Bytes): bool
```

Bytes.endsWith reports whether b ends with suffix.

<!-- end generated -->

## Details

- Compares bytes exactly, with no regard for text: a suffix may start in the middle of a UTF-8
  character's bytes, and the two spellings of "é" differ.
- The empty `Bytes` is a suffix of every value, so `Bytes.endsWith(b, String.toBytes(""))` is always
  true.

## Basic examples

```ant
let data = String.toBytes("main.ant")
Bytes.endsWith(data, String.toBytes(".ant")) // true
Bytes.endsWith(data, String.toBytes(".go"))  // false
```

Call it through `.` on any `Bytes`:

```ant
let jpeg = Bytes.fromArray([0xff, 0xd8, 0xff, 0xe0, 0xff, 0xd9])
jpeg.endsWith(Bytes.fromArray([0xff, 0xd9])) // true
jpeg.endsWith(Bytes.fromArray([0xff, 0xe0])) // false
```

## Scope

Every value ends with the empty `Bytes`, and the empty `Bytes` ends with nothing else:

```ant
let none = String.toBytes("")
Bytes.endsWith(Bytes.fromArray([1, 2]), none) // true
Bytes.endsWith(none, none)                    // true
Bytes.endsWith(none, Bytes.fromArray([2]))    // false
```

A suffix longer than the value never matches, even when the value is inside it:

```ant
let tail = Bytes.fromArray([0xff, 0xd9])
Bytes.endsWith(tail, Bytes.fromArray([0xd8, 0xff, 0xd9])) // false
Bytes.endsWith(Bytes.fromArray([0xd8, 0xff, 0xd9]), tail) // true
```

Bytes from any source compare the same way, here from hex text and from a slice:

```ant
let b = Result.unwrap(Encoding.fromHex("cafebabe"))
b.endsWith(Result.unwrap(Encoding.fromHex("babe"))) // true
b.endsWith(Bytes.slice(b, 2, 4))                    // true
b.endsWith(Bytes.slice(b, 0, 2))                    // false
```

The suffix may be part of a UTF-8 character. "ñ" is the two bytes `c3 b1`, and the last of them
followed by "o" is a suffix of "año":

```ant
let word = String.toBytes("año")
word                                         // Bytes(61 c3 b1 6f)
word.endsWith(Bytes.fromArray([0xb1, 0x6f])) // true
word.endsWith(String.toBytes("ño"))          // true
```

In a pipeline, the value is the first argument:

```ant
String.toBytes("index.html") | Bytes.endsWith(String.toBytes(".html")) // true
```

## Applications

Check that a JPEG file is complete: it starts with the marker `ff d8` and ends with `ff d9`:

```ant
func isWholeJpeg(data: Bytes): bool {
  let start = Bytes.fromArray([0xff, 0xd8])
  let end = Bytes.fromArray([0xff, 0xd9])
  data.startsWith(start) && data.endsWith(end)
}

isWholeJpeg(Bytes.fromArray([0xff, 0xd8, 0x00, 0xff, 0xd9])) // true
isWholeJpeg(Bytes.fromArray([0xff, 0xd8, 0x00, 0x00]))       // false
```

Strip a network line ending, `\r\n` or `\n`, from a line read as bytes:

```ant
func chomp(line: Bytes): Bytes {
  let crlf = Bytes.fromArray([0x0d, 0x0a])
  let lf = Bytes.fromArray([0x0a])
  if line.endsWith(crlf) {
    return Bytes.slice(line, 0, len(line) - 2)
  }
  if line.endsWith(lf) {
    return Bytes.slice(line, 0, len(line) - 1)
  }
  line
}

chomp(String.toBytes("HELO ant\r\n")) // Bytes(48 45 4c 4f 20 61 6e 74)
chomp(String.toBytes("QUIT\n"))       // Bytes(51 55 49 54)
chomp(String.toBytes("QUIT"))         // Bytes(51 55 49 54)
```

## Properties and relations

`Bytes.endsWith` agrees with slicing the last bytes of `b`:

```ant
let b = Bytes.fromArray([1, 2, 3, 4])
for s in [[4], [3, 4], [2, 4], [], [1, 2, 3, 4], [0, 1, 2, 3, 4]] {
  let suffix = Bytes.fromArray(s)
  let n = len(b)
  let m = len(suffix)
  assert(b.endsWith(suffix) == (m <= n && Bytes.slice(b, n - m, n) == suffix))
}
```

A value always ends with what [`Bytes.concat`](concat.md) puts last:

```ant
let head = String.toBytes("ant")
let tail = String.toBytes("hill")
assert(Bytes.concat([head, tail]).endsWith(tail))
assert(Bytes.concat([head, tail]).endsWith(Bytes.concat([head, tail])))
```

On the UTF-8 bytes of two strings, it gives the same answer as
[`String.endsWith`](../String/endsWith.md):

```ant
let s = "año"
for t in ["o", "ño", "ñ", ""] {
  assert(s.endsWith(t) == Bytes.endsWith(String.toBytes(s), String.toBytes(t)))
}
```

## Possible issues

The value comes first and the suffix second. Swapped, the call asks whether the suffix ends with the
value:

```ant
let data = String.toBytes("main.ant")
let ext = String.toBytes(".ant")
Bytes.endsWith(ext, data) // false
Bytes.endsWith(data, ext) // true
```

The precomposed "é" and "e" followed by a combining accent are different bytes, so text that looks
the same can fail to match. Normalize the strings with [`String.normalize`](../String/normalize.md)
before converting them:

```ant
let word = "cafe\u{301}"
let nfc = String.normalize(word, "NFC")
String.toBytes(word).endsWith(String.toBytes("é")) // false
String.toBytes(nfc).endsWith(String.toBytes("é"))  // true
```

A string is not `Bytes`. Convert it with [`String.toBytes`](../String/toBytes.md), or compare the
strings with [`String.endsWith`](../String/endsWith.md):

```ant
let data = String.toBytes("main.ant")
println(Bytes.endsWith(data, ".ant")) // error: cannot use string where Bytes is expected
```

An array of integers is not `Bytes` either. Convert it with [`Bytes.fromArray`](fromArray.md):

```ant
let data = Bytes.fromArray([0xff, 0xd8, 0xff, 0xd9])
println(Bytes.endsWith(data, [0xff, 0xd9])) // error: cannot use Array[int] where Bytes is expected
```

## See also

[Bytes.startsWith](startsWith.md) · [Bytes.indexOf](indexOf.md) · [Bytes.slice](slice.md) ·
[Bytes.concat](concat.md) · [String.endsWith](../String/endsWith.md) ·
[String.toBytes](../String/toBytes.md)
