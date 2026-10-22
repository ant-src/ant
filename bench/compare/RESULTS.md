# ant versus antc

Recorded on 2026-10-22 with `./run.sh` (hyperfine, wall time per process, mean of 15 runs, or of 5
for the parallel sweep). Machine: Apple M4 Pro, 10 performance and 4 efficiency cores, macOS;
Go 1.27.1; clang -O2 with Boehm GC from Homebrew.

- **ant**: this repository's tree-walking interpreter in Go. Goroutines are Go goroutines and run
  in parallel; there is no global lock. Each run includes parsing and checking.
- **antc**: the self-hosted compiler on the `antc` branch of ant-spec (`make fixpoint`), which
  emits C over a tagged-value runtime. Goroutines are threads behind one global lock (GIL), so
  only one runs Ant code at a time. Times are for the built binary; compile time is excluded.

The programs are `bench/*.ant` with two dialect changes that antc needs: `for _ in` is `for i in`
and `func Point.add` is a plain function. `par_fib` and `par_map` are new: N goroutines that
share nothing, N from the command line. Every program prints the same output under both.

## Single goroutine

| Program | ant | antc | ant / antc |
| --- | ---: | ---: | ---: |
| fib | 18.6 ms | 3.3 ms | 5.7× |
| loop | 133.6 ms | 12.7 ms | 10.5× |
| array | 90.9 ms | 12.1 ms | 7.5× |
| map | 57.4 ms | 16.4 ms | 3.5× |
| strings | 20.7 ms | 15.3 ms | 1.4× |
| records | 120.0 ms | 34.4 ms | 3.5× |
| closures | 122.4 ms | 15.2 ms | 8.1× |
| seq | 44.5 ms | 9.4 ms | 4.7× |
| json | 16.6 ms | 12.3 ms | 1.4× |
| goroutines | 90.1 ms | 56.5 ms | 1.6× |

Compiled code beats tree walking by 4 to 10× where the interpreter does the work (calls, loops,
closures). The gap narrows to 1.4× where both sit in library code (strings, json).
`goroutines` (2 k goroutines, 100 k contended mutex sections) is closer than its CPU suggests:
ant burns 247 ms of user time to antc's 21 ms, because it spins on the contended mutex across
cores while antc hands one lock around.

## Parallel scaling

Wall time for N goroutines, each doing the same independent work: `par_fib` computes `fib(28)`,
`par_map` builds and reads back a 200 k-entry map with string keys.

| N | ant par_fib | antc par_fib | ant par_map | antc par_map |
| ---: | ---: | ---: | ---: | ---: |
| 1 | 103 ms | 11 ms | 93 ms | 27 ms |
| 2 | 120 ms | 16 ms | 113 ms | 57 ms |
| 4 | 179 ms | 29 ms | 140 ms | 111 ms |
| 8 | 343 ms | 56 ms | 200 ms | 222 ms |
| 10 | 435 ms | 70 ms | 234 ms | 279 ms |
| 14 | 532 ms | 99 ms | 305 ms | 398 ms |

antc behaves as a GIL predicts: user time equals wall time and wall time grows linearly in N.

ant runs in parallel (14 goroutines of `par_fib` use 3.5 s of CPU in 0.53 s) but scales poorly,
only 2.7× over serial on 14 cores. That is not the interpreter: a CPU profile of `par_fib` with
8 goroutines puts interpreter code at about 7% of samples. The rest is the Go runtime:
`runtime.lock2` on the heap lock (46%), `madvise` while allocating spans (24%) and scheduler
wake-ups. The interpreter allocates about 6 times per call (see `../BASELINE.md`), the live heap
stays tiny, and so the GC runs about 300 cycles in one 350 ms run. A larger GC target shows it:

| GOGC | par_fib N = 1 | N = 8 | N = 14 |
| ---: | ---: | ---: | ---: |
| 100 (default) | 108 ms | 343 ms | 531 ms |
| 400 | 110 ms | 202 ms | 334 ms |
| 1600 | 115 ms | 183 ms | 263 ms |

So ant passes antc on allocation-heavy parallel work (`par_map` at N ≥ 8) despite being 3.5× slower
serially. On call-heavy work, antc's 9× single-thread lead is more than ant's parallelism gains
back at 14 cores. Cutting the per-call allocations, the two candidates in `../BASELINE.md`, is
what would let ant's lack of a GIL pay off. A higher default GC target (`debug.SetGCPercent` or a
memory limit in `cmd/ant`) is a cheap partial fix.

## After PLAN4.md, Phases 28 to 34

Rerun on 2026-10-22 on the `perf` branch, same machine and method. `ant run` now sets
`GOGC=400` (Phase 29).

| Program | ant before | ant now | antc | ant / antc now |
| --- | ---: | ---: | ---: | ---: |
| fib | 18.6 ms | 18.3 ms | 3.6 ms | 5.1× |
| loop | 133.6 ms | 128.1 ms | 14.2 ms | 9.0× |
| array | 90.9 ms | 74.7 ms | 13.3 ms | 5.6× |
| map | 57.4 ms | 44.5 ms | 19.4 ms | 2.3× |
| strings | 20.7 ms | 16.1 ms | 16.5 ms | 1.0× |
| records | 120.0 ms | 97.7 ms | 33.3 ms | 2.9× |
| closures | 122.4 ms | 98.4 ms | 16.2 ms | 6.1× |
| seq | 44.5 ms | 34.4 ms | 10.0 ms | 3.5× |
| json | 16.6 ms | 16.3 ms | 14.0 ms | 1.2× |
| goroutines | 90.1 ms | 71.2 ms | 55.2 ms | 1.3× |

The geometric mean against antc went from 3.8× to 3.0×. Times per process include about 3.7 ms
of startup, which is a fifth of `fib`'s run; in-process, `fib` went from 14.9 ms to 13.4 ms.

| N | ant par_fib before | now | antc | ant par_map before | now | antc |
| ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| 1 | 103 ms | 98 ms | 10 ms | 93 ms | 61 ms | 28 ms |
| 2 | 120 ms | 105 ms | 17 ms | 113 ms | 63 ms | 57 ms |
| 4 | 179 ms | 125 ms | 31 ms | 140 ms | 75 ms | 124 ms |
| 8 | 343 ms | 186 ms | 59 ms | 200 ms | 92 ms | 236 ms |
| 10 | 435 ms | 204 ms | 73 ms | 234 ms | 106 ms | 291 ms |
| 14 | 532 ms | 298 ms | 104 ms | 305 ms | 138 ms | 407 ms |

`par_fib` with 10 goroutines is now 4.8 times faster than serial (was 2.4×), and `par_map` beats
antc from 4 goroutines up, by 2.7× at 10. On call-heavy parallel work antc still leads: its
single-core lead is 9.6× on `fib(28)` and ant's parallelism makes up about 3.4× of it.

What is left is dispatch: `loop` runs at about 64 ns per iteration, and the Phase 34 experiment
shows that removing every boxed int would buy about 1.15× of it. The tree walker's own overhead
(a type switch per node, a `(value, ctl)` pair per result) is the next lever: PLAN4.md section 3.8,
closure compilation.
