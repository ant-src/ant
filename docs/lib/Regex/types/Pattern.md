<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.Pattern

```ant signature
type Regex.Pattern
```

Regex.Pattern is a compiled regular expression, from a literal /.../flags or
[`Regex.compile`](../compile.md). It is immutable and compares by its normalized source and flags,
so it serves as a map key.

<!-- end generated -->

## Details

- A pattern comes only from a literal, which `ant check` validates, or from
  [Regex.compile](../compile.md), which returns a `Result`. It has no constructor and no fields;
  read its parts with [Regex.source](../source.md) and [Regex.flags](../flags.md).
- Two patterns are `==` when their normalized sources and their flags are. The flags are compared as
  a set, but sources that match the same text, such as `[0-9]` and `\d`, are still different
  patterns.
- A pattern renders as a literal that reads back to it, as `/a\/b/i`. The empty pattern renders as
  `/(?:)/`.

## Basic examples

A literal is a `Regex.Pattern`, and so is the `Ok` value of `Regex.compile`:

```ant
let digits: Regex.Pattern = /\d+/
digits                 // /\d+/
Regex.compile(`\d+`)   // Ok{ value: /\d+/ }
digits.test("room 42") // true
```

Annotate a parameter with `Regex.Pattern` to take a pattern:

```ant
func count(p: Regex.Pattern, s: string): int {
  p.findAll(s).length()
}

println(count(/o/, "foo boo"))
println(count(/\bb/i, "Bob builds boats"))
```

```text
4
3
```

## Scope

### Flags

The flags are part of the pattern. Their order does not matter:

```ant
let ant = /ant/is
ant                        // /ant/is
ant == /ant/si             // true
ant == /ant/i              // false
Regex.compile("ant", "si") // Ok{ value: /ant/is }
```

### Rendering

`println`, interpolation, and `String.from` show the pattern as a literal, slashes and flags
included. A `/` in the source is written `\/`, and a line break `\n`:

```ant
"${/a+/m}"                          // "/a+/m"
String.from(/a\/b/i)                // "/a\\/b/i"
Result.unwrap(Regex.compile("a/b")) // /a\/b/
Regex.compile("a\nb")               // Ok{ value: /a\nb/ }
Regex.compile("")                   // Ok{ value: /(?:)/ }
```

`Json.stringify` writes that rendering as a JSON string:

```ant
Json.stringify(/a\/b/i)          // "\"/a\\\\/b/i\""
Json.stringify(Map{ id: /\d+/ }) // "{\"id\":\"/\\\\d+/\"}"
```

### In collections and records

A pattern serves as a map key and a set member. Equal patterns are one key:

```ant
let hits = Map{ /a/: 1 }
hits[/a/] = 2
hits                      // Map{ /a/: 2 }
Set.of([/a/i, /a/i, /a/]) // Set.of([/a/i, /a/])
```

A record field can hold a pattern:

```ant
type Rule = { name: string, pattern: Regex.Pattern }

let rule = Rule{ name: "number", pattern: /^\d+$/ }
rule                      // Rule{ name: "number", pattern: /^\d+$/ }
rule.pattern.test("2026") // true
```

### Unions and `is`

`Regex.Pattern` has a tag, so a union can hold it beside other kinds and `is` tells them apart:

```ant
func matches(want: Regex.Pattern | string, s: string): bool {
  if want is Regex.Pattern { want.test(s) } else { want == s }
}

println(matches(/^v\d+$/, "v12"))
println(matches("v12", "v12"))
println(matches("v1", "v12"))
```

```text
true
true
false
```

## Applications

Classify input by the first rule whose pattern matches it:

```ant
type Rule = { kind: string, pattern: Regex.Pattern }

let rules = [
  Rule{ kind: "number", pattern: /^\d+$/ },
  Rule{ kind: "word", pattern: /^[a-z]+$/i },
  Rule{ kind: "space", pattern: /^\s+$/ },
]

func classify(s: string): string {
  rules.find({ x.pattern.test(s) })?.kind ?? "other"
}

for s in ["42", "Ant", "  ", "a1"] {
  println("${s}: ${classify(s)}")
}
```

```text
42: number
Ant: word
  : space
a1: other
```

Compile a user's pattern once, and report a bad one as an error:

```ant
func grep(source: string, lines: Array[string]): Result[Array[string], Error] {
  let p = try Regex.compile(source).wrapErr("bad pattern ${source}")
  Result.ok(lines.filter({ p.test(x) }))
}

let lines = ["error: disk full", "ok", "error: timeout"]
println(grep("^error", lines))
println(grep("(", lines) is Err)
```

```text
Ok{ value: ["error: disk full", "error: timeout"] }
true
```

Count matches per pattern in a map keyed by the pattern:

```ant
let patterns = [/\d/, /[aeiou]/, /\s/]
let text = "ant 1.0 runs"
let counts = Map.fromEntries(patterns.map({ Entry{ key: x, value: x.findAll(text).length() } }))
for { key, value } in counts {
  println("${key} ${value}")
}
```

```text
/\d/ 2
/[aeiou]/ 2
/\s/ 2
```

## Properties and relations

A pattern's source and flags compile back to an equal pattern:

```ant
for p in [/a\/b/i, /^\d+$/m, /x.y/s, Result.unwrap(Regex.compile(""))] {
  assert(Regex.compile(p.source(), p.flags()) == Ok{ value: p })
}
```

A `\/` in a literal and a `/` in a compiled string give one pattern:

```ant
assert(/a\/b/ == Result.unwrap(Regex.compile("a/b")))
assert(Regex.source(/a\/b/) == "a/b")
```

## Possible issues

Equality compares the sources, not what the patterns match. Two patterns that match the same text
can differ:

```ant
let digit = /[0-9]/
digit == /\d/                                   // false
Regex.compile("(?i)ant") == Ok{ value: /ant/i } // false
```

`Regex.compile` returns a `Result`, not a pattern. Unwrap it, or use `try`:

```ant
let p: Regex.Pattern = Regex.compile("a+")
// error: cannot use Result[Regex.Pattern, Error] where Regex.Pattern is expected
```

A string is not a pattern, even one that looks like one:

```ant
let p: Regex.Pattern = "a+" // error: cannot use string where Regex.Pattern is expected
```

`Regex` is a namespace, not a type. The type is `Regex.Pattern`:

```ant
let p: Regex = /a/ // error: Regex is a namespace, not a type
```

A pattern has no constructor and no fields:

```ant
let p = Regex.Pattern("a+") // error: Regex.Pattern is a type, not a function
let q = /a+/
println(q.source) // error: Regex.Pattern is not a record, so it has no field source
```

An empty pattern cannot be a literal, since `//` begins a comment; write `Regex.compile("")`. A
pattern that starts with `*` cannot be a literal either, since `/*` begins one:

```ant
Regex.compile("") // Ok{ value: /(?:)/ }
```

There is no `g` flag. `Regex.findAll` and `Regex.replace` act on every match:

```ant
let p = /a/g // error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

```ant
let p = /a/
p.findAll("banana").length() // 3
p.replace("banana", "o")     // "bonono"
```

`Json.parse` gives back the rendering as a string, not a pattern. To restore a pattern from JSON,
store its source and its flags as separate fields and pass them to `Regex.compile`:

```ant
let text = Json.stringify(/a+/i)
Json.parse(text) // Ok{ value: "/a+/i" }
```

## See also

[Regex.compile](../compile.md) · [Regex.source](../source.md) · [Regex.flags](../flags.md) ·
[Regex.test](../test.md) · [Regex.Match](Match.md) · [Regex](../index.md) ·
[Regular expression literals](../../../language/literals/regex.md)
