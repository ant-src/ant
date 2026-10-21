# Benchmark baseline

Recorded on 2026-10-22 with `go test -run '^$' -bench . -benchmem -count=1 ./bench`.

Machine: Apple M4 Pro, 14 cores, arm64, macOS; Go 1.27.1. Commit: the one that adds this file.

| Program | Time per run | Bytes per run | Allocations per run | What dominates |
| --- | ---: | ---: | ---: | --- |
| fib | 15.1 ms | 14.4 MB | 450 k | 75 k calls: 200 ns and 6 allocations per call |
| loop | 143.8 ms | 48.1 MB | 6.0 M | 2 M iterations: 72 ns and 3 allocations per iteration |
| array | 83.1 ms | 118.8 MB | 4.7 M | 200 k pushes, 400 k index operations, map and filter |
| map | 53.7 ms | 87.1 MB | 2.5 M | 100 k sets and gets with template-string keys |
| strings | 16.7 ms | 35.4 MB | 859 k | 50 k templates, join, 5 k concatenations, split |
| records | 144.2 ms | 151.2 MB | 7.2 M | 600 k record allocations and 300 k calls |
| closures | 108.1 ms | 168.5 MB | 6.4 M | 1 M closure calls and 200 k lambda calls |
| seq | 49.8 ms | 73.7 MB | 2.8 M | 400 k yields through filter and map stages |
| goroutines | 78.0 ms | 26.8 MB | 856 k | 2 k goroutines, 100 k mutex sections |
| json | 13.7 ms | 23.6 MB | 476 k | 5 round trips of a 2 k-entry document |

## Reading the numbers

The interpreter does roughly 70 ns of work per statement and allocates about three times per
loop iteration, which is where the time goes: an `int` crossing into a `value.Value` interface
allocates when it is outside Go's small-integer cache, every shared slot store allocates a
cell, and every call allocates its environment and frame. Nothing here is a surprise for a
tree-walking interpreter over interface values.

## Candidates, in order

PLAN2.md allows at most two optimizations, each justified by these numbers:

1. **The cell per shared store** (`value.Slot`): a tagged one-word representation for small
   integers and booleans would remove the allocation for the commonest stores. It touches only
   the `value` package.
2. **The environment per call**: `fib` pays six allocations per call; a frame with inline
   storage for a handful of locals would remove most of them.

A change is accepted only with a `benchstat` comparison against this baseline showing the gain
and no regression elsewhere.
