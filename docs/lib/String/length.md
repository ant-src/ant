<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.length

```ant signature
func String.length(s: string): int
```

String.length returns the number of code points in s, not the graphemes a reader sees.

<!-- end generated -->

## Details

- `String.length` counts the code points of `s`, and `""` has length `0`. Lengths, indices, and
  slices of a string all count code points. A string is stored as UTF-8, so its length in bytes,
  `Bytes.length(String.toBytes(s))`, can be larger than its length in code points
  ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- What a reader sees as one character can be several code points, such as an `e` followed by a
  combining accent, or a family emoji joined from three people. `String.graphemes` splits text into
  those units.
- The built-in `len` counts a string's code points as `String.length` does
  ([spec: Prelude](https://github.com/ant-src/spec/blob/master/ant-lang.md#prelude)).
- `String.length` takes a `string`, so a value of a named primitive type over `string` needs
  `string(id)` first, while `len(id)` unwraps it
  ([spec: Named primitive types](https://github.com/ant-src/spec/blob/master/ant-lang.md#named-primitive-types)).

## Basic examples

Count the characters of a word:

```ant
String.length("ant")   // 3
String.length("hello") // 5
```

The empty string has length `0`:

```ant
String.length("") // 0
```

## Scope

A letter outside ASCII, written as one code point, counts once:

```ant
String.length("héllo") // 5
String.length("日本語")   // 3
String.length("🐜")     // 1
```

An escape such as `\t` or `\n` is one code point, and so is a line break inside a string literal
that spans lines:

```ant
String.length("a\tb\n") // 4
String.length("ab
cd") // 5
```

In a raw string, a backslash is a character of its own:

```ant
String.length(`a\nb`) // 4
```

An accent written as a combining mark after its letter is a code point of its own:

```ant
String.length("\u{E9}")   // 1
String.length("e\u{301}") // 2
```

A flag is two regional indicator symbols, and a family emoji is three people joined by two
zero-width joiners:

```ant
String.length("🇨🇭")            // 2
String.length("👨‍👩‍👧")         // 5
len(String.graphemes("👨‍👩‍👧")) // 1
```

Call it through `.`, in a pipe, or pass it as a function value:

```ant
"ant".length()                        // 3
"ant" | String.length()               // 3
["a", "bb", "ccc"].map(String.length) // [1, 2, 3]
```

## Applications

Check that a user name has between 3 and 16 characters:

```ant
func validName(name: string): bool {
  var n = String.length(String.trim(name))
  n >= 3 && n <= 16
}

validName("Ann")   // true
validName("  Al ") // false
validName("")      // false
```

Shorten a title to a limit, ending it with an ellipsis when it was cut:

```ant
func shorten(s: string, limit: int): string {
  if String.length(s) <= limit {
    return s
  }
  String.slice(s, 0, limit - 1) + "…"
}

shorten("Ant", 10)                   // "Ant"
shorten("An anthill of tunnels", 10) // "An anthil…"
```

Find the longest word of a sentence:

```ant
String.fields("ants dig long tunnels").maxBy(String.length) // "tunnels"
```

Underline a heading with as many characters as it has:

```ant
var title = "Größe"
println(title)
println(String.repeat("=", String.length(title)))
```

```text
Größe
=====
```

Sort words from the shortest to the longest:

```ant
["tunnel", "ant", "hill"] | Array.sortBy(String.length) // ["ant", "hill", "tunnel"]
```

## Properties and relations

`len` and `String.length` give the same count:

```ant
for s in ["", "ant", "héllo", "👨‍👩‍👧"] {
  assert(len(s) == String.length(s))
}
```

The length is the number of elements of `String.chars` and of `String.toCodePoints`:

```ant
var s = "e\u{301}🇨🇭 ant"
assert(len(String.chars(s)) == String.length(s))
assert(len(String.toCodePoints(s)) == String.length(s))
```

The length of a concatenation is the sum of the lengths:

```ant
var a = "ant"
var b = "hill🐜"
assert(String.length(a + b) == String.length(a) + String.length(b))
```

`String.isEmpty` is true exactly when the length is `0`:

```ant
for s in ["", " ", "ant"] {
  assert(String.isEmpty(s) == (String.length(s) == 0))
}
```

`String.slice` from `0` to the length gives back the whole string:

```ant
var s = "anthill 🐜"
assert(String.slice(s, 0, String.length(s)) == s)
```

A string has at least as many bytes as code points, and as many only when it is all ASCII:

```ant
for s in ["ant", "héllo", "🐜"] {
  assert(Bytes.length(String.toBytes(s)) >= String.length(s))
}
Bytes.length(String.toBytes("ant"))   // 3
Bytes.length(String.toBytes("héllo")) // 6
Bytes.length(String.toBytes("🐜"))     // 4
```

## Possible issues

`String.length` counts code points, not graphemes, so a skin tone or an accent that is a mark of its
own adds to the count. Count graphemes with `len` of `String.graphemes`:

```ant
String.length("👍🏽")                  // 2
len(String.graphemes("👍🏽"))          // 1
String.length("cafe\u{301}")         // 5
len(String.graphemes("cafe\u{301}")) // 4
```

The two spellings of `é` have different lengths. `String.normalize` brings them to one:

```ant
String.length("e\u{301}")                          // 2
String.length(String.normalize("e\u{301}", "NFC")) // 1
```

The length is not the size in bytes. A limit on stored or sent bytes needs the byte count:

```ant
var s = "日本語"
String.length(s)                // 3
Bytes.length(String.toBytes(s)) // 9
```

Changing case can change the length, since `String.toUpper` uses the full mappings:

```ant
String.length("ß")                 // 1
String.length(String.toUpper("ß")) // 2
```

A value of a named primitive type is not a `string`, so passing it is a compile error:

<!-- test: error cannot use Name where string is expected -->

```ant
type Name = string

var n = Name("Ann")
println(String.length(n))
```

Convert it with `string(n)`, or use `len`, which unwraps it:

```ant
type Name = string

var n = Name("Ann")
println(String.length(string(n)))
println(len(n))
```

```text
3
3
```

## See also

[String.graphemes](graphemes.md) · [String.chars](chars.md) · [String.toCodePoints](toCodePoints.md)
· [String.isEmpty](isEmpty.md) · [String.slice](slice.md) · [String.toBytes](toBytes.md) ·
[String.normalize](normalize.md) · [Bytes.length](../Bytes/length.md) ·
[Array.length](../Array/length.md)
