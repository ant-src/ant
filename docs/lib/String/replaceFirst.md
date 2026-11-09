<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.replaceFirst

```ant signature
func String.replaceFirst(s: string, old: string, new: string): string
```

String.replaceFirst returns s with the first occurrence of old replaced by new. It panics if old is
empty.

<!-- end generated -->

## Details

- When `old` does not occur, the result equals `s`.
- Matching is exact, code point by code point, with no case folding or normalization.
- To replace every occurrence, use [`String.replace`](replace.md).

## Basic examples

Replace the first `-` and keep the rest:

```ant
String.replaceFirst("a-b-c", "-", "+") // "a+b-c"
```

When `old` does not occur, the string comes back unchanged:

```ant
String.replaceFirst("abc", "x", "y") // "abc"
```

An empty `new` removes the first occurrence:

```ant
String.replaceFirst("2026-11-08", "-", "") // "202611-08"
```

## Scope

`old` can be any string that is not empty, `new` can be any string, and `new` may contain `old`:

```ant
String.replaceFirst("a::b::c", "::", "/") // "a/b::c"
String.replaceFirst("aXa", "a", "aa")     // "aaXa"
```

The first occurrence is the leftmost one, so the match starts at the first position where `old`
begins:

```ant
String.replaceFirst("aaa", "aa", "b") // "ba"
```

In the empty string, `old` never occurs, so the result is `""`:

```ant
String.replaceFirst("", "x", "y") // ""
```

The whole string can be `old`:

```ant
String.replaceFirst("abc", "abc", "xyz") // "xyz"
```

Unicode text works like any other. The match is exact, so case matters and a precomposed `é` does
not contain the code point `e`:

```ant
String.replaceFirst("🐜🐝🐜", "🐜", "🐞") // "🐞🐝🐜"
String.replaceFirst("Café", "c", "K")     // "Café"
String.replaceFirst("é", "e", "a")        // "é"
```

The search matches code points, not graphemes, so it can strip the first combining accent from its
letter. The code point 769 is the accent U+0301 that stays on `t`:

```ant
let s = String.replaceFirst("e\u{301}t\u{301}", "\u{301}", "")
String.toCodePoints(s) // [101, 116, 769]
```

Call it through `.`, in a pipe, or with named arguments:

```ant
"x.y.z".replaceFirst(".", "!")                      // "x!y.z"
"x.y.z" | String.replaceFirst(".", "!")             // "x!y.z"
String.replaceFirst(s: "a b c", old: " ", new: "_") // "a_b c"
```

## Applications

Fill `?` placeholders in order, one value each:

```ant
func fill(template: string, values: Array[string]): string {
  var out = template
  for v in values {
    out = String.replaceFirst(out, "?", v)
  }
  out
}

println(fill("? + ? = ?", ["1", "2", "3"]))
```

```text
1 + 2 = 3
```

Turn only the first `_` of a `snake_case` name into a space; later underscores stay:

```ant
let names = ["first_name", "zip_code_ext", "id"]
for name in names {
  println(name.replaceFirst("_", " "))
}
```

```text
first name
zip code_ext
id
```

Mark the first `# ` in a document, in a pipe:

```ant
let doc = "# Intro\ntext\n# Usage\n"
let marked = doc | String.replaceFirst("# ", "# [start] ")
print(marked)
```

```text
# [start] Intro
text
# Usage
```

## Properties and relations

When `old` occurs, the result is the text before it, then `new`, then the text after it, as
[`String.cut`](cut.md) splits it:

```ant
let s = "user=ann=admin"
let c = String.cut(s, "=")
assert(c != null)
assert(String.replaceFirst(s, "=", ": ") == c.before + ": " + c.after)
```

When `old` does not occur, the result is `s`:

```ant
let s = "plain"
assert(!String.contains(s, "="))
assert(String.replaceFirst(s, "=", ": ") == s)
```

The length changes by the difference between `new` and `old` when `old` occurs, and does not change
otherwise:

```ant
let s = "a-b-c"
let r = String.replaceFirst(s, "-", "--")
assert(String.length(r) == String.length(s) - 1 + 2)
```

When `old` occurs at most once, `String.replaceFirst` and [`String.replace`](replace.md) agree:

```ant
for s in ["k=v", "kv", "=", ""] {
  assert(String.count(s, "=") <= 1)
  assert(String.replaceFirst(s, "=", ":") == String.replace(s, "=", ":"))
}
```

## Possible issues

An empty `old` panics:

```ant
String.replaceFirst("abc", "", "x") // panic: String.replaceFirst needs an old string that is not empty
```

`String.replaceFirst` returns a new string. Calling it without using the result changes nothing:

```ant
var s = "a-b"
String.replaceFirst(s, "-", "+")
println(s)
s = String.replaceFirst(s, "-", "+")
println(s)
```

```text
a-b
a+b
```

`old` is plain text, not a pattern: `.` matches a dot only. To replace the first match of a regular
expression, use [`Regex.replaceFirst`](../Regex/replaceFirst.md):

```ant
String.replaceFirst("a1b22", "[0-9]+", "#") // "a1b22"
```

`String.replaceFirst` matches anywhere, not only at the start:

```ant
String.replaceFirst("C# # notes", "# ", "## ") // "C## # notes"
```

To change only a leading prefix, test it with [`String.startsWith`](startsWith.md) first:

```ant
func promote(s: string): string {
  if String.startsWith(s, "# ") { "#" + s } else { s }
}

println(promote("C# # notes"))
println(promote("# x"))
```

```text
C# # notes
## x
```

Each call searches the whole string again from the front, including text that an earlier call
inserted. In the `fill` helper above, a value that contains `?` is itself replaced by the next
value:

<!-- test: before
func fill(template: string, values: Array[string]): string {
  var out = template
  for v in values {
    out = String.replaceFirst(out, "?", v)
  }
  out
}
-->

```ant
fill("? ?", ["?", "x"]) // "x ?"
```

A loop that replaces until `old` is gone never ends when `new` contains `old`. To replace every
occurrence, use [`String.replace`](replace.md).

There is no `String.replaceLast`. To replace the last occurrence, find it with
[`String.lastIndexOf`](lastIndexOf.md) and rebuild the string around it:

```ant
func replaceLast(s: string, old: string, new: string): string {
  let i = String.lastIndexOf(s, old)
  if i == null {
    return s
  }
  String.slice(s, 0, i) + new + String.slice(s, i + String.length(old), String.length(s))
}

println(replaceLast("a-b-c", "-", "+"))
```

```text
a-b+c
```

## See also

[String.replace](replace.md) · [Regex.replaceFirst](../Regex/replaceFirst.md) · [String.cut](cut.md)
· [String.indexOf](indexOf.md) · [String.lastIndexOf](lastIndexOf.md) ·
[String.trimPrefix](trimPrefix.md) · [String.count](count.md) · [String.contains](contains.md)
