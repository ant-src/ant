<!-- status: reviewed -->

# string

`string` is the type of text: an immutable sequence of Unicode code points, stored as UTF-8. The
functions on it are in the [String](../../lib/String/index.md) namespace.

## Details

- A string never changes. Every `String` function that edits text, such as `trim` or `replace`,
  returns a new string.
- Lengths, indices, and slices count code points, not bytes. There is no `s[i]` and no character
  type: a character is a string of one code point, as `for` and `String.chars` give.
- One character on screen can be several code points, such as an `e` and a combining accent, or an
  emoji joined from several. `String.graphemes` splits text into those units.
- `==` and `<` compare code point by code point, with no case folding or normalization. `+` joins
  two strings and nothing else.
- Text that comes from outside the program, a file, standard input, or an argument, is decoded as
  UTF-8, with each invalid sequence replaced by U+FFFD. Read `Bytes` to keep the bytes as they were.

## Basic examples

A literal, its length, and a string built from other values:

```ant
let name = "Ant"
String.length(name)                    // 3
"Hello, ${name}"                       // "Hello, Ant"
"${name} has ${name.length()} letters" // "Ant has 3 letters"
```

Two strings join with `+`:

```ant
"ant" + "hill" // "anthill"
```

## Scope

### Code points and bytes

A code point outside ASCII takes more than one byte in UTF-8. `length` counts code points, and the
`Bytes` of the string count bytes:

```ant
let word = "café"
word.length()           // 4
word.toBytes().length() // 5
word.chars()            // ["c", "a", "f", "é"]
word.toCodePoints()     // [99, 97, 102, 233]
```

### Iteration

`for` over a string binds each code point as a one-character string:

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

### Slices and searches

Indices count code points, from 0, and a slice ends before its second index:

```ant
let text = "naïve café"
text.slice(0, 5)     // "naïve"
text.indexOf("café") // 6
text.indexOf("tea")  // null
text.cut(" ")        // String.Cut{ before: "naïve", after: "café" }
```

### Comparison and sorting

`<` orders strings by code point, so every upper-case ASCII letter comes before every lower-case
one. `string` is `Ordered`, so `Array.sort` takes an array of strings:

```ant
"apple" < "banana" // true
"Zebra" < "apple"  // true
```

To ignore case, sort by `foldCase`:

```ant
let words = ["pear", "apple", "Fig"]
words.sort()                   // ["Fig", "apple", "pear"]
words.sortBy({ x.foldCase() }) // ["apple", "Fig", "pear"]
```

### Conversions

`String.from` renders any value as a string, as `${...}` does. `toInt` and `toFloat` read a number,
and give `null` for text that is not one:

```ant
String.from(42)  // "42"
String.from(1.5) // "1.5"
"42".toInt()     // 42
"4x2".toInt()    // null
"2.5".toFloat()  // 2.5
```

### Inside a container

A string renders as itself at the top level and quoted inside an array, a map, or a record:

```ant
let tags = ["red", "a \"b\""]
println(tags)
println(tags[0])
println(Map{ "lang": "Ant" })
```

```text
["red", "a \"b\""]
red
Map{ "lang": "Ant" }
```

### Map keys and set members

Strings compare by value, so they work as map keys and set members:

```ant
let ages = Map{ "Ann": 31, "Bob": 27 }
ages["Ann"]                    // 31
Set.of(["x", "y", "x"]).size() // 2
```

### Nullable strings

`string | null` is a separate type. `??` gives a default, and `?.` skips the call on `null`:

```ant
let nick: string | null = null
nick ?? "anonymous" // "anonymous"
nick?.toUpper()     // null
```

### Named types over `string`

A `type` declaration over `string` makes a distinct type. Its constructor tags a value, and
`string(x)` gives the plain string back:

```ant
type Email = string

let email = Email("ann@example.com")
email                          // Email("ann@example.com")
string(email).endsWith(".com") // true
```

## Applications

### Reading lines of input

Trim text read from outside, then split it into lines. `lines` drops a `\r` before each `\n` and
drops the empty piece after a final `\n`:

```ant
let text = "  a\nb \n"
text.trim().lines() // ["a", "b"]
```

### Counting words

`fields` splits at runs of white space, so extra spaces make no empty words:

```ant
let text = "  the quick   brown fox  "
text.fields().length() // 4
```

### Capitalizing a word

```ant
func capitalize(word: string): string {
  if word.isEmpty() { return word }
  word.slice(0, 1).toUpper() + word.slice(1, word.length())
}

capitalize("ant")  // "Ant"
capitalize("élan") // "Élan"
capitalize("")     // ""
```

### Building text from parts

Collect the pieces in an array and join them once:

```ant
let rows = [["Ann", "31"], ["Bob", "27"]]
let table = rows.map({ x | String.join("\t") }) | String.join("\n")
println(table)
```

```text
Ann	31
Bob	27
```

### Reversing what a reader sees

Reverse the graphemes, not the code points, so an accent stays on its letter. Reversing the code
points moves the accent to the front:

```ant
let word = "cafe\u{301}"
word.graphemes().reverse() | String.concat() // "éfac"
word.chars().reverse() | String.concat()     // "\u{301}efac"
```

## Properties and relations

The length is the number of code points, and splitting then joining gives the text back:

```ant
for s in ["", "ant", "naïve café", "a,b,,c"] {
  assert(s.length() == s.chars().length())
  assert(s.length() == len(s))
  assert((s.split(",") | String.join(",")) == s)
  assert(String.fromBytes(s.toBytes()) == s)
  assert((s.chars() | String.concat()) == s)
}
println("ok")
```

```text
ok
```

## Possible issues

### There is no indexing

```ant
let word = "ant"
let first = word[0] // error: string cannot be indexed
```

Slice the string, or take its code points as an array:

```ant
let word = "ant"
word.slice(0, 1) // "a"
word.chars()[0]  // "a"
```

### `+` does not convert

```ant
let n = 3
let label = "items: " + n // error: + needs both operands of one type, got string and int
```

Interpolate instead:

```ant
let n = 3
"items: ${n}" // "items: 3"
```

### Two spellings of one character

`é` can be one code point or an `e` and a combining accent. They look the same, but `==` tells them
apart. Normalize text that can come in either spelling:

```ant
let composed = "\u{e9}"
let decomposed = "e\u{301}"
composed == decomposed                          // false
decomposed.length()                             // 2
String.normalize(decomposed, "NFC") == composed // true
```

### Length is not what a reader counts

An emoji joined from several people is one grapheme, but several code points:

```ant
let family = "👨‍👩‍👧"
family.length()             // 5
family.graphemes().length() // 1
```

### `==` is case-sensitive

`foldCase` maps every case variant to one form, so `"Straße"` and `"STRASSE"` match:

```ant
"Apple" == "apple"                          // false
"Apple".foldCase() == "apple".foldCase()    // true
"Straße".foldCase() == "STRASSE".foldCase() // true
```

### Case mapping can change the length

```ant
"straße".toUpper()          // "STRASSE"
"straße".toUpper().length() // 7
```

### A named type over `string` is not a `string`

`String` functions take a plain `string`:

```ant
type Email = string
let email = Email("ann@example.com")
let domain = email.cut("@") // error: Email has no field or function cut
```

Unwrap it with `string(email)` first.

## See also

- [String literals and interpolation](../literals/strings.md)
- [Raw strings](../literals/raw-strings.md)
- [String](../../lib/String/index.md)
- [String.length](../../lib/String/length.md)
- [String.graphemes](../../lib/String/graphemes.md)
- [String.normalize](../../lib/String/normalize.md)
- [String.from](../../lib/String/from.md)
- [string(x)](../../lib/prelude/string.md)
- [byte](byte.md)
- [Bytes](../../lib/Bytes/index.md)
- [Comparison operators](../operators/comparison.md)
- [Strings and bytes](../../guide/strings.md)
