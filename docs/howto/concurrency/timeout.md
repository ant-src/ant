<!-- status: reviewed -->

# Give up after a timeout or deadline

Stop waiting once a time limit passes, and report why.

1. Make a context with [Context.new](../../lib/Context/new.md), giving it the time limit as a
   duration.
2. Pass the context as `ctx:` to every blocking call that should give up: a channel send or receive,
   [Time.sleep](../../lib/Time/sleep.md), and the calls of `Http`, `Net`, `File`, and `Cmd`.
3. Handle the giving up. A call that returns a `Result` returns an `Err` whose message is the
   context's reason, a call that returns a value or `null` returns `null`, a `void` call returns
   early, and `Channel.send` returns `false`.
4. Read [Context.reason](../../lib/Context/reason.md) to say why the work stopped.

## Wait for a reply, at most a while

[Channel.receive](../../lib/Channel/receive.md) returns `null` when the context ends first. Turn
that into an `Err` that names the reason:

```ant
func lookup(name: string, ctx: Context): Result[string, Error] {
  let replies: Channel[string] = Channel.new(1)
  go {
    Time.sleep(200ms)
    replies.send("${name}: 10.0.0.7")
  }
  let reply = replies.receive(ctx: ctx) ?? return Result.fail("lookup ${name}: ${ctx.reason()}")
  Result.ok(reply)
}

println(lookup("db", Context.new(2s)))
println(lookup("db", Context.new(50ms)))
```

```text
Ok{ value: "db: 10.0.0.7" }
Err{ error: Error{ message: "lookup db: timeout", kind: null } }
```

The context does not stop the goroutine that sends. The buffer of one lets its send finish when
nobody receives.

`null` also means the channel closed; when both can happen, ask
[Context.isDone](../../lib/Context/isDone.md) to tell them apart.

## Wait on several channels, at most a while

In a [`select`](../../language/keywords/select.md), give the context its own clause with
[Context.wait](../../lib/Context/wait.md). It runs when the context ends before any other head can
proceed:

```ant
let lines: Channel[string] = Channel.new()
let t = Context.new(50ms)
select {
  let line = Channel.receive(lines) { println(line ?? "closed") }
  Context.wait(t) { println("gave up, ${t.reason()}") }
}
```

```text
gave up, timeout
```

## Bound a call that returns a Result

A call that returns a `Result` gives up with an `Err` whose message is the reason. Add what the call
was doing with [Result.wrapErr](../../lib/Result/wrapErr.md). Here a local server takes a second to
answer `/report`:

```ant
let listener = try Net.listen("127.0.0.1:0")
go {
  _ = Http.serve(listener, {
    Time.sleep(1s)
    Http.Reply{ status: 200, headers: Map{}, body: "report" }
  })
}
let url = "http://${Net.address(listener)}/report"

let ctx = Context.new(100ms)
println(Http.get(url, ctx: ctx) | Result.wrapErr("get report"))
```

```text
Err{ error: Error{ message: "get report: timeout", kind: null } }
```

For one request on its own, [Http.get](../../lib/Http/get.md) also takes `timeout:`, which gives the
same `Err`. Use a context when the limit covers other work too; when both are given, the earlier one
wins.

## Share one budget across several steps

Pass one context to every step, so that the limit covers them all. A `void` call such as
`Time.sleep` returns early, so ask [Context.isDone](../../lib/Context/isDone.md) after it:

```ant
let ctx = Context.new(380ms)
for step in ["download", "unpack", "install"] {
  Time.sleep(150ms, ctx: ctx)
  if ctx.isDone() {
    println("${step}: gave up, ${ctx.reason()}")
    break
  }
  println("${step}: done")
}
```

```text
download: done
unpack: done
install: gave up, timeout
```

## Give one step a shorter limit

Make a child context with `parent:`. It ends at its own limit or when its parent ends, whichever
comes first, and its end leaves the parent running:

```ant
let replies: Channel[string] = Channel.new()
let overall = Context.new(5s)
let cache = Context.new(50ms, parent: overall)
let cached = replies.receive(ctx: cache) ?? "nothing"
println("cache: ${cached}, ${cache.reason()}")
println("overall done: ${overall.isDone()}")
```

```text
cache: nothing, timeout
overall done: false
```

## Give up at a deadline

`Context.new` takes a duration. For a deadline at a fixed instant, take the time left with
[Time.until](../../lib/Time/until.md). A duration that is zero or negative makes `Context.new`
panic, so check a deadline that has passed first, with [Time.sign](../../lib/Time/sign.md):

```ant
func contextUntil(deadline: Time.Instant): Result[Context, Error] {
  let left = Time.until(Time.now(), deadline)
  if left.sign() <= 0 {
    return Result.fail("the deadline ${deadline} has passed")
  }
  Result.ok(Context.new(left))
}

let soon = try contextUntil(Time.now().add(50ms))
soon.wait()
println(soon.reason())

let past = try Time.parseInstant("2020-01-01T00:00:00Z")
println(contextUntil(past))
```

```text
timeout
Err{ error: Error{ message: "the deadline 2020-01-01T00:00:00Z has passed", kind: null } }
```

[Context.deadline](../../lib/Context/deadline.md) gives the instant back, the earliest of the
context and its parents.

## Tell a timeout from Ctrl-C

Make the context a child of [Os.interrupt](../../lib/Os/interrupt.md) to stop at the time limit or
at Ctrl-C, whichever comes first. The reason says which one it was:

```ant
let ctx = Context.new(100ms, parent: Os.interrupt())
ctx.wait()
let message = if ctx.reason() == "interrupt" { "stopped by the user" } else { "out of time" }
println(message)
```

```text
out of time
```

## See also

[Context.new](../../lib/Context/new.md) · [Context.reason](../../lib/Context/reason.md) ·
[Context.isDone](../../lib/Context/isDone.md) · [Context.wait](../../lib/Context/wait.md) ·
[Context.deadline](../../lib/Context/deadline.md) · [Channel.receive](../../lib/Channel/receive.md)
· [Time.sleep](../../lib/Time/sleep.md) · [Time.until](../../lib/Time/until.md) ·
[Time.sign](../../lib/Time/sign.md) · [Time.now](../../lib/Time/now.md) ·
[Time.add](../../lib/Time/add.md) · [Time.parseInstant](../../lib/Time/parseInstant.md) ·
[Channel.new](../../lib/Channel/new.md) · [Http.get](../../lib/Http/get.md) ·
[Http.serve](../../lib/Http/serve.md) · [Net.listen](../../lib/Net/listen.md) ·
[Net.address](../../lib/Net/address.md) · [Result.wrapErr](../../lib/Result/wrapErr.md) ·
[Os.interrupt](../../lib/Os/interrupt.md) · [Cancel a group of goroutines](cancel.md) ·
[Take the first result of several](first-result.md) ·
[Stop or time out a process](../processes/stop.md)
