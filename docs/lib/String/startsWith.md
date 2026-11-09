<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.startsWith

```ant signature
func String.startsWith(s: string, t: string): bool
```

String.startsWith reports whether s begins with t.

<!-- end generated -->

## Details

- Every string starts with `""`.
- The match is exact, code point by code point: it does not ignore case or normalize, so the two
  spellings of "é" differ.

## Basic examples

```ant
String.startsWith("https://ant-lang.org", "https://") // true
String.startsWith("https://ant-lang.org", "http://")  // false
```

Call it through `.` on any string:

```ant
var flag = "--verbose"
flag.startsWith("--") // true
flag.startsWith("-")  // true
flag.startsWith("-v") // false
```

## Scope

Every string starts with the empty string, and the empty string starts with nothing else:

```ant
String.startsWith("ant", "") // true
String.startsWith("", "")    // true
String.startsWith("", "a")   // false
```

A prefix longer than the string never matches, even when the string is inside it:

```ant
String.startsWith("ant", "anthill") // false
String.startsWith("anthill", "ant") // true
```

A string starts with itself:

```ant
String.startsWith("ant", "ant") // true
```

Text in any script compares by code points:

```ant
String.startsWith("Ωμέγα", "Ωμ") // true
String.startsWith("東京都", "東京")   // true
String.startsWith("ñandú", "ñ")  // true
```

In a pipeline, the string is the first argument:

```ant
"# title" | String.startsWith("#") // true
```

Naming the arguments makes clear which string is the prefix, in either order:

```ant
String.startsWith(s: "v1.2.0", t: "v") // true
String.startsWith(t: "v", s: "v1.2.0") // true
```

## Applications

Drop the comment lines of a config file:

```ant
var text = "# settings\nport=8080\n# debug=true\nhost=localhost"
String.lines(text).filter({ !x.trimStart().startsWith("#") }) // ["port=8080", "host=localhost"]
```

Split command-line arguments into flags and operands:

```ant
var args = ["-v", "main.ant", "--out", "build", "util.ant"]
var flags = args.filter({ x.startsWith("-") })
var operands = args.filter({ !x.startsWith("-") })
println(flags)
println(operands)
```

```text
["-v", "--out"]
["main.ant", "build", "util.ant"]
```

Pick a handler by the scheme of a URL:

```ant
func scheme(url: string): string {
  if url.startsWith("https://") {
    "secure web"
  } else if url.startsWith("http://") {
    "web"
  } else if url.startsWith("file://") {
    "local file"
  } else {
    "unknown"
  }
}

println(scheme("https://ant-lang.org"))
println(scheme("http://example.com"))
println(scheme("file:///tmp/x"))
println(scheme("ftp://example.com"))
```

```text
secure web
web
local file
unknown
```

## Properties and relations

`String.startsWith` agrees with slicing the first code points of `s`:

```ant
var s = "anthill"
for t in ["ant", "anth", "", "anthill", "anthills", "hill"] {
  var m = t.length()
  assert(s.startsWith(t) == (m <= s.length() && s.slice(0, m) == t))
}
```

`String.startsWith` holds exactly when [`String.indexOf`](indexOf.md) finds `t` at index 0:

```ant
var s = "banana"
for t in ["ban", "an", "", "x"] {
  assert(s.startsWith(t) == (s.indexOf(t) == 0))
}
```

A string always starts with what is prepended to it, and [`String.trimPrefix`](trimPrefix.md)
removes exactly that:

```ant
var s = "hill"
var t = "ant"
assert((t + s).startsWith(t))
assert(String.trimPrefix(t + s, t) == s)
```

Reversing the code points of both strings turns `startsWith` into [`String.endsWith`](endsWith.md):

```ant
func reverse(s: string): string {
  String.concat(s.chars().reverse())
}

for t in ["re", "report", "port", ""] {
  var s = "report.tar.gz"
  assert(s.startsWith(t) == reverse(s).endsWith(reverse(t)))
}
```

On the UTF-8 bytes, [`Bytes.startsWith`](../Bytes/startsWith.md) gives the same answer:

```ant
var s = "año"
for t in ["a", "añ", "ñ", ""] {
  assert(s.startsWith(t) == Bytes.startsWith(String.toBytes(s), String.toBytes(t)))
}
```

## Possible issues

The string comes first and the prefix second. Swapped, the call asks the opposite question:

```ant
String.startsWith("http", "http://ant-lang.org") // false
String.startsWith("http://ant-lang.org", "http") // true
```

A short prefix also matches longer words. `"http"` matches an `https` URL too, so include the
separator when it matters:

```ant
var url = "https://ant-lang.org"
url.startsWith("http")    // true
url.startsWith("http://") // false
```

Case matters, so lower the string, or fold it with [`String.foldCase`](foldCase.md), and compare it
with a lower-case prefix:

```ant
var header = "Content-Type: text/html"
header.startsWith("content-type")                  // false
header.toLower().startsWith("content-type")        // true
String.foldCase(header).startsWith("content-type") // true
```

Leading white space is part of the string. Trim it with [`String.trimStart`](trimStart.md) first:

```ant
var line = "  # note"
line.startsWith("#")             // false
line.trimStart().startsWith("#") // true
```

The precomposed "é" and "e" followed by a combining accent are different code points. Normalize both
sides with [`String.normalize`](normalize.md) when text can arrive in either spelling:

```ant
var word = "e\u{301}cole"
word.startsWith("é")                                                   // false
String.normalize(word, "NFC").startsWith(String.normalize("é", "NFC")) // true
```

`String.startsWith` works on code points, not on what a reader sees as one character. A family emoji
is three people joined by U+200D, so it starts with the first of them:

```ant
String.startsWith("👨‍👩‍👧", "👨") // true
```

There is no character type and no `'c'` literal; a character is a string of one code point:

```ant
println(String.startsWith("abc", 'a')) // error: unexpected character '\''
```

A nullable prefix is a compile error. Narrow it first:

```ant
var prefix: string | null = null
println(String.startsWith("main.ant", prefix))
// error: cannot use string | null where string is expected
```

`String.startsWith` takes one prefix, not an array of them; use [`Array.some`](../Array/some.md):

```ant
println(String.startsWith("https://x", ["http://", "https://"]))
// error: cannot use Array[string] where string is expected
```

```ant
["http://", "https://"].some({ "https://x".startsWith(x) }) // true
```

## See also

[String.endsWith](endsWith.md) · [String.trimPrefix](trimPrefix.md) · [String.contains](contains.md)
· [String.indexOf](indexOf.md) · [String.cut](cut.md) · [String.slice](slice.md) ·
[Bytes.startsWith](../Bytes/startsWith.md)
