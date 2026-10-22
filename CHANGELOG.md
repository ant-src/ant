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
