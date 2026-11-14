<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.any

```ant signature
func Parallel.any[T, E](fs: Array[func(Context): Result[T, E]], ctx: Context | null = null): Result[T, Array[E]] | null
```

Parallel.any calls every function of fs at once with the scope context and returns the first Ok one
returns, cancelling the scope for the rest. An Err cancels nothing; when every function returns Err,
it returns Err of the errors in input order, and Err(\[\]) when fs is empty. It returns null when
the scope ends before an Ok or the last Err.

<!-- end generated -->

## Details

- The call returns only after every callback has ended, the losers included. A loser stops early
  only if it passes the scope context to its waits.
- When `ctx` is done already and `fs` is not empty, no callback runs and the call returns `null`.
- A panic in any callback, a loser after the winner included, is raised in the caller.

## Basic examples

The first `Ok` wins. An `Err` from another callback does not stop the call:

```ant
func primary(ctx: Context): Result[string, Error] { Result.fail("primary is down") }
func mirror(ctx: Context): Result[string, Error] { Result.ok("reply from the mirror") }

println(Parallel.any([primary, mirror]))
```

```text
Ok{ value: "reply from the mirror" }
```

When every callback fails, the result is `Err` of all the errors:

```ant
func primary(ctx: Context): Result[string, Error] { Result.fail("primary is down") }
func mirror(ctx: Context): Result[string, Error] { Result.fail("mirror is down") }

println(Parallel.any([primary, mirror]))
```

```text
Err{ error: [Error{ message: "primary is down" }, Error{ message: "mirror is down" }] }
```

## Scope

### Lambdas and the scope context

A lambda callback gets the scope context as `x`. Once the fast lookup wins, the call cancels the
scope, and the slow lookup's sleep returns early:

```ant
func lookup(server: string, delay: Time.Duration, ctx: Context): Result[string, Error] {
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() {
    println("${server} stopped: ${ctx.reason()}")
    return Result.fail("${server}: ${ctx.reason()}")
  }
  Result.ok("answer from ${server}")
}

let answer = Parallel.any([
  { lookup("slow", 1s, x) },
  { lookup("fast", 10ms, x) },
])
println(answer)
```

```text
slow stopped: cancelled
Ok{ value: "answer from fast" }
```

### A fast failure waits for a slow success

An `Err` cancels nothing, so a quick failure leaves the others running:

```ant
func failFast(ctx: Context): Result[int, string] { Result.err("refused") }
func answerLater(ctx: Context): Result[int, string] {
  Time.sleep(20ms, ctx: ctx)
  Result.ok(42)
}

println(Parallel.any([failFast, answerLater]))
```

```text
Ok{ value: 42 }
```

### Errors in input order

The error list follows `fs`. Here the first callback fails last, and its error still comes first:

```ant
func slowFailure(ctx: Context): Result[int, string] {
  Time.sleep(20ms, ctx: ctx)
  Result.err("slow")
}
func quickFailure(ctx: Context): Result[int, string] { Result.err("quick") }

println(Parallel.any([slowFailure, quickFailure]))
```

```text
Err{ error: ["slow", "quick"] }
```

### Empty input

An empty list has no callback to succeed, so the result is `Err([])`:

```ant
let none: Array[func(Context): Result[int, string]] = []
println(Parallel.any(none))
```

```text
Err{ error: [] }
```

## Named arguments

### ctx

The scope is a child of `ctx`. When `ctx` ends before any callback succeeds, the call returns
`null`:

```ant
func lookup(server: string, ctx: Context): Result[string, Error] {
  Time.sleep(1s, ctx: ctx)
  if ctx.isDone() { return Result.fail("${server}: ${ctx.reason()}") }
  Result.ok("answer from ${server}")
}

let ctx = Context.new(20ms)
let answer = Parallel.any([{ lookup("east", x) }, { lookup("west", x) }], ctx: ctx)
println(answer)
println(ctx.reason())
```

```text
null
timeout
```

A context that is done already runs no callback:

```ant
var calls = 0
func count(ctx: Context): Result[int, string] {
  calls += 1
  Result.ok(calls)
}

let done = Context.new()
done.cancel()
println(Parallel.any([count], ctx: done))
println(calls)
```

```text
null
0
```

## Applications

### Read the first file that exists

Look for a file in several places at once. The `Err` of every miss lists all the places tried:

<!-- test: file local.json
{"port": 8080}
-->

```ant
func readFirst(paths: Array[string]): Result[string, Error] {
  let readers = paths.map(func(path: string): func(Context): Result[string, Error] {
    { File.read(path) }
  })
  let found = Parallel.any(readers) ?? return Result.fail("cancelled")
  found | Result.mapErr({ Error{ message: x.map({ x.message }) | String.join("; ") } })
}

println(readFirst(["missing.json", "local.json"]))
println(readFirst(["a.json", "b.json"]))
```

```text
Ok{ value: "{\"port\": 8080}\n" }
Err{ error: Error{ message: "open a.json: no such file or directory; open b.json: no such file or directory" } }
```

### Ask replicas with a deadline

Ask every replica, take the first answer, and give up after a deadline. The function owns its
deadline context and cancels it on return:

```ant
func ask(replica: string, delay: Time.Duration, ctx: Context): Result[string, Error] {
  Time.sleep(delay, ctx: ctx)
  if ctx.isDone() { return Result.fail("${replica}: ${ctx.reason()}") }
  Result.ok("quote from ${replica}")
}

func quote(delays: Array[Time.Duration]): Result[string, Error] {
  let deadline = Context.new(100ms)
  defer { deadline.cancel() }
  let reply = Parallel.any([
    { ask("a", delays[0], x) },
    { ask("b", delays[1], x) },
  ], ctx: deadline) ?? return Result.fail("no replica answered in time")
  reply | Result.mapErr({ Error{ message: "all ${x.length()} replicas failed" } })
}

println(quote([1s, 10ms]))
println(quote([1s, 1s]))
```

```text
Ok{ value: "quote from b" }
Err{ error: Error{ message: "no replica answered in time" } }
```

## Properties and relations

`Parallel.race` returns the first value of any kind, `Parallel.any` the first `Ok`. With a fast
`Err` and a slower `Ok`, they disagree:

```ant
func refuse(ctx: Context): Result[int, string] { Result.err("refused") }
func answer(ctx: Context): Result[int, string] {
  Time.sleep(20ms, ctx: ctx)
  Result.ok(42)
}

let first = Parallel.race([refuse, answer])
let firstOk = Parallel.any([refuse, answer])
assert(first is Err && first.error == "refused")
assert(firstOk is Ok && firstOk.value == 42)
```

When every callback fails, the result holds one error per callback:

```ant
func refuse(ctx: Context): Result[int, string] { Result.err("refused") }

let fs = [refuse, refuse, refuse]
let result = Parallel.any(fs)
assert(result is Err && result.error.length() == fs.length())
```

## Possible issues

### A callback that ignores its context holds the call

The winner returns at once, but the call waits for the loser, which sleeps without the scope
context. Pass `x`, or the `ctx` parameter, to every wait:

```ant
var stubbornFinished = false
func stubborn(ctx: Context): Result[string, Error] {
  Time.sleep(50ms)
  stubbornFinished = true
  Result.ok("stubborn")
}
func quick(ctx: Context): Result[string, Error] { Result.ok("quick") }

println(Parallel.any([stubborn, quick]))
println(stubbornFinished)
```

```text
Ok{ value: "quick" }
true
```

### A callback that waits on the caller's ctx misses the winner

The call cancels only its scope context, so a callback that sleeps on the caller's `deadline` does
not see the fast winner and holds the call for the full second. Pass the scope context to its waits
instead: the `ctx` parameter of a named function, or `x` in a lambda:

```ant
var slowFinished = false
let deadline = Context.new(5s)
defer { deadline.cancel() }

func slow(ctx: Context): Result[string, Error] {
  Time.sleep(1s, ctx: deadline)
  slowFinished = true
  Result.fail("slow replica")
}
func fast(ctx: Context): Result[string, Error] { Result.ok("fast replica") }

println(Parallel.any([slow, fast], ctx: deadline))
println(slowFinished)
```

```text
Ok{ value: "fast replica" }
true
```

### A loser's panic outweighs the winner

A panic in any callback reaches the caller, even after another callback has won. Here `Panic.catch`
stops it in the caller:

```ant
func quick(ctx: Context): Result[string, Error] { Result.ok("quick") }
func broken(ctx: Context): Result[string, Error] {
  Time.sleep(10ms, ctx: ctx)
  panic("broken replica")
}

let caught = Panic.catch({ Parallel.any([quick, broken]) })
if caught is Err {
  println("caught: ${caught.error.message}")
}
```

```text
caught: broken replica
```

### A lambda that only builds a Result has no type

A lambda whose body is only `Result.fail` or `Result.ok` leaves a type parameter of the `Result`
unknown. Give the callbacks a declared return type, as the named functions on this page do:

```ant
let answer = Parallel.any([
  { Result.fail("primary is down") },
  // error: cannot infer the type parameter T of Result.fail; annotate the result
  { Result.ok("mirror") },
])
```

## See also

[Parallel.race](race.md) · [Parallel.run](run.md) · [Parallel.map](map.md) ·
[Parallel.each](each.md) · [Context.new](../Context/new.md) · [Context.cancel](../Context/cancel.md)
· [Result.mapErr](../Result/mapErr.md) · [Concurrency guide](../../guide/concurrency.md)
