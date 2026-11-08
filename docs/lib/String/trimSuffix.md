<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trimSuffix

```ant signature
func String.trimSuffix(s: string, suffix: string): string
```

String.trimSuffix returns s with suffix removed once from its end, or s unchanged when it does not
end with suffix.

<!-- end generated -->

## Details

- `String.trimSuffix` removes `suffix` from the end of `s` once, when `s` ends with it, and returns
  `s` unchanged otherwise. It never panics
  ([spec: String](https://github.com/ant-src/spec/blob/master/ant-lang.md#string)).
- The match is exact, code point by code point, as in [`String.endsWith`](endsWith.md): it does not
  ignore case and does not normalize
  ([spec: Dynamic semantics](https://github.com/ant-src/spec/blob/master/ant-lang.md#12-dynamic-semantics)).
- Both parameters are `string`. A `string | null` must be narrowed first
  ([spec: Narrowing](https://github.com/ant-src/spec/blob/master/ant-lang.md#narrowing)).

## Basic examples

```ant
String.trimSuffix("main.ant", ".ant") // "main"
String.trimSuffix("main.ant", ".go")  // "main.ant"
```

Call it through `.` on any string:

```ant
var name = "report.tar.gz"
name.trimSuffix(".gz")     // "report.tar"
name.trimSuffix(".tar.gz") // "report"
name.trimSuffix(".tar")    // "report.tar.gz"
```

## Scope

The suffix comes off once, even when the string ends with several copies of it:

```ant
String.trimSuffix("ha!!!", "!") // "ha!!"
String.trimSuffix("abab", "ab") // "ab"
```

An empty suffix changes nothing, an empty string stays empty, and a suffix equal to the whole string
leaves the empty string:

```ant
String.trimSuffix("ant", "")    // "ant"
String.trimSuffix("", "x")      // ""
String.trimSuffix("ant", "ant") // ""
```

A suffix longer than the string never matches:

```ant
String.trimSuffix("ant", "giant") // "ant"
```

Text in any script trims by code points:

```ant
String.trimSuffix("東京都", "都")    // "東京"
String.trimSuffix("Ωμέγα", "γα") // "Ωμέ"
```

A combining accent is a code point of its own, so it can come off alone and leave the bare letter:

```ant
String.trimSuffix("cafe\u{301}", "\u{301}") // "cafe"
```

In a pipeline, the string is the first argument:

```ant
"index.html" | String.trimSuffix(".html") // "index"
```

## Named arguments

`String.trimSuffix` has no defaults, but naming the arguments makes clear which string is the
suffix:

```ant
String.trimSuffix(s: "photo.jpeg", suffix: ".jpeg") // "photo"
String.trimSuffix(suffix: ".jpeg", s: "photo.jpeg") // "photo"
```

## Applications

Drop the extension from the names that have it:

```ant
var files = ["main.ant", "util.ant", "notes.txt"]
files.map({ x.trimSuffix(".ant") }) // ["main", "util", "notes.txt"]
```

Drop whatever extension a file name has, with [`Path.ext`](../Path/ext.md). A name without one comes
back unchanged, since its extension is `""`:

```ant
var f = "photo.jpeg"
f.trimSuffix(Path.ext(f)) // "photo"
var g = "Makefile"
g.trimSuffix(Path.ext(g)) // "Makefile"
```

Read a duration such as `250ms` by trimming the unit and parsing the rest:

```ant
func millis(s: string): int | null {
  if !s.endsWith("ms") {
    return null
  }
  String.toInt(s.trimSuffix("ms"))
}

millis("250ms") // 250
millis("2s")    // null
```

Remove one trailing slash from a URL path, but keep the root:

```ant
func clean(path: string): string {
  if path == "/" { path } else { path.trimSuffix("/") }
}

clean("/docs/") // "/docs"
clean("/docs")  // "/docs"
clean("/")      // "/"
```

Tell whether the suffix was there by comparing the result with the input:

```ant
var word = "walked"
var stem = word.trimSuffix("ed")
if stem != word {
  println("past tense of ${stem}")
}
```

```text
past tense of walk
```

## Properties and relations

Whatever is appended comes back off:

```ant
for s in ["ant", "", "hill"] {
  for t in ["hill", "", "s"] {
    assert(String.trimSuffix(s + t, t) == s)
  }
}
```

When `s` ends with `suffix`, the result and the suffix join back into `s`; otherwise `s` comes back
unchanged:

```ant
var s = "report.tar.gz"
for t in [".gz", ".tar", "", s, "x"] {
  var r = s.trimSuffix(t)
  if s.endsWith(t) {
    assert(r + t == s)
  } else {
    assert(r == s)
  }
}
```

It agrees with slicing off the last code points:

```ant
var s = "año.txt"
for t in [".txt", "o.txt", "ñ", ""] {
  var n = s.length()
  var m = t.length()
  var expected = if s.endsWith(t) { s.slice(0, n - m) } else { s }
  assert(s.trimSuffix(t) == expected)
}
```

Reversing the code points of both strings turns it into [`String.trimPrefix`](trimPrefix.md):

```ant
func reverse(s: string): string {
  String.concat(s.chars().reverse())
}

for t in ["gz", "tar.gz", "tar", ""] {
  var s = "report.tar.gz"
  assert(reverse(s.trimSuffix(t)) == reverse(s).trimPrefix(reverse(t)))
}
```

## Possible issues

`String.trimSuffix` removes the suffix once. To remove every trailing copy, loop while the string
still ends with it:

```ant
var s = "done!!!"
s.trimSuffix("!") // "done!!"
while s.endsWith("!") {
  s = s.trimSuffix("!")
}
s // "done"
```

The suffix is one string, not a set of characters. To drop trailing white space of any kind, use
[`String.trimEnd`](trimEnd.md):

```ant
String.trimSuffix("line \t\n", "\n") // "line \t"
String.trimEnd("line \t\n")          // "line"
```

Removing `"\n"` from a line that ends with `"\r\n"` leaves the `"\r"`:

```ant
String.trimSuffix("line\r\n", "\n") // "line\r"
```

To drop the whole line ending, use [`String.trimEnd`](trimEnd.md), or split the text with
[`String.lines`](lines.md):

```ant
String.trimEnd("line\r\n") // "line"
```

Case matters, so `".md"` does not match `".MD"`:

```ant
String.trimSuffix("README.MD", ".md") // "README.MD"
```

The match works on code points, not on what a reader sees as one character. A family emoji is three
people joined by U+200D, so trimming the last person leaves a joiner at the end:

```ant
String.trimSuffix("👨‍👩‍👧", "👧").endsWith("\u{200D}") // true
```

The string comes first and the suffix second. Swapped, the call returns the suffix unchanged:

```ant
String.trimSuffix(".ant", "main.ant") // ".ant"
```

A nullable suffix is a compile error. Narrow it first:

<!-- test: error cannot use string | null where string is expected -->

```ant
var ext: string | null = null
println(String.trimSuffix("main.ant", ext))
```

## See also

[String.trimPrefix](trimPrefix.md) · [String.endsWith](endsWith.md) · [String.trimEnd](trimEnd.md) ·
[String.trim](trim.md) · [String.slice](slice.md) · [String.cut](cut.md) ·
[String.replace](replace.md) · [Path.ext](../Path/ext.md)
