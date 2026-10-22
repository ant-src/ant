# Plan 5: the everyday library

Plan 3 gave Ant the world outside: processes, streams, paths, time. What is still missing is the
small change every program reaches for in its first hour. A total cannot be printed to two
decimals, a hex string cannot be parsed, a line cannot be matched against a pattern, an element
cannot be removed from the middle of an array, a log cannot be taken. A response cannot be
fetched with `PUT`, a token cannot be base64-decoded, a webhook cannot be checked with an HMAC.

This plan takes these from Go's standard library, as Plan 3 did: each builtin is a thin wrapper
over a Go package, and the Go function's shape is the best evidence of what a program needs. The
spec comes first and the implementation second.

The servers and the waiting are out of scope: an HTTP server, sockets, timeouts on a blocking
call, cancelling a goroutine. They depend on one design question, how Ant waits for one of
several things without channels, and that question gets a plan of its own (section 3.9).

## 1. Where we start

Section 11 of `ant-lang.md` is the core library: `Array`, `Map`, `String`, `Iter`, `Math`, `Bits`,
`Result`, `Record` and the sync types. `stdlib.md` has `File`, `Http`, `Path`, `Time`, `Os` and
`Json`. Its rule is that an operation that can fail returns `Result[_, Error]` and nothing panics
on bad input. The exceptions are `Json.stringify` on a cycle and `Time`, which is computation and
panics where the core library would.

The gaps, by what goes wrong in a program:

- **Numbers to text and back.** `String.from` is the only rendering. It gives no fixed number of
  decimals, no base other than ten and no padding. `String.toInt` reads decimal only.
- **Strings.** It lacks `lastIndexOf`, one-sided trims, a single replacement, splitting on white
  space or on lines, splitting at the first separator, a count, and any test of what kind of
  character a string holds.
- **Patterns.** There are no regular expressions.
- **Arrays.** Only the ends can change, through `push` and `pop`. There is no way to find an
  index by predicate, or to take the minimum, unique elements, chunks or groups.
- **Math.** There are no logarithms, exponentials or trigonometry, no test for NaN or infinity, no
  integer random and no clamp. `Bits` has no leading or trailing zero count.
- **Encodings and digests.** There is no base64, hex, URL escaping, CSV, SHA-2, HMAC or secure
  random bytes.
- **HTTP.** There are only `get` and `post`: no other method, no headers on `get`, no timeout.

In the implementation, each namespace is a record of `*interp.Builtin` values in `internal/lib`,
declared in `internal/prelude/prelude.ant`. The checker types a call from that declaration.
Defaults on builtins exist since Phase 22, through `Builtin.Optional`. Opaque library types exist
in two forms: `File.Handle` as an object compared by identity, and the `time` kind compared by
value. `TestLibraryCoverage` requires a corpus program for every prelude function.

## 2. Definition of done

- `ant-lang.md` section 11 and `stdlib.md` list every new function with its signature, its
  comment and its ◆ where it has one. The new namespaces are named in section 11's list.
- The prelude declares each function, a builtin implements each, and `TestLibraryCoverage` and
  `TestPanicCoverage` have no pending entry. Every new function is used by a program under
  `tests/run`, and every ◆ is reached by one.
- The regex literal is in section 2 of `ant-lang.md`, and every lexing case of section 4.4 has a
  program under `tests/syntax` or `tests/run`.
- Two new examples: one that uses patterns and formatting on text, one that calls an HTTP API
  with headers, a method and a signature.
- The suite is green under `-race` on both platforms, and the coverage gate stays at 90% overall
  and 80% per package.
- `CHANGELOG.md`, the README and the package documentation in `ant.go` list the additions.

## 3. What Go provides and what Ant takes

Each signature is proposed spec text, with the Go function behind it so the mapping can be
checked. The semantics follow Go's unless the comment says otherwise. As in the rest of the
library, indices and lengths count code points, never bytes.

### 3.1 Numbers to text and back (core, `String`)

```
String.fromInt[T: Integer](x: T, base: int = 10): string        // FormatInt, FormatUint for unsigned; lower-case digits; ◆ base outside 2 ..= 36
String.fromFloat(x: float64, digits: int): string                // strconv.FormatFloat(x, 'f', digits): fixed point, half to even; ◆ digits < 0
String.toInt(s: string, base: int = 10): int | null              // strconv.ParseInt; no prefix, optional sign; ◆ base outside 2 ..= 36
String.padStart(s: string, width: int, pad: string = " "): string   // repeats pad on the left to width code points; ◆ pad is empty
String.padEnd(s: string, width: int, pad: string = " "): string
```

`String.fromInt` formats a signed value with `strconv.FormatInt` and an unsigned one with
`strconv.FormatUint`, so the largest `uint64` renders as `18446744073709551615` and not as `-1`.
The corpus covers the edges: `Math.minInt` in bases 2, 10 and 36, the largest `uint64` in base
16, and `int8(-128)` in base 2, which renders as `-10000000`. `String.toInt` gains its `base` as
a default, so existing calls are unchanged. `fromFloat` of
NaN or an infinity renders as `String.from` does. With these, `${String.fromFloat(total, 2)}` and
`String.padStart(String.fromInt(n, 16), 8, "0")` cover what `%.2f` and `%08x` do. A format
string is left out (section 3.9).

### 3.2 Strings (core, `String`)

```
type Cut = { before: string, after: string }

String.lastIndexOf(s: string, t: string): int | null    // strings.LastIndex, in code points
String.trimStart(s: string): string                     // strings.TrimLeftFunc(unicode.IsSpace)
String.trimEnd(s: string): string
String.trimPrefix(s: string, prefix: string): string    // strings.TrimPrefix: once, when present
String.trimSuffix(s: string, suffix: string): string
String.replaceFirst(s: string, old: string, new: string): string   // strings.Replace(n = 1); ◆ old is empty
String.fields(s: string): Array[string]                 // strings.Fields: split on runs of white space
String.lines(s: string): Array[string]                  // split on \n, a \r before it dropped, no final empty line
String.cut(s: string, sep: string): String.Cut | null   // strings.Cut; ◆ sep is empty
String.count(s: string, t: string): int                 // strings.Count, non-overlapping; ◆ t is empty
String.isDigit(s: string): bool                         // unicode.IsDigit over every code point; false on ""
String.isLetter(s: string): bool                        // unicode.IsLetter, likewise
String.isSpace(s: string): bool                         // unicode.IsSpace, likewise
String.isUpper(s: string): bool                         // unicode.IsUpper, likewise
String.isLower(s: string): bool                         // unicode.IsLower, likewise
```

Ant has no character type: `for c in s` yields one-code-point strings. So the class tests take a
string and hold for every code point in it. `String.isDigit(c)` tests a character, and
`String.isDigit(s)` tests that a whole field is digits (decision 3). `String.lines` is the
companion of `File.read`, and it splits `\r\n` and `\n` line endings alike.

### 3.3 Arrays (core, `Array`)

```
Array.insert[T](xs: Array[T], i: int, v: T): void                // mutates; slices.Insert; ◆ i outside 0 ..= length
Array.removeAt[T](xs: Array[T], i: int): T                       // mutates; the removed element; ◆ i out of range
Array.first[T](xs: Array[T]): T | null                           // null when empty
Array.last[T](xs: Array[T]): T | null
Array.lastIndexOf[T](xs: Array[T], v: T): int | null             // last index, by ==
Array.findIndex[T](xs: Array[T], f: func(T): bool): int | null   // slices.IndexFunc
Array.findLast[T](xs: Array[T], f: func(T): bool): T | null
Array.findLastIndex[T](xs: Array[T], f: func(T): bool): int | null
Array.countWhere[T](xs: Array[T], f: func(T): bool): int
Array.min[T: Ordered](xs: Array[T]): T | null                    // slices.Min without the panic; a NaN in xs gives NaN
Array.max[T: Ordered](xs: Array[T]): T | null
Array.minBy[T, K: Ordered](xs: Array[T], key: func(T): K): T | null   // the first of equal minima
Array.maxBy[T, K: Ordered](xs: Array[T], key: func(T): K): T | null   // the first of equal maxima
Array.unique[T](xs: Array[T]): Array[T]                          // by ==, the first of each kept, in order
Array.uniqueBy[T, K](xs: Array[T], key: func(T): K): Array[T]    // by == on the key, likewise
Array.flatMap[T, U](xs: Array[T], f: func(T): Array[U]): Array[U]
Array.chunk[T](xs: Array[T], n: int): Array[Array[T]]            // slices.Chunk; the last may be shorter; ◆ n < 1
Array.groupBy[T, K](xs: Array[T], key: func(T): K): Map[K, Array[T]]   // keys in first-seen order
```

The array functions come in the families the core library already has. `first` and `last` pair like
`push` and `pop`. `indexOf`, `find` and `findIndex` each get a `last` form, as `String.indexOf` gets
`String.lastIndexOf`. `min`, `max` and `unique` each get a `By` form taking a key, as `sort` has
`sortBy`. A predicate count is `countWhere`, not `count`, because `String.count` counts a substring
which is a different kind of argument (decision 15).

### 3.4 Math and Bits (core)

```
Math.log(x: float64): float64              // natural; math.Log
Math.log2(x: float64): float64
Math.log10(x: float64): float64
Math.exp(x: float64): float64
Math.sin(x: float64): float64              // and cos, tan, asin, acos, atan
Math.atan2(y: float64, x: float64): float64
Math.hypot(x: float64, y: float64): float64
Math.cbrt(x: float64): float64
Math.isNaN(x: float64): bool
Math.isFinite(x: float64): bool            // neither NaN nor an infinity
Math.clamp[T: Ordered](x: T, lo: T, hi: T): T   // ◆ lo > hi
Math.randomInt(from: int, to: int): int    // uniform from inclusive to exclusive, math/rand/v2; ◆ to <= from
Math.NaN: float64
Math.Infinity: float64
Math.maxFloat64: float64

Bits.leadingZeros[T: Integer](a: T): int    // math/bits.LeadingZeros, at the type's width
Bits.trailingZeros[T: Integer](a: T): int
Bits.rotateLeft[T: Integer](a: T, n: int): T   // math/bits.RotateLeft; n may exceed the width; ◆ n < 0
Bits.rotateRight[T: Integer](a: T, n: int): T  // likewise
```

The rotations pair like the shifts and panic on a negative count as they do. `Math.mod` is not
needed, since `%` already takes floats. `Math.randomInt` can't be seeded, like `Math.random`. A
seeded generator is left out (section 3.9).

### 3.5 Regex: a literal and a core namespace

Regular expressions get a literal, `/.../flags`, as in JavaScript, with Go's RE2 syntax inside.
They also get a namespace in the core library over `regexp`. A literal is language, so `Regex`
lives in section 11 of `ant-lang.md` beside `String`, not in `stdlib.md`.

```ant
var date = /^(\d{4})-(\d{2})-(\d{2})$/
var word = /\bant\b/i

if Regex.test(word, line) { ... }
var m = Regex.find(date, s) ?? return Result.err(Error{ message: "not a date" })
var p = try Regex.compile(pattern)                   // a pattern from data
```

**Why a literal.** There are three reasons:

- Ant has no raw strings, and its escape list is closed. `"\d"` is a syntax error, so every
  pattern written as a string doubles its backslashes: `Regex.compile("^(\\d{4})-(\\d{2})$")`.
  A literal reads as written.
- A literal is checked when the source is parsed. `/(/` is a syntax error in `ant check` and
  `ant run` alike, before `main` starts, so a pattern in the source never needs a
  `Result.unwrap`.
- A JavaScript programmer writes `/.../` without thinking. This follows section 2's policy for
  JavaScript habits: they work, or they fail with a clear message, never silently.

`Regex.compile` stays for patterns that come from data, and returns `Result`. A literal for code
and a function for data are the same split as a string literal and `String.fromBytes`, so the
principle of one obvious way holds.

**Lexing.** `/` begins a regular expression unless the last token before it, skipping line breaks,
is an operand: an identifier, a literal of any kind, `true`, `false`, `null`, `void`, `)`, `]` or
`}`. After an operand, `/` is division, as before. The decision needs the previous token only, so
the grammar keeps one token of lookahead and takes no feedback from the checker. The rule differs
from the line-ending rule in one respect: `return`, `break` and `continue` can end an expression but
are not operands, so `return /x/` is a regular expression. When there is no previous token, at the
start of a file or of an interpolation, which the lexer reads as a fresh token stream, `/` begins a
regular expression, so `${ /x/ }` holds a pattern.

Line breaks are skipped because the lexer emits a line break whatever the nesting, and the parser
ignores it inside parentheses and brackets. The lexer cannot track that nesting itself, since it
cannot tell a block's `{` from a record's. So these are division today and stay division:

```ant
var r = (total
  / count)
var xs = [total
  / count, 1]
```

The cost is that a pattern cannot start a line that follows an operand. That loses nothing: at
statement level such a line is a syntax error today, and a pattern alone as a statement would do
nothing. With that rule, no program changes meaning. Every `/` that now begins a regular expression
follows a token after which division is a syntax error today.

Inside the literal, `\/` is a `/` and does not end it, and a `/` inside a character class `[...]`
does not end it either, as in JavaScript. The literal cannot span lines. `//` and `/*` still begin
comments, so an empty pattern and a pattern starting with `*` cannot be written as literals. An
empty pattern is `Regex.compile("")`, and a pattern starting with `*` is invalid anyway. `/=`
after an operand stays the compound assignment.

**Syntax and flags.** The pattern is RE2 as Go's `regexp/syntax` parses it. Matching is therefore
linear in the input. RE2 rejects lookahead, lookbehind and backreferences, so a JavaScript pattern
that uses them is a syntax error at the literal, with a message naming the construct. JavaScript's
`\u0041` is also a syntax error, and its message points to `\x{0041}`. `(?<name>...)` names a group
as in JavaScript.

A few patterns are valid in both and match differently, and the spec lists them: `\s` is `[\t\n\f\r
]`, without `\v` or Unicode spaces; `.` excludes only `\n`, where JavaScript also excludes `\r`,
U+2028 and U+2029. `\d`, `\w` and `\b` are ASCII in both. So the claim is that a JavaScript habit
either works, fails with a message, or is one of the listed differences.

The flags are `i`, `m` and `s`, each at most once, mapped to RE2's `(?i)`, `(?m)` and `(?s)`, and
sorted when the pattern is built, so `/a/si == /a/is`. JavaScript's
other flags are syntax errors whose message names the Ant way: `g` is `Regex.findAll` or
`Regex.replace`, `y` is a `^` anchor on a `String.slice`, and `u`, `v` and `d` are already how Ant
behaves.

**Value.** A literal evaluates to a `Regex.Pattern`. Each literal is compiled once, so a literal
inside a loop costs nothing per iteration. A pattern is its source and its sorted flags, and
patterns compare by both, so `/a/i == /a/i` holds and a pattern can be a map key. The source is
normalized the same way for a literal and for `Regex.compile`: an escaped `/` loses its backslash,
since RE2 reads `\/` and `/` alike. So `/a\/b/` has the source `a/b`, and both
`Regex.compile("a+", "i") == /a+/i` and `Regex.compile("a/b") == /a\/b/` hold.

A pattern renders as a literal that would read back to it: every `/` in the source is written `\/`,
a line break as `\n` or `\r`, and an empty source as `/(?:)/`, as JavaScript does. So
`Regex.compile("a/b")` renders as `/a\/b/`, never `/a/b/`, and `Regex.compile("")` never as
`//`. `Json.stringify` writes that rendering as a JSON string, `"/a+/i"`. `Json.parse` gives it
back as a string, not a pattern.

```
type Pattern                                                      // a compiled expression: literal or Regex.compile
type Match = { text: string, start: int, end: int, groups: Array[string | null], named: Map[string, string | null] }

Regex.compile(source: string, flags: string = ""): Result[Regex.Pattern, Error]   // regexp.Compile; Err on bad syntax or flags
Regex.test(p: Regex.Pattern, s: string): bool                     // MatchString
Regex.find(p: Regex.Pattern, s: string): Regex.Match | null       // FindStringSubmatchIndex
Regex.findAll(p: Regex.Pattern, s: string): Array[Regex.Match]    // FindAllStringSubmatchIndex
Regex.replace(p: Regex.Pattern, s: string, replacement: string): string   // ReplaceAllString: $1 and ${name} expand
Regex.replaceFirst(p: Regex.Pattern, s: string, replacement: string): string   // the first match only, likewise
Regex.replaceWith(p: Regex.Pattern, s: string, f: func(Regex.Match): string): string
Regex.split(p: Regex.Pattern, s: string): Array[string]           // Split(s, -1)
Regex.escape(s: string): string                                   // QuoteMeta
Regex.source(p: Regex.Pattern): string                            // normalized: without slashes, flags, or \ before /
Regex.flags(p: Regex.Pattern): string                             // in the order i, m, s
```

`start` and `end` are code-point indices into `s`, so `String.slice(s, m.start, m.end)` is
`m.text`. `groups[0]` is the whole match, and a group that did not take part is `null`. `named`
has an entry for each named group. The replacement syntax is Go's, `$1` and `${name}`, not
JavaScript's `$<name>`. A plain string does not interpolate `$`, so `"$1"` needs no escape.

`Regex` is a namespace that declares types, like `Time`, and not a type with functions, like
`Mutex`. Section 6 lets a type carry functions but not types, and `Regex.Match` has to live
somewhere. So the annotation is `Regex.Pattern` (decision 5).

### 3.6 Encodings (stdlib.md)

Three small namespaces, named after what they encode. Bytes go in and text comes out, with
`String.toBytes` and `String.fromBytes` between them and strings.

```
Encoding.toBase64(bytes: Array[byte], url: bool = false): string                 // base64.StdEncoding or URLEncoding, padded
Encoding.fromBase64(s: string, url: bool = false): Result[Array[byte], Error]     // accepts missing padding
Encoding.toHex(bytes: Array[byte]): string                                       // lower case
Encoding.fromHex(s: string): Result[Array[byte], Error]                          // either case

type Url.Parts = { scheme: string, opaque: string, user: string, host: string, port: string,
                   path: string, query: string, fragment: string }
Url.parse(s: string): Result[Url.Parts, Error]                   // url.Parse; absent parts are ""
Url.stringify(parts: Url.Parts): string                          // url.URL.String
Url.escape(s: string): string                                    // url.QueryEscape
Url.unescape(s: string): Result[string, Error]                   // url.QueryUnescape
Url.parseQuery(query: string): Result[Map[string, Array[string]], Error]   // url.ParseQuery, keys in order of appearance
Url.stringifyQuery(params: Map[string, Array[string]]): string  // url.Values.Encode, but in insertion order

Csv.parse(text: string, sep: string = ","): Result[Array[Array[string]], Error]   // encoding/csv, FieldsPerRecord = -1; ◆ invalid sep
Csv.stringify(rows: Array[Array[string]], sep: string = ","): string             // quotes where needed; \n line ends; ◆ invalid sep
```

`Url.Parts` mirrors `url.URL`, `Opaque` included. `mailto:alice@example.com` parses with
`scheme` `mailto`, `opaque` `alice@example.com`, and an empty `path`. Without `opaque`, that URL
would lose its content. `user` is the userinfo as written, a password included, `query` is the
raw query, and `path` is decoded, as Go has them.

`Url` pairs `parse` with `stringify` as `Json` and `Csv` do, for the URL and for its query. The
query takes the same map type it gives, so `parseQuery` and `stringifyQuery` round-trip, and a
repeated key such as `tag=a&tag=b` survives. A one-value query is `Map{ "q": ["ant"] }`. A union
value type would read better but cannot work, since a `Map[string, string]` is not a `Map[string,
string | Array[string]]` (section 4, Assignability). `stringifyQuery` keeps the map's order because
Ant maps are ordered, and a signed request needs a known order. JSON is already its own namespace,
and XML is left out.

A separator is valid when it is one code point other than `"`, `\r`, `\n`, U+0000 and U+FFFD. This
is the rule of Go's `validDelim`, which the reader and the writer both apply. The separator is part
of the program, not of the data, so an invalid one is a ◆ in both functions, checked before any row
is read or written. With a valid separator, `Csv.stringify` cannot fail, since it writes to a
string, and so it returns a `string`. So `Csv`, like `Crypto` with `randomBytes`, is a namespace of
computation that panics where the core library would, and `stdlib.md`'s opening paragraph names both
beside `Time`.

### 3.7 Crypto (stdlib.md)

```
Crypto.sha256(bytes: Array[byte]): Array[byte]              // crypto/sha256
Crypto.sha512(bytes: Array[byte]): Array[byte]
Crypto.sha1(bytes: Array[byte]): Array[byte]                // for interoperability, as git uses it
Crypto.md5(bytes: Array[byte]): Array[byte]                 // likewise; the comment says it is not secure
Crypto.hmacSha1(key: Array[byte], message: Array[byte]): Array[byte]     // for TOTP and OAuth 1
Crypto.hmacSha256(key: Array[byte], message: Array[byte]): Array[byte]
Crypto.hmacSha512(key: Array[byte], message: Array[byte]): Array[byte]
Crypto.equal(a: Array[byte], b: Array[byte]): bool          // subtle.ConstantTimeCompare
Crypto.randomBytes(n: int): Array[byte]                     // crypto/rand; ◆ n < 0
Crypto.uuid(): string                                       // a version 4 UUID from crypto/rand, lower case
```

A digest is bytes, so a hex digest is `Crypto.sha256(String.toBytes(s)) | Encoding.toHex()`.
Encryption, signatures and keys are left out. They need key types, and no program asks for them.

### 3.8 Http.request (stdlib.md)

```
Http.request(method: string, url: string, body: string = "",
             headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null): Result[Http.Response, Error]
Http.get(url: string, headers: Map[string, string] = Map{}, timeout: Time.Duration | null = null): Result[Http.Response, Error]
Http.post(url: string, body: string, headers: Map[string, string] = Map{},
          timeout: Time.Duration | null = null): Result[Http.Response, Error]
```

This is `http.NewRequestWithContext` with the run's context, and `timeout` sets a deadline on top of
that. A timeout is an `Err` with the message `timeout`. `get` and `post` take the same optional
`headers` and `timeout`, and the spec defines them as `request` with the method filled in. `post`'s
`headers` becomes optional, which breaks no call. The response's headers stay one string per name,
joined with `, ` as now. A binary body is left out.

The timeout is converted to a Go `time.Duration` by the rules of `Time.total` without a
`relativeTo`. Days count as 24 hours, and hours and smaller units are exact. A `Duration` with
years, months or weeks has no fixed length, so the call returns `Err` before sending anything, as
`Time.total` would need a `relativeTo`. A zero or negative timeout is also an `Err`, since a request
that has already timed out is a mistake. A timeout longer than Go's maximum, about 292 years, is
clamped to it, as `Time.sleep` clamps a huge count. `null` means no deadline beyond the run's. These
are `Err` and not ◆ because `Http` is I/O and follows `stdlib.md`'s rule.

### 3.9 Considered and left out

- **A format string** (`fmt.Sprintf`). Template strings and section 3.1 cover the uses, and a
  format string is a second, unchecked language inside strings.
- **`Set` and a heap.** `Map[K, bool]` is a set. A heap is twenty lines of Ant on an array, and
  can be added when a program wants one.
- **More `Iter`**: `zip`, `takeWhile`, `skip`. They wait for a program that needs them.
- **A seeded random generator.** It would be an object like the sync types. It waits for a
  program that needs repeatable output.
- **Error wrapping** (`errors.Is`, `%w`). It changes `Error`, which every namespace returns, so
  it needs its own design.
- **A test framework** (`testing`): an `ant test` command and assertions. Programs are tested
  through the corpus and their trailers for now, and a test runner gets a plan of its own.
- **Command-line flags** (`flag`). An Ant module can parse `Os.args()`, and the examples will
  show how.
- **Compression, archives, templates, structured logging, SQL**: each is a plan of its own when
  a program needs it.
- **Plan 6, the waiting**: an HTTP server, sockets, a deadline on any blocking call, cancelling a
  goroutine, timers. Go does all of this with `select` over channels and `context`, and Ant has
  neither. The candidates are a `Future`-like sync object that a goroutine completes,
  `Context.withTimeout` as a sync object every blocking call accepts, or a `select` over sync
  objects. Choosing between them is a language decision. Plan 5 adds only the one timeout that
  needs no such design, on `Http.request`.

## 4. Mechanisms

### 4.1 Builtins

Everything in sections 3.1 to 3.8 except the regex literal, which section 4.2 covers, is an ordinary
builtin over a Go function, declared in the
prelude. The `Builtin.Optional` of Phase 22 carries the defaults. Code-point indices in `Regex` are
converted from Go's byte offsets with a counter that advances through the input past each match, in
order. Group offsets are not in order, since an outer group starts before an inner one, so each
group's offset is counted from the start of its match. A `findAll` over a large text is therefore
linear in the text plus the total length of the matches, not quadratic.

### 4.2 Regex literals and patterns

The lexer gains a `Regex` token kind, with the source and flags as its text. Before reading a `/`,
it checks whether the last token, skipping line breaks, is an operand (section 3.5). That is a small
function beside `lastContinues`, which answers a different question: `return` ends a line but is not
an operand. The check runs before the operator table, so `(/=/)` is a pattern while `x /= 2` stays
an assignment. The lexer finds the end of the literal with the class and escape rules, removes the
backslash of an escaped `/`, and checks and sorts the flags. A lexer option turns literals off,
which gives today's lexer for the differential fuzz test of section 4.4.

The parser builds a `RegexLit` node and calls `regexp.Compile` once, keeping the `*regexp.Regexp` on
the node. That one call is the validation and the compilation, so there is one code path, and the
interpreter only wraps the compiled value when the module is initialized. An error is reported at
the literal's position, with Go's message translated where it names RE2 internals. Lookaround,
backreferences and `\u` get their own messages, since those are the errors a JavaScript programmer
will meet.

A pattern is a new value kind, `regex`, beside `time` and `sync` in the value package. It holds the
normalized source, the sorted flags, and the `*regexp.Regexp` compiled from `(?flags)source`.
Equality and hashing use the source and flags. `Regex.compile` normalizes the same way and builds
the same value. A `*regexp.Regexp` is safe for concurrent use, so a shared pattern needs no lock.
The checker types a literal as `Regex.Pattern`, and `String.from` and `Json.stringify` follow the
rendering rule of section 3.5.

### 4.3 Record results

The runtime has no untagged record: every `Record` carries a `Tag`, and `value.NewRecord(nil, …)`
crashes. So every record a builtin returns has a declared type with a stable tag, as `Entry` has
`value.EntryTag`. `String.Cut`, `Url.Parts` and `Regex.Match` are declared in the prelude like
`Os.Output`, and `internal/lib` builds their tags once and registers them in the tag table, as
`lib.go` does for `Entry`. `String.Cut` is the first type declared on a core namespace, which
section 11 already allows: each namespace is "a plain record of functions and types".

### 4.4 Tests

Corpus programs under `tests/run`: one per namespace or per group of functions, tagged `11 String`,
`11 Array`, `11 Math and Bits`, `11 Regex`, `stdlib Encoding`, `stdlib Url`, `stdlib Csv`, `stdlib
Crypto` and `stdlib Http`. A ◆ is reached with a `// panics:` line, as now. `TestPanicCoverage`
reads only chapter 11 of `ant-lang.md` today, so a tag naming `Csv.parse` or `Crypto.randomBytes`
would be rejected as unknown, and a generic tag could not hold each function to a program. Phase 36
extends its scanner to the code blocks of `stdlib.md`. There the ◆ often sits on a continuation line
of a wrapped signature, as in `Time.plainTime`, `Time.with` and `Time.roundDuration`, so the scanner
joins a signature's lines before reading its name. The ◆ functions of `Time` already there come
under the gate at the same time, and any without a program goes on `pendingPanics` until Phase 37
gives it one.

The literal gets programs under `tests/syntax` and `tests/run`, tagged `2 Literals`, one per lexing
case. These stay division: `a / b / c`, `f(x) / 2`, `{ x } / 2`, `xs[0] / 2`, `x /= 2`, `}` followed
by `/` on the next line, and the two bracketed divisions of section 3.5, `(total` and `[total` with
`/ count` on the next line. These are a pattern: `/x/` after `(`, `,`, `=`, `|`, `??` and `return`,
`(/=/)`, and `${ /x/ }` inside a template string. `/[/]/` and `/a\/b/` are one literal. An
unterminated literal, a literal across a line, an unknown flag, a repeated flag, `/(?=a)/`,
`/(a)\1/`, `/\u0041/` and `/(/` are each a syntax error with its message. Rendering gets cases for
`Regex.compile("a/b")`, `Regex.compile("")`, a source with a line break and `/a/si`.

The corpus has no pattern in it today, so its passing proves nothing about division. A fuzz test
does that instead: on every input for which the new lexer emits no `Regex` token, its token stream
must equal the stream of the lexer with literals turned off, which is today's lexer. The lexer's
fuzz seeds gain inputs with slashes, brackets and line breaks.

Digests are checked against the known vectors of RFC 4231 for HMAC and FIPS 180-4 for SHA.
`Http.request` uses the corpus's echo server, and the timeout uses a new `/slow` route that waits
longer than the timeout. Its conversion rules each get a case: a timeout in days, one in months,
zero, a negative one, and one beyond Go's range.

## 5. Phases

Phases continue from Plan 4's numbering. Plan 4 defined a Phase 35, unboxed numbers, gated on a
Phase 34 whose gate failed, so it never ran; this plan starts at 36 to keep the numbers unique. Each
phase is one commit or a few, with the suite green
under `-race` and the coverage gate where it was.

- **Phase 36: spec.** The signatures of section 3 and their paragraphs go into `ant-lang.md` section
  11 and `stdlib.md`. The regex literal goes into section 2: the literal, the `/` rule and the
  listed differences from JavaScript. The `regex` kind goes where the time kind is: section 4's
  prose on equality by value and immutability, and the Values table of section 12, with one value
  per literal, built at load. One commit in the spec repository. In this repository, the
  `TestPanicCoverage` scanner is extended to `stdlib.md` (section 4.4). The new headings go on
  `pendingSpec` and the new ◆ functions on `pendingPanics`, since a spec-only commit leaves the
  prelude, and so `TestLibraryCoverage`, untouched. Done when every decision of section 6 is written
  down and the suite is green with those two lists.
- **Phase 37: strings and numbers.** Sections 3.1 and 3.2, with their corpus programs, and a
  program for each `Time` ◆ that Phase 36 put on `pendingPanics`. Done when `String` and `Time`
  have no pending entry.
- **Phase 38: arrays and math.** Sections 3.3 and 3.4. Done when `Array`, `Math` and `Bits` have
  no pending entry.
- **Phase 39: the regex literal.** The token, the operand rule, the `RegexLit` node with its
  validation and messages, the `regex` value kind, the checker's type, rendering. The
  syntax corpus of section 4.4 and the differential fuzz test. Done when every lexing case has a
  program and the fuzz test has run clean for ten minutes.
- **Phase 40: the Regex namespace.** The functions of section 3.5, with the code-point indices.
  Done when the corpus matches, replaces and splits non-ASCII text at the right indices, and
  `Regex.compile` builds values equal to the literals.
- **Phase 41: encodings and crypto.** Sections 3.6 and 3.7, with the known test vectors.
- **Phase 42: Http.request.** Section 3.8 and the `/slow` route, with `get` and `post`
  re-expressed in the spec.
- **Phase 43: examples and documentation.** `examples/logstat.ant` reads an access log from standard
  input, parses lines with a `Regex`, groups by status with `Array.groupBy`, and prints a table
  padded and to two decimals. `examples/webhook.ant` posts a JSON payload with `Http.request`, a
  `PUT`, an HMAC signature header and a timeout to the test server, and checks the response.
  `CHANGELOG.md`, the package documentation in `ant.go`, the README, and the status here.

## 6. Decisions

Open, each with the recommendation the plan is written to, except 4, which is taken:

1. **Scope.** Add these functions now, from a survey of Go (recommended), rather than one at a
   time as programs ask. `stdlib.md` says "a function is added when a program needs it", and
   this plan treats the first hour of every program as that need. The sentence stays.
2. **Number formatting** as `String.fromInt(x, base)`, `String.fromFloat(x, digits)` and the
   pads (recommended), rather than a format-string function.
3. **Character classes** as `String.isDigit(s)` and the like, true when every code point holds
   (recommended), rather than a `Char` namespace over integer code points.
4. **A regex literal** `/.../flags` with RE2 inside and the operand rule for `/`, rather than
   patterns only as strings to `Regex.compile`, or raw strings as a general fix for the escapes.
   Taken on 2026-10-22, in conversation.
5. **`Regex` as a core namespace** declaring `Regex.Pattern` and `Regex.Match` (recommended),
   rather than `Regex` as a type like `Mutex`, which would need section 6 to let a type carry a
   type, or a prelude `RegexMatch`.
6. **Patterns compared by normalized source and sorted flags**, as a new `regex` value kind
   (recommended), rather than by identity like `File.Handle`, which would make `/a/ == /a/` false.
7. **Flags `i`, `m`, `s` only**, with JavaScript's others as syntax errors that name the Ant way
   (recommended), rather than accepting and ignoring `g` and `u`.
8. **Go's replacement syntax** `$1` and `${name}` (recommended), rather than JavaScript's
   `$<name>`.
9. **Indices in code points** in `Regex.Match` (recommended), as everywhere in `String`, rather
   than Go's byte offsets.
10. **Encodings split into `Encoding`, `Url` and `Csv`** (recommended), rather than one namespace
   or five (`Base64`, `Hex`, ...).
11. **`Crypto` includes `sha1` and `md5`** for interoperability, marked insecure (recommended),
   rather than SHA-2 only.
12. **`Http.request` with a `Time.Duration` timeout** (recommended), rather than milliseconds as
   an `int`, or no timeout until Plan 6.
13. **Left out**: format strings, `Set`, a heap, more `Iter`, seeded random, a test framework,
    error wrapping, flags, compression, templates, logging and SQL (recommended, section 3.9).
14. **Plan 6 for the waiting**: server, sockets, deadlines and cancellation designed together
    (recommended), rather than an `Http.serve` now that blocks forever and cannot be stopped.
15. **`Array.countWhere`** for a predicate count (recommended), rather than `Array.count` with the
    difference documented. `contains` and `indexOf` already differ between `Array` (an element) and
    `String` (a substring), but both take a value. A predicate is a different kind of argument, as
    `find` beside `indexOf` shows, and the name says so.
16. **Queries as `Map[string, Array[string]]`** both ways (recommended), so they round-trip and keep
    repeated keys, rather than `Map[string, string]` in and arrays out.
17. **Families completed** (recommended): `first` beside `last`, `last` forms of `indexOf`, `find`
    and `findIndex`, `By` forms of `min`, `max` and `unique`, `rotateRight` beside `rotateLeft`,
    `Regex.replaceFirst`, `Url.stringify`, HMAC for every SHA digest, and the `request` options on
    `get` and `post`. The library's rule is consistency within a family, not one way to do it.

## 7. Status

Written on 2026-10-22 for review, and revised the same day after a review: `String.cut` returns a
declared `String.Cut`, since the runtime has no untagged records; the `Http.request` timeout has
conversion rules; the CSV separator follows Go's `validDelim` in both directions; `fromInt` uses
`FormatUint` for unsigned values; `Url.Parts` has `opaque`; the panic gate reads `stdlib.md`. The
test framework (`ant test` and a `Test` namespace) was dropped from the plan the same day, at
Anton's request. The regex literal of section 3.5 was added the same day, at Anton's request, with a
phase of its own; decisions 4 to 9 cover its details, and 4 is taken. A second review of the literal
found that the operand test must skip line breaks, since `(total` followed by `/ count)` on the next
line is division today; it also fixed normalization, rendering, flag order and the one compile. A
consistency pass the same day completed the function families (decision 17), renamed the predicate
count to `countWhere` and made queries round-trip. A third review the same day corrected references
to the repository and the spec: there is no `ant fmt` and no `doc.go`, `try` is a prefix, the phases
start at 36 after Plan 4's unused 35, and Phase 36 names the pending lists a spec-only commit
breaks.

Phase 36 was done on 2026-10-22: one commit in the spec repository, with sections 2, 4, 5, 11 and
12 of `ant-lang.md` and the opening paragraph, `Http`, `Csv`, `Encoding`, `Url` and `Crypto` of
`stdlib.md`, and one here, with the panic scanner extended to `stdlib.md` and `pendingSpec` and
`pendingPanics` filled. Two things differ from the plan as written. The nine `Time` panic programs
and `json_cycle.ant` each already triggered one ◆ function, so they were retagged with its name
instead of waiting for Phase 37, which leaves `Time.with`, `since`, `addDuration`,
`compareDuration` and `roundDuration` for that phase. And the parser's test that reads every `ant`
block of the spec could not read the section 2 example with a literal in it, so it gained a
`pendingBlocks` list of the same kind as the others, which Phase 39 empties. Decisions 1 to 3 and
5 to 17 are written into the spec as recommended, so they are taken with this phase.

Phase 37 was done on 2026-10-22: sections 3.1 and 3.2 in `internal/lib/string.go` and the
prelude, `String.Cut` as the first type on a core namespace, two corpus programs for the
functions, one per ◆, and a program for each of the five `Time` ◆ functions Phase 36 left
pending. `String.toInt` keeps its own digit check in front of the parse, now for any base, so
that a prefix, a space or an underscore is `null` whatever the parser would take. Nothing else
differs from the plan.

Phase 38 was done on 2026-10-22: sections 3.3 and 3.4 in `array.go`, `math.go` and `bits.go`,
with `Insert` and `Remove` on the array value, which publish a fresh slice so a concurrent
reader keeps a consistent view, and three corpus programs plus one per ◆. One thing was found
on the way: the checker synthesized a lambda's body only when the expected result was a bare
type parameter, so `Array.flatMap(xs, { [x] })` failed to infer `U` from `Array[U]`. It now
synthesizes whenever the expected result mentions an unsolved parameter, and the Inference
section of the spec says so in one clause.

Phase 39 was done on 2026-10-22: the `Regex` token with the operand rule in the lexer, the
`RegexLit` node built by the parser from `value.NewRegex`, which is the one path that
normalizes, sorts the flags, compiles and rewords RE2's errors, the `regex` value kind with its
equality, hash, rendering and JSON cases, the checker's type for a literal, and `is` on
`Regex.Pattern`. The lexer keeps the source as written and leaves the normalization to
`NewRegex`, so a literal and `Regex.compile` share every step. Thirteen programs under
`tests/syntax` and two under `tests/run` cover the lexing cases of section 4.4, and
`FuzzRegexDivision` is the differential test, which ran clean for ten minutes. The prelude's
list of bound names in `names.go` is hand-written, so `Regex` was added there; Phase 41 adds
the four stdlib namespaces the same way.

Phase 40 was done on 2026-10-22: the eleven functions of section 3.5 in `internal/lib/regex.go`
over Go's `regexp`, with `Regex.Match` as a tagged record and the code point indices from one
cursor per call, and a corpus program that matches, replaces and splits non-ASCII text at the
right indices and compares `Regex.compile`'s values with literals. Nothing differs from the
plan.

Phase 41 was done on 2026-10-22: `Csv`, `Encoding`, `Url` and `Crypto` in four files under
`internal/lib`, a program per namespace with the FIPS 180-4, RFC 2202 and RFC 4231 vectors,
and one per ◆, which empties `pendingSpec` and `pendingPanics`. Two small choices on the way:
`Url.parseQuery` reads the pairs itself, since Go's `ParseQuery` returns an unordered map, and
`Csv.stringify` writes its rows itself, since Go's writer also quotes a field that is `\.`, a
PostgreSQL convention the spec does not have.

Phase 42 was done on 2026-10-22: `Http.request` with `get` and `post` as its forms, the
timeout as a deadline on the run's context with the conversion rules of section 3.8, the
test server's `/echo` answering any method with `X-Method` set and a `/slow` route that
answers once the client has given up, and a corpus program for the methods, the headers and
every timeout rule. Nothing differs from the plan.
