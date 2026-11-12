<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Regex.flags

```ant signature
func Regex.flags(p: Regex.Pattern): string
```

Regex.flags returns the flags of p, "" or some of i, m, and s in that order.

<!-- end generated -->

## Details

- The result is sorted, whatever order the literal or [Regex.compile](compile.md) gave the flags in.
- Flag groups inside the pattern, such as `(?i)`, belong to the source, not to the flags.
- With [Regex.source](source.md), the result is everything [Regex.compile](compile.md) needs to
  build the same pattern again.

## Basic examples

The flags of a literal:

```ant
Regex.flags(/ant/i) // "i"
Regex.flags(/ant/)  // ""
```

`i` ignores case, `m` makes `^` and `$` match at line ends, and `s` lets `.` match a line break:

```ant
Regex.flags(/^\w+$/m) // "m"
Regex.flags(/a.b/s)   // "s"
```

## Scope

The flags come back in the order `i`, `m`, `s`, however they were written:

```ant
/a/si | Regex.flags() // "is"
Regex.flags(/a.b/smi) // "ims"
```

A pattern from [Regex.compile](compile.md) has the flags of its second argument, sorted the same
way, and none when it is left out:

```ant
Regex.flags(Result.unwrap(Regex.compile("a+", "mi"))) // "im"
Regex.flags(Result.unwrap(Regex.compile("a+")))       // ""
```

Called through `.`, on a pattern held in a name:

```ant
let line = /^\s*#/m
line.flags()               // "m"
line.flags().contains("m") // true
line.flags().contains("i") // false
```

## Applications

Make a case-insensitive copy of any pattern, and leave one that ignores case already as it is:

```ant
func ignoringCase(p: Regex.Pattern): Regex.Pattern {
  if p.flags().contains("i") { return p }
  Result.unwrap(Regex.compile(p.source(), "${p.flags()}i"))
}

let word = ignoringCase(/^ant$/m)
println(word)
println(word.test("bee\nANT"))
println(ignoringCase(/x/i))
```

```text
/^ant$/im
true
/x/i
```

Save a pattern as data, with its source and flags, and compile it back:

```ant
type Rule = { source: string, flags: string }

let date = /^\d+\/\d+$/m
let rule = Rule{ source: date.source(), flags: date.flags() }
println(Json.stringify(rule))
let restored = Regex.compile(rule.source, rule.flags)
println(restored == Ok{ value: date })
```

```text
{"source":"^\\d+/\\d+$","flags":"m"}
true
```

## Properties and relations

The source and the flags rebuild an equal pattern, and two patterns that differ only in the order of
their flags are equal:

```ant
let p = /a\/b/si
assert(Regex.compile(p.source(), p.flags()) == Ok{ value: p })
assert(/a/si == /a/is)
assert(Regex.flags(/a/si) == Regex.flags(/a/is))
```

The flags are the letters after the last `/` of the pattern's rendering:

```ant
String.from(/a.b/smi) // "/a.b/ims"
```

## Possible issues

A flag group in the pattern changes how it matches, but not its flags. A check of the flags alone
misses it:

```ant
let p = /(?i)ant/
p.flags()     // ""
p.test("ANT") // true
```

Compare the flags against the sorted letters. The order written in the literal is not kept:

```ant
Regex.flags(/a/si) == "si" // false
Regex.flags(/a/si) == "is" // true
```

To ask about one flag, test for its letter:

```ant
Regex.flags(/a/si).contains("i") // true
```

There is no `g` flag. A literal with one does not compile, and the message names the function to use
instead:

```ant
let every = /ant/g
// error: there is no g flag: Regex.findAll and Regex.replace work on every match
```

[Regex.compile](compile.md) is `Err` on an unknown flag, a capital letter, or a repeated one:

```ant
Regex.compile("a", "I")  // Err{ error: Error{ message: "unknown flag 'I'; the flags are i, m and s" } }
Regex.compile("a", "ii") // Err{ error: Error{ message: "the flag i is repeated" } }
```

## See also

[Regex.source](source.md) · [Regex.compile](compile.md) · [Regex.Pattern](types/Pattern.md) ·
[Regex literals](../../language/literals/regex.md)
