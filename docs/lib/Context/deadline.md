<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Context.deadline

```ant signature
func Context.deadline(ctx: Context): Time.Instant | null
```

Context.deadline returns the earliest deadline of ctx and its parents, or null if none has one.

<!-- end generated -->

## Details

- A child never reports a later deadline than its parent: with a longer timeout, it reports the
  parent's.
- The deadline stays after the context is done, so it does not tell whether the context is done; use
  [Context.isDone](isDone.md) for that.

## Basic examples

A context with a timeout has a deadline, the moment the timeout passes:

```ant
let ctx = Context.new(timeout: 2s)
Context.deadline(ctx) is Time.Instant // true
```

A context without a timeout, and without a parent that has one, has none:

```ant
let ctx = Context.new()
println(Context.deadline(ctx))
```

```text
null
```

## Scope

The deadline lies the timeout after the call to [Context.new](new.md):

```ant
let before = Time.now()
let ctx = Context.new(timeout: 1h)
let deadline = Context.deadline(ctx) ?? return
let timeout = Time.until(before, deadline, largestUnit: Time.Hour{})
println(Time.roundDuration(timeout, smallestUnit: Time.Minute{}))
```

```text
PT1H
```

A child without a timeout reports its parent's deadline:

```ant
let request = Context.new(timeout: 5s)
let step = Context.new(parent: request)
step.deadline() == request.deadline() // true
```

A child with a shorter timeout has its own, earlier deadline:

```ant
let request = Context.new(timeout: 1h)
let step = Context.new(timeout: 1s, parent: request)
println(step.deadline() == request.deadline())
let stepAt = step.deadline() ?? return
let requestAt = request.deadline() ?? return
println(Time.compare(stepAt, requestAt))
```

```text
false
-1
```

A child with a longer timeout keeps its parent's deadline, the earlier one:

```ant
let request = Context.new(timeout: 1s)
let step = Context.new(timeout: 1h, parent: request)
step.deadline() == request.deadline() // true
```

The deadline reaches every generation:

```ant
let outer = Context.new(timeout: 30s)
let middle = Context.new(parent: outer)
let inner = Context.new(parent: middle)
inner.deadline() == outer.deadline() // true
```

A parent without a deadline gives its child none:

```ant
let parent = Context.new()
println(Context.new(parent: parent).deadline())
```

```text
null
```

A cancelled context keeps its deadline:

```ant
let ctx = Context.new(timeout: 1h)
ctx.cancel()
ctx.isDone()           // true
ctx.deadline() != null // true
```

So does one whose deadline has passed:

```ant
let ctx = Context.new(timeout: 20ms)
ctx.wait()
println(ctx.reason())
println(ctx.deadline() != null)
```

```text
timeout
true
```

## Applications

### Skip work that cannot finish in time

Compare the time left with what a step needs, and skip it before starting. A context without a
deadline runs every step:

```ant
func runStep(name: string, needs: Time.Duration, ctx: Context): void {
  let deadline = ctx.deadline()
  if deadline != null && Time.compareDuration(Time.until(Time.now(), deadline), needs) < 0 {
    println("skip ${name}: not enough time")
    return
  }
  println("run ${name}")
}

let ctx = Context.new(timeout: 1s)
runStep("index", 100ms, ctx)
runStep("reindex", 1h, ctx)
runStep("reindex", 1h, Context.new())
```

```text
run index
skip reindex: not enough time
run reindex
```

### Tell a server how long to try

Turn the time left into whole seconds, such as for a timeout header on a request, and send nothing
when there is no deadline:

```ant
func timeoutHeader(ctx: Context): string | null {
  let deadline = ctx.deadline() ?? return null
  let left = Time.until(Time.now(), deadline)
  "${int(Math.ceil(left.total(Time.Second{})))}"
}

println(timeoutHeader(Context.new(timeout: 2s)))
println(timeoutHeader(Context.new()))
```

```text
2
null
```

## Properties and relations

A child's deadline is never after its parent's:

```ant
let parent = Context.new(timeout: 10s)
let parentAt = parent.deadline() ?? return
for timeout in [1s, 10s, 1m] {
  let childAt = Context.new(timeout: timeout, parent: parent).deadline() ?? return
  assert(Time.compare(childAt, parentAt) <= 0)
}
```

The deadline of a context made with a timeout lies between the times read just before and just after
the call plus the timeout:

```ant
let before = Time.now()
let ctx = Context.new(timeout: 5s)
let after = Time.now()
let deadline = ctx.deadline() ?? return
assert(Time.compare(deadline, before.add(5s)) >= 0)
assert(Time.compare(deadline, after.add(5s)) <= 0)
```

## Possible issues

### The result can be `null`

`Context.deadline` returns `null` when no context in the chain has a timeout, so the result is not a
`Time.Instant` until the `null` is handled:

```ant
let ctx = Context.new(timeout: 1s)
let at: Time.Instant = ctx.deadline()
// error: cannot use Time.Instant | null where Time.Instant is expected
```

Handle it with `??` or a guard:

```ant
let ctx = Context.new(timeout: 1s)
let at = ctx.deadline() ?? return
println(Time.compare(at, Time.now()))
```

```text
1
```

### Durations do not compare with `<`

The time left is a `Time.Duration`, which `<` does not order:

```ant
func isShort(ctx: Context): bool {
  let deadline = ctx.deadline() ?? return false
  let left = Time.until(Time.now(), deadline)
  left < 100ms
  // error: < applies to a single numeric kind, string, or a named primitive type over one, not to Time.Duration
}
```

Use [Time.compareDuration](../Time/compareDuration.md):

```ant
let ctx = Context.new(timeout: 1s)
let deadline = ctx.deadline() ?? return
let left = Time.until(Time.now(), deadline)
println(Time.compareDuration(left, 100ms) < 0)
```

```text
false
```

### Time left after the deadline is negative

Once the deadline passes, the time left is negative, and a new context with it as the timeout
panics:

```ant
let ctx = Context.new(timeout: 10ms)
ctx.wait()
let deadline = ctx.deadline() ?? return
let left = Time.until(Time.now(), deadline)
let next = Context.new(timeout: left) // panic: Context.new: the timeout is zero or negative
```

To give work the rest of a deadline, pass the context as the parent instead; a child of a done
parent is done at once:

```ant
let ctx = Context.new(timeout: 10ms)
ctx.wait()
let next = Context.new(parent: ctx)
println(next.reason())
```

```text
timeout
```

## See also

[Context.new](new.md) · [Context.isDone](isDone.md) · [Context.reason](reason.md) ·
[Context.wait](wait.md) · [Time.now](../Time/now.md) · [Time.until](../Time/until.md) ·
[Time.compareDuration](../Time/compareDuration.md) · [Time.Instant](../Time/types/Instant.md) ·
[Context](index.md)
