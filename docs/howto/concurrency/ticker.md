<!-- status: reviewed -->

# Run something periodically

Run a piece of work on a fixed beat, such as a report every minute, and stop it with a context. This
program runs every 20 milliseconds and stops after three runs:

```ant
func report(run: int) {
  println("report ${run}")
}

let ctx = Context.new()
defer { ctx.cancel() }
let ticks = Time.ticker(20ms, ctx)
var runs = 0
for _ in ticks.values(ctx) {
  runs += 1
  report(runs)
  if runs == 3 { ctx.cancel() }
}
println("stopped after ${runs} runs")
```

```text
report 1
report 2
report 3
stopped after 3 runs
```

1. Create a context with [Context.new](../../lib/Context/new.md), and cancel it with `defer` so the
   ticker stops however the block that owns it ends, an early `return` included; a ticker left
   running keeps ticking until the run ends. Pass a timeout, such as `Context.new(1h)`, to stop at a
   deadline instead.
2. Start the ticker with [Time.ticker](../../lib/Time/ticker.md), passing the period and the
   context. The first tick comes one period after the call.
3. Loop over the ticks with [Channel.values](../../lib/Channel/values.md), passing the same context.
   The ticker never closes its channel, so the context is what ends the loop. Each tick is the time
   it was sent, a `Time.Instant`; bind it, as in `for at in ticks.values(ctx)`, to use it in the
   work.
4. Do the work in the loop body. Cancel the context with
   [Context.cancel](../../lib/Context/cancel.md) to stop.

The beat is counted from the call, so time spent in the body does not shift it. When a run takes
longer than the period, the ticker holds one tick and drops the rest: the next run starts at once,
and the missed ones do not pile up.

## Run once at the start

The first tick comes one period after the call. To run at once and then on every tick, call the work
before the loop. Here an hourly sync runs at the start and stops at Ctrl-C, which
[Os.interrupt](../../lib/Os/interrupt.md) turns into a done context:

<!-- test: before
go {
  Time.sleep(100ms)
  Cmd.read(Cmd.new("kill", ["-INT", "${Os.pid()}"])).unwrap()
}
-->

```ant
func sync() {
  println("synced")
}

let stop = Os.interrupt()
let ticks = Time.ticker(1h, stop)
sync()
for _ in ticks.values(stop) {
  sync()
}
println("stopped")
```

```text
synced
stopped
```

## Tick beside other work

To do something on every tick while waiting for other events, receive the ticks in a `select` beside
the other channels. Here a progress line prints every 50 milliseconds while a worker sends its
results:

<!-- test: check-only how many progress lines print depends on timing -->

```ant
let jobs = ["a.csv", "b.csv", "c.csv", "d.csv"]
let results: Channel[string] = Channel.new()
go {
  for job in jobs {
    Time.sleep(30ms)
    results.send(job)
  }
  results.close()
}

let ctx = Context.new()
defer { ctx.cancel() }
let ticks = Time.ticker(50ms, ctx)
var done = 0
while true {
  select {
    let job = Channel.receive(results) {
      if job == null { break }
      done += 1
    }
    Channel.receive(ticks) { println("${done} of ${jobs.length()} done") }
  }
}
println("all ${done} done")
```

It prints lines such as `1 of 4 done` and `3 of 4 done`, then `all 4 done`.

1. Make the results channel with [Channel.new](../../lib/Channel/new.md), and close it with
   [Channel.close](../../lib/Channel/close.md) when the worker finishes.
2. In a `select`, receive from the results with [Channel.receive](../../lib/Channel/receive.md) and
   leave the loop when the receive gives `null`, which means the channel is closed.
3. In the same `select`, receive from the ticker and do the periodic work.

## Wait between runs instead

When the pause should come after each run, not on a fixed beat, use
[Time.sleep](../../lib/Time/sleep.md) in the loop instead of a ticker. Pass the context so the sleep
ends early when it is done:

```ant
let ctx = Context.new()
var runs = 0
while !ctx.isDone() {
  runs += 1
  println("run ${runs}")
  if runs == 3 { ctx.cancel() }
  Time.sleep(20ms, ctx: ctx)
}
```

```text
run 1
run 2
run 3
```

Each run then starts a full period after the previous one ends, so a slow run delays the next.
[Context.isDone](../../lib/Context/isDone.md) checks the context between runs.
