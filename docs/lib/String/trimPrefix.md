<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.trimPrefix

```ant signature
func String.trimPrefix(s: string, prefix: string): string
```

String.trimPrefix returns s with prefix removed once from its start, or s unchanged when it does not
start with prefix.

<!-- end generated -->

## Details

- Removes `prefix` once, when `s` starts with it; otherwise returns `s` unchanged.
- `prefix` is a string to match exactly, not a set of characters, and it is removed once, not
  repeatedly.

## Basic examples

Remove a leading `v` from a version tag:

```ant
String.trimPrefix("v1.4.2", "v") // "1.4.2"
```

A string that does not start with the prefix comes back unchanged:

```ant
String.trimPrefix("1.4.2", "v") // "1.4.2"
```

Call it through `.` on any string:

```ant
"https://ant-lang.org".trimPrefix("https://") // "ant-lang.org"
```

## Scope

The prefix is removed once, even when it repeats:

```ant
String.trimPrefix("vvx", "v")    // "vx"
String.trimPrefix("////a", "//") // "//a"
```

An empty prefix changes nothing, and nothing is removed from an empty string:

```ant
String.trimPrefix("abc", "") // "abc"
String.trimPrefix("", "")    // ""
String.trimPrefix("", "a")   // ""
```

The whole string as the prefix leaves the empty string. A prefix longer than `s` matches nothing:

```ant
String.trimPrefix("abc", "abc") // ""
String.trimPrefix("ab", "abc")  // "ab"
```

The prefix must be at the very start. A match later in the string does not count:

```ant
String.trimPrefix("a-b-", "b-") // "a-b-"
```

Non-ASCII text works the same way, by code points:

```ant
String.trimPrefix("🐜🐝🐞", "🐜")       // "🐝🐞"
String.trimPrefix("über-cool", "über-") // "cool"
```

In a pipeline, the string is the first argument. Named arguments use `s:` and `prefix:`:

```ant
"--verbose" | String.trimPrefix("--")         // "verbose"
String.trimPrefix(s: "#heading", prefix: "#") // "heading"
```

## Applications

Read the value of a `--name=value` flag:

```ant
func flag(arg: string, name: string): string | null {
  var p = "--${name}="
  if arg.startsWith(p) { arg.trimPrefix(p) } else { null }
}

println(flag("--port=8080", "port"))
println(flag("--host=x", "port"))
```

```text
8080
null
```

Take the token out of an `Authorization` header:

```ant
var header = "Bearer abc.def.ghi"
header.trimPrefix("Bearer ") // "abc.def.ghi"
```

Turn Markdown headings into plain titles, leaving other lines alone:

```ant
var text = "# Intro\nSome text.\n# Usage"
for line in String.lines(text) {
  println(line.trimPrefix("# "))
}
```

```text
Intro
Some text.
Usage
```

Strip a common directory from a list of paths:

```ant
var files = ["src/main.ant", "src/lib/geo.ant", "README.md"]
files.map({ x.trimPrefix("src/") }) // ["main.ant", "lib/geo.ant", "README.md"]
```

## Properties and relations

When `s` starts with `prefix`, putting the prefix back gives `s`. When it does not, the result is
`s`:

```ant
for s in ["v1.0", "1.0", "v", ""] {
  var t = String.trimPrefix(s, "v")
  if s.startsWith("v") {
    assert("v" + t == s)
  } else {
    assert(t == s)
  }
}
```

When the prefix is present, the result is the slice of `s` behind it:

```ant
var s = "prefix-rest"
var p = "prefix-"
assert(s.startsWith(p))
assert(s.trimPrefix(p) == s.slice(p.length(), s.length()))
```

The result differs from `s` exactly when `s` starts with a prefix that is not empty:

```ant
for s in ["ant", "hill", ""] {
  assert((s.trimPrefix("an") != s) == s.startsWith("an"))
}
```

[`String.trimSuffix`](trimSuffix.md) does the same at the other end:

```ant
var s = "<b>"
s.trimPrefix("<").trimSuffix(">") // "b"
```

## Possible issues

The prefix is a string, not a set of characters to strip. `"xy"` matches only `x` followed by `y`:

```ant
String.trimPrefix("xxyz", "xy") // "xxyz"
String.trimPrefix("xyz", "xy")  // "z"
```

It removes one copy. To remove every leading copy, loop with [`String.startsWith`](startsWith.md):

```ant
var s = "0007"
while s.startsWith("0") {
  s = s.trimPrefix("0")
}
println(s)
```

```text
7
```

The match is case-sensitive. Lowering `s` first makes the prefix match, but it lowers the rest of
`s` too:

```ant
String.trimPrefix("Http://Ant-Lang.org", "http://")           // "Http://Ant-Lang.org"
String.trimPrefix("Http://Ant-Lang.org".toLower(), "http://") // "ant-lang.org"
```

To keep the case of the rest, compare only the start of `s` with [`String.equalFold`](equalFold.md):

```ant
var s = "Http://Ant-Lang.org"
var p = "http://"
if s.length() >= p.length() && String.equalFold(s.slice(0, p.length()), p) {
  s.slice(p.length(), s.length())
} else {
  s
} // "Ant-Lang.org"
```

The match counts code points, not what a reader sees. Removing `e` from an `e` with a combining
accent leaves the accent alone:

```ant
String.trimPrefix("e\u{301}tude", "e").startsWith("\u{301}") // true
```

To remove white space, use [`String.trimStart`](trimStart.md), which takes no prefix:

```ant
String.trimStart("  \tindent") // "indent"
```

Testing the result with `startsWith` does not tell whether a prefix was removed: `"vv1"` gives
`"v1"`, which still starts with `v`. Test `s` with [`String.startsWith`](startsWith.md) first, or
compare the result with `s`:

```ant
String.trimPrefix("vv1", "v") // "v1"
String.trimPrefix("v1", "v")  // "1"
```

## See also

[String.trimSuffix](trimSuffix.md) · [String.startsWith](startsWith.md) ·
[String.trimStart](trimStart.md) · [String.trim](trim.md) · [String.slice](slice.md) ·
[String.cut](cut.md) · [String.replaceFirst](replaceFirst.md)
