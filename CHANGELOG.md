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

Decisions the specification leaves open are recorded in `PLAN.md` and questions for the
specification in `SPEC-FEEDBACK.md`.
