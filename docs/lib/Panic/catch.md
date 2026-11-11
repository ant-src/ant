<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic.catch

```ant signature
func Panic.catch[T](f: func(): T): Result[T, Panic]
```

Panic.catch calls f and returns Ok of its value, or an Err of the panic that left f once the
deferred blocks inside f have run. A caught panic is not reported. Os.exit and the end of the run
pass through. It belongs at the edge of a unit of work, not in place of a Result.

<!-- end generated -->

## Details

- The deferred blocks inside `f` run before `Panic.catch` returns. When a deferred block panics too,
  the `Err` holds that later panic.
- `Os.exit` is not a panic: it ends the program from inside `f`, and no deferred block runs.
- A goroutine that `f` starts is not covered. A panic in it ends that goroutine and is reported.

## Basic examples

A block that returns normally gives `Ok` of its value. A panic gives `Err` of the panic:

```ant
Panic.catch({ 6 * 7 })                                // Ok{ value: 42 }
Panic.catch({ [1, 2][5] }) is Err                     // true
Panic.catch({ 1 / 0 }) | Result.mapErr({ x.message }) // Err{ error: "division by zero" }
```

Read the message of the panic from the `Err`:

```ant
let r = Panic.catch({ panic("boom") })
if r is Err {
  println("caught: ${r.error.message}")
}
println("still running")
```

```text
caught: boom
still running
```

## Scope

A function with no parameters can be passed by name:

```ant
func load(): string {
  panic("config missing")
}

let r = Panic.catch(load)
if r is Err {
  println(r.error.message)
}
```

```text
config missing
```

A block that returns `void` gives `Ok` of `void`:

```ant
let r = Panic.catch({ println("hello") })
println(r)
```

```text
hello
Ok{ value: void }
```

A `Result` that `f` returns stays a value. Only a panic becomes the outer `Err`:

```ant
func lookup(): Result[int, string] {
  Result.err("not found")
}

println(Panic.catch(lookup))
```

```text
Ok{ value: Err{ error: "not found" } }
```

Every panic is caught the same way, whether it comes from `panic`, `assert`, or an operation that
fails at run time:

```ant
func message(p: Panic): string {
  p.message
}

let xs = [1, 2]
Panic.catch({ panic("stop") }) | Result.mapErr(message)     // Err{ error: "stop" }
Panic.catch({ assert(xs[0] > 1) }) | Result.mapErr(message) // Err{ error: "assertion failed: xs[0] > 1 (1 > 1)" }
Panic.catch({ xs[2] }) | Result.mapErr(message)             // Err{ error: "index 2 out of range for an array of length 2" }
```

The panic carries a trace, innermost frame first. Each frame names a function, its module's path,
and a line and column:

```ant
func check(n: int) {
  if n < 0 {
    panic("negative: ${n}")
  }
}

let r = Panic.catch({ check(-1) })
if r is Err {
  let frame = r.error.trace[0]
  println("${frame.function} at ${frame.path}:${frame.line}:${frame.column}")
}
```

```text
check at main.ant:3:10
```

Deferred blocks inside `f` run before `Panic.catch` returns:

```ant
let steps: Array[string] = []
let r = Panic.catch({
  defer { steps.push("cleanup") }
  steps.push("work")
  panic("failed")
})
println(steps)
println(r is Err)
```

```text
["work", "cleanup"]
true
```

## Applications

Run every job, so that a bug in one does not stop the others:

```ant
type Job = { id: int, input: string }

func handle(job: Job): int {
  let n = job.input.toInt() ?? panic("bad input: ${job.input}")
  100 / n
}

let jobs = [Job{ id: 1, input: "4" }, Job{ id: 2, input: "0" }, Job{ id: 3, input: "x" }]
for job in jobs {
  let done = Panic.catch({ handle(job) })
  if done is Err {
    println("job ${job.id}: ${done.error.message}")
  } else {
    println("job ${job.id}: ${done.value}")
  }
}
```

```text
job 1: 25
job 2: division by zero
job 3: bad input: x
```

Fall back to a default when a calculation panics, with `Result.unwrapOr`:

```ant
func average(xs: Array[int]): int {
  xs.reduce(0, func(sum: int, x: int): int { sum + x }) / xs.length()
}

Panic.catch({ average([2, 4]) }) | Result.unwrapOr(0) // 3
Panic.catch({ average([]) }) | Result.unwrapOr(0)     // 0
```

Catch a panic, log it, and pass it on with `Panic.raise`, which keeps the original trace:

```ant
func step(): int {
  panic("disk full")
}

let r = Panic.catch({
  let inner = Panic.catch(step)
  if inner is Err {
    println("step failed: ${inner.error.message}")
    Panic.raise(inner.error)
  }
  0
})
if r is Err {
  println("outer saw: ${r.error.message}, raised in ${r.error.trace[0].function}")
}
```

```text
step failed: disk full
outer saw: disk full, raised in step
```

## Properties and relations

When `f` does not panic, `Panic.catch(f)` is `Ok` of what `f` returns:

```ant
func f(): Array[int] {
  [3, 1, 2].sort()
}

let r = Panic.catch(f)
assert(r is Ok && r.value == f())
```

When `f` panics, the `Err` holds the message `f` panicked with:

```ant
let r = Panic.catch({ panic("same text") })
assert(r is Err && r.error.message == "same text")
```

## Possible issues

A function that takes arguments does not fit. Wrap the call in a block:

```ant
func parse(s: string): int {
  s.toInt() ?? panic("not a number: ${s}")
}

let r = Panic.catch(parse) // error: cannot use func(string): int where func(): T is expected
```

`Panic.catch({ parse("x") })` compiles.

A panic in a deferred block replaces the one in progress, so the `Err` holds only the last:

```ant
let r = Panic.catch({
  defer { panic("in cleanup") }
  panic("in work")
})
r.mapErr({ x.message }) // Err{ error: "in cleanup" }
```

A goroutine that `f` starts runs on its own. Its panic is reported on standard error, and
`Panic.catch` still returns `Ok`:

```ant
let wg = WaitGroup.new()
wg.add(1)
let r = Panic.catch({
  go {
    defer { wg.done() }
    panic("in goroutine")
  }
  "started"
})
wg.wait()
println(r)
```

```text
Ok{ value: "started" }
```

Catch the panic inside the goroutine instead: `go { Panic.catch({ work() }) }`.

`Os.exit` is not a panic. It ends the program from inside `f`, and the lines after never run:

<!-- test: exit 3 -->

```ant
let r = Panic.catch({ Os.exit(3) })
println("not printed")
```

A panic can leave data half changed. What `f` wrote before it panicked stays written:

```ant
let balances = Map{ ann: 100, bob: 0 }
let r = Panic.catch({
  balances["ann"] = 50
  panic("transfer interrupted")
})
balances["ann"] // 50
balances["bob"] // 0
```

Do not use `Panic.catch` for failures a caller is meant to handle. A panic carries only a message,
so there is nothing to branch on. Return a `Result` instead:

```ant
func parsePort(s: string): Result[int, string] {
  let n = s.toInt() ?? return Result.err("not a number: ${s}")
  Result.ok(n)
}

parsePort("80")   // Ok{ value: 80 }
parsePort("http") // Err{ error: "not a number: http" }
```

## See also

[Panic.raise](raise.md) · [Panic](index.md) · [Panic.Frame](types/Frame.md) ·
[panic](../prelude/panic.md) · [assert](../prelude/assert.md) ·
[defer](../../language/keywords/defer.md) · [Result](../Result/index.md) ·
[Result.unwrapOr](../Result/unwrapOr.md) · [Result.mapErr](../Result/mapErr.md)
