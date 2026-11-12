<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.source

```ant signature
func Regex.source(p: Regex.Pattern): string
```

Regex.source returns the normalized source of p, in which \\/ is written /.

<!-- end generated -->

## Details

- The source holds the pattern without the slashes around it and without its flags; read the flags
  with [Regex.flags](flags.md).
- The source is normalized: outside a `\Q...\E` span, `\/` loses its backslash and a line break
  becomes `\n` or `\r`; an empty pattern becomes `(?:)`.
- Two patterns are `==` exactly when their sources and their flags are.

## Basic examples

Read the source of a literal:

```ant
Regex.source(/ant/)  // "ant"
Regex.source(/\d+/)  // "\\d+"
Regex.source(/a\/b/) // "a/b"
```

The flags are not part of the source:

```ant
Regex.source(/hello/i) // "hello"
Regex.flags(/hello/i)  // "i"
```

## Scope

A pattern from `Regex.compile` has a source like a literal's. A `/` needs no escape in the string:

```ant
Result.unwrap(Regex.compile("a/b")).source() // "a/b"
```

No literal can write an empty pattern, so build it with `Regex.compile`. Its source is `(?:)`, the
same as the source of `/(?:)/`:

```ant
Result.unwrap(Regex.compile("")).source() // "(?:)"
```

A line break in the compiled string becomes the escape `\n`, so the source stays on one line:

```ant
Result.unwrap(Regex.compile("a\nb")).source() // "a\\nb"
```

A `/` inside a character class stays as written:

```ant
/[a/b]/.source() // "[a/b]"
```

So does a flag group such as `(?i)`:

```ant
/(?i)ant/.source() // "(?i)ant"
```

The text of a `\Q...\E` span is taken as itself, so a `\/` inside it keeps its backslash:

```ant
/\Qa\/b\E/.source() // "\\Qa\\/b\\E"
```

Call it through `.` or in a pipe:

```ant
let word = /\w+/
word.source()         // "\\w+"
word | Regex.source() // "\\w+"
```

## Applications

Wrap a pattern so that it matches only whole words, keeping its flags:

```ant
func wholeWord(p: Regex.Pattern): Regex.Pattern {
  Result.unwrap(Regex.compile("\\b(?:${p.source()})\\b", p.flags()))
}

let cat = wholeWord(/cat/i)
cat.test("a Cat sat")   // true
cat.test("concatenate") // false
```

Combine several patterns into one alternation:

```ant
let keywords = [/if/, /else/, /for/] | Array.map(Regex.source) | String.join("|")
keywords                           // "if|else|for"
Regex.compile("^(?:${keywords})$") // Ok{ value: /^(?:if|else|for)$/ }
```

Store a pattern as data, with its source and flags in separate fields, and build it again later:

```ant
type SavedRule = { name: string, source: string, flags: string }

let rule = /^[a-z]+\/\d+$/i
let saved = SavedRule{ name: "id", source: rule.source(), flags: rule.flags() }
println(Json.stringify(saved))
let restored = Result.unwrap(Regex.compile(saved.source, saved.flags))
println(restored == rule)
```

```text
{"name":"id","source":"^[a-z]+/\\d+$","flags":"i"}
true
```

## Properties and relations

Compiling the source with the flags gives back an equal pattern:

```ant
let p = /^\d{4}-\d{2}$/m
assert(Regex.compile(p.source(), p.flags()) == Ok{ value: p })
```

Patterns are equal when their sources and flags are, so `/a\/b/` and a compiled `a/b` are one
pattern:

```ant
let p = /a\/b/
let q = Result.unwrap(Regex.compile("a/b"))
assert(p.source() == q.source())
assert(p == q)
```

A pattern renders as a literal, with slashes and flags, while its source has neither:

```ant
let p = /a+/i
"${p}"     // "/a+/i"
p.source() // "a+"
```

## Possible issues

The source alone drops the flags. A pattern compiled from it no longer ignores case:

```ant
let p = /ant/i
let q = Result.unwrap(Regex.compile(p.source()))
p.test("ANT") // true
q.test("ANT") // false
```

A backslash in the source appears doubled when the string is shown, because the REPL shows a string
as a string literal. The source itself holds one backslash:

```ant
let src = /\d/.source()
src          // "\\d"
src.length() // 2
```

## See also

[Regex.flags](flags.md), [Regex.compile](compile.md), [Regex.escape](escape.md),
[Regex.Pattern](types/Pattern.md), [Regex](index.md)
