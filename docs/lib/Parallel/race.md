<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.race

```ant signature
func Parallel.race[T](fs: Array[func(Context): T], ctx: Context | null = null): T | null
```

Parallel.race calls every function of fs at once with the scope context and returns the first value
one returns, an Err included, cancelling the scope for the rest. It returns null when the scope ends
before any function returns, so a T that includes null is wrapped in a record. It panics if fs is
empty.

<!-- end generated -->

## Details

- Each function runs on a goroutine of its own and receives the scope context, `x` in a lambda. Pass
  it to every wait in the function, so that the function stops once another one wins.
- `race` returns only after every function has ended, its deferred blocks included. A function that
  ignores its context holds the call until it ends.
- The first value wins, whatever it is: an `Err` ends the race as an `Ok` does. To wait for the
  first success, use [Parallel.any](any.md).
- If a function panics, `race` waits for the others and raises that panic in the caller, even when
  another function already won.
- When `ctx` is done already, no function runs and `race` returns `null`.

## Basic examples

The function that returns first wins; the scope context cuts the other one's wait short:

```ant
let winner = Parallel.race([
  {
    Time.sleep(1s, ctx: x)
    "slow"
  },
  {
    Time.sleep(10ms, ctx: x)
    "fast"
  },
])
println(winner)
```

```text
fast
```

A race of one function returns its value:

```ant
println(Parallel.race([{ 6 * 7 }]))
```

```text
42
```

## Scope

The losers see their context end with the reason `cancelled`. A deferred block in a function runs
before `race` returns, so the caller reads what it wrote with no lock:

```ant
type Log = { loser: string | null }

let log = Log{ loser: null }
let winner = Parallel.race([
  { "first" },
  {
    defer { log.loser = x.reason() }
    Time.sleep(1h, ctx: x)
    "second"
  },
])
println(winner)
println(log.loser)
```

```text
first
cancelled
```

A named function takes the context as a parameter, and the lambda passes `x` on:

```ant
type Reply = { server: string, body: string }

func ask(server: string, latency: Time.Duration, ctx: Context): Reply {
  Time.sleep(latency, ctx: ctx)
  Reply{ server, body: "pong" }
}

let reply = Parallel.race([{ ask("eu", 1s, x) }, { ask("us", 10ms, x) }])
println(reply?.server)
```

```text
us
```

An `Err` is a value like any other, and wins when it comes first:

```ant
func connect(ctx: Context): Result[string, Error] {
  Result.fail("connection refused")
}

func download(ctx: Context): Result[string, Error] {
  Time.sleep(50ms, ctx: ctx)
  Result.ok("data")
}

let outcome = Parallel.race([connect, download])
println(outcome)
```

```text
Err{ error: Error{ message: "connection refused", kind: null } }
```

The functions share one result type. For results of different kinds, annotate the array with a
union, and an `is` check tells the winner apart:

```ant
type Timeout = {}

let attempts: Array[func(Context): int | Timeout] = [
  {
    Time.sleep(1h, ctx: x)
    42
  },
  {
    Time.sleep(20ms, ctx: x)
    Timeout{}
  },
]
let outcome = Parallel.race(attempts)
if outcome is Timeout {
  println("timed out")
}
```

```text
timed out
```

## Named arguments

`ctx:` bounds the whole race. The scope context is its child, so a deadline on `ctx` ends every
function, and `race` returns `null` when it passes before any function returns:

```ant
let ctx = Context.new(20ms)
let winner = Parallel.race([
  {
    Time.sleep(1h, ctx: x)
    "done"
  },
], ctx: ctx)
println(winner ?? "no winner")
println(ctx.reason())
```

```text
no winner
timeout
```

A `ctx` that is done already runs no function at all:

```ant
let ctx = Context.new()
ctx.cancel()
let winner = Parallel.race([
  {
    println("never printed")
    1
  },
], ctx: ctx)
println(winner)
```

```text
null
```

## Applications

Run a reader and a watchdog together and stop both when either ends. Here the reader ends when its
input closes, and the watchdog gives up after a second:

```ant
func readLoop(input: Channel[string], ctx: Context): string {
  for line in input.values(ctx: ctx) {
    println("read ${line}")
  }
  "input closed"
}

func watchdog(limit: Time.Duration, ctx: Context): string {
  Time.sleep(limit, ctx: ctx)
  "too slow"
}

let input: Channel[string] = Channel.new(2)
input.send("hello")
input.send("bye")
input.close()

let ended = Parallel.race([{ readLoop(input, x) }, { watchdog(1s, x) }])
println(ended)
```

```text
read hello
read bye
input closed
```

Search several sources at once and keep the first answer:

```ant
type Source = { name: string, latency: Time.Duration }

func search(source: Source, ctx: Context): string {
  Time.sleep(source.latency, ctx: ctx)
  "${source.name}: 3 results"
}

let sources = [Source{ name: "disk", latency: 200ms }, Source{ name: "cache", latency: 5ms }]
let answer = sources
  | Array.map(func(s: Source): func(Context): string { { search(s, x) } })
  | Parallel.race()
println(answer)
```

```text
cache: 3 results
```

Inside a function that returns a `Result`, turn a race that ended with no winner into an `Err` with
`?? return`:

```ant
func ask(server: string, ctx: Context): string {
  Time.sleep(if server == "eu" { 1s } else { 10ms }, ctx: ctx)
  "pong from ${server}"
}

func ping(ctx: Context): Result[string, Error] {
  let reply = Parallel.race([{ ask("eu", x) }, { ask("us", x) }], ctx: ctx) ?? return Result.fail("cancelled")
  Result.ok(reply.toUpper())
}

println(ping(Context.new()))
let done = Context.new()
done.cancel()
println(ping(done))
```

```text
Ok{ value: "PONG FROM US" }
Err{ error: Error{ message: "cancelled", kind: null } }
```

## Properties and relations

`race` cancels its scope context when it returns, and never cancels `ctx`:

```ant
type Seen = { scope: Context | null }

let seen = Seen{ scope: null }
let ctx = Context.new()
let winner = Parallel.race([
  {
    seen.scope = x
    1
  },
], ctx: ctx)
let scope = seen.scope
assert(winner == 1)
assert(scope != null && scope.isDone())
assert(!ctx.isDone())
```

## Possible issues

A function that ignores its context holds the call: here the loser waits without `ctx`, so `race`
returns only after it ends, though `"fast"` won long before:

```ant
let winner = Parallel.race([
  { "fast" },
  {
    Time.sleep(50ms)
    println("slow ignored its context")
    "slow"
  },
])
println(winner)
```

```text
slow ignored its context
fast
```

When `T` includes `null`, a `null` winner looks like a race that ended with no winner. Wrap the
value in a record:

```ant
type Found = { user: string | null }

func findUser(id: int, ctx: Context): Found {
  let users = Map{ 1: "ann" }
  Found{ user: users[id] }
}

let found = Parallel.race([{ findUser(2, x) }])
println(if found == null { "cancelled" } else { found.user ?? "no such user" })
```

```text
no such user
```

A panic in any function reaches the caller, even after another function won:

```ant
let winner = Parallel.race([{ "won" }, { panic("boom") }]) // panic: boom
```

An empty array panics, since there is no value to return:

```ant
let none: Array[func(Context): int] = []
let winner = Parallel.race(none) // panic: Parallel.race needs at least one function
```

Every function must return the type the first one sets:

```ant
let winner = Parallel.race([{ 1 }, { "one" }])
// error: cannot use string where int is expected
```

## See also

[Parallel.any](any.md) · [Parallel.run](run.md) · [Parallel.map](map.md) · [Parallel.each](each.md)
· [Context.new](../Context/new.md) · [Context.cancel](../Context/cancel.md) ·
[select](../../language/keywords/select.md)
