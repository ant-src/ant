<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex

Regex holds the functions of regular expressions, written as a literal /.../flags or compiled from
data with [`Regex.compile`](compile.md). The syntax is RE2, so a match takes time linear in the
input; indices into the searched string count code points.

<!-- end generated -->

## Details

- Write a pattern fixed in the source as a literal, such as `/\d+/i`; `ant check` rejects a bad one.
  Compile a pattern from data with [Regex.compile](compile.md), which returns a `Result`.
- RE2 has no lookahead, lookbehind, or backreferences.
- `\d`, `\w`, `\s`, and `\b` match ASCII only. Use `\pL`, `\pN`, or `\p{Greek}` for Unicode classes.
- A pattern is immutable and compares by its source and flags, so it can be a map key.

## Basic examples

`Regex.test` asks whether a pattern matches anywhere in a string:

```ant
Regex.test(/\d+/, "room 42") // true
Regex.test(/^\d+$/, "42a")   // false
```

`Regex.find` returns the leftmost match, with its groups, or `null`:

```ant
let email = /(?<user>\w+)@(?<host>[\w.]+)/
Regex.find(email, "to ann@example.com")?.named // Map{ "user": "ann", "host": "example.com" }
Regex.find(email, "no address")                // null
```

`Regex.replace` replaces every match, and `$n` names a group:

```ant
Regex.replace(/\s+/, "a   b \t c", " ")                    // "a b c"
Regex.replace(/(\w+)@(\w+)/, "ann@home bob@work", "$2:$1") // "home:ann work:bob"
```

`Regex.compile` builds a pattern from a string and returns a `Result`:

```ant
Regex.compile("a+", "i") // Ok{ value: /a+/i }
Regex.compile("[a-z")    // Err{ error: Error{ message: "invalid regular expression: missing closing ]: `[a-z`" } }
```

## Scope

### Flags

The flags `i`, `m`, and `s` follow the closing slash, in any order. `i` ignores case, `m` makes `^`
and `$` match at line ends, and `s` lets `.` match a line break:

```ant
Regex.test(/ant/i, "ANT")  // true
Regex.test(/^b/, "a\nb")   // false
Regex.test(/^b/m, "a\nb")  // true
Regex.test(/a.b/, "a\nb")  // false
Regex.test(/a.b/s, "a\nb") // true
Regex.flags(/ant/si)       // "is"
```

### Groups

`groups[0]` is the whole match and `groups[n]` the `n`th group, `null` when the group took no part.
`named` holds the named groups:

```ant
let m = Regex.find(/(\d+)(px|em)?/, "width 12")
m?.groups // ["12", "12", null]
```

### Unicode

The ASCII classes skip letters and digits outside ASCII. The Unicode classes include them, and the
indices count code points:

```ant
Regex.findAll(/\w+/, "héllo") | Array.map({ x.text })  // ["h", "llo"]
Regex.findAll(/\pL+/, "héllo") | Array.map({ x.text }) // ["héllo"]
Regex.test(/\d/, "٣")                                  // false
Regex.test(/\pN/, "٣")                                 // true
Regex.find(/b/, "→b")?.start                           // 1
```

### Dot calls and pipes

A pattern takes the dot form, since `Regex` is its namespace. The pattern is the first argument, so
a pipe carries a pattern into a `Regex` function, not the string it searches; pipe the results
onward instead:

```ant
let word = /[a-z]+/
word.test("42 apples")                                // true
word.findAll("one two three") | Array.map({ x.text }) // ["one", "two", "three"]
```

### Patterns as values

A pattern compares by value and serves as a map key:

```ant
let kinds = Map{ /^\d+$/: "number", /^[a-z]+$/: "word" }
let digit = /[0-9]/
kinds[/^\d+$/]   // "number"
digit == /[0-9]/ // true
digit == /\d/    // false
```

`is Regex.Pattern` tells a pattern from the other members of a union:

```ant
func describe(rule: Regex.Pattern | string): string {
  if rule is Regex.Pattern { "pattern ${rule}" } else { "text ${rule}" }
}

println(describe(/^a+$/i))
println(describe("abc"))
```

```text
pattern /^a+$/i
text abc
```

## Functions by job

### Building a pattern

- [Regex.compile](compile.md): a pattern from a source string and flags, or an `Err`
- [Regex.escape](escape.md): a string with every metacharacter escaped, to match it literally

### Matching

- [Regex.test](test.md): whether the pattern matches anywhere
- [Regex.find](find.md): the leftmost match, or `null`
- [Regex.findAll](findAll.md): every match, left to right, without overlap

### Replacing

- [Regex.replace](replace.md): every match replaced, with `$n` and `${name}` expanded
- [Regex.replaceFirst](replaceFirst.md): the first match replaced, expanded the same way
- [Regex.replaceWith](replaceWith.md): every match replaced by what a function returns for it

### Splitting

- [Regex.split](split.md): the pieces of a string between the matches

### Reading a pattern

- [Regex.source](source.md): the normalized source
- [Regex.flags](flags.md): the flags, sorted

### Types

- [Regex.Pattern](types/Pattern.md): a compiled regular expression
- [Regex.Match](types/Match.md): one match, with its text, indices, and groups

## Applications

Parse log lines into records with named groups:

```ant
type LogLine = { level: string, message: string }

let line = /^\[(?<level>[A-Z]+)\] (?<message>.*)$/

func parse(text: string): LogLine | null {
  let m = line.find(text)
  if m == null {
    return null
  }
  LogLine{ level: m.named["level"] ?? "", message: m.named["message"] ?? "" }
}

println(parse("[WARN] disk almost full"))
println(parse("garbage"))
```

```text
LogLine{ level: "WARN", message: "disk almost full" }
null
```

Validate a whole input by anchoring the pattern with `^` and `$`:

```ant
let isoDate = /^\d{4}-\d{2}-\d{2}$/
["2026-11-12", "12/11/2026", "2026-11-12x"].map({ isoDate.test(x) }) // [true, false, false]
```

Mask every number but its last two digits:

```ant
Regex.replaceWith(/\d+/, "card 4111111111111111, pin 1234", {
  let n = x.text.length()
  "*".repeat(n - 2) + x.text.slice(n - 2, n)
}) // "card **************11, pin **34"
```

Split a list whose separators vary:

```ant
Regex.split(/\s*[,;]\s*/, "red, green ;blue,  gray") // ["red", "green", "blue", "gray"]
```

Turn a title into a URL slug: replace each run of other characters with a dash, then drop the dashes
at the ends:

```ant
let title = "  Hello, World: Ant 1.0!  "
let dashed = Regex.replace(/[^a-z0-9]+/, title.toLower(), "-")
Regex.replace(/^-|-$/, dashed, "") // "hello-world-ant-1-0"
```

## Properties and relations

`Regex.find` gives `null` exactly when `Regex.test` is `false`, and a match's indices slice out its
text:

```ant
let p = /o+/
for s in ["foo", "bar", "bόo"] {
  let m = p.find(s)
  assert((m != null) == p.test(s))
  if m != null {
    assert(s.slice(m.start, m.end) == m.text)
  }
}
```

A pattern compiled from `Regex.escape(s)` matches `s` itself:

```ant
for s in ["a.b", "1+1=2", "(x)"] {
  let p = s | Regex.escape() | Regex.compile() | Result.unwrap()
  assert(p.test(s))
  assert(Regex.replace(p, s, "") == "")
}
```

A literal and `Regex.compile` with the same source and flags give the same pattern:

```ant
assert(Regex.compile(`\d+`, "i") == Ok{ value: /\d+/i })
```

## Possible issues

`Regex` is a namespace, not a type. Annotate with `Regex.Pattern`:

```ant
func hasDigit(p: Regex, s: string): bool { p.test(s) } // error: Regex is a namespace, not a type
```

There is no `g` flag. `Regex.findAll`, `Regex.replace`, and `Regex.split` already work on every
match:

```ant
let all = /a/g // error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

RE2 has no lookahead, lookbehind, or backreferences, and a literal with one does not compile:

```ant
let p = /\d+(?=px)/
// error: lookahead is not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences
```

A line that starts with a literal continues the line before it when that line ends in an operand,
such as `)`, and the `/` reads as division. Start the line with `Regex.test(...)` or name the
pattern with `let`:

```ant
let s = "a\nb"
println(s.length())
/^b/m.test(s) // error: unexpected character '^'
```

`Regex.compile` returns a `Result`, so handle the `Err` before matching:

```ant
let r = Regex.compile("a+")
let ok = r.test("aa") // error: Result[Regex.Pattern, Error] has no field or function test
```

Unwrap it with `try` in a function that returns a `Result`:

```ant
func matches(source: string, s: string): Result[bool, Error] {
  let p = try Regex.compile(source)
  Ok{ value: p.test(s) }
}

println(matches("a+", "caat"))
println(matches("[a-z", "caat"))
```

```text
Ok{ value: true }
Err{ error: Error{ message: "invalid regular expression: missing closing ]: `[a-z`" } }
```

`//` begins a comment, so the empty pattern has no literal. Write `Regex.compile("")`:

```ant
Regex.compile("") // Ok{ value: /(?:)/ }
```

In a string, `${1}` interpolates the number 1 instead of naming group 1. Use a raw string or `\${`:

```ant
Regex.replace(/(\w+)/, "ann", "${1}x") // "1x"
Regex.replace(/(\w+)/, "ann", `${1}x`) // "annx"
```

`$1x` names a group `1x`, which does not exist, so it expands to an empty string. Write `${1}x` in a
raw string to put text after a group:

```ant
Regex.replace(/(\w+)/, "ann", "$1x") // ""
```

`i` folds case one code point at a time, so it does not match `ß` against `SS`. Compare whole
strings with `String.equalFold` instead:

```ant
Regex.test(/straße/i, "STRASSE") // false
"straße".equalFold("STRASSE")    // true
```

## See also

- [Regular expression literals](../../language/literals/regex.md): the `/.../flags` syntax
- [String](../String/index.md): `String.contains`, `String.replace`, and `String.split` for fixed
  text, with no pattern
- [Text guide](../guides/text.md): working with strings across namespaces
