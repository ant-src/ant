<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.test

```ant signature
func Regex.test(p: Regex.Pattern, s: string): bool
```

Regex.test reports whether p matches anywhere in s.

<!-- end generated -->

## Details

- The pattern matches anywhere in `s` unless it is anchored: anchor it with `^` and `$` to match the
  whole string.
- Without the `m` flag, `$` matches only at the very end of `s`, not before a final line break.
- `\d`, `\w`, `\s`, and `\b` are ASCII. Use `\pL`, `\pN`, or `\p{Greek}` for other scripts.

## Basic examples

Check whether a string contains digits:

```ant
Regex.test(/\d+/, "order 66")  // true
Regex.test(/\d+/, "no digits") // false
```

Anchor the pattern to check the whole string:

```ant
Regex.test(/^\d+$/, "66")  // true
Regex.test(/^\d+$/, "66a") // false
```

## Scope

Call it with a dot on the pattern, a literal or a name:

```ant
/\d/.test("a1") // true
let date = /^\d{4}-\d{2}-\d{2}$/
date.test("2026-11-12")  // true
date.test("2026-11-12x") // false
```

An empty string matches a pattern that can match the empty string, and no pattern that needs a
character:

```ant
Regex.test(/x*/, "") // true
Regex.test(/./, "")  // false
```

The `i` flag ignores case, either after the literal or as a `(?i)` group inside it:

```ant
Regex.test(/ant/i, "ANTHILL")    // true
Regex.test(/(?i)ant/, "ANTHILL") // true
Regex.test(/ant/, "ANTHILL")     // false
```

The `m` flag lets `^` and `$` match at every line, and the `s` flag lets `.` match a line break:

```ant
Regex.test(/^b/, "a\nb")   // false
Regex.test(/^b/m, "a\nb")  // true
Regex.test(/a.b/, "a\nb")  // false
Regex.test(/a.b/s, "a\nb") // true
```

`\d` matches only ASCII digits. `\pN` matches a digit of any script, and `\pL` a letter:

```ant
Regex.test(/\d/, "٣")          // false
Regex.test(/\pN/, "٣")         // true
Regex.test(/^\pL+$/, "日本語") // true
```

`\b` marks a word boundary, so the pattern matches a whole word only:

```ant
Regex.test(/\bant\b/, "an ant") // true
Regex.test(/\bant\b/, "giant")  // false
```

A pattern from data comes from [Regex.compile](compile.md), which returns a `Result`. Unwrap it
before the test:

```ant
let p = Regex.compile("a+b").unwrap()
Regex.test(p, "caab") // true
```

## Applications

Keep the log lines that start with a time:

```ant
let lines = ["09:14 start", "loading config", "09:15 ready", "done"]
lines.filter({ /^\d{2}:\d{2} /.test(x) }) // ["09:14 start", "09:15 ready"]
```

Find the first capitalized word:

```ant
["ant", "Bee", "cat"].find({ /^[A-Z]/.test(x) }) // "Bee"
```

Check a password against a list of rules, and name the first one it breaks:

```ant
type Rule = { name: string, pattern: Regex.Pattern }

let rules = [
  Rule{ name: "a digit", pattern: /\d/ },
  Rule{ name: "an uppercase letter", pattern: /\p{Lu}/ },
  Rule{ name: "at least 8 characters", pattern: /^.{8,}$/ },
]

func check(password: string): string {
  let broken = rules.find({ !x.pattern.test(password) })
  if broken == null { "ok" } else { "needs ${broken.name}" }
}

println(check("hunter2"))
println(check("Hunter22"))
```

```text
needs an uppercase letter
ok
```

Test lines against a pattern given by the user, and report a bad pattern as an error:

```ant
func grep(source: string, lines: Array[string]): Result[Array[string], Error] {
  let p = try (Regex.compile(source) | Result.wrapErr("pattern ${source}"))
  Result.ok(lines.filter({ p.test(x) }))
}

println(grep("^a", ["ant", "bee", "aphid"]))
println(grep("(", ["ant"]))
```

```text
Ok{ value: ["ant", "aphid"] }
Err{ error: Error{ message: "pattern (: invalid regular expression: missing closing ): `(`" } }
```

## Properties and relations

`Regex.test` is true exactly when [Regex.find](find.md) finds a match:

```ant
let p = /\d+/
for s in ["a1", "abc", ""] {
  assert(p.test(s) == (Regex.find(p, s) != null))
}
```

A pattern from [Regex.escape](escape.md) matches its text literally, as `String.contains` does:

```ant
let text = "1+1"
let p = Regex.compile(Regex.escape(text)).unwrap()
for s in ["1+1=2", "11", ""] {
  assert(p.test(s) == s.contains(text))
}
```

The empty pattern matches every string:

```ant
let p = Regex.compile("").unwrap()
assert(["", "ant"].every({ p.test(x) }))
```

## Possible issues

An unanchored pattern matches part of the string, so it passes input it is meant to reject. Anchor a
check of the whole string:

```ant
Regex.test(/\d+/, "12abc")   // true
Regex.test(/^\d+$/, "12abc") // false
```

`$` does not match before a final line break. Trim a line read with its line break first:

```ant
Regex.test(/^ok$/, "ok\n")        // false
Regex.test(/^ok$/, "ok\n".trim()) // true
```

Text from a user is a pattern of its own, where `+` and `.` are metacharacters. Escape it with
[Regex.escape](escape.md) to match it literally:

```ant
let input = "1+1"
Regex.test(Regex.compile(input).unwrap(), "1+1=2")               // false
Regex.test(Regex.compile(Regex.escape(input)).unwrap(), "1+1=2") // true
```

The pattern comes first. The arguments the other way round are a compile error:

```ant
let ok = Regex.test("order 66", /\d+/)
// error: cannot use string where Regex.Pattern is expected
// error: cannot use Regex.Pattern where string is expected
```

Strings have no `test`; call it on the pattern:

```ant
let ok = "order 66".test(/\d+/) // error: string has no field or function test
```

[Regex.compile](compile.md) returns a `Result`, not a pattern:

```ant
let p = Regex.compile("a+b")
let ok = Regex.test(p, "aab") // error: cannot use Result[Regex.Pattern, Error] where Regex.Pattern is expected
```

A literal is checked when the source is parsed, so a bad one is a syntax error, not `false`:

```ant
let ok = Regex.test(/(/, "a") // error: invalid regular expression: missing closing ): `(`
```

There is no `g` flag: `Regex.test` looks for one match, and [Regex.findAll](findAll.md) returns them
all:

```ant
let ok = Regex.test(/a/g, "a") // error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

The `i` flag folds case one code point at a time, so `ß` does not match `SS`:

```ant
Regex.test(/straße/i, "STRASSE") // false
```

## See also

[Regex.find](find.md), [Regex.findAll](findAll.md), [Regex.compile](compile.md),
[Regex.escape](escape.md), [Regex.Pattern](types/Pattern.md),
[String.contains](../String/contains.md), [String.startsWith](../String/startsWith.md)
