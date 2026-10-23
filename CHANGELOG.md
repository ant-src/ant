# Changelog

## Unreleased

The first implementation of Ant draft 0.1: the `ant` command with `run`, `check` and a REPL,
and the `github.com/ant-src/ant` module with `Load`, `Run`, `Check`, `Start`, `Func`,
`Namespace` and `Session`.

- A tree-walking interpreter of chapter 12 with real goroutines and memory safety under data
  races: one-word shared slots, atomic array headers, a lock per map.
- A bidirectional type checker of chapters 4 to 10, with generics solved from the expected type
  and the arguments, union discriminators, flow-based narrowing, and exhaustive `is` chains.
- The whole standard library of chapter 11 and the I/O namespaces of `stdlib.md`, declared in
  Ant notation in `internal/prelude/prelude.ant`.
- A test corpus traceable to every heading of the specification, every library function and
  every panic; ten example programs; fuzzers for every layer; a generated-expression test of the
  checker against the runtime; statement coverage above 90%.

- The world outside: `Os.exec` without a shell, `Os.environ`, `Os.cwd`, the
  standard streams as `File.Handle`s through `Os.stdin`, `Os.stdout` and `Os.stderr`;
  `File.stat`, `mkdir`, `removeAll`, `rename`, `walk`, `tempDir`, `append`, `readAll`,
  `writeText` and `flush`; a `Path` namespace; `Math.random`; defaults on library parameters;
  `ant.WithStdin`; and a `// stdin:` trailer in the corpus.
- Time values after JavaScript's Temporal: `Instant`, `ZonedDateTime`, `PlainDate`, `PlainTime`,
  `PlainDateTime`, `Zone` and `Duration` as a `time` kind compared by value, with RFC 9557
  strings, IANA zones from a generated table, gaps and overlaps resolved by disambiguation,
  calendar and exact arithmetic, rounding, totals, balancing and transitions, in
  `internal/temporal`. `Time.now()` returns an `Instant`.
- The coverage gate runs `go test` itself with `-run`, and three more examples: a task runner
  over `Os.exec`, a filter over standard input, and an agenda across zones.

- The everyday library (Plan 5). A regular expression literal `/.../flags` with RE2 inside: a
  `/` begins one unless the token before it is an operand, so every existing program keeps its
  meaning, which a differential fuzz test checks; a pattern is a value compared by source and
  flags, rendered as a literal that reads back. A `Regex` namespace: `compile`, `test`, `find`,
  `findAll`, `replace`, `replaceFirst`, `replaceWith`, `split`, `escape`, `source` and `flags`,
  with `Regex.Match` in code point indices.
- `String.fromInt` and `fromFloat`, `toInt` with a base, `padStart` and `padEnd`, one-sided
  trims, `trimPrefix` and `trimSuffix`, `lastIndexOf`, `count`, `replaceFirst`, `cut` with
  `String.Cut`, `fields`, `lines` and the class tests `isDigit`, `isLetter`, `isSpace`,
  `isUpper` and `isLower`.
- `Array.first`, `last`, `insert`, `removeAt`, `flatMap`, `findIndex`, `findLast`,
  `findLastIndex`, `countWhere`, `lastIndexOf`, `min`, `max`, `minBy`, `maxBy`, `unique`,
  `uniqueBy`, `chunk` and `groupBy`; the checker infers a lambda's result through `Array[U]`.
- `Math.log`, `log2`, `log10`, `exp`, `sin`, `cos`, `tan`, `asin`, `acos`, `atan`, `atan2`,
  `hypot`, `cbrt`, `isNaN`, `isFinite`, `clamp`, `randomInt`, `NaN`, `Infinity` and
  `maxFloat64`; `Bits.rotateLeft`, `rotateRight`, `leadingZeros` and `trailingZeros`.
- `Http.request` with any method, headers and a `Time.Duration` timeout, and `get` and `post`
  with the same options.
- Four I/O namespaces: `Encoding` for base64 and hex, `Url` with `Url.Parts` and ordered query
  maps, `Csv` with a separator, and `Crypto` with SHA-2, SHA-1, MD5, HMAC, a constant-time
  compare, random bytes and UUIDs.
- The panic gate reads `stdlib.md` as well as chapter 11, and two more examples: a log summary
  with patterns and padding, and a signed webhook delivery.

- The waiting (Plan 6). A `Context` is done once, by a call, a deadline or its parent, and
  every blocking library call takes one as its last parameter and gives up with its reason:
  `Chan`, `WaitGroup.wait`, `Time.sleep`, `Http.request`, `Os.exec`, the `File` reads, `Net`,
  and `Http.serve`. `WaitGroup` is rebuilt on a channel closed at zero.
- `Chan[T]`, a channel as a sync object: `new`, `send`, `receive`, `close`, `isClosed`,
  `length` and `values` for `for`. Section 9's Waiting gives the mailbox, the first of several,
  the pool, the future and the ticker.
- `Net`: `listen`, `address`, `accept`, `close`, `dial` and `peer`, with a connection as a
  `File.Handle` whose writes flush and whose reads a context ends through the deadline.
- `Http.serve` over a listener with `Http.Request`, each handler on its own goroutine, a panic
  or a bad status answered with 500, and a five-second drain when the context is done.
- `Os.interrupt`, a context the first Ctrl-C or SIGTERM completes when the program asked for
  it, the second ending the run; `ant.WithInterrupts` for hosts, and the `ant` command relaying
  its signals. Three more examples: a worker pool with a deadline, a socket echo server, and a
  JSON API that drains on an interrupt.

- `select` (Plan 7), a keyword: a clause is one of five blocking calls, `Chan.receive`,
  `Chan.send`, `Context.wait`, `WaitGroup.wait` or `Time.sleep`, written as the call and not
  performed, with a block each, `var x =` binding the result and `else` for nothing ready. The
  heads' operands evaluate once before the wait, one ready head is chosen at random, and the
  blocks return, break and continue like any other, so a state machine is a loop around a
  `select`. A `select` is an expression typed like an `if`. The `Select` namespace of Plan 6 is
  gone.

- The review (Plan 8). A handle's reads, writes and close lock apart, so a reader and a writer
  share a connection and a close from a third goroutine ends a read in progress; a read ended by
  its context keeps what it consumed for the next read, standard input stays open and the late
  line arrives whole, and `File.pipe` flushes what it copied. A timeout context unregisters
  from the run when its deadline passes; a blocking call with a done context gives up at once;
  `Context.deadline` is right past 2262; a count above 2^31 − 1 is a panic row; a channel keeps
  a closed signal beside a channel of values that is never closed, so a close under a blocked
  sender is race-free; a goroutine whose wait the run's end released runs nothing more; the
  host's deadline reads `cancelled`; an interrupt that ends the run is `ErrInterrupted` and
  exit 130. `Http.request` follows no redirect and refuses an empty method, `Host` is among a
  handler's headers, the status bound is 200 to 599, `serve` closes what the drain left and the
  listener with it, acceptors with their own contexts do not disturb each other, and a userinfo
  whose escape does not decode is kept. Regex: the first participating group fills a duplicate
  name, a POSIX name inside a class, `split` of an empty string, and a compiled source with a
  line break or nothing in it reads back equal; strict base64 padding; a NaN bound in
  `Math.clamp` is a panic. The checker holds a generic function's own type parameters rigid, so
  a lambda inside it cannot pass with a body of another type.
