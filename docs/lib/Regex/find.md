<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.find

```ant signature
func Regex.find(p: Regex.Pattern, s: string): Regex.Match | null
```

Regex.find returns the leftmost match of p in s, or null when there is none.

<!-- end generated -->

## Details

- `start` and `end` count code points, not bytes, so `String.slice(s, m.start, m.end)` is `m.text`.
- A pattern that can match the empty string, such as `/x*/`, matches at position 0 of any string, so
  the result is an empty match rather than `null`.

## Basic examples

Find the first run of digits:

```ant
Regex.find(/\d+/, "ab12") // Regex.Match{ text: "12", start: 2, end: 4, groups: ["12"], named: Map{} }
```

No match gives `null`:

```ant
Regex.find(/\d+/, "abc") // null
```

Read the matched text with `?.`, and give a fallback with `??`:

```ant
Regex.find(/\d+/, "room 42")?.text ?? "none" // "42"
Regex.find(/\d+/, "lobby")?.text ?? "none"   // "none"
```

## Scope

Only the first match is returned, even when there are more:

```ant
Regex.find(/\d+/, "1 22 333")?.text // "1"
```

A group's text is in `groups`, after the whole match at index 0:

```ant
Regex.find(/(\w+)@(\w+)/, "mail ann@home")?.groups // ["ann@home", "ann", "home"]
```

A group that took no part in the match is `null`:

```ant
Regex.find(/(a)|(b)/, "b")?.groups // ["b", null, "b"]
Regex.find(/(x)?y/, "y")?.groups   // ["y", null]
```

A named group is also in `named`, in the pattern's order:

```ant
let m = Regex.find(/(?<year>\d{4})-(?<month>\d\d)/, "since 2026-11")
m?.named          // Map{ "year": "2026", "month": "11" }
m?.named["month"] // "11"
```

An empty string has no match for a pattern that needs a character:

```ant
Regex.find(/\d+/, "") // null
```

Indices count code points, so `é` and `🐜` each take one position:

```ant
Regex.find(/ant/, "café 🐜 ant") // Regex.Match{ text: "ant", start: 7, end: 10, groups: ["ant"], named: Map{} }
```

Flags change how the pattern matches: `i` ignores case, `m` lets `^` match after a line break, and
`s` lets `.` match a line break:

```ant
Regex.find(/ANT/i, "an Ant")?.text // "Ant"
Regex.find(/^b/m, "a\nb")?.start   // 2
Regex.find(/^b/, "a\nb")           // null
Regex.find(/a.b/s, "a\nb")?.text   // "a\nb"
Regex.find(/a.b/, "a\nb")          // null
```

A pattern built from data with `Regex.compile` works the same as a literal:

```ant
let p = Result.unwrap(Regex.compile("a.c"))
Regex.find(p, "xabc")?.text // "abc"
```

The pattern comes first, so a call through `.` goes on the pattern:

```ant
/\d+/.find("a 7")?.text // "7"
```

## Applications

Parse a version number out of a line of text, with a guard for the line that has none:

```ant
type Version = { major: int, minor: int, patch: int }

func parseVersion(s: string): Version | null {
  let m = Regex.find(/(\d+)\.(\d+)\.(\d+)/, s)
  if m == null { return null }
  let n = m.groups.map({ x?.toInt() ?? 0 })
  Version{ major: n[1], minor: n[2], patch: n[3] }
}

parseVersion("ant version 0.1.0 (darwin)") // Version{ major: 0, minor: 1, patch: 0 }
parseVersion("dev build")                  // null
```

Split a `key=value` setting with named groups, and return an error for a line that is not one:

```ant
type Setting = { key: string, value: string }

func parseSetting(line: string): Result[Setting, Error] {
  let m = Regex.find(/^\s*(?<key>\w+)\s*=\s*(?<value>.*)$/, line)
  if m == null {
    return Result.fail("not a setting: ${line}")
  }
  Result.ok(Setting{ key: m.named["key"] ?? "", value: m.named["value"] ?? "" })
}

parseSetting("  port = 8080") // Ok{ value: Setting{ key: "port", value: "8080" } }
parseSetting("# comment")     // Err{ error: Error{ message: "not a setting: # comment" } }
```

Cut a string at the first match, using `start` and `end`:

```ant
func splitFirst(s: string, sep: Regex.Pattern): Array[string] {
  let m = Regex.find(sep, s)
  if m == null { return [s] }
  [s.slice(0, m.start), s.slice(m.end, s.length())]
}

splitFirst("name:  Ann: admin", /:\s*/) // ["name", "Ann: admin"]
splitFirst("plain", /:\s*/)             // ["plain"]
```

## Properties and relations

`text` is the slice of `s` between `start` and `end`, and is `groups[0]`:

```ant
let s = "price: 12 €, or 30 €"
let m = Regex.find(/\d+ €/, s)
assert(m != null)
assert(s.slice(m.start, m.end) == m.text)
assert(m.groups[0] == m.text)
```

`Regex.find` gives `null` exactly when `Regex.test` is false:

```ant
let p = /[aeiou]/
for s in ["rhythm", "ant", ""] {
  assert((Regex.find(p, s) == null) == !Regex.test(p, s))
}
```

It returns the first match `Regex.findAll` returns:

```ant
let p = /\d+/
let s = "a1b22c333"
assert(Regex.find(p, s) == Regex.findAll(p, s).first())
```

## Possible issues

The result may be `null`, so reading a field without narrowing is a compile error:

```ant
let m = Regex.find(/\d+/, "x 42")
let digits = m.text // error: Regex.Match | null has no field text: narrow the value first
```

The pattern comes first, so the call goes on the pattern, not the string:

```ant
"a 7".find(/\d+/) // error: string has no field or function find
```

```ant
/\d+/.find("a 7")?.text // "7"
```

A pattern that matches the empty string finds an empty match at the start instead of `null`. Use `+`
instead of `*` to require at least one character:

```ant
Regex.find(/\d*/, "abc 12")?.text // ""
Regex.find(/\d+/, "abc 12")?.text // "12"
```

`\w`, `\d`, and `\s` match ASCII only, so `\w+` stops at an accented letter. Use `\pL` for letters
in any script:

```ant
Regex.find(/\w+/, "héllo")?.text  // "h"
Regex.find(/\pL+/, "héllo")?.text // "héllo"
```

## See also

[Regex.findAll](findAll.md), [Regex.test](test.md), [Regex.replaceWith](replaceWith.md),
[Regex.compile](compile.md), [Regex.Match](types/Match.md), [Regex.Pattern](types/Pattern.md),
[String.slice](../String/slice.md), [String.indexOf](../String/indexOf.md)
