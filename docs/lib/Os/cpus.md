<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Os.cpus

```ant signature
func Os.cpus(): int
```

Os.cpus returns the number of CPUs the program may use at once: the host's logical CPUs, or fewer
under a container's CPU limit, and at least 1. $GOMAXPROCS, set to a positive number, is the count
instead. It can change while the program runs, as a container's limit can. It is the limit to give
[`Parallel.map`](../Parallel/map.md) for CPU-bound work.

<!-- end generated -->

## Details

- The count is never below 1, so `Os.cpus()` is always a valid `limit` for
  [Parallel.map](../Parallel/map.md) and [Parallel.each](../Parallel/each.md).
- The count can change while the program runs. Call `Os.cpus` where you size the work, rather than
  once at startup.

## Basic examples

Print the number of CPUs the program may use. On a laptop with 8 logical CPUs it prints `8`:

<!-- test: check-only the count depends on the host -->

```ant
println(Os.cpus())
```

Give it to `Parallel.map` as the limit, so at most one callback per CPU runs at once:

```ant
func sumOfSquares(n: int, _: Context): int {
  Array.range(1, n + 1).map({ x * x }).sum() ?? 0
}

let totals = Parallel.map([10, 100, 1000], sumOfSquares, limit: Os.cpus())
println(totals)
```

```text
[385, 338350, 333833500]
```

## Scope

Set the `GOMAXPROCS` environment variable to choose the count, here fewer CPUs:

<!-- test: env GOMAXPROCS=2 -->

```ant
println(Os.cpus())
```

```text
2
```

## Applications

### A worker count the user can override

Read the count from an environment variable, and fall back to the CPUs when it is missing or not a
number. `Math.max` keeps the count a valid limit when the variable holds 0 or a negative number:

<!-- test: env WORKERS=3 -->

```ant
let workers = Math.max(Os.env("WORKERS")?.toInt() ?? Os.cpus(), 1)
println("using ${workers} workers")
```

```text
using 3 workers
```

### One chunk per CPU

Split the input into as many chunks as there are CPUs, and give each chunk its own callback. The
result is the same for any count:

```ant
func chunkSum(xs: Array[int], _: Context): int { xs.sum() ?? 0 }

let numbers = Array.range(1, 1001)
let cpus = Os.cpus()
let size = (numbers.length() + cpus - 1) / cpus
let sums = Parallel.map(numbers.chunk(size), chunkSum, limit: cpus) ?? []
println(sums.sum())
```

```text
500500
```

### Process a stream of jobs

`Parallel.each` takes the same limit for work that arrives on a channel:

```ant
let jobs: Channel[int] = Channel.new(capacity: 3)
go {
  defer { jobs.close() }
  for n in [3, 5, 7] { jobs.send(n) }
}
let done = Atomic.new(0)
Parallel.each(jobs, func(n: int, _: Context) { done.add(n) }, limit: Os.cpus())
println(done.load())
```

```text
15
```

## Properties and relations

The count is at least 1:

```ant
assert(Os.cpus() >= 1)
```

The limit changes how many callbacks run at once, not the result:

```ant
func cube(n: int, _: Context): int { n * n * n }

let xs = [1, 2, 3, 4]
assert(Parallel.map(xs, cube, limit: Os.cpus()) == xs.map({ x * x * x }))
```

## Possible issues

### A limit of zero

`Os.cpus()` is at least 1, but `Os.cpus() - 1` is 0 on a single CPU, and `Parallel.map` panics on a
limit below 1:

<!-- test: env GOMAXPROCS=1 -->

```ant
let xs = [1, 2, 3]
let doubled = Parallel.map(xs, func(n: int, _: Context) { n * 2 }, limit: Os.cpus() - 1)
// panic: Parallel.map needs a limit of at least 1, got 0
```

To leave one CPU free for the rest of the program, keep the result at 1 or more with `Math.max`:

<!-- test: env GOMAXPROCS=1 -->

```ant
let workers = Math.max(Os.cpus() - 1, 1)
println(workers)
```

```text
1
```

### Nested calls

The limit counts per call. A `Parallel.map` with `limit: Os.cpus()` inside another one runs up to
the square of the count at once. Give the inner call a limit of 1, or run it with `Array.map`:

```ant
let rows = [[1, 2], [3, 4], [5, 6]]
let sums = Parallel.map(rows, func(row: Array[int], _: Context): int {
  row.map({ x * 10 }).sum() ?? 0
}, limit: Os.cpus())
println(sums)
```

```text
[30, 70, 110]
```

## See also

[Parallel.map](../Parallel/map.md) · [Parallel.each](../Parallel/each.md) · [Os.arch](arch.md) ·
[Os.env](env.md) · [Os](index.md)
