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

- An empty `suffix` returns `s`.
- The match is exact, code point by code point, as in [`String.endsWith`](endsWith.md): it does not
  ignore case or normalize.

## Basic examples

```ant
String.trimSuffix("main.ant", ".ant") // "main"
String.trimSuffix("main.ant", ".go")  // "main.ant"
```

Call it through `.` on any string:

```ant
let name = "report.tar.gz"
name.trimSuffix(".gz")     // "report.tar"
name.trimSuffix(".tar.gz") // "report"
name.trimSuffix(".tar")    // "report.tar.gz"
```

## Scope

The suffix comes off once, even when the string ends with several copies of it:

```ant
"ha!!!".trimSuffix("!") // "ha!!"
"abab".trimSuffix("ab") // "ab"
```

An empty suffix changes nothing, an empty string stays empty, and a suffix equal to the whole string
leaves the empty string:

```ant
"ant".trimSuffix("")    // "ant"
"".trimSuffix("x")      // ""
"ant".trimSuffix("ant") // ""
```

A suffix longer than the string never matches:

```ant
"ant".trimSuffix("giant") // "ant"
```

Text in any script trims by code points:

```ant
"東京都".trimSuffix("都") // "東京"
"Ωμέγα".trimSuffix("γα")  // "Ωμέ"
```

A combining accent is a code point of its own, so it can come off alone and leave the bare letter:

```ant
"cafe\u{301}".trimSuffix("\u{301}") // "cafe"
```

In a pipeline, the string is the first argument. Named arguments use `s:` and `suffix:`:

```ant
"index.html" | String.trimSuffix(".html")           // "index"
String.trimSuffix(s: "photo.jpeg", suffix: ".jpeg") // "photo"
```

## Applications

Drop the extension from the names that have it:

```ant
let files = ["main.ant", "util.ant", "notes.txt"]
files.map({ x.trimSuffix(".ant") }) // ["main", "util", "notes.txt"]
```

Drop whatever extension a file name has, with [`Path.ext`](../Path/ext.md). A name without one comes
back unchanged, since its extension is `""`:

```ant
let f = "photo.jpeg"
f.trimSuffix(Path.ext(f)) // "photo"
let g = "Makefile"
g.trimSuffix(Path.ext(g)) // "Makefile"
```

Read a count with a unit, such as `250ms`, by trimming the unit and parsing the rest:

```ant
func millis(s: string): int | null {
  if !s.endsWith("ms") {
    return null
  }
  s.trimSuffix("ms").toInt()
}

millis("250ms") // 250
millis("2s")    // null
```

To read duration text as a `Time.Duration`, use [`Time.parseDuration`](../Time/parseDuration.md),
which reads Go notation such as `"250ms"` or `"2h30m"` as well as ISO 8601:

```ant
Time.parseDuration("250ms") // Ok{ value: PT0.25S }
Time.parseDuration("2s")    // Ok{ value: PT2S }
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
let word = "walked"
let stem = word.trimSuffix("ed")
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
    assert((s + t).trimSuffix(t) == s)
  }
}
```

When `s` ends with `suffix`, the result and the suffix join back into `s`; otherwise `s` comes back
unchanged:

```ant
let s = "report.tar.gz"
for t in [".gz", ".tar", "", s, "x"] {
  let r = s.trimSuffix(t)
  if s.endsWith(t) {
    assert(r + t == s)
  } else {
    assert(r == s)
  }
}
```

It agrees with slicing off the last code points:

```ant
let s = "año.txt"
for t in [".txt", "o.txt", "ñ", ""] {
  let n = s.length()
  let m = t.length()
  let expected = if s.endsWith(t) { s.slice(0, n - m) } else { s }
  assert(s.trimSuffix(t) == expected)
}
```

Reversing the code points of both strings turns it into [`String.trimPrefix`](trimPrefix.md):

```ant
func reverse(s: string): string {
  s.chars().reverse() | String.concat()
}

for t in ["gz", "tar.gz", "tar", ""] {
  let s = "report.tar.gz"
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
"line \t\n".trimSuffix("\n") // "line \t"
"line \t\n".trimEnd()        // "line"
```

Removing `"\n"` from a line that ends with `"\r\n"` leaves the `"\r"`:

```ant
"line\r\n".trimSuffix("\n") // "line\r"
```

To drop the whole line ending, use [`String.trimEnd`](trimEnd.md), or split the text with
[`String.lines`](lines.md):

```ant
"line\r\n".trimEnd() // "line"
```

Case matters, so `".md"` does not match `".MD"`:

```ant
"README.MD".trimSuffix(".md") // "README.MD"
```

The match works on code points, not on what a reader sees as one character. A family emoji is three
people joined by U+200D, so trimming the last person leaves a joiner at the end:

```ant
"👨‍👩‍👧".trimSuffix("👧").endsWith("\u{200D}") // true
```

The string comes first and the suffix second. Swapped, the call returns the suffix unchanged:

```ant
String.trimSuffix(".ant", "main.ant") // ".ant"
```

A nullable suffix is a compile error. Narrow it first:

```ant
let ext: string | null = null
println(String.trimSuffix("main.ant", ext))
// error: cannot use string | null where string is expected
```

## See also

[String.trimPrefix](trimPrefix.md) · [String.endsWith](endsWith.md) · [String.trimEnd](trimEnd.md) ·
[String.trim](trim.md) · [String.slice](slice.md) · [String.cut](cut.md) ·
[String.replace](replace.md) · [Path.ext](../Path/ext.md) ·
[Time.parseDuration](../Time/parseDuration.md)
