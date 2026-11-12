<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.split

```ant signature
func Regex.split(p: Regex.Pattern, s: string): Array[string]
```

Regex.split returns the pieces of s between the matches of p. A pattern that never matches gives
\[s\], an empty s gives \[""\], and an empty match at the start or end of s adds no empty piece.

<!-- end generated -->

## Details

- Capture groups do not add pieces: the pieces are only the text between whole matches.
- A pattern that can match the empty string splits between code points, not between graphemes.
- `\s`, `\w`, and `\d` are ASCII. Use `\pL` and the other Unicode classes, or `String.fields`, for
  Unicode text.

## Basic examples

Split at commas with any spaces around them:

```ant
Regex.split(/\s*,\s*/, "red , green,blue") // ["red", "green", "blue"]
```

Split at any run of digits:

```ant
Regex.split(/\d+/, "a1b22c333d") // ["a", "b", "c", "d"]
```

## Scope

A pattern that never matches gives the whole string, and an empty string gives one empty piece,
whatever the pattern:

```ant
Regex.split(/,/, "abc") // ["abc"]
Regex.split(/,/, "")    // [""]
Regex.split(/.*/, "")   // [""]
```

A match that is not empty, at either end or next to another one, leaves an empty piece:

```ant
Regex.split(/,/, ",a,,b,") // ["", "a", "", "b", ""]
Regex.split(/-/, "-")      // ["", ""]
```

An empty match at the start or the end adds no piece; one in the middle splits there:

```ant
Regex.split(/a*/, "baaab")    // ["b", "b"]
Regex.split(/\b/, "hi there") // ["hi", " ", "there"]
```

The empty pattern splits between code points:

```ant
Regex.split(/(?:)/, "a😀b") // ["a", "😀", "b"]
```

Flags apply as they do in any match. With `i` the separator matches in any case, and with `m` the
`^` anchor also matches at the start of each line:

```ant
Regex.split(/\s+and\s+/i, "salt AND pepper and oil") // ["salt", "pepper", "oil"]
Regex.split(/^/m, "a\nb\nc")                         // ["a\n", "b\n", "c"]
```

A pattern from data, built with `Regex.compile`, splits the same way:

```ant
let sep = Regex.compile("[;|]") | Result.unwrap()
Regex.split(sep, "a;b|c") // ["a", "b", "c"]
```

Call it through `.` on the pattern, or with named arguments:

```ant
/\s*;\s*/.split("a ; b;c")           // ["a", "b", "c"]
Regex.split(p: /-/, s: "2026-11-12") // ["2026", "11", "12"]
```

## Applications

Split text into lines, whether they end in `\n` or `\r\n`:

```ant
Regex.split(/\r?\n/, "one\r\ntwo\nthree") // ["one", "two", "three"]
```

Parse settings written as `key = value` pairs with loose spacing:

```ant
let text = "host = example.com; port=8080 ;debug = true"
Regex.split(/\s*;\s*/, text).map({ Regex.split(/\s*=\s*/, x) }) // [["host", "example.com"], ["port", "8080"], ["debug", "true"]]
```

Count the words of a sentence, splitting at every run of characters that are not letters:

```ant
let words = Regex.split(/[^\pL]+/, "Ant, bee & ant: über-cool!")
  .filter({ !x.isEmpty() })
println(words)
println(words.length())
```

```text
["Ant", "bee", "ant", "über", "cool"]
5
```

Split a version string at dots and dashes, and read the numeric parts:

```ant
Regex.split(/[.-]/, "1.24.3-rc2").map({ x.toInt() }) // [1, 24, 3, null]
```

## Properties and relations

When the separator is a fixed string, `Regex.escape` turns it into a pattern that splits as
`String.split` does:

```ant
let s = "a.b..c."
let dot = Regex.escape(".") | Regex.compile() | Result.unwrap()
assert(Regex.split(dot, s) == s.split("."))
```

When the pattern never matches the empty string, the pieces are the text that `Regex.findAll` skips:
putting the matches back between them rebuilds `s`:

```ant
let p = /\s*,\s*/
let s = "a , b,,c"
let pieces = Regex.split(p, s)
let matches = Regex.findAll(p, s)
assert(pieces.length() == matches.length() + 1)
var rebuilt = pieces[0]
for i in Iter.range(0, matches.length()) {
  rebuilt += matches[i].text + pieces[i + 1]
}
assert(rebuilt == s)
```

## Possible issues

On a string, `.split` is `String.split`, which takes a string separator, not a pattern:

```ant
let parts = "a1b22c".split(/\d+/) // error: cannot use Regex.Pattern where string is expected
```

Call `Regex.split` with the pattern first, or call `.split` on the pattern:

```ant
let digits = /\d+/
Regex.split(digits, "a1b22c") // ["a", "b", "c"]
digits.split("a1b22c")        // ["a", "b", "c"]
```

Capture groups do not add the matched separators to the result. Use `Regex.findAll` to keep them:

```ant
Regex.split(/([+-])/, "1+2-3")                 // ["1", "2", "3"]
Regex.findAll(/[+-]/, "1+2-3").map({ x.text }) // ["+", "-"]
```

Separators at the ends of the string leave empty pieces. Trim the string first, or filter them out:

```ant
Regex.split(/\s+/, "  a  b  ")        // ["", "a", "b", ""]
Regex.split(/\s+/, "  a  b  ".trim()) // ["a", "b"]
```

`\s` matches only ASCII whitespace, so a no-break space does not split. `String.fields` splits at
Unicode whitespace:

```ant
Regex.split(/\s+/, "a\u{a0}b c") // ["a\u{a0}b", "c"]
"a\u{a0}b c".fields()            // ["a", "b", "c"]
```

Likewise `\W` is every character but an ASCII letter, digit, or `_`, so it splits inside words with
accents:

```ant
Regex.split(/\W+/, "café au lait")     // ["caf", "au", "lait"]
Regex.split(/[^\pL]+/, "café au lait") // ["café", "au", "lait"]
```

`Regex.split` has no limit and splits at every match, so a value that contains the separator breaks
into more pieces:

```ant
Regex.split(/\s*=\s*/, "url = a=b") // ["url", "a", "b"]
```

To split at the first match only, find it with `Regex.find` and slice around it:

```ant
func splitFirst(s: string): Array[string] {
  let m = /\s*=\s*/.find(s)
  if m == null {
    return [s]
  }
  return [s.slice(0, m.start), s.slice(m.end, s.length())]
}

println(splitFirst("url = a=b"))
println(splitFirst("debug"))
```

```text
["url", "a=b"]
["debug"]
```

A pattern that matches the empty string splits a combining mark off its base character:

```ant
Regex.split(/(?:)/, "e\u{301}x") // ["e", "\u{301}", "x"]
```

## See also

[Regex.findAll](findAll.md), [Regex.replace](replace.md), [Regex.escape](escape.md),
[Regex.compile](compile.md), [String.split](../String/split.md),
[String.fields](../String/fields.md), [String.lines](../String/lines.md)
