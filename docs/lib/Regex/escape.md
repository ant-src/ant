<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.escape

```ant signature
func Regex.escape(s: string): string
```

Regex.escape returns s with every metacharacter escaped, so a pattern compiled from it matches s
literally.

<!-- end generated -->

## Details

- Pass the result to [Regex.compile](compile.md), alone or as part of a larger source.
- It escapes text for a pattern only. A replacement string has its own escape, `$$` for a `$`.

## Basic examples

Escape the metacharacters of a string. A string value shows quoted, so each backslash appears
doubled:

```ant
Regex.escape("a.b*c")     // "a\\.b\\*c"
Regex.escape("1+1=2?")    // "1\\+1=2\\?"
Regex.escape("(see p.3)") // "\\(see p\\.3\\)"
```

Compile the escaped text to match it literally:

```ant
let p = Result.unwrap(Regex.compile(Regex.escape("1+1=2?")))
p                   // /1\+1=2\?/
p.test("is 1+1=2?") // true
p.test("11=2")      // false
```

## Scope

Every metacharacter gets a backslash, the backslash included:

```ant
Regex.escape(`\.+*?()|[]{}^$`) // "\\\\\\.\\+\\*\\?\\(\\)\\|\\[\\]\\{\\}\\^\\$"
```

Text without metacharacters, Unicode included, comes back as it is:

```ant
Regex.escape("hello world")   // "hello world"
Regex.escape("héllo 日本 🐜") // "héllo 日本 🐜"
Regex.escape("")              // ""
```

A `/` needs no escape in a compiled pattern, so the pattern renders it as a literal would:

```ant
Regex.escape("a/b")                               // "a/b"
Result.unwrap(Regex.compile(Regex.escape("a/b"))) // /a\/b/
```

The escaped text can be one part of a larger source. Here it sits between word boundaries, and the
`i` flag still applies to it:

```ant
let word = Result.unwrap(Regex.compile("\\b${Regex.escape("Go")}\\b", "i"))
word.replaceWith("go, Go, GO and gopher", { "[${x.text}]" }) // "[go], [Go], [GO] and gopher"
```

## Applications

Find any of several literal words, such as language names with `+` and `#` in them:

```ant
let names = ["c++", "c#", "f*"]
let alternatives = names | Array.map({ Regex.escape(x) }) | String.join("|")
let p = try Regex.compile("(?:${alternatives})")
println(p.findAll("I write c++, c# and f*, not c") | Array.map({ x.text }))
```

```text
["c++", "c#", "f*"]
```

Search for text a user typed, ignoring case. The user's text is data, so a `.` in it is a dot:

```ant
func countMentions(text: string, query: string): Result[int, Error] {
  let p = try Regex.compile(Regex.escape(query), "i")
  Result.ok(p.findAll(text).length())
}

println(try countMentions("Version 1.2, version 1.2.3, version 132", "VERSION 1.2"))
```

```text
2
```

## Properties and relations

The pattern compiled from `Regex.escape(s)` matches `s` itself, from end to end:

```ant
for s in ["a.b*c", `C:\temp`, "[1] (x)", "$5 ^ 2", "ünï|cödé"] {
  let p = Result.unwrap(Regex.compile("^${Regex.escape(s)}$"))
  assert(p.test(s))
  assert(p.find(s)?.text == s)
}
```

Escaping text without metacharacters changes nothing, so escaping it twice changes nothing either:

```ant
let s = "plain words"
assert(Regex.escape(s) == s)
assert(Regex.escape(Regex.escape(s)) == s)
```

## Possible issues

Text compiled without escaping is read as a pattern. It may compile and match the wrong strings:

```ant
let p = Result.unwrap(Regex.compile("1+1=2?"))
p.test("11=2")   // true
p.test("1+1=2?") // false
```

Or it may not compile at all:

```ant
Regex.compile("1+1=2?)") // Err{ error: Error{ message: "invalid regular expression: unexpected ): `1+1=2?)`" } }
```

Escaping text that is escaped already escapes its backslashes, and the pattern then looks for them:

```ant
let twice = Regex.escape(Regex.escape("a.b"))
twice                                           // "a\\\\\\.b"
Result.unwrap(Regex.compile(twice)).test("a.b") // false
```

`Regex.escape` does not escape a replacement. There a `$` is written `$$`, and a backslash is kept
as it is, so `$5` still reads as a group:

```ant
Regex.replace(/USD/, "cost: USD", Regex.escape("$5"))      // "cost: \\"
Regex.replace(/USD/, "cost: USD", "$5".replace("$", "$$")) // "cost: $5"
```

[Regex.replaceWith](replaceWith.md) puts the function's result in place as written, with no
expansion:

```ant
Regex.replaceWith(/USD/, "cost: USD", { "$5" }) // "cost: $5"
```

## See also

[Regex.compile](compile.md) · [Regex.test](test.md) · [Regex.findAll](findAll.md) ·
[Regex.replace](replace.md) · [Regex.replaceWith](replaceWith.md) ·
[String.replace](../String/replace.md) · [Regex.source](source.md) ·
[String.contains](../String/contains.md) · [Regex.Pattern](types/Pattern.md)
