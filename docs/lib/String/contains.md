<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# String.contains

```ant signature
func String.contains(s: string, t: string): bool
```

String.contains reports whether t occurs in s.

<!-- end generated -->

## Details

- Compares code points exactly: it does not ignore case and does not normalize, so `"\u{E9}"` and
  `"e\u{301}"` are different strings.

## Basic examples

```ant
String.contains("seafood", "foo") // true
String.contains("seafood", "bar") // false
```

Call it through `.` on any string:

```ant
var greeting = "hello world"
greeting.contains("o w")   // true
greeting.contains("Hello") // false
```

## Scope

With an empty `t` or an empty `s`:

```ant
String.contains("abc", "") // true
String.contains("", "")    // true
String.contains("", "a")   // false
```

`t` may be longer than one code point, and may be all of `s`:

```ant
String.contains("ant", "ant")  // true
String.contains("ant", "ants") // false
```

Non-ASCII text works the same way, by code points:

```ant
String.contains("naïve café", "ï") // true
String.contains("日本語のテキスト", "テキ")  // true
```

An emoji joined from several code points contains each of them:

```ant
String.contains("👨‍👩‍👧", "👩") // true
```

In a pipeline, the string is the first argument:

```ant
"key=value" | String.contains("=") // true
```

## Applications

Keep the log lines that mention an error:

```ant
var log = "ok\nERROR: disk full\nok\nERROR: timeout"
String.lines(log).filter({ x.contains("ERROR") }) // ["ERROR: disk full", "ERROR: timeout"]
```

Search without regard to case by folding both sides with [`String.foldCase`](foldCase.md):

```ant
func mentions(text: string, word: string): bool {
  text.foldCase().contains(word.foldCase())
}

println(mentions("Ant runs on STRASSE 5", "straße"))
println(mentions("Ant runs on Main St", "straße"))
```

```text
true
false
```

Reject a file name that tries to leave its directory:

```ant
func safeName(name: string): Result[string, string] {
  if name.contains("/") || name.contains("..") {
    Result.err("bad name: ${name}")
  } else {
    Result.ok(name)
  }
}

println(safeName("notes.txt"))
println(safeName("../secret"))
```

```text
Ok{ value: "notes.txt" }
Err{ error: "bad name: ../secret" }
```

## Properties and relations

For a non-empty `t`, `String.contains` agrees with [`String.indexOf`](indexOf.md) and
[`String.count`](count.md):

```ant
var s = "banana"
for t in ["an", "nab", "a", "x"] {
  assert(s.contains(t) == (s.indexOf(t) != null))
  assert(s.contains(t) == (s.count(t) > 0))
}
```

A string contains each of its prefixes, suffixes, and slices:

```ant
var s = "anthill"
for i in Iter.range(0, s.length() + 1) {
  assert(s.contains(s.slice(0, i)))
  assert(s.contains(s.slice(i, s.length())))
  for j in Iter.range(i, s.length() + 1) {
    assert(s.contains(s.slice(i, j)))
  }
}
```

`String.startsWith` and `String.endsWith` imply `String.contains`, and `String.concat` keeps every
part:

```ant
var a = "ant"
var b = "hill"
var s = String.concat([a, b])
assert(s.startsWith(a) && s.contains(a))
assert(s.endsWith(b) && s.contains(b))
```

## Possible issues

The test is case-sensitive. Fold the case of both sides, as in the `mentions` example above:

```ant
String.contains("Ant", "ant")            // false
String.contains("Ant".foldCase(), "ant") // true
```

`String.contains` compares code points, not what a reader sees. An `e` with a combining accent
contains a plain `e`, and does not contain the single code point `\u{E9}` until
[`String.normalize`](normalize.md) composes it:

```ant
var word = "cafe\u{301}"
word.contains("e")                               // true
word.contains("\u{E9}")                          // false
String.normalize(word, "NFC").contains("\u{E9}") // true
```

A substring is not a word. To test for a whole word, split the text first and use
[`Array.contains`](../Array/contains.md):

```ant
var text = "the cat sat"
text.contains("he")                 // true
String.fields(text).contains("he")  // false
String.fields(text).contains("cat") // true
```

A nullable string, such as the result of `Map` indexing, is not a `string`, so it is a compile
error:

```ant
var user = Map{ name: "Ann" }
println(String.contains(user["name"], "A"))
// error: cannot use string | null where string is expected
```

Narrow it first:

```ant
var user = Map{ name: "Ann" }
var name = user["name"]
if name != null {
  println(name.contains("A"))
}
```

```text
true
```

Or give a default with `??`:

```ant
var user = Map{ name: "Ann" }
(user["name"] ?? "").contains("A") // true
```

`String.contains` takes a string, not a predicate. To test each character with a function, use
[`String.chars`](chars.md) and [`Array.some`](../Array/some.md):

```ant
println(String.contains("abc", { x == "b" }))
// error: a lambda needs an expected function type, but string is expected
```

```ant
"a1b".chars().some({ String.isDigit(x) }) // true
```

## See also

[String.startsWith](startsWith.md) · [String.endsWith](endsWith.md) · [String.indexOf](indexOf.md) ·
[String.lastIndexOf](lastIndexOf.md) · [String.count](count.md) · [String.cut](cut.md) ·
[String.equalFold](equalFold.md) · [String.normalize](normalize.md) ·
[String.graphemes](graphemes.md) · [Array.contains](../Array/contains.md)
