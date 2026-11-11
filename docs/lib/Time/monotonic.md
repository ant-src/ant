<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.monotonic

```ant signature
func Time.monotonic(): int
```

Time.monotonic returns milliseconds from an arbitrary origin, for measuring elapsed time. Only the
difference of two readings means anything. Changes to the wall clock do not affect it, and a later
reading in the same run, on any goroutine, is never smaller than an earlier one.

<!-- end generated -->

## Details

- Subtract two readings taken in the same run to get the milliseconds between them.
- The result is an `int` of whole milliseconds, not a `Time.Duration`. Wrap a difference in
  `Time.duration(milliseconds: ...)` to pass or print it as a duration.
- Anything shorter than a millisecond can measure as `0`. Time many repetitions and divide.

## Basic examples

Take a reading before the work and one after, and subtract:

<!-- test: check-only the elapsed time changes from run to run -->

```ant
let start = Time.monotonic()
Time.sleep(Time.duration(milliseconds: 50))
println("slept ${Time.monotonic() - start} ms")
```

It prints a line such as `slept 51 ms`.

## Scope

The difference is an `int`, so integer arithmetic and comparison apply to it directly:

<!-- test: check-only the elapsed time changes from run to run -->

```ant
let start = Time.monotonic()
let squares = Iter.range(0, 100_000).toArray().map({ x * x })
let elapsed = Time.monotonic() - start
println("${squares.length()} squares")
if elapsed > 100 {
  println("slow: ${elapsed} ms")
}
```

Turn a difference into a `Time.Duration` with `Time.duration`:

```ant
let elapsed = 1500
Time.duration(milliseconds: elapsed) // PT1.5S
```

## Applications

Time any function and pass its result through:

<!-- test: check-only the elapsed time changes from run to run -->

```ant
func timed[T](label: string, f: func(): T): T {
  let start = Time.monotonic()
  let result = f()
  println("${label}: ${Time.monotonic() - start} ms")
  result
}

let total = timed("sum", { Iter.range(0, 1_000_000).toArray().sum() })
println(total)
```

Poll until a time budget runs out. A reading plus a span gives a deadline in the same units:

<!-- test: check-only the number of polls depends on timing -->

```ant
let deadline = Time.monotonic() + 200
var polls = 0
while Time.monotonic() < deadline {
  polls += 1
  Time.sleep(Time.duration(milliseconds: 50))
}
println("gave up after ${polls} polls")
```

Measure an operation too fast for one reading by timing many runs:

<!-- test: check-only the elapsed time changes from run to run -->

```ant
let runs = 100_000
let start = Time.monotonic()
for _ in Iter.range(0, runs) {
  ["a", "b"].map({ x.toUpper() })
}
let total = Time.monotonic() - start
println("${float64(total) * 1000.0 / float64(runs)} µs per call")
```

## Possible issues

The difference is an `int`, not a `Time.Duration`, so it does not pass where a duration is expected:

```ant
let start = Time.monotonic()
Time.sleep(Time.monotonic() - start) // error: cannot use int where Time.Duration is expected
```

Wrap it in `Time.duration` first. Here the loop sleeps for what is left of a 100 ms frame; a frame
that ran over gives a negative sleep, which returns at once:

```ant
for frame in Iter.range(0, 3) {
  let start = Time.monotonic()
  println("frame ${frame}")
  let left = 100 - (Time.monotonic() - start)
  Time.sleep(Time.duration(milliseconds: left))
}
```

```text
frame 0
frame 1
frame 2
```

A reading is not a timestamp. Its origin is not the Unix epoch, so `Time.fromEpochMillis` turns it
into an instant that means nothing. Use [Time.now](now.md) for the current time:

<!-- test: check-only the reading changes from run to run -->

```ant
println(Time.fromEpochMillis(Time.monotonic()))
println(Time.now())
```

## See also

[Time.now](now.md) · [Time.sleep](sleep.md) · [Time.duration](duration.md) ·
[Time.epochMillis](epochMillis.md) · [Time.Duration](types/Duration.md)
