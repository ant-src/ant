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
- The match is exact, code point by code point as `==` compares: there is no case folding or
  normalization, so the two spellings of "é" differ.

## Basic examples

```ant
String.startsWith("https://ant-lang.org", "https://") // true
String.startsWith("https://ant-lang.org", "http://")  // false
```

Call it through `.` on any string:

```ant
let flag = "--verbose"
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
String.startsWith("Ωμέγα", "Ωμ")    // true
String.startsWith("東京都", "東京") // true
String.startsWith("ñandú", "ñ")     // true
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

On a string that can be null, use `?.` and supply a default with `??`:

```ant
let name: string | null = null
name?.startsWith("x") ?? false // false
```

## Applications

Drop the comment lines of a config file:

```ant
let text = "# settings\nport=8080\n# debug=true\nhost=localhost"
text.lines().filter({ !x.trimStart().startsWith("#") }) // ["port=8080", "host=localhost"]
```

Split command-line arguments into flags and operands:

```ant
let args = ["-v", "main.ant", "--out", "build", "util.ant"]
let flags = args.filter({ x.startsWith("-") })
let operands = args.filter({ !x.startsWith("-") })
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
let s = "anthill"
for t in ["ant", "anth", "", "anthill", "anthills", "hill"] {
  let m = t.length()
  assert(s.startsWith(t) == (m <= s.length() && s.slice(0, m) == t))
}
```

`String.startsWith` holds exactly when [`String.indexOf`](indexOf.md) finds `t` at index 0:

```ant
let s = "banana"
for t in ["ban", "an", "", "x"] {
  assert(s.startsWith(t) == (s.indexOf(t) == 0))
}
```

A string always starts with what is prepended to it, and [`String.trimPrefix`](trimPrefix.md)
removes exactly that:

```ant
let s = "hill"
let t = "ant"
assert((t + s).startsWith(t))
assert((t + s).trimPrefix(t) == s)
```

Reversing the code points of both strings turns `startsWith` into [`String.endsWith`](endsWith.md):

```ant
func reverse(s: string): string {
  String.concat(s.chars().reverse())
}

for t in ["re", "report", "port", ""] {
  let s = "report.tar.gz"
  assert(s.startsWith(t) == reverse(s).endsWith(reverse(t)))
}
```

On the UTF-8 bytes, [`Bytes.startsWith`](../Bytes/startsWith.md) gives the same answer:

```ant
let s = "año"
for t in ["a", "añ", "ñ", ""] {
  assert(s.startsWith(t) == Bytes.startsWith(s.toBytes(), t.toBytes()))
}
```

## Possible issues

The string comes first and the prefix second. Swapped, the call asks whether the prefix begins with
the string, which is false whenever the string is longer:

```ant
String.startsWith("http", "http://ant-lang.org") // false
String.startsWith("http://ant-lang.org", "http") // true
```

A short prefix also matches longer words. `"http"` matches an `https` URL too, so include the
separator when it matters:

```ant
let url = "https://ant-lang.org"
url.startsWith("http")    // true
url.startsWith("http://") // false
```

Case matters. Fold both sides with [`String.foldCase`](foldCase.md) to compare without regard to
case:

```ant
let header = "Content-Type: text/html"
header.startsWith("content-type")                       // false
header.foldCase().startsWith("Content-Type".foldCase()) // true
```

Folding changes more than case, so fold the prefix too:

```ant
"Straße".foldCase().startsWith("straß")            // false
"Straße".foldCase().startsWith("STRAß".foldCase()) // true
```

Leading white space is part of the string. Trim it with [`String.trimStart`](trimStart.md) first:

```ant
let line = "  # note"
line.startsWith("#")             // false
line.trimStart().startsWith("#") // true
```

The precomposed "é" and "e" followed by a combining accent are different code points. Normalize both
sides with [`String.normalize`](normalize.md) when text can arrive in either spelling:

```ant
let word = "e\u{301}cole"
word.startsWith("é")                                   // false
word.normalize("NFC").startsWith("é".normalize("NFC")) // true
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
let prefix: string | null = null
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
