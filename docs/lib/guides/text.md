<!-- status: reviewed -->

# Working with text

Text in Ant is a `string`: an immutable sequence of Unicode code points, stored as UTF-8. Four
namespaces work with it:

- [String](../String/index.md) holds the operations on text: measuring, searching, splitting,
  trimming, case, and conversion to and from numbers.
- [Regex](../Regex/index.md) matches patterns, written as a literal such as `/\d+/` or compiled from
  data.
- [Bytes](../Bytes/index.md) holds binary data, and the UTF-8 bytes of a string when you need them.
- [Encoding](../Encoding/index.md) writes bytes as hex or Base64 text and reads them back.

Lengths, indices, and slices count code points, not bytes. What a reader sees as one character can
be several code points, which `String.graphemes` keeps together:

```ant
let word = "naïve"
word.length()             // 5
word.toBytes().length()   // 6
"👍🏽".length()             // 2
"👍🏽".graphemes().length() // 1
```

There is no character type. A `for` loop gives each code point as a one-code-point string, and the
class tests such as `String.isDigit` take one:

```ant
for c in "a1b2" {
  if c.isDigit() {
    println(c)
  }
}
```

```text
1
2
```

## Writing strings

A string between double quotes takes escapes and interpolates any value with `${...}`. A raw string
between backticks keeps every character as written, which suits paths and templates:

```ant
let n = 3
"${n} ants\tin a row" // "3 ants\tin a row"
`C:\temp\${n}`        // "C:\\temp\\\${n}"
```

Both kinds may span lines, and each line break is one `\n`:

```ant
let banner = "Ant
  ===
"
print(banner)
```

```text
Ant
  ===
```

There is no format string. Interpolate the result of `String.fromInt` or `String.fromFloat`, and pad
it to a width:

```ant
let total = 12.5
"total: ${String.fromFloat(total, 2).padStart(8)}" // "total:    12.50"
"#${String.fromInt(255, 16).padStart(6, "0")}"     // "#0000ff"
```

## String or Regex

For fixed text, use the `String` functions. They match code point by code point, with no pattern
syntax to escape:

```ant
let line = "error: disk full"
line.startsWith("error:")    // true
line.replace("full", "busy") // "error: disk busy"
line.cut(": ")?.after        // "disk full"
```

For a shape rather than fixed text, such as "digits", "a word", or "any run of spaces", use a
pattern:

```ant
let line = "took 120 ms,  then 45 ms"
Regex.test(/\d+ ms/, line)                         // true
Regex.findAll(/\d+/, line) | Array.map({ x.text }) // ["120", "45"]
Regex.replace(/\s+/, line, " ")                    // "took 120 ms, then 45 ms"
```

A pattern is the first argument of every `Regex` function that matches (`test`, `find`, `findAll`,
`replace`, `replaceFirst`, `replaceWith`, `split`), so the string goes second. A pipe carries the
pattern, not the string, into a `Regex` function; pipe the results onward instead.

A replacement refers to a group by number or name. In a double-quoted replacement, `${...}`
interpolates; write `\${name}` or use a raw string:

```ant
Regex.replace(/(\d+)-(\d+)-(\d+)/, "2026-11-14", `$3.$2.$1`)                // "14.11.2026"
Regex.replace(/(?<y>\d+)-(?<m>\d+)-(?<d>\d+)/, "2026-11-14", "\${d}.\${m}") // "14.11"
```

## Strings and bytes

A `string` is text, and `Bytes` is whatever a file, a connection, or a digest holds. Text becomes
bytes through `String.toBytes`, and bytes become text through `String.fromBytes`, which is `null`
when they are not UTF-8, or `String.fromBytesLossy`, which writes U+FFFD for each bad sequence:

```ant
let data = "héllo".toBytes()
data                                                    // Bytes(68 c3 a9 6c 6c 6f)
String.fromBytes(data)                                  // "héllo"
Bytes.fromArray([0x61, 0xff]) | String.fromBytes()      // null
Bytes.fromArray([0x61, 0xff]) | String.fromBytesLossy() // "a�"
```

To put bytes where only text fits, such as JSON, a URL, or a header, encode them as hex or Base64.
The decoders return a `Result`, since the text may not be valid:

```ant
let data = "hi?".toBytes()
Encoding.toHex(data)                                          // "68693f"
Encoding.toBase64(data)                                       // "aGk/"
"aGk/" | Encoding.fromBase64() | Result.map(String.fromBytes) // Ok{ value: "hi?" }
Encoding.fromHex("6g")                                        // Err{ error: Error{ message: "invalid byte: U+0067 'g'", kind: null } }
```

## Putting them together

Parse a log line with named groups, then clean up the parts with `String` functions:

```ant
type LogLine = { level: string, message: string }

func parse(line: string): LogLine | null {
  let m = Regex.find(/^\[(?<level>\w+)\]\s*(?<message>.*)$/, line)
  if m == null {
    return null
  }
  LogLine{
    level: (m.named["level"] ?? "").toUpper(),
    message: (m.named["message"] ?? "").trim(),
  }
}

for line in "[warn] disk almost full  \nnoise\n[info]  started\n".lines() {
  let entry = parse(line)
  if entry != null {
    println("${entry.level.padEnd(5)} ${entry.message}")
  }
}
```

```text
WARN  disk almost full
INFO  started
```

Hash a message and show the digest as hex, the way checksums are published:

```ant
let digest = "hello" | String.toBytes() | Crypto.sha256() | Encoding.toHex()
println(digest)
```

```text
2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
```

Compare names the way a reader would, regardless of case and of how an accent was typed:

```ant
func sameName(a: string, b: string): bool {
  a.trim().normalize("NFC").equalFold(b.trim().normalize("NFC"))
}

println(sameName("José", " JOSE\u{301} "))
println(sameName("Straße", "STRASSE"))
```

```text
true
true
```

## Functions by job

### Building text

- [String.from](../String/from.md): any value, as `println` and interpolation render it
- [String.fromInt](../String/fromInt.md): an integer in a base
- [String.fromFloat](../String/fromFloat.md): a float with a fixed number of decimals
- [String.repeat](../String/repeat.md): `n` copies of a string
- [String.concat](../String/concat.md): pieces one after the other
- [String.join](../String/join.md): pieces with a separator between them
- [String.padStart](../String/padStart.md), [String.padEnd](../String/padEnd.md): padding up to a
  width

### Size and pieces

- [String.length](../String/length.md): the number of code points
- [String.isEmpty](../String/isEmpty.md): whether there are none
- [String.slice](../String/slice.md): the code points from one index up to another
- [String.chars](../String/chars.md): one string per code point
- [String.graphemes](../String/graphemes.md): one string per character as a reader sees it

### Searching

- [String.contains](../String/contains.md): whether fixed text occurs
- [String.startsWith](../String/startsWith.md), [String.endsWith](../String/endsWith.md): whether
  text begins or ends with fixed text
- [String.indexOf](../String/indexOf.md), [String.lastIndexOf](../String/lastIndexOf.md): the first
  or last index of fixed text, or `null`
- [String.count](../String/count.md): how many times fixed text occurs, without overlap
- [Regex.test](../Regex/test.md): whether a pattern matches anywhere
- [Regex.find](../Regex/find.md): the leftmost match, with its groups, or `null`
- [Regex.findAll](../Regex/findAll.md): every match, left to right

### Splitting

- [String.split](../String/split.md): the pieces between each separator
- [String.cut](../String/cut.md): the two sides of the first separator, as
  [String.Cut](../String/types/Cut.md)
- [String.fields](../String/fields.md): the runs between white space
- [String.lines](../String/lines.md): the lines of a text
- [Regex.split](../Regex/split.md): the pieces between the matches of a pattern

### Trimming

- [String.trim](../String/trim.md), [String.trimStart](../String/trimStart.md),
  [String.trimEnd](../String/trimEnd.md): white space removed at both ends, the start, or the end
- [String.trimPrefix](../String/trimPrefix.md), [String.trimSuffix](../String/trimSuffix.md): a
  prefix or suffix removed once

### Replacing

- [String.replace](../String/replace.md), [String.replaceFirst](../String/replaceFirst.md): every or
  the first occurrence of fixed text replaced
- [Regex.replace](../Regex/replace.md), [Regex.replaceFirst](../Regex/replaceFirst.md): every or the
  first match replaced, with `$1` and `${name}` expanded
- [Regex.replaceWith](../Regex/replaceWith.md): every match replaced by what a function returns

### Case and normalization

- [String.toUpper](../String/toUpper.md), [String.toLower](../String/toLower.md): upper or lower
  case
- [String.foldCase](../String/foldCase.md): a key in which case does not matter
- [String.equalFold](../String/equalFold.md): whether two strings are equal without regard to case
- [String.normalize](../String/normalize.md): one of the four Unicode normalization forms

### Character classes

- [String.isDigit](../String/isDigit.md), [String.isNumber](../String/isNumber.md): decimal digits,
  or any numbers
- [String.isLetter](../String/isLetter.md), [String.isUpper](../String/isUpper.md),
  [String.isLower](../String/isLower.md): letters, upper-case, or lower-case letters
- [String.isSpace](../String/isSpace.md): white space
- [String.isPunct](../String/isPunct.md), [String.isSymbol](../String/isSymbol.md),
  [String.isMark](../String/isMark.md): punctuation, symbols, or marks such as combining accents
- [String.isControl](../String/isControl.md), [String.isGraphic](../String/isGraphic.md),
  [String.isPrint](../String/isPrint.md): control, graphic, or printable code points
- [String.category](../String/category.md): the general category of one code point
- [String.inScript](../String/inScript.md): whether one code point belongs to a script

### Numbers from text

- [String.toInt](../String/toInt.md), [String.toUint](../String/toUint.md),
  [String.toFloat](../String/toFloat.md): a number read from text, or `null`

### Patterns

- [Regex.compile](../Regex/compile.md): a pattern from a string, or an `Err`
- [Regex.escape](../Regex/escape.md): text with every metacharacter escaped, to match it literally
- [Regex.source](../Regex/source.md), [Regex.flags](../Regex/flags.md): the source and flags of a
  pattern
- [Regex.Pattern](../Regex/types/Pattern.md): a compiled regular expression
- [Regex.Match](../Regex/types/Match.md): one match, with its text, indices, and groups

### Code points and bytes

- [String.codePoint](../String/codePoint.md), [String.fromCodePoint](../String/fromCodePoint.md):
  one code point to an `int` and back
- [String.toCodePoints](../String/toCodePoints.md),
  [String.fromCodePoints](../String/fromCodePoints.md): every code point to an `int` and back
- [String.toBytes](../String/toBytes.md): the UTF-8 bytes of a string
- [String.fromBytes](../String/fromBytes.md): bytes as text, or `null` when they are not UTF-8
- [String.fromBytesLossy](../String/fromBytesLossy.md): bytes as text, with U+FFFD for what is not
  UTF-8

### Hex and Base64

- [Encoding.toHex](../Encoding/toHex.md), [Encoding.fromHex](../Encoding/fromHex.md): bytes as two
  hex digits each, and back
- [Encoding.toBase64](../Encoding/toBase64.md), [Encoding.fromBase64](../Encoding/fromBase64.md):
  bytes as Base64, standard or URL-safe, and back

### Building and changing bytes

- [Bytes.new](../Bytes/new.md): a number of zero bytes
- [Bytes.fromArray](../Bytes/fromArray.md): bytes from an array of integers
- [Bytes.concat](../Bytes/concat.md): an array of `Bytes` joined in order
- [Bytes.clone](../Bytes/clone.md): a copy
- [Bytes.push](../Bytes/push.md), [Bytes.extend](../Bytes/extend.md): append one byte, or other
  bytes
- [Bytes.pop](../Bytes/pop.md), [Bytes.truncate](../Bytes/truncate.md): remove the last byte, or
  keep the first ones
- [Bytes.fill](../Bytes/fill.md), [Bytes.copy](../Bytes/copy.md): overwrite a range with one byte,
  or with other bytes

### Reading bytes

- [Bytes.length](../Bytes/length.md): the number of bytes
- [Bytes.slice](../Bytes/slice.md): a copy of a range
- [Bytes.indexOf](../Bytes/indexOf.md): the index of other bytes from an offset, or `null`
- [Bytes.startsWith](../Bytes/startsWith.md), [Bytes.endsWith](../Bytes/endsWith.md): whether a
  prefix or suffix is there
- [Bytes.same](../Bytes/same.md): whether two names hold one collection
- [Bytes.toArray](../Bytes/toArray.md): the bytes as an `Array[byte]`

### Numbers in bytes

- [Bytes.read](../Bytes/read.md), [Bytes.write](../Bytes/write.md): a fixed-size number at an index,
  in a stated byte order
- [Bytes.append](../Bytes/append.md): a number added at the end
- [Bytes.sizeOf](../Bytes/sizeOf.md): the bytes a number type takes
- [Bytes.Order](../Bytes/types/Order.md): [Bytes.BigEndian](../Bytes/types/BigEndian.md) or
  [Bytes.LittleEndian](../Bytes/types/LittleEndian.md)

## Mistakes to avoid

A string cannot be indexed. Slice it, or index its `chars`:

```ant
let s = "ant"
println(s[0]) // error: string cannot be indexed
```

```ant
let s = "ant"
s.slice(0, 1) // "a"
s.chars()[0]  // "a"
```

`join` is a `String` function that takes an array, so a dot call on the array does not find it. End
the chain with a pipe:

```ant
let parts = ["a", "b"]
println(parts.join(",")) // error: Array[string] has no field or function join
```

```ant
let parts = ["a", "b"]
parts | String.join(",") // "a,b"
```

`==` compares code points, so it neither ignores case nor joins the two spellings of `é`. Compare
through `String.equalFold` and `String.normalize`:

```ant
"Ant" == "ANT"                                           // false
"Ant".equalFold("ANT")                                   // true
"\u{E9}" == "e\u{301}"                                   // false
"\u{E9}".normalize("NFC") == "e\u{301}".normalize("NFC") // true
```

The `i` flag of a pattern folds case one code point at a time, so it misses what `String.equalFold`
accepts:

```ant
Regex.test(/straße/i, "STRASSE") // false
"straße".equalFold("STRASSE")    // true
```

`String.length` counts code points, which differs from the bytes a file or a protocol counts. Count
the bytes of `String.toBytes`:

```ant
"日本".length()           // 2
"日本".toBytes().length() // 6
```

`\d`, `\w`, and `\s` in a pattern match ASCII only. Use `\p{Nd}` for digits and `\pL` for letters in
any script:

```ant
Regex.test(/^\w+$/, "héllo")  // false
Regex.test(/^\pL+$/, "héllo") // true
Regex.test(/^\d+$/, "٣٤")     // false
Regex.test(/^\p{Nd}+$/, "٣٤") // true
```

## See also

- [Strings](../../guide/strings.md): how strings, characters, and bytes fit together
- [String literals](../../language/literals/strings.md),
  [Raw strings](../../language/literals/raw-strings.md),
  [Regular expression literals](../../language/literals/regex.md)
- [Numbers](numbers.md): parsing and printing numbers
- [Data formats](data-formats.md): JSON, CSV, and digests
