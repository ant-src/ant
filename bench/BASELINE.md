# Benchmark baseline

Recorded on 2026-10-22 at the start of PLAN4.md (Phase 28) with
`go test -run '^$' -bench . -benchmem -count 6 ./bench`, medians from `benchstat`.

Machine: Apple M4 Pro, 14 cores (10 performance, 4 efficiency), arm64, macOS; Go 1.27.1.

| Program | Time per run | Bytes per run | Allocations per run | What dominates |
| --- | ---: | ---: | ---: | --- |
| fib | 14.9 ms | 13.8 MiB | 450 k | 75 k calls: 200 ns and 6 allocations per call |
| loop | 157.0 ms | 45.8 MiB | 6.0 M | 2 M iterations: 78 ns and 3 allocations per iteration |
| array | 85.2 ms | 113.3 MiB | 4.7 M | 200 k pushes, 400 k index operations, map and filter |
| map | 61.0 ms | 83.1 MiB | 2.5 M | 100 k sets and gets with template-string keys |
| strings | 20.8 ms | 33.8 MiB | 859 k | 50 k templates, join, 5 k concatenations, split |
| records | 169.4 ms | 144.2 MiB | 7.2 M | 600 k record allocations and 300 k calls |
| closures | 108.4 ms | 160.6 MiB | 6.4 M | 1 M closure calls and 200 k lambda calls |
| seq | 59.4 ms | 70.3 MiB | 2.8 M | 400 k yields through filter and map stages |
| goroutines | 78.0 ms | 25.6 MiB | 856 k | 2 k goroutines, 100 k mutex sections |
| json | 14.7 ms | 22.7 MiB | 496 k | 5 round trips of a 2 k-entry document |
| par_fib n=1 | 116.1 ms | 94.2 MiB | 3.1 M | one goroutine computing fib(28) |
| par_fib n=4 | 217.0 ms | 376.7 MiB | 12.3 M | four, sharing nothing |
| par_fib n=10 | 378.6 ms | 941.8 MiB | 30.9 M | ten: 3.1× over serial, ideal 10× |
| par_map n=1 | 102.2 ms | 130.9 MiB | 3.6 M | one goroutine building a 200 k-entry map |
| par_map n=4 | 143.3 ms | 523.5 MiB | 14.4 M | four, sharing nothing |
| par_map n=10 | 224.7 ms | 1.28 GiB | 36.0 M | ten: 4.5× over serial |

Run-to-run noise on this machine is up to ±9% on the parallel rows, so phase gates compare
`benchstat` medians of at least 10 interleaved runs (`scripts/bench.sh`).

## Reading the numbers

The interpreter does roughly 70 ns of work per statement and allocates about three times per
loop iteration, which is where the time goes: an `int` crossing into a `value.Value` interface
allocates when it is outside Go's small-integer cache, every shared slot store allocates a
cell, and every call allocates its environment and frame. Nothing here is a surprise for a
tree-walking interpreter over interface values.

## Candidates, in order

PLAN2.md allowed at most two optimizations; PLAN4.md lifts the limit and takes these two among
others:

1. **The cell per shared store** (`value.Slot`): a tagged one-word representation for small
   integers and booleans would remove the allocation for the commonest stores. It touches only
   the `value` package.
2. **The environment per call**: `fib` pays six allocations per call; a frame with inline
   storage for a handful of locals would remove most of them.

A change is accepted only with a `benchstat` comparison against this baseline showing the gain
and no regression elsewhere.
