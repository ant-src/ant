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

- Replaces only the first occurrence of `old`; later ones stay. When `old` does not occur, the
  result equals `s`.
- Panics if `old` is empty.
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
var s = String.replaceFirst("e\u{301}t\u{301}", "\u{301}", "")
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
var names = ["first_name", "zip_code_ext", "id"]
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
var doc = "# Intro\ntext\n# Usage\n"
var marked = doc | String.replaceFirst("# ", "# [start] ")
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
var s = "user=ann=admin"
var c = String.cut(s, "=")
assert(c != null)
assert(String.replaceFirst(s, "=", ": ") == c.before + ": " + c.after)
```

When `old` does not occur, the result is `s`:

```ant
var s = "plain"
assert(!String.contains(s, "="))
assert(String.replaceFirst(s, "=", ": ") == s)
```

The length changes by the difference between `new` and `old` when `old` occurs, and does not change
otherwise:

```ant
var s = "a-b-c"
var r = String.replaceFirst(s, "-", "--")
assert(String.length(r) == String.length(s) - 1 + 2)
```

When `old` occurs at most once, `String.replaceFirst` and [`String.replace`](replace.md) agree:

```ant
for s in ["k=v", "kv", "=", ""] {
  assert(String.count(s, "=") <= 1)
  assert(String.replaceFirst(s, "=", ":") == String.replace(s, "=", ":"))
}
```

When `new` does not contain `old`, as here, calling it `String.count(s, old)` times replaces every
occurrence, as `String.replace` does:

```ant
var s = "a-b-c-d"
var r = s
var n = String.count(s, "-")
while n > 0 {
  r = String.replaceFirst(r, "-", "+")
  n = n - 1
}
assert(r == String.replace(s, "-", "+"))
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

The first occurrence can be anywhere in `s`, not only at the start:

```ant
String.replaceFirst("x-a", "a", "b") // "x-b"
```

To change only a leading prefix, test it with [`String.startsWith`](startsWith.md), or remove it
with [`String.trimPrefix`](trimPrefix.md) and concatenate:

```ant
var s = "C# # notes"
String.replaceFirst(s, "# ", "## ")    // "C## # notes"
String.startsWith(s, "# ")             // false
"## " + String.trimPrefix("# x", "# ") // "## x"
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

A loop that replaces until `old` is gone never ends when `new` contains `old`. Bound the loop by
`String.count`, as in the properties above, or use [`String.replace`](replace.md).

## See also

[String.replace](replace.md) · [Regex.replaceFirst](../Regex/replaceFirst.md) · [String.cut](cut.md)
· [String.indexOf](indexOf.md) · [String.trimPrefix](trimPrefix.md) · [String.count](count.md) ·
[String.contains](contains.md)
