<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Time.ticker

```ant signature
func Time.ticker(d: Time.Duration, ctx: Context | null = null): Channel[Time.Instant]
```

Time.ticker returns a channel that receives the current instant every d, counted from the call. It
holds one tick and drops a tick that finds it full, so a slow reader gets one tick, not a burst. It
stops when ctx is done, or when the run ends if ctx is null, and is never closed. It panics if d is
zero or negative or has years, months or weeks.

<!-- end generated -->

## Details

- The first tick comes `d` after the call, not at once. The beat is counted from the call, so time
  the reader spends between receives does not shift it.
- The channel is never closed. A receive, a `for` over `Channel.values`, or a `select` clause on it
  waits for the next tick, so give the wait a context or another clause that ends it.
- Once `ctx` is done, no receive gets a tick, not even the one the channel held.
- Only the ticker sends on the channel: `Channel.send` and `Channel.close` on it panic.

## Basic examples

Receive three ticks, ten milliseconds apart, then stop the ticker:

```ant
let ctx = Context.new()
let ticks = Time.ticker(Time.duration(milliseconds: 10), ctx)
for i in Iter.range(1, 4) {
  ticks.receive()
  println("tick ${i}")
}
ctx.cancel()
```

```text
tick 1
tick 2
tick 3
```

Each tick is the instant it was sent:

<!-- test: check-only the instants change from run to run -->

```ant
let ticks = Time.ticker(Time.duration(seconds: 1))
println(ticks.receive())
```

It waits a second and prints an instant such as `2026-11-11T21:27:08.050436Z`.

## Scope

Read ticks with `Channel.values`. The loop ends when its context is done, or here when it breaks:

```ant
let ctx = Context.new(Time.duration(seconds: 1))
let ticks = Time.ticker(Time.duration(milliseconds: 10), ctx)
var polls = 0
for at in ticks.values(ctx) {
  polls += 1
  if polls == 3 { break }
}
println(polls)
```

```text
3
```

A ticker is a `select` clause like any channel. Here it beats beside a worker's results and prints
its progress until the results run out:

<!-- test: check-only how many progress lines print depends on timing -->

```ant
let jobs = ["parse", "check", "build", "test"]
let results: Channel[string] = Channel.new()
go {
  for job in jobs {
    Time.sleep(Time.duration(milliseconds: 20))
    results.send(job)
  }
  results.close()
}

let ctx = Context.new()
defer { ctx.cancel() }
let ticks = Time.ticker(Time.duration(milliseconds: 30), ctx)
var done = 0
while true {
  select {
    var job = Channel.receive(results) {
      if job == null { break }
      done += 1
    }
    Channel.receive(ticks) { println("${done} of ${jobs.length()} done") }
  }
}
println("all ${jobs.length()} done")
```

It prints lines such as `1 of 4 done` and `2 of 4 done`, then `all 4 done`.

## Named arguments

### ctx

Without `ctx`, the ticker runs until the run ends. With it, the ticker stops when the context is
done, and drops the tick it held:

```ant
let ctx = Context.new()
let ticks = Time.ticker(Time.duration(milliseconds: 10), ctx)
Time.sleep(Time.duration(milliseconds: 50))
println(ticks.length())
ctx.cancel()
println(ticks.length())
```

```text
1
0
```

## Applications

Limit a loop to one step per tick, such as one request every 20 milliseconds:

```ant
let ctx = Context.new()
defer { ctx.cancel() }
let ticks = Time.ticker(Time.duration(milliseconds: 20), ctx)
for page in ["/", "/docs", "/blog"] {
  ticks.receive()
  println("fetch ${page}")
}
```

```text
fetch /
fetch /docs
fetch /blog
```

Poll a condition on every tick until it holds or a timeout passes:

```ant
func waitFor(ready: func(): bool, timeout: Time.Duration): bool {
  let ctx = Context.new(timeout)
  defer { ctx.cancel() }
  let ticks = Time.ticker(Time.duration(milliseconds: 10), ctx)
  while !ready() {
    select {
      Context.wait(ctx) { return false }
      Channel.receive(ticks) {}
    }
  }
  true
}

let start = Time.monotonic()
println(waitFor({ Time.monotonic() - start > 30 }, Time.duration(seconds: 1)))
println(waitFor({ false }, Time.duration(milliseconds: 50)))
```

```text
true
false
```

## Properties and relations

The first tick comes no sooner than `d` after the call, and each tick is later than the one before:

```ant
let d = Time.duration(milliseconds: 20)
let start = Time.now()
let ticks = Time.ticker(d)
let first = ticks.receive() ?? start
let second = ticks.receive() ?? start
assert(first.compare(start.add(d)) >= 0)
assert(second.compare(first) > 0)
```

A reader slower than the beat finds one tick waiting, not one for every period it missed:

```ant
let ctx = Context.new()
let ticks = Time.ticker(Time.duration(milliseconds: 5), ctx)
Time.sleep(Time.duration(milliseconds: 50))
assert(ticks.length() == 1)
ctx.cancel()
```

## Possible issues

The period must be above zero:

```ant
let ticks = Time.ticker(Time.duration(seconds: 0))
// panic: Time.ticker needs a period above zero, got PT0S
```

Years, months and weeks have no fixed length, so a period with them panics. Write the period in days
or smaller units:

```ant
let ticks = Time.ticker(Time.duration(weeks: 1))
// panic: Time.ticker needs a duration without years, months or weeks, which have no fixed length
```

The channel only receives, so sending on it panics:

```ant
let ticks = Time.ticker(Time.duration(seconds: 1))
let sent = ticks.send(Time.now()) // panic: Channel.send: a ticker's channel only receives
```

`Channel.close` on it panics the same way.

The channel is never closed, so a receive on a stopped ticker waits forever. Give the receive a
context, which makes it return `null`:

```ant
let ctx = Context.new()
let ticks = Time.ticker(Time.duration(milliseconds: 10), ctx)
ctx.cancel()
let tick = ticks.receive(Context.new(Time.duration(milliseconds: 50)))
println(tick)
```

```text
null
```

## See also

[Time.sleep](sleep.md) · [Time.now](now.md) · [Time.duration](duration.md) ·
[Channel.receive](../Channel/receive.md) · [Channel.values](../Channel/values.md) ·
[Context.new](../Context/new.md) · [Time.Duration](types/Duration.md)
