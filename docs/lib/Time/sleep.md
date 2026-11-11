<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.sleep

```ant signature
func Time.sleep(d: Time.Duration, ctx: Context | null = null): void
```

Time.sleep blocks the calling goroutine for d, or until ctx is done when ctx is not null. Hours and
smaller are exact and a day is 24 hours; a zero or negative d returns at once, and a span beyond
about 292 years is taken as that. It panics if d has years, months or weeks.

<!-- end generated -->

## Details

- Only the calling goroutine waits; other goroutines keep running.
- A zero or negative `d` returns at once.
- With `ctx:`, the sleep ends early when the context is done; check `ctx.isDone()` afterward to tell
  which happened.
- A `d` with years, months or weeks panics, since those have no fixed length. Days count as 24
  hours.

## Basic examples

Pause the script for half a second:

```ant
println("wait")
Time.sleep(Time.duration(milliseconds: 500))
println("go")
```

```text
wait
go
```

## Scope

Any span of hours and smaller units works, combined in one duration:

```ant
Time.sleep(Time.duration(seconds: 1, milliseconds: 250))
println("slept 1.25 s")
```

```text
slept 1.25 s
```

A duration parsed from text works the same way:

```ant
let d = Time.parseDuration("PT0.05S").unwrap()
Time.sleep(d)
println("slept ${d}")
```

```text
slept PT0.05S
```

A zero or negative duration does not wait at all:

```ant
let start = Time.monotonic()
Time.sleep(Time.duration(seconds: -5))
Time.sleep(Time.duration(seconds: 0))
println(Time.monotonic() - start < 100)
```

```text
true
```

A sleep blocks only the goroutine that calls it, so goroutines that sleep run side by side. The one
with the shorter sleep finishes first:

```ant
let done: Channel[string] = Channel.new()
go {
  Time.sleep(Time.duration(milliseconds: 50))
  Channel.send(done, "slow")
}
go {
  Time.sleep(Time.duration(milliseconds: 10))
  Channel.send(done, "fast")
}
println(Channel.receive(done))
println(Channel.receive(done))
```

```text
fast
slow
```

In a `select`, a sleep is a timeout beside the other clauses. A head is written in the namespace
form and is not performed until it is chosen; here the receive never becomes ready in time, so the
sleep's clause wins:

```ant
let results: Channel[string] = Channel.new()
go {
  Time.sleep(Time.duration(seconds: 5))
  Channel.send(results, "done")
}
let msg = select {
  let r = Channel.receive(results) { r ?? "closed" }
  Time.sleep(Time.duration(milliseconds: 20)) { "timed out" }
}
println(msg)
```

```text
timed out
```

## Named arguments

### ctx

With `ctx:`, the sleep ends when the context is done, whichever comes first. A one-hour sleep under
a 20-millisecond timeout ends after 20 milliseconds:

```ant
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
Time.sleep(Time.duration(hours: 1), ctx: ctx)
println(ctx.reason())
```

```text
timeout
```

Another goroutine can wake the sleeper by cancelling the context:

```ant
let ctx = Context.new()
go {
  Time.sleep(Time.duration(milliseconds: 20))
  ctx.cancel()
}
Time.sleep(Time.duration(minutes: 10), ctx: ctx)
println(ctx.reason())
```

```text
cancelled
```

A context that is already done makes the sleep return at once:

```ant
let ctx = Context.new()
ctx.cancel()
Time.sleep(Time.duration(hours: 1), ctx: ctx)
println("returned")
```

```text
returned
```

## Applications

Retry a failing call, doubling the wait after each failure:

```ant
var calls = 0
func fetch(): Result[string, Error] {
  calls += 1
  if calls < 3 { Result.fail("busy") } else { Result.ok("data") }
}

func retry[T](attempts: int, f: func(): Result[T, Error]): Result[T, Error] {
  var wait = Time.duration(milliseconds: 10)
  for attempt in Iter.range(1, attempts) {
    let result = f()
    if result.isOk() { return result }
    println("attempt ${attempt} failed, waiting ${wait}")
    Time.sleep(wait)
    wait = wait.addDuration(wait)
  }
  f()
}

println(retry(5, fetch).unwrap())
```

```text
attempt 1 failed, waiting PT0.01S
attempt 2 failed, waiting PT0.02S
data
```

Poll for a condition, but give up when a context runs out. The sleep takes the same context, so the
loop never waits past the deadline:

```ant
func ready(): bool { false }

let ctx = Context.new(timeout: Time.duration(milliseconds: 100))
while !ready() && !ctx.isDone() {
  Time.sleep(Time.duration(milliseconds: 30), ctx: ctx)
}
println(if ready() { "ready" } else { "gave up: ${ctx.reason()}" })
```

```text
gave up: timeout
```

## Properties and relations

A sleep lasts at least `d`, as [Time.monotonic](monotonic.md) measures it:

```ant
let start = Time.monotonic()
Time.sleep(Time.duration(milliseconds: 30))
assert(Time.monotonic() - start >= 30)
```

A sleep with a context returns once the context is done, well before `d`:

```ant
let ctx = Context.new(timeout: Time.duration(milliseconds: 20))
let start = Time.monotonic()
Time.sleep(Time.duration(seconds: 30), ctx: ctx)
assert(ctx.isDone())
assert(Time.monotonic() - start < 30_000)
```

## Possible issues

`Time.sleep` takes a `Time.Duration`, not a number of milliseconds:

```ant
Time.sleep(1000) // error: cannot use int where Time.Duration is expected
```

Write the unit with [Time.duration](duration.md):

```ant
Time.sleep(Time.duration(milliseconds: 1000))
```

Months, years and weeks have no fixed length, so a sleep that has them panics, even though the
checker accepts it:

```ant
Time.sleep(Time.duration(weeks: 1))
// panic: Time.sleep needs a duration without years, months or weeks, which have no fixed length
```

Use days or hours instead, which `Time.sleep` takes as fixed spans:

<!-- test: check-only it waits a week -->

```ant
Time.sleep(Time.duration(days: 7))
```

A `Time.sleep` head in a `select` takes no context, since the context is a clause of its own:

```ant
let c: Channel[int] = Channel.new()
let ctx = Context.new()
let msg = select {
  let v = Channel.receive(c) { "got ${v}" }
  Time.sleep(Time.duration(milliseconds: 20), ctx: ctx) { "timed out" } // error: a select clause's head takes no context: the context is a clause of its own, Context.wait(ctx) { ... }
}
```

Write the context as its own [Context.wait](../Context/wait.md) clause:

```ant
let c: Channel[int] = Channel.new()
let ctx = Context.new()
ctx.cancel()
let msg = select {
  let v = Channel.receive(c) { "got ${v}" }
  Time.sleep(Time.duration(seconds: 5)) { "timed out" }
  Context.wait(ctx) { "stopped: ${ctx.reason()}" }
}
println(msg)
```

```text
stopped: cancelled
```

## See also

[Time.ticker](ticker.md) · [Time.duration](duration.md) · [Time.monotonic](monotonic.md) ·
[Context.new](../Context/new.md) · [Context.wait](../Context/wait.md) ·
[Time.Duration](types/Duration.md)
