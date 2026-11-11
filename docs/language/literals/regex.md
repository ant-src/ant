<!-- status: reviewed -->

# Regular expression literals

A pattern between slashes, with optional flags after the closing slash, is a regular expression
literal: `/^\d+$/` and `/ant/i` are `Regex.Pattern` values, as in JavaScript.

## Details

- The pattern is RE2 syntax. `ant check` and `ant run` reject an invalid pattern before the program
  runs, so an invalid literal never reaches run time. Use `Regex.compile` for a pattern from data;
  it returns a `Result`.
- A backslash belongs to the pattern: `\d` needs no doubling. `\/` is a `/`, and so is a `/` inside
  a character class `[...]`; neither ends the literal.
- The flags are `i` (ignore case), `m` (`^` and `$` match at line ends), and `s` (`.` matches a line
  break), each at most once, in any order.
- A literal stays on one line. It cannot be empty or start with `*`, since `//` and `/*` begin
  comments; write those with `Regex.compile`.
- A pattern is immutable and compared by its source and flags, so two equal literals are `==` and a
  pattern can be a map key.

## Basic examples

A literal is a pattern you pass to the `Regex` functions:

```ant
Regex.test(/^\d{3}-\d{4}$/, "555-1234") // true
Regex.test(/^ant$/i, "ANT")             // true
Regex.test(/^ant$/, "ANT")              // false
```

A pattern renders as the literal that reads back to it:

```ant
[/colou?r/, /a\/b/i] // [/colou?r/, /a\/b/i]
```

## Scope

### Flags

`i` ignores case, `m` lets `^` and `$` match at each line, and `s` lets `.` match a line break:

```ant
let text = "first\nsecond"
Regex.test(/^SECOND/i, "second")  // true
Regex.test(/^second/, text)       // false
Regex.test(/^second/m, text)      // true
Regex.test(/first.second/, text)  // false
Regex.test(/first.second/s, text) // true
```

### Slashes and backslashes

`\/` and a `/` inside `[...]` both match a slash. Escapes such as `\d` and `\.` go into the pattern
as written:

```ant
Regex.test(/^\/usr\//, "/usr/bin") // true
Regex.test(/[/]/, "a/b")           // true
Regex.test(/\.ant$/, "main.ant")   // true
```

A backslash in the input is matched with `\\`, and a raw string holds the input without escapes:

```ant
/^C:\\temp$/.test(`C:\temp`) // true
```

### Groups

Groups, named groups, and flag groups such as `(?i)` work inside a literal:

```ant
let date = /(?<year>\d{4})-(?<month>\d\d)/.find("due 2026-11-11")
date?.named["year"]          // "2026"
date?.groups[2]              // "11"
Regex.test(/(?i)ant/, "ANT") // true
```

### Unicode

`\d`, `\w`, `\s`, and `\b` match ASCII only. `\pL` matches a letter in any script, and `\p{Greek}` a
script by name:

```ant
let text = "héllo wörld"
Regex.findAll(/\w+/, text).map({ x.text })               // ["h", "llo", "w", "rld"]
Regex.findAll(/\pL+/, text).map({ x.text })              // ["héllo", "wörld"]
Regex.test(/\bllo/, "héllo")                             // true
Regex.findAll(/\p{Greek}+/, "alpha αβγ").map({ x.text }) // ["αβγ"]
```

`é` is not a word character to `\b`, so a boundary falls inside `héllo`.

### Where a literal can start

`/` begins a literal wherever no operand comes before it: after `(`, `,`, `=`, `return`, or `{`, and
at the start of `${...}`:

```ant
func isScript(name: string): bool {
  return /\.ant$/.test(name)
}

let names = ["main.ant", "notes.txt"]
println(names.filter({ /\.txt$/.test(x) }))
println(isScript("main.ant"))
println("pattern ${/a\/b/}")
```

```text
["notes.txt"]
true
pattern /a\/b/
```

After an operand, such as a name or `)`, `/` is division:

```ant
let total = 10
let count = 2
total / count // 5
```

### Patterns as values

A pattern is a value like any other. Annotate it as `Regex.Pattern`, keep it in a collection, or use
it as a map key:

```ant
let rules: Map[Regex.Pattern, string] = Map{ /^\d+$/: "number", /^\w+$/: "word" }
rules[/^\d+$/] // "number"
```

## Applications

### Validate input

```ant
func checkUsername(name: string): Result[string, Error] {
  if !/^[a-z][a-z0-9_]{2,15}$/.test(name) {
    return Result.fail("username ${name}: use 3 to 16 lowercase letters, digits, or _")
  }
  Result.ok(name)
}

println(checkUsername("ant_lang"))
println(checkUsername("9lives"))
```

```text
Ok{ value: "ant_lang" }
Err{ error: Error{ message: "username 9lives: use 3 to 16 lowercase letters, digits, or _" } }
```

### Rewrite with groups

`$1` and `$2` in a replacement refer to the groups:

```ant
/(\w+)@(\w+)/.replace("ann@home bob@work", "$2:$1") // "home:ann work:bob"
```

### Split on a pattern

```ant
Regex.split(/\s*,\s*/, "red , green,blue") // ["red", "green", "blue"]
Regex.split(/\d+/, "a1b22c")               // ["a", "b", "c"]
```

### Pull out every match

```ant
let text = "Call 555-1234 or 555-9876"
Regex.findAll(/\d{3}-\d{4}/, text).map({ x.text }) // ["555-1234", "555-9876"]
```

## Properties and relations

A literal and `Regex.compile` with the same source and flags build equal patterns. Flags compare in
any order, and `\/` in a literal is a plain `/` in the source:

```ant
assert(Regex.compile("a+", "i") == Ok{ value: /a+/i })
assert(/a/si == /a/is)
assert(Regex.compile("a/b") == Ok{ value: /a\/b/ })
assert(/a\/b/i.source() == "a/b")
assert(/a/si.flags() == "is")
```

## Possible issues

### An invalid pattern is a compile error

```ant
let open = /(/ // error: invalid regular expression: missing closing ): `(`
```

The same pattern from data is an `Err`, not a compile error:

```ant
Regex.compile("(") // Err{ error: Error{ message: "invalid regular expression: missing closing ): `(`" } }
```

### No lookahead or backreferences

RE2 has no lookahead, lookbehind, or backreferences, so a match takes time linear in the input:

```ant
let price = /\d+(?=€)/
// error: lookahead is not supported: a pattern is RE2, which has no lookahead, lookbehind or backreferences
```

### No `g`, `y`, `u`, `v`, or `d` flags

Ant has no `g` flag. `Regex.findAll` and `Regex.replace` work on every match:

```ant
let vowels = /[aeiou]/g // error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

```ant
/[aeiou]/.replace("banana", "_") // "b_n_n_"
```

JavaScript's other flags are errors too: for `y`, anchor with `^` on a `String.slice` of the input;
`u`, `v`, and `d` are how Ant already behaves, so drop them:

```ant
let word = /\w+/u // error: there is no u flag: a pattern is Unicode and a match carries its indices already
```

### A line that starts with a literal

`/` after an operand is division, even when a line break comes between them. A statement that starts
with a literal on the line after a name or a value is a division with no left operand:

```ant
let word = "ant"
/a/.test(word) // error: unexpected '/'
```

Bind the result, or write the call in namespace form, so the line starts with a name:

```ant
let word = "ant"
let found = /a/.test(word)
println(found)
println(Regex.test(/n/, word))
```

```text
true
true
```

### An empty pattern

`//` begins a comment, so an empty pattern is written with `Regex.compile`:

```ant
Regex.compile("") // Ok{ value: /(?:)/ }
```

### Ignoring case is not case folding

`i` compares one code point at a time, so `ß` does not match `SS`. Compare such text with
`String.equalFold`:

```ant
Regex.test(/straße/i, "STRASSE") // false
"straße".equalFold("STRASSE")    // true
```

## See also

[Regex](../../lib/Regex/index.md) · [Regex.Pattern](../../lib/Regex/types/Pattern.md) ·
[Regex.compile](../../lib/Regex/compile.md) · [Regex.test](../../lib/Regex/test.md) ·
[Regex.find](../../lib/Regex/find.md) · [Regex.replace](../../lib/Regex/replace.md) ·
[Raw strings](raw-strings.md) · [String literals](strings.md) ·
[Arithmetic operators](../operators/arithmetic.md)
