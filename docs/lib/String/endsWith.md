<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.endsWith

```ant signature
func String.endsWith(s: string, t: string): bool
```

String.endsWith reports whether s ends with t.

<!-- end generated -->

## Details

- Compares code points exactly: it does not ignore case and does not normalize, so the two spellings
  of "é" differ.

## Basic examples

```ant
String.endsWith("main.ant", ".ant") // true
String.endsWith("main.ant", ".go")  // false
```

Call it through `.` on any string:

```ant
var name = "report.tar.gz"
name.endsWith(".gz")    // true
name.endsWith(".tar")   // false
name.endsWith("tar.gz") // true
```

## Scope

Every string ends with the empty string, and the empty string ends with nothing else:

```ant
String.endsWith("ant", "") // true
String.endsWith("", "")    // true
String.endsWith("", "t")   // false
```

A suffix longer than the string never matches, even when the string is inside it:

```ant
String.endsWith("ant", "giant") // false
String.endsWith("giant", "ant") // true
```

Text in any script compares by code points:

```ant
String.endsWith("Ωμέγα", "γα") // true
String.endsWith("東京都", "都")    // true
String.endsWith("naïve", "ve") // true
```

In a pipeline, the string is the first argument:

```ant
"index.html" | String.endsWith(".html") // true
```

## Named arguments

`String.endsWith` has no defaults, but naming both arguments makes clear which string is the suffix:

```ant
String.endsWith(s: "photo.jpeg", t: ".jpeg") // true
String.endsWith(t: ".jpeg", s: "photo.jpeg") // true
```

## Applications

Keep the files with one extension:

```ant
var files = ["main.ant", "notes.txt", "util.ant", "ant.png"]
files.filter({ x.endsWith(".ant") }) // ["main.ant", "util.ant"]
```

Accept any of several suffixes with `Array.some`:

```ant
func isImage(name: string): bool {
  [".png", ".jpg", ".gif"].some({ name.endsWith(x) })
}

isImage("ant.png") // true
isImage("ant.svg") // false
```

Join shell lines that end with a backslash:

```ant
var lines = ["cc -o ant \\", "  main.c \\", "  util.c", "echo done"]
var commands: Array[string] = []
var pending = ""
for line in lines {
  if line.endsWith("\\") {
    pending += String.trimSuffix(line, "\\")
  } else {
    commands.push(pending + line)
    pending = ""
  }
}
for c in commands {
  println(c)
}
```

```text
cc -o ant   main.c   util.c
echo done
```

## Properties and relations

`String.endsWith` agrees with slicing the last code points of `s`:

```ant
var s = "report.tar.gz"
for t in [".gz", ".tar", "", "report.tar.gz", "x.report.tar.gz"] {
  var n = s.length()
  var m = t.length()
  assert(s.endsWith(t) == (m <= n && s.slice(n - m, n) == t))
}
```

A string always ends with what is appended to it, and `String.trimSuffix` removes exactly that:

```ant
var s = "ant"
var t = "hill"
assert((s + t).endsWith(t))
assert(String.trimSuffix(s + t, t) == s)
```

Reversing the code points of both strings turns `endsWith` into
[`String.startsWith`](startsWith.md):

```ant
func reverse(s: string): string {
  String.concat(s.chars().reverse())
}

for t in ["gz", "tar.gz", "tar", ""] {
  var s = "report.tar.gz"
  assert(s.endsWith(t) == reverse(s).startsWith(reverse(t)))
}
```

On the UTF-8 bytes, [`Bytes.endsWith`](../Bytes/endsWith.md) gives the same answer:

```ant
var s = "año"
for t in ["o", "ño", "ñ", ""] {
  assert(s.endsWith(t) == Bytes.endsWith(String.toBytes(s), String.toBytes(t)))
}
```

## Possible issues

The string comes first and the suffix second. Swapped, the call asks the opposite question:

```ant
String.endsWith(".ant", "main.ant") // false
String.endsWith("main.ant", ".ant") // true
```

Case matters, so lower both sides, or fold them with [`String.foldCase`](foldCase.md), to ignore it:

```ant
var name = "PHOTO.JPG"
name.endsWith(".jpg")                  // false
name.toLower().endsWith(".jpg")        // true
String.foldCase(name).endsWith(".jpg") // true
```

The precomposed "é" and "e" followed by a combining accent are different code points: the string
`"cafe\u{301}"` ends with the accent, not with "é". Normalize both sides with
[`String.normalize`](normalize.md) when text can arrive in either spelling:

```ant
var word = "cafe\u{301}"
word.endsWith("\u{301}")                                             // true
word.endsWith("é")                                                   // false
String.normalize(word, "NFC").endsWith(String.normalize("é", "NFC")) // true
```

`String.endsWith` works on code points, not on what a reader sees as one character. A family emoji
is three people joined by U+200D, so it ends with the last of them:

```ant
String.endsWith("👨‍👩‍👧", "👧") // true
```

[`String.graphemes`](graphemes.md) shows the one character a reader sees:

```ant
String.graphemes("👨‍👩‍👧").length() // 1
```

There is no character type and no `'c'` literal; a character is a string of one code point:

<!-- test: error unexpected character -->

```ant
println(String.endsWith("abc", 'c'))
```

A nullable suffix is a compile error. Narrow it first:

<!-- test: error cannot use string | null where string is expected -->

```ant
var ext: string | null = null
println(String.endsWith("main.ant", ext))
```

`String.endsWith` takes one suffix, not an array of them; use `Array.some`, as `isImage` above does:

<!-- test: error cannot use Array[string] where string is expected -->

```ant
println(String.endsWith("ant.png", [".png", ".jpg"]))
```

## See also

[String.startsWith](startsWith.md) · [String.trimSuffix](trimSuffix.md) ·
[String.contains](contains.md) · [String.lastIndexOf](lastIndexOf.md) · [String.slice](slice.md) ·
[Bytes.endsWith](../Bytes/endsWith.md) · [Path.ext](../Path/ext.md)
