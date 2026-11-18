<!-- status: reviewed -->

# Strings and bytes

Ant has one type for text and one for binary data. A `string` is an immutable sequence of Unicode
code points, stored as UTF-8. `Bytes` is a mutable sequence of bytes, for what files, connections,
and digests hold. This chapter shows how a string counts its characters, how to write one, what
happens to line breaks, and when to reach for `Bytes` instead.

## Code points

A string's length, its indices, and its slices count code points, not bytes. A code point is one
Unicode character number, such as U+00EF for `ï`. UTF-8 stores it in one to four bytes, so the byte
count is often larger:

```ant
let s = "naïve café"
s.length()           // 10
s.toBytes().length() // 12
s.slice(0, 5)        // "naïve"
s.indexOf("café")    // 6
```

There is no character type. A character is a string of one code point, and a `for` loop over a
string gives one such string per code point:

```ant
var vowels = 0
for c in "naïve" {
  if "aeiouï".contains(c) { vowels += 1 }
}
println(vowels)
```

```text
3
```

A string has no `s[i]`. Finding a code point by index means walking the UTF-8 from the start, so a
loop over indices would be quadratic. Slice the string, or take its `chars` once and index the
array:

```ant
let word = "ant"
word.slice(1, 2)    // "n"
word.chars()        // ["a", "n", "t"]
word.chars()[2]     // "t"
word.toCodePoints() // [97, 110, 116]
"é".codePoint()     // 233
```

`String.fromCodePoint` goes the other way. It returns `null` for a number that is not a Unicode
scalar value, such as a surrogate:

```ant
String.fromCodePoint(0x1F600) // "😀"
String.fromCodePoint(0xD800)  // null
```

## Graphemes

What a reader sees as one character can be several code points: an `e` followed by a combining
accent, a thumb with a skin tone, a flag made of two regional letters, or a family emoji joined from
three people. Unicode calls such a unit an extended grapheme cluster, and `String.graphemes` splits
text into them:

```ant
let word = "e\u{301}te"
word.length()                // 4
word.chars()                 // ["e", "\u{301}", "t", "e"]
word.graphemes().length()    // 3
word.graphemes()[0].length() // 2
"🇨🇭".length()                // 2
"🇨🇭".graphemes()             // ["🇨🇭"]
"👍🏽".graphemes()             // ["👍🏽"]
```

Code point indices can cut through a grapheme. Text that you cut, reverse, or count for display goes
through `graphemes`:

```ant
func reverse(s: string): string {
  s.graphemes().reverse() | String.concat()
}

"🇨🇭e\u{301}".chars().reverse() | String.concat() // "\u{301}e🇭🇨"
reverse("🇨🇭e\u{301}") == "e\u{301}🇨🇭"            // true
```

Reversing the code points moves the accent off its letter and turns the Swiss flag into two loose
letter symbols. Reversing the graphemes keeps both whole.

The same `é` has two spellings: the single code point U+00E9, and `e` followed by U+0301. `==`
compares code points, so the two are different strings. Bring text that can arrive in either
spelling to one form with `String.normalize` before you compare it or use it as a key:

```ant
let composed = "caf\u{E9}"
let decomposed = "cafe\u{301}"
composed == decomposed                                   // false
composed.length()                                        // 4
decomposed.length()                                      // 5
composed.normalize("NFC") == decomposed.normalize("NFC") // true
```

To compare text without regard to case, use `String.equalFold`. It applies the full case folding, so
`ß` matches `SS`. For a map key in which case does not matter, store `String.foldCase` of the text:

```ant
"Straße".equalFold("STRASSE") // true
"Straße".foldCase()           // "strasse"
```

## Writing strings

A string is written between double quotes. It takes the escapes `\n`, `\t`, `\r`, `\\`, `\"`, `\$`,
and `\u{...}` with a hexadecimal code point:

```ant
"tab\there"     // "tab\there"
"say \"hi\""    // "say \"hi\""
"\u{1F41C} ant" // "🐜 ant"
```

### Interpolation

`${...}` inside a string interpolates any expression. Ant renders the value with `String.from`, the
same format `println` writes, so numbers, arrays, records, and `null` need no conversion:

```ant
type Point = { x: int, y: int }
let p = Point{ x: 1, y: 2 }
let n = 3
"${n} item${if n == 1 { "" } else { "s" }}" // "3 items"
"at ${p}"                                   // "at Point{ x: 1, y: 2 }"
"doubled: ${[1, 2].map({ x * 2 })}"         // "doubled: [2, 4]"
"missing: ${null}"                          // "missing: null"
```

Braces nest inside `${...}`, so a lambda, a record literal, or another string fits there. A `$` that
is not followed by `{` is an ordinary character, and `\$` writes a literal `$` before a brace:

```ant
"costs $5"            // "costs $5"
"write \${name} here" // "write \${name} here"
```

The value of the last case is the string `write ${name} here`. A string inside an array, or as a
case value here, renders quoted with escapes, so it reads back as a literal; `println` of the string
itself writes it as is.

There is no format string. For a fixed number of decimals or a zero-padded width, interpolate the
result of `String.fromFloat` or `String.fromInt`:

```ant
let price = 4.5
"total: ${String.fromFloat(price, 2)}"         // "total: 4.50"
"#${String.fromInt(255, 16).padStart(6, "0")}" // "#0000ff"
```

### Strings across lines

A string may span lines. Each line break inside the quotes is one `\n` in the value, and the
indentation of the following lines is part of the string:

```ant
let usage = "usage: greet NAME
  -h  show help
"
print(usage)
```

```text
usage: greet NAME
  -h  show help
```

Ant removes no indentation, so start the continuation lines at the left margin when the text should
not be indented.

### Joining pieces

An array has no `join`. To build a string from many pieces, end the chain with `| String.join(sep)`,
or `| String.concat()` when nothing goes between them:

```ant
["a", "b", "c"] | String.join(", ")         // "a, b, c"
[1, 2, 3].map({ "${x}" }) | String.concat() // "123"
```

## Raw strings

A raw string is written between backticks, as in Go. It has no escapes and no interpolation: every
character between the backticks stands for itself. Use it for text full of backslashes, quotes, or
`${`, such as a Windows path, a regular expression, or a shell script:

```ant
`C:\temp\new`  // "C:\\temp\\new"
`say "hi"`     // "say \"hi\""
`echo ${HOME}` // "echo \${HOME}"
`\d+\.\d+`     // "\\d+\\.\\d+"
```

A raw string may span lines like a string. It cannot hold a backtick; join raw strings and a string
with `+` for that:

```ant
`run ` + "`ant fmt`" + ` first` // "run `ant fmt` first"
```

Pick the form by its content:

| The text has                                  | Write                    |
| --------------------------------------------- | ------------------------ |
| values to insert                              | `"Hello, ${name}"`       |
| special characters, such as a tab or an emoji | `"\t"`, `"\u{1F41C}"`    |
| backslashes, quotes, or a `${` to keep as is  | `` `C:\temp` ``          |
| a regular expression known when you write it  | a literal, `/\d+/`       |
| a regular expression built from strings       | ``Regex.compile(`\d+`)`` |

## Line endings

Ant accepts a line feed, a carriage return and line feed, or a lone carriage return as a line break
in source. Inside a string or a raw string, each of them is one `\n`, so a multi-line literal has
the same value whether the file was saved on Windows or on Linux.

Text read at run time is another matter, since a file written on Windows can hold `\r\n`.
`String.lines` handles it. It splits at each `\n`, drops a `\r` right before one, and gives no empty
last line for a final line break:

```ant
"one\r\ntwo\r\n".lines() // ["one", "two"]
"one\ntwo".lines()       // ["one", "two"]
"one\n\n".lines()        // ["one", ""]
"".lines()               // []
```

`String.split(text, "\n")` does none of this, so its pieces keep the `\r` and end with an empty
string:

```ant
"one\r\ntwo\r\n".split("\n") // ["one\r", "two\r", ""]
```

## Text from outside the program

Every string holds valid Unicode. Text that reaches a program as a string, from a file, standard
input, the network, an argument, or a child process, is decoded as UTF-8, and each invalid byte, or
each sequence cut short, becomes one replacement character U+FFFD. Read the data as `Bytes` when you
must keep it exactly:

```ant
let path = "data.bin"
try File.writeBytes(path, Bytes.fromArray([0x61, 0xff, 0x62]))
let text = try File.read(path)
let raw = try File.readBytes(path)
println(text)
println(raw)
```

```text
a�b
Bytes(61 ff 62)
```

`String.fromBytes` decodes bytes you hold and returns `null` when they are not UTF-8.
`String.fromBytesLossy` replaces the broken sequences, as the readers do:

```ant
let good = "héllo".toBytes()
let bad = Bytes.fromArray([0x61, 0xff])
good                       // Bytes(68 c3 a9 6c 6c 6f)
String.fromBytes(good)     // "héllo"
String.fromBytes(bad)      // null
String.fromBytesLossy(bad) // "a�"
```

A byte order mark at the start of a file is kept, as the code point U+FEFF. Remove it with
`String.trimPrefix(text, "\u{FEFF}")` when a format does not expect it.

## Bytes

`Bytes` holds binary data: an image, a digest, a network frame, a file that is not text. It has no
literal. Bytes come from I/O, from `String.toBytes`, from `Encoding.fromHex` and
`Encoding.fromBase64`, from `Bytes.new`, and from `Bytes.fromArray`, which takes plain integer
literals:

```ant
let png = Bytes.fromArray([0x89, 0x50, 0x4e, 0x47])
png                 // Bytes(89 50 4e 47)
png.length()        // 4
png[0]              // 137
Encoding.toHex(png) // "89504e47"
Bytes.new(3)        // Bytes(00 00 00)
```

A string escape writes a code point, never a byte. `"\u{89}"` is U+0089, which UTF-8 stores in two
bytes:

```ant
"\u{89}".toBytes() // Bytes(c2 89)
```

Unlike a string, `Bytes` is mutable, and like an array it is shared: every name for it sees a write
or growth. Call `clone` for a copy you can change on its own:

```ant
let frame = Bytes.new()
let alias = frame
let copy = frame.clone()
frame.push(byte(1))
frame.append(type(uint16), uint16(513), Bytes.BigEndian{})
alias                                          // Bytes(01 02 01)
copy                                           // Bytes()
frame.read(type(uint16), 1, Bytes.BigEndian{}) // 513
```

`append`, `read`, and `write` take the number type and the byte order on every call, so a binary
format reads the same on every machine. The [Bytes](../lib/Bytes/index.md) page lists the rest.

`String.toBytes` copies, and so do `String.fromBytes` and `String.fromBytesLossy`. A string never
changes when the `Bytes` it came from does.

## Habits from other languages

Ant has no character literal. Single quotes are not a string:

```ant
let c = 'a' // error: unexpected character '\''
```

Write `"a"`. A one-code-point string is a character.

Unlike in JavaScript or Python, a string cannot be indexed. Index its `chars`, or slice it:

```ant
let s = "ant"
println(s[0]) // error: string cannot be indexed
```

Backticks are not JavaScript template literals. They make a raw string, so `${...}` stays as
written:

```ant
let name = "Ann"
`Hello, ${name}` // "Hello, \${name}"
"Hello, ${name}" // "Hello, Ann"
```

`+` on a string takes only another string. Interpolate a number instead of adding it:

```ant
let n = 3
println("n=" + n) // error: + needs both operands of one type, got string and int
```

```ant
let n = 3
"n=${n}" // "n=3"
```

Ant has no `\x` escape. Write the code point with `\u{...}`, and build bytes with `Bytes.fromArray`:

```ant
let s = "\x41" // error: bad escape '\x'
```

The length of a string is neither Go's byte count nor JavaScript's count of UTF-16 units. It counts
code points, so an emoji outside the Basic Multilingual Plane is one, and a flag is two:

```ant
"😀".length()           // 1
"😀".toBytes().length() // 4
"🇨🇭".length()           // 2
```

A byte from `Bytes` is a `byte`, and arithmetic needs both operands of one type. Convert with
`int(...)` before mixing it with an `int`:

```ant
let b = "ab".toBytes()
int(b[0]) + 1 // 98
```

## Picking the tool

| You want to                               | Use                                         |
| ----------------------------------------- | ------------------------------------------- |
| Count, slice, or loop over characters     | `length`, `slice`, `for c in s`             |
| Index characters by position              | `s.chars()[i]`                              |
| Count or cut what a reader sees           | `String.graphemes`                          |
| Compare text that may be spelled two ways | `String.normalize(s, "NFC")` on both sides  |
| Compare text ignoring case                | `String.equalFold`, `String.foldCase`       |
| Insert values into text                   | `"${value}"`                                |
| Join many pieces                          | `String.join`, `String.concat`              |
| Keep backslashes, quotes, or `${` as is   | a raw string, `` `...` ``                   |
| Split text into lines                     | `String.lines`                              |
| Keep data exactly as it is                | `Bytes`, `File.readBytes`                   |
| Turn bytes into text                      | `String.fromBytes`, `String.fromBytesLossy` |
| Put bytes into text, such as JSON         | `Encoding.toBase64`, `Encoding.toHex`       |

## See also

- Reference: [String literals](../language/literals/strings.md),
  [Raw strings](../language/literals/raw-strings.md),
  [Line endings](../language/lexical/line-endings.md), [string](../language/types/string.md)
- [String](../lib/String/index.md), [Bytes](../lib/Bytes/index.md),
  [Encoding](../lib/Encoding/index.md)
- [Working with text](../lib/guides/text.md) for the `String` and `Regex` functions by job
- [Idiomatic Ant](idiomatic.md)
