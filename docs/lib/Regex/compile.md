<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.compile

```ant signature
func Regex.compile(source: string, flags: string = ""): Result[Regex.Pattern, Error]
```

Regex.compile compiles source, in RE2 syntax, with flags, some of i, m, and s in any order. It
returns Err on bad syntax, an unknown flag, or a repeated one. flags defaults to "".

<!-- end generated -->

## Details

- Use a literal such as `/\d+/i` for a pattern fixed in the source; `Regex.compile` is for a pattern
  that comes from data, such as user input or a configuration file.
- Returns `Err` instead of panicking, for bad syntax and for lookahead, lookbehind, and
  backreferences, which RE2 does not have.
- The result is the same pattern a literal with that source and those flags gives, and compares `==`
  to it. A `/` needs no escape in `source`.
- `//` begins a comment, so the empty pattern has no literal; write `Regex.compile("")`.

## Basic examples

Compile a pattern from a string:

```ant
Regex.compile("a+b") // Ok{ value: /a+b/ }
```

Pass flags as the second argument:

```ant
Regex.compile("ant", "i") // Ok{ value: /ant/i }
```

Bad syntax gives an `Err`, not a panic:

```ant
Regex.compile("(ab") // Err{ error: Error{ message: "invalid regular expression: missing closing ): `(ab`" } }
```

## Scope

Write a source with backslashes as a raw string, so each backslash reaches the pattern as written:

```ant
Regex.compile(`^\d{4}-\d{2}$`) // Ok{ value: /^\d{4}-\d{2}$/ }
```

The empty source compiles to the empty pattern, which renders as `/(?:)/`:

```ant
Regex.compile("") // Ok{ value: /(?:)/ }
```

A `/` and a line break are ordinary characters in `source`. The pattern renders them escaped, so the
rendering reads back as a literal:

```ant
Regex.compile("a/b")  // Ok{ value: /a\/b/ }
Regex.compile("a\nb") // Ok{ value: /a\nb/ }
```

Named groups, inline flag groups, and Unicode classes are all RE2 syntax:

```ant
Regex.compile(`(?<year>\d{4})`) // Ok{ value: /(?<year>\d{4})/ }
Regex.compile(`(?i)ant`)        // Ok{ value: /(?i)ant/ }
Regex.compile(`\p{Greek}+`)     // Ok{ value: /\p{Greek}+/ }
```

Lookahead, lookbehind, and backreferences are not RE2, so they give an `Err`:

```ant
for source in ["a(?=b)", `(a)\1`] {
  let r = Regex.compile(source)
  if r is Err {
    println(r.error.message)
  }
}
```

```text
lookahead is not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences
backreferences are not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences
```

Call it in a pipe, and work on the pattern inside the `Result` with `Result.map`:

```ant
`\pL+` | Regex.compile() | Result.map({ x.test("λόγος") }) // Ok{ value: true }
```

## Named arguments

`flags` is a string of `i`, `m`, and `s`, in any order: `i` ignores case, `m` makes `^` and `$`
match at line ends, and `s` lets `.` match a line break. The pattern keeps them sorted:

```ant
Regex.compile("^a.", flags: "si") // Ok{ value: /^a./is }
Regex.compile("^a.", flags: "m")  // Ok{ value: /^a./m }
```

With `m`, `^` matches after a line break too:

```ant
Regex.compile("^b", "m") | Result.map({ x.test("a\nb") }) // Ok{ value: true }
```

An unknown flag or a repeated one gives an `Err`:

```ant
for flags in ["x", "ii", "g"] {
  let r = Regex.compile("a", flags: flags)
  if r is Err {
    println(r.error.message)
  }
}
```

```text
unknown flag 'x'; the flags are i, m and s
the flag i is repeated
there is no g flag: Regex.findAll and Regex.replace work on every match
```

## Applications

Compile a pattern the user passes on the command line, and report a bad one with `Result.wrapErr`:

<!-- test: args ^b
-->

```ant
func grep(lines: Array[string], source: string): Result[Array[string], Error] {
  let p = try (Regex.compile(source) | Result.wrapErr("pattern ${source}"))
  Result.ok(lines.filter({ p.test(x) }))
}

let lines = ["apple", "banana", "blueberry", "cherry"]
for line in try grep(lines, Os.args()[0]) {
  println(line)
}
```

```text
banana
blueberry
```

Search for text the user typed, ignoring case. `Regex.escape` makes every character match itself, so
the source is always valid and `Result.unwrap` cannot panic:

```ant
func countMentions(text: string, query: string): int {
  let p = query | Regex.escape() | Regex.compile("i") | Result.unwrap()
  p.findAll(text).length()
}

println(countMentions("C++ and c++ and C", "c++"))
```

```text
2
```

Build one pattern that matches any word of a list exactly:

```ant
let words = ["c++", "f#", "go"]
let alternatives = words | Array.map(Regex.escape) | String.join("|")
let lang = Regex.compile("^(?:${alternatives})$") | Result.unwrap()
println(lang)
println(["go", "f#", "rust"].map({ lang.test(x) }))
```

```text
/^(?:c\+\+|f#|go)$/
[true, true, false]
```

## Properties and relations

`Regex.compile` gives the pattern a literal with the same source and flags gives:

```ant
assert(Regex.compile(`\d+`, "i") == Ok{ value: /\d+/i })
assert(Regex.compile("a/b") == Ok{ value: /a\/b/ })
```

The order of the flags does not matter:

```ant
assert(Regex.compile("a", "si") == Regex.compile("a", "is"))
```

[Regex.source](source.md) and [Regex.flags](flags.md) give back what compiles to the same pattern:

```ant
let p = /a\/b+$/im
assert(Regex.compile(p.source(), p.flags()) == Ok{ value: p })
```

A pattern compiled from [Regex.escape](escape.md) of a string matches that string:

```ant
let s = "1+1=2 (really?)"
let p = Regex.compile("^${Regex.escape(s)}$") | Result.unwrap()
assert(p.test(s))
```

## Possible issues

In a string, `\d` is not an escape Ant knows, so it is a compile error:

```ant
let digits = Regex.compile("\d+") // error: bad escape '\d'
```

Write the source as a raw string, `` `\d+` ``, or double each backslash, `"\\d+"`:

```ant
Regex.compile(`\d+`)  // Ok{ value: /\d+/ }
Regex.compile("\\d+") // Ok{ value: /\d+/ }
```

`Result.unwrap` panics on an `Err`, so unwrap only a source you know is valid. For data, use `try`
or check the `Result`:

```ant
Regex.compile("[a-") | Result.unwrap()
// panic: Result.unwrap on an Err: Error{ message: "invalid regular expression: missing closing ]: `[a-`" }
```

A source fixed in the program is better written as a literal: `ant check` reports its syntax error,
and there is no `Result` to unwrap:

```ant
let p = /(ab/ // error: missing closing )
```

## See also

[Regex.Pattern](types/Pattern.md) · [Regex.escape](escape.md) · [Regex.source](source.md) ·
[Regex.flags](flags.md) · [Regex.test](test.md) ·
[regular expression literals](../../language/literals/regex.md) ·
[raw strings](../../language/literals/raw-strings.md) · [Result](../Result/index.md)
