<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.run

```ant signature
func Parallel.run(fs: Array[func(Context): void], ctx: Context | null = null): void
```

Parallel.run calls every function of fs at once with the scope context, or none when ctx is done
already. Its functions return void, unlike other callbacks whose value the library ignores, so that
they need not share a result type; each hands back its result by writing a field.

<!-- end generated -->

## Details

- Every function starts at once on a goroutine of its own, and the call returns only after all of
  them have ended, deferred blocks included. Nothing reports which ones ran: a function that must
  say so writes a field.
- Each function gets the call's scope context as `x`, a child of `ctx`. The call cancels it when it
  returns and when a function panics, but never cancels `ctx` itself.
- When `ctx` is done already, no function runs. An empty list returns at once.
- A panic in one function cancels the scope, waits for the others, and then raises that panic in the
  caller. A later panic in the same call is reported and dropped.
- Two functions that write the same field, array, or map still need a lock. A field or slot that
  only one function writes can be read after the call with no lock.

## Basic examples

Run two loads at once, each writing its own field:

```ant
type Loaded = { user: string | null, orders: Array[string] }

let loaded = Loaded{ user: null, orders: [] }
Parallel.run([
  { loaded.user = "ann" },
  { loaded.orders = ["book", "pen"] },
])
println(loaded)
```

```text
Loaded{ user: "ann", orders: ["book", "pen"] }
```

Each function receives the scope context, which it passes to its waits. When `ctx` times out, the
scope context is done too, so a wait given `x` returns early:

```ant
type Page = { title: string, cancelled: bool }

let page = Page{ title: "", cancelled: false }
Parallel.run([
  { page.title = "Home" },
  {
    Time.sleep(1s, ctx: x)
    page.cancelled = x.isDone()
  },
], ctx: Context.new(30ms))
println(page)
```

```text
Page{ title: "Home", cancelled: true }
```

## Scope

The functions need not share a result type. Here one writes a string and the other runs a
`Parallel.map` of its own under the scope context:

```ant
type Profile = { name: string, scores: Array[int] }

func tenfold(n: int, work: Context): int { n * 10 }

let profile = Profile{ name: "", scores: [] }
Parallel.run([
  { profile.name = "Ann" },
  { profile.scores = Parallel.map([1, 2, 3], tenfold, limit: 2, ctx: x) ?? [] },
])
println(profile)
```

```text
Profile{ name: "Ann", scores: [10, 20, 30] }
```

A named function that takes a `Context` and returns nothing goes in the list as it is:

```ant
type Stats = { lines: int }

let stats = Stats{ lines: 0 }
func countLines(work: Context) {
  stats.lines = "a\nb\nc".split("\n").length()
}

Parallel.run([countLines])
println(stats.lines)
```

```text
3
```

Each function may write its own slot of one array:

```ant
let words = ["ant", "bee", "wasp"]
let sizes = [0, 0, 0]
Parallel.run([
  { sizes[0] = words[0].length() },
  { sizes[1] = words[1].length() },
  { sizes[2] = words[2].length() },
])
println(sizes)
```

```text
[3, 3, 4]
```

An empty list runs nothing and returns at once:

```ant
Parallel.run([])
println("done")
```

```text
done
```

## Named arguments

### ctx

With `ctx:`, the scope context is a child of `ctx`, so a timeout or a cancel of `ctx` reaches every
function. A `ctx` that is done already runs none of them:

```ant
type Flags = { ran: bool }

let flags = Flags{ ran: false }
let ctx = Context.new()
ctx.cancel()
Parallel.run([{ flags.ran = true }], ctx: ctx)
println(flags.ran)
```

```text
false
```

The call never cancels `ctx`, so the caller can go on using it:

```ant
let ctx = Context.new()
Parallel.run([{ Time.sleep(1ms, ctx: x) }], ctx: ctx)
println(ctx.isDone())
```

```text
false
```

## Applications

### Fail fast on the first error

`Parallel.run` looks at no result, so it does not stop on an `Err`. To stop the others, cancel a
context you own, and pass it as `ctx:`:

```ant
func fetch(name: string, delay: Time.Duration, work: Context): Result[string, Error] {
  Time.sleep(delay, ctx: work)
  if work.isDone() { return Result.fail("${name}: ${work.reason()}") }
  if name == "prices" { return Result.fail("prices: unavailable") }
  Result.ok("${name} data")
}

type Report = { stock: Result[string, Error] | null, prices: Result[string, Error] | null }

let report = Report{ stock: null, prices: null }
let ctx = Context.new()
Parallel.run([
  { report.stock = fetch("stock", 1s, x) },
  {
    let prices = fetch("prices", 10ms, x)
    report.prices = prices
    if prices is Err { ctx.cancel() }
  },
], ctx: ctx)
println(report.stock)
println(report.prices)
```

```text
Err{ error: Error{ message: "stock: cancelled" } }
Err{ error: Error{ message: "prices: unavailable" } }
```

### Catch a panic from any of the functions

The panic reaches the caller after the other functions have ended, so their deferred blocks have
run. [Panic.catch](../Panic/catch.md) turns it into a value:

```ant
type Log = { slow: string }

let log = Log{ slow: "" }
func slow(work: Context) {
  defer { log.slow = "stopped: ${work.reason()}" }
  work.wait()
}

let caught = Panic.catch({
  Parallel.run([slow, { panic("boom") }])
})
println(caught is Err)
println(log.slow)
```

```text
true
stopped: cancelled
```

## Properties and relations

Functions that each write one slot of an array give what [Parallel.map](map.md) returns:

```ant
let words = ["ant", "bee", "wasp"]
let sizes = [0, 0, 0]
let fs: Array[func(Context): void] = []
for i in Iter.range(0, words.length()) {
  fs.push({ sizes[i] = words[i].length() })
}
Parallel.run(fs)
assert(sizes == Parallel.map(words, func(w: string, work: Context) { w.length() }, limit: words.length()))
```

Every write a function makes is visible once the call returns:

```ant
type Pair = { a: int, b: int }

let pair = Pair{ a: 0, b: 0 }
Parallel.run([{ pair.a = 1 }, { pair.b = 2 }])
assert(pair == Pair{ a: 1, b: 2 })
```

## Possible issues

### A function that returns a value is a compile error

The list holds functions that return `void`. A lambda whose last expression has a value does not
fit:

```ant
func load(): int { 1 }

Parallel.run([{ load() }]) // error: cannot use int where void is expected
```

Store the value in a field, or discard it with `_ = load()`.

### Only the first panic reaches the caller

When two functions panic, the caller gets the first panic. The second is reported and dropped:

```ant
let caught = Panic.catch({
  Parallel.run([
    { panic("first") },
    {
      Time.sleep(20ms)
      panic("second")
    },
  ])
})
println(caught | Result.mapErr({ x.message }))
```

```text
Err{ error: "first" }
```

### Wait on x, not on the caller's context

A function that waits on the caller's context does not see the scope end, so a sibling's panic waits
for it. Here the call holds until the timeout:

```ant
let started = Time.monotonic()
let ctx = Context.new(50ms)
_ = Panic.catch({
  Parallel.run([{ panic("boom") }, { ctx.wait() }], ctx: ctx)
})
println(Time.monotonic() - started >= 40)
```

```text
true
```

Wait on `x` instead, and the call ends as soon as the panic cancels the scope:

```ant
let started = Time.monotonic()
let ctx = Context.new(50ms)
_ = Panic.catch({
  Parallel.run([{ panic("boom") }, { x.wait() }], ctx: ctx)
})
println(Time.monotonic() - started >= 40)
```

```text
false
```

### A wait that gives up returns, and the function goes on

A `void` wait such as `Time.sleep` returns early when `x` is done, and the code after it still runs.
Ask `x.isDone()` before writing a result that the cancelled work did not produce:

```ant
type Page = { title: string | null }

let page = Page{ title: null }
Parallel.run([
  {
    Time.sleep(1s, ctx: x)
    if x.isDone() { return }
    page.title = "Home"
  },
], ctx: Context.new(20ms))
println(page.title)
```

```text
null
```

### Two functions that share a value need a lock

Writes to different fields or slots need no lock. Two functions that push to one array, or put keys
in one map, race; guard the shared value with a [Mutex](../Mutex/index.md):

```ant
let seen: Array[string] = []
let lock = Mutex.new()
Parallel.run([
  { lock.withLock({ seen.push("ann") }) },
  { lock.withLock({ seen.push("bob") }) },
])
println(seen.sort())
```

```text
["ann", "bob"]
```

## See also

[Parallel.map](map.md) · [Parallel.each](each.md) · [Parallel.race](race.md) ·
[Parallel.any](any.md) · [Context.new](../Context/new.md) · [Panic.catch](../Panic/catch.md) ·
[WaitGroup](../WaitGroup/index.md) · [go](../../language/keywords/go.md)
