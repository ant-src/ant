<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.reason

```ant signature
func Context.reason(ctx: Context): string | null
```

Context.reason returns why ctx is done, or null while it is not. The reason is "cancelled",
"timeout", "interrupt", or the parent's reason.

<!-- end generated -->

## Details

- A context is done once, so its reason never changes: a cancel after a timeout leaves `"timeout"`.
- A child that its parent ends takes the parent's reason, so the reason says how the work ended, not
  which context ended it.
- The `Err` of a blocking call that a context ends has the context's reason as its message.

## Basic examples

A context is running until something ends it, and its reason is `null` until then:

```ant
let ctx = Context.new()
Context.reason(ctx) // null
ctx.cancel()
Context.reason(ctx) // "cancelled"
```

A context whose timeout passes is done with the reason `"timeout"`:

```ant
let ctx = Context.new(timeout: 50ms)
ctx.wait()
println(Context.reason(ctx))
```

```text
timeout
```

## Scope

### A child of a cancelled parent

Cancelling a parent ends its children with the parent's reason, and a child created under a done
parent is done at once:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
parent.cancel()
child.reason()                                     // "cancelled"
Context.new(timeout: 10s, parent: parent).reason() // "cancelled"
```

### A child of a parent that timed out

The parent's reason reaches every descendant, here a grandchild with no timeout of its own:

```ant
let parent = Context.new(timeout: 50ms)
let grandchild = Context.new(parent: Context.new(parent: parent))
grandchild.wait()
println(grandchild.reason())
```

```text
timeout
```

### An interrupt

The context from [Os.interrupt](../Os/interrupt.md) is done with the reason `"interrupt"` when the
program gets Ctrl-C, SIGINT, or, where the host has it, SIGTERM. Here the program sends SIGTERM to
itself:

```ant
let stop = Os.interrupt()
let work = Context.new(parent: stop)
let signal = try Cmd.new("kill", ["-TERM", "${Os.pid()}"]).start()
try signal.wait()
work.wait()
println(stop.reason())
println(work.reason())
```

```text
interrupt
interrupt
```

### A reason in an error

A blocking call that returns a `Result` gives up with an `Err` whose message is the reason. Here no
client connects before the timeout:

```ant
let listener = try Net.listen("127.0.0.1:0")
defer { listener.close() }
println(listener.accept(ctx: Context.new(timeout: 20ms)))
```

```text
Err{ error: Error{ message: "timeout" } }
```

## Applications

### Tell a close from a timeout

`Channel.receive` returns `null` both when the channel is closed and drained and when the context
ends first. The reason tells them apart:

```ant
func collect(replies: Channel[string], ctx: Context): Result[Array[string], Error] {
  let got: Array[string] = []
  while true {
    let reply = replies.receive(ctx) ?? break
    got.push(reply)
  }
  let reason = ctx.reason()
  if reason != null { return Result.fail("collect: ${reason}") }
  Result.ok(got)
}

let replies: Channel[string] = Channel.new(2)
replies.send("a")
replies.send("b")
replies.close()
println(collect(replies, Context.new(timeout: 1s)))

let silent: Channel[string] = Channel.new()
println(collect(silent, Context.new(timeout: 50ms)))
```

```text
Ok{ value: ["a", "b"] }
Err{ error: Error{ message: "collect: timeout" } }
```

### Report why a worker stopped

A worker checks [Context.isDone](isDone.md) between steps and passes its context into its waits.
When it stops, the reason says whether time ran out or someone cancelled it:

```ant
func work(steps: int, ctx: Context): string {
  var done = 0
  while done < steps && !ctx.isDone() {
    Time.sleep(20ms, ctx)
    done += 1
  }
  let reason = ctx.reason() ?? return "finished ${done} steps"
  "stopped after ${done} steps: ${reason}"
}

println(work(3, Context.new(timeout: 1s)))
let cancelled = Context.new()
cancelled.cancel()
println(work(3, cancelled))
```

```text
finished 3 steps
stopped after 0 steps: cancelled
```

## Properties and relations

A context has a reason exactly when [Context.isDone](isDone.md) reports it done:

```ant
let ctx = Context.new()
assert(ctx.isDone() == (ctx.reason() != null))
ctx.cancel()
assert(ctx.isDone() == (ctx.reason() != null))
```

The first reason stays, whatever ends the context later:

```ant
let ctx = Context.new(timeout: 20ms)
ctx.wait()
ctx.cancel()
assert(ctx.reason() == "timeout")
```

A child's end does not reach its parent:

```ant
let parent = Context.new()
let child = Context.new(parent: parent)
child.cancel()
child.reason()  // "cancelled"
parent.reason() // null
```

## Possible issues

### The reason may be null

The result is `string | null`, so a context that the code knows is done still needs a default or a
check before its reason is a `string`:

```ant
let ctx = Context.new()
ctx.cancel()
let reason: string = ctx.reason() // error: cannot use string | null where string is expected
println(reason)
```

Give it a default with `??`:

```ant
let ctx = Context.new()
ctx.cancel()
ctx.reason() ?? "running" // "cancelled"
```

### A cancelled child looks like a cancelled parent

A child cancelled by its own call and one ended by a cancelled parent both have the reason
`"cancelled"`. Ask the parent when it matters which one ended:

```ant
let parent = Context.new()
let own = Context.new(parent: parent)
own.cancel()
let inherited = Context.new(parent: parent)
parent.cancel()
own.reason() == inherited.reason() // true
```

## See also

[Context.isDone](isDone.md) · [Context.cancel](cancel.md) · [Context.wait](wait.md) ·
[Context.new](new.md) · [Context.deadline](deadline.md) · [Os.interrupt](../Os/interrupt.md) ·
[Context](index.md)
