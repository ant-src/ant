# Plan 4: a fast interpreter

Plans 1 to 3 built a correct interpreter and left speed alone on purpose: PLAN2.md allowed at
most two optimizations and made neither. This plan makes the interpreter fast without changing
what a program does. The specification stays the only source of truth, and every phase ends in a
commit with every test green under the race detector. A change lands only with a `benchstat`
comparison showing its gain and no regression above 5% on any program.

The point of having no global lock is that goroutines run in parallel. Today that buys little:
14 goroutines finish only 2.7 times faster than one, and antc, which has a GIL, beats ant by 5 to
6 times on parallel `fib` because it is 9 times faster on one core. This plan fixes both halves.

## 1. Where we start

Measured on 2026-10-22, Apple M4 Pro (10 performance and 4 efficiency cores), Go 1.27.1, with
`bench/compare/run.sh` (see `bench/compare/RESULTS.md`) and allocation and CPU profiles of
`bench/*.ant`.

| | Now |
| --- | --- |
| Single-goroutine time against antc, geometric mean of 10 programs | 3.8× slower |
| Worst ratios | `loop` 10.5×, `closures` 8.1×, `array` 7.5×, `fib` 5.7× |
| Allocations per call (`fib`) | 6 |
| Allocations per loop iteration (`loop`) | 3 |
| `par_fib`, speedup of 10 goroutines over 1 (ideal 10) | 2.4× |
| `par_fib` with 8 goroutines: share of CPU in interpreter code | about 7% |
| GC cycles in one 350 ms run of `par_fib 8` | about 300 |
| Startup, `ant run` of an empty `main` | 3.7 ms |

The profiles give the causes:

1. **Every call allocates its environment piece by piece.** `newEnv` makes the `Env`, a `plain`
   slice and a `shared` slice. `callFunc` adds a `filled []bool`, and `postfix` adds the argument
   slice. That is all six of `fib`'s allocations per call (`internal/interp/env.go`,
   `internal/interp/eval.go` `callFunc`).
2. **Every number outside 0 to 255 is boxed.** `Value` is `any`, so `int64` and `float64` results
   go through `runtime.convT64`. All of `loop`'s 3 allocations per iteration come from that, in
   `value.int64Op` and `eval`.
3. **Every store into a shared slot allocates a cell.** A record field, an array element, a
   captured local and a module variable are each a `value.Slot`. A `Slot` is an atomic pointer to
   an immutable `cell`, which is what section 9's memory-safety guarantee rests on. A store of a
   number allocates twice: the box and then the cell.
4. **Blocks and loop iterations open environments.** `evalBlock` calls `newEnv` for any block with
   locals. `forLoop` calls it for every iteration, and a `for` over `Iter.range` calls every
   element through a `Builtin` named `yield` and through `callBuiltin`.
5. **Dispatch is by name and by string.** Operators reach `value.Binary` as strings like `"+"`, and
   each step compares against a chain of string cases. A field read looks the name up in the
   tag's `map[string]int` on every access.
6. **Allocation rate turns into lock contention under parallelism.** With the default `GOGC=100`
   and a heap of a few megabytes, the GC runs constantly. Goroutines then queue on the Go heap
   lock (`runtime.lock2`, 46% of samples) and in `madvise`. Raising `GOGC` alone takes
   `par_fib 14` from 531 ms to 263 ms. It changes nothing on one goroutine.
7. **Two global locks in the library.** A contended `Mutex.withLock` flushes stdout before
   blocking, and `Instance.Flush` takes the run-wide `outMu`. The `goroutines` benchmark spends
   247 ms of CPU on a 90 ms run.

Not a cause: startup (3.7 ms), map lookups apart from the per-key bucket slice, and shared
counters. The only atomic every goroutine writes is `in.active`, once per `go`.

## 2. Definition of done

1. **Single-core speed.** The geometric mean against antc is at most 2.0× (now 3.8×), and no
   program is worse than 4× (now 10.5×). `fib` allocates at most one object per call, and a
   `while` loop over `int` locals allocates nothing per iteration.
2. **Parallel scaling.** `par_fib` with 10 goroutines takes at most 1.5 times as long as with one,
   a speedup of at least 6.5× (now 2.4×). `par_fib` and `par_map` beat antc at every N ≥ 2.
3. **Benchmarks in-process and in CI.** `bench/bench_test.go` runs `par_fib` and `par_map` at
   N = 1, 4 and 10 as sub-benchmarks. `bench/BASELINE.md` is re-recorded at the end of every
   phase. CI runs the benchmarks once as a smoke test, without timing gates, because CI machines
   are too noisy for them.
4. **Nothing observable changes.** Every test passes under `-race`, the fuzzers and the
   conformance suite included. Traces, panic messages and output order are byte-identical.
   Section 9 still holds: every slot two goroutines can reach is published by one atomic
   pointer store of an immutable cell.
5. **The public embedding API is unchanged**, unless decision 2 says otherwise.

## 3. Mechanisms

### 3.1 Measuring

- `ant run --cpuprofile FILE` and `--memprofile FILE`, so anyone can profile a program without the
  harness this plan was written with.
- `par_fib` and `par_map` move into `bench/` with N as an argument. Their sub-benchmarks set
  `WithArgs`.
- `scripts/bench.sh OLD NEW` runs `go test -bench` ten times at each of two commits and prints
  `benchstat`. Every phase's commit message quotes the result.
- `bench/compare/run.sh` stays the cross-implementation check and is rerun at the end of the plan.

### 3.2 GC target and the stdout lock

`cmd/ant` calls `debug.SetGCPercent(400)` at startup unless `GOGC` is set in the environment. The
library does not touch GC settings, because the embedding program owns its process; the README
tells embedders what the setting is worth. Measured: `par_fib 8` goes from 343 ms to 202 ms with no
single-core change. Section 3.6 shrinks the heap churn underneath, so the setting matters less
over time.

`Instance` keeps an atomic `dirty` flag that `Print` sets and `Flush` clears. A blocking builtin
then takes `outMu` only when there is output to flush. The flush-before-blocking rule stays as it
is, without the run-wide lock on the uncontended path.

### 3.3 One allocation per call

- `Env` gets inline storage, `plain [4]value.Value` and `shared [2]value.Slot`, used when the
  scope fits. The resolver's counts already say whether it does. A frame then costs one
  allocation for most functions, and larger frames keep the slice path.
- `callFunc` takes a fast path when every argument is positional and the count matches: no
  `filled` slice and no name matching. Named arguments and defaults keep the present code.
- Arguments go to a per-goroutine argument stack. `postfix` evaluates into
  `g.args[top:top+n]`, the callee copies them into its frame, and the stack pops on return. A
  `Builtin` receives the subslice. The contract becomes "a builtin must not keep `args` after it
  returns": `internal/lib` is audited for that, and `hostBuiltin` copies for host functions.
- `defer g.pop()` becomes an explicit pop on the normal path. The unwinding path already restores
  the stack where a panic is recovered.

Expected: `fib` from 6 allocations per call to 1; `closures` and `records` gain from the same
path.

### 3.4 Flatter scopes

The resolver lays every local of a function out in the function's own frame, with each block
taking a range of slot indexes, unless the block owns a shared slot. Sections 5 and 12 require a
fresh binding for each iteration of a `for`, which matters when a closure captures it; such a block keeps its own
environment, and every other block, which is nearly all of them, opens none. `ast.BlockScope`
gains an `Inline` flag, and `evalBlock` and `forLoop` skip `newEnv` when it is set. The tests on
closures in loops are the guard.

### 3.5 Faster dispatch

- **Operators as integers.** The parser stores an `ast.Op` enum next to the operator text, which
  the printer and the error messages keep using. `arith`, `value.Binary` and `value.Compare`
  switch on it.
- **A fast path for `int` and `float64`.** `binary` tries `l.(int64)` and `r.(int64)`, then the
  same for `float64`, before falling into `value.Binary`. That skips the `Tagged` checks and the
  type switch over ten kinds. The overflow checks stay.
- **Inline caches for fields.** `ast.Field` carries an atomic pointer to `{tag, index}`. A read
  whose record has the cached tag goes straight to `fields[index]`, and a miss looks the name up
  and replaces the cache. The cache is atomic because one AST serves every goroutine and every
  `Run`. Assignments to fields use the same cache.
- **Native sequences.** The library's sequence builtins (`Iter.range`, `Iter.fromArray`, and every
  `seq(...)` in `internal/lib/iter.go`) implement `interp.NativeSeq`, `Each(g, push func(Value)
  bool)`. `forLoop` drives a native sequence through `run` directly, with no `yield` builtin and
  no `callBuiltin` per element. A user-written sequence keeps the current path, so nothing
  observable changes: a native sequence's `yield` was never reachable from Ant code.
- **Map buckets without a slice per key.** `value.Map.index` becomes `map[uint64]int32`, with the
  head of a collision chain threaded through `mapEntry.next`. That removes the `[]int` allocated
  per new key.

### 3.6 Stores without a fresh cell

This is the first candidate in `bench/BASELINE.md`. A cell has to be immutable, but it does not
have to be new:

- **Self-cells.** `Record`, `Array`, `Map`, `Func`, `Tagged`, `Tag` and the sync objects embed a
  `cell` that holds the object itself, filled in at construction. Storing a reference publishes a
  pointer to that cell and allocates nothing.
- **Interned cells** for `null`, `void`, `true`, `false` and the `int` values -128 to 1023, built
  once at package init.
- Every other store allocates a cell as today. `CompareAndSwap` and `Modify` compare cell
  pointers only to detect a concurrent store, so sharing a cell between slots is safe: a CAS that
  sees an interned cell compares the values with `Equal` before swapping, as it does now.

Expected: `records`, `array` and `closures` lose most of their store allocations, and
`goroutines` loses the counter's cells.

### 3.7 Unboxed numbers (behind decision 2)

Sections 3.2 to 3.6 leave one cause: an `int` or `float64` in a `Value` is a heap box. Removing
it means changing the representation inside the interpreter:

```go
type V struct {
	k Kind   // kind, or KindRef for everything behind ref
	n uint64 // the bits of every numeric kind and of bool
	r any    // string, *Record, *Array, … when k is KindRef or KindString
}
```

Plain locals, arguments and the argument stack hold `V` by value and never box. A `Slot` still
publishes an immutable cell, now `cell{v V}`, so section 9 holds as before. A number stored into a
field or an element costs one allocation instead of two, or none when it falls in the interned
range.

This is the largest change in the plan. It touches `value` (2,100 lines), `internal/interp`
(1,800) and `internal/lib` (2,900). It is done as a spike first: `value.V` in `value` plus
`fib`, `loop` and `par_fib` running end to end on a branch, measured. It proceeds only if `loop`
is at least 2 times faster than after section 3.6. The public API (`ant.Value`, `HostFunc`,
`Runtime.Call`) keeps `any` and converts at the boundary in `host.go`.

### 3.8 Considered and left for later

- **Closure compilation**: turning each `ast.Expr` into a Go `func(*Goroutine, *Env) V` once at
  load. Tree walkers usually gain 1.5 to 3× from it. It goes later because it multiplies with
  sections 3.3 to 3.7 rather than replacing them, and it is easier to write once `V` exists.
- **A bytecode VM or a JIT.** Out of proportion for an interpreter that also has to be
  embeddable and race-safe. antc is the fast implementation.
- **Caching boxed integers beyond 1023** with `unsafe` interface construction. It is fragile
  across Go releases and covers only part of the range. Section 3.7 is the real fix.
- **Per-goroutine heaps or arenas.** The Go runtime does not offer them in a usable form.

## 4. Phases

Each phase ends in a commit whose message quotes its `benchstat`, with every test green under
`-race`.

- **Phase 28: measuring.** Section 3.1: `--cpuprofile`, `--memprofile`, `par_fib` and `par_map` in
  `bench/`, `scripts/bench.sh`, the CI smoke run, and `BASELINE.md` re-recorded with the parallel
  rows.
- **Phase 29: runtime hygiene.** Section 3.2. Gate: `par_fib` at N = 10 at least 1.7 times
  faster, single-core unchanged.
- **Phase 30: calls.** Section 3.3. Gate: `fib` at most 2 allocations per call and at least 30%
  faster.
- **Phase 31: scopes.** Section 3.4. Gate: no `newEnv` in a profile of `for` loops whose body
  captures nothing, plus a test that closures created in a loop still see one binding per
  iteration.
- **Phase 32: dispatch.** Section 3.5. Gate: `loop`, `records` and `seq` each at least 25% faster.
- **Phase 33: cells.** Section 3.6. Gate: `records` and `array` allocations at least halved.
- **Phase 34: the unboxed spike.** The spike in section 3.7 and a written result. It stops here
  unless the gate holds.
- **Phase 35: unboxed numbers.** The rest of section 3.7, if Phase 34 passed. Then
  `bench/compare/run.sh` is rerun, `bench/compare/RESULTS.md` updated, and section 2 checked
  item by item.

## 5. Decisions

1. **The GC setting in `cmd/ant`.** Proposed: `SetGCPercent(400)` unless `GOGC` is set. The
   alternative is a soft memory limit (`SetMemoryLimit`) with `GOGC=off`, which uses memory
   better but behaves worse for programs that are mostly live data. To decide before Phase 29.
2. **Unboxed values.** Proposed: run the Phase 34 spike, and on success change the internal
   representation while the public API keeps `any`. The alternative is to also expose `V` to
   embedders, which avoids the conversion at the host boundary but breaks every `HostFunc`. To
   decide before Phase 34.
3. **PLAN2's limit of two optimizations** is lifted by this plan. The rule that replaces it is the
   one above: every change needs a `benchstat` showing its gain and no regression above 5%.

## 6. Status

Written on 2026-10-22 from the measurements in section 1. Work happens on branch `perf`.
Decisions 1 and 2 were taken as proposed: `SetGCPercent(400)` in `cmd/ant`, and `any` stays in
the public API.

Phase 28 is done: `ant run --cpuprofile/--memprofile`, `par_fib` and `par_map` in `bench/` with
sub-benchmarks at n = 1, 4 and 10, `scripts/bench.sh`, the CI bench smoke, and `BASELINE.md`
re-recorded.

Phase 29 is done: `cmd/ant` sets `SetGCPercent(400)` unless `GOGC` is set, and `Flush` takes
`outMu` only when output is waiting. `ant run`, 15 runs each: `par_fib 10` 443 → 231 ms (1.92×),
`par_map 10` 245 → 199 ms, `par_fib 1` and `loop` unchanged.

Phase 30 is done in part. Done: environments of up to four plain slots are one allocation, sized
exactly (`env1` to `env4`); a fixed inline array was tried first and made every frame 144 bytes,
which cost more than it saved. A call of an Ant function with no named arguments evaluates onto
the goroutine's argument stack and binds from it; every `recover` resets the stack. Positional
calls of the right arity skip `filled`. Not done: `defer g.pop()` stays, since open-coded defers
cost about nothing here. benchstat against Phase 29, 8 runs: `fib` −12% (450 k → 150 k
allocations, 2 per call), `par_fib` −11 to −13%, `records` −3%, allocations −30% geomean, no
regression. The 30% time gate is missed: a profile of `fib(32)` now shows dispatch (`eval`,
`binary`, `arith`, `ident`), which is Phase 32's.

Phase 31 is done: a block or loop scope with no captured binding opens no environment. Its slots
go into the nearest enclosing environment of the same function (`BlockScope.Inline` and `Base`),
siblings share a range, and entering the block resets the range. Locals are not hoisted within a
block, so a plain slot cannot be read before its binding and the reset is only defensive.
`tests/run/loop_scopes.ant` covers fresh bindings per iteration, sibling and nested blocks, and
recursion through a block. benchstat against Phase 30, 8 runs: `map` −11%, `par_map` −11 to
−14%, `records` −8%, `strings` −9%, `array` and `seq` −5%, no regression. A memory profile of
`records` shows `newEnv` only under `frame`.
