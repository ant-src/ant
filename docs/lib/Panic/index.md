<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic

```ant signature
type Panic = { message: string, trace: Array[Panic.Frame] }
```

Panic is a panic that [`Panic.catch`](catch.md) stopped: its message and the goroutine's stack when
it began, innermost frame first.

<!-- end generated -->

A panic stops the current goroutine. [`panic`](../prelude/panic.md),
[`assert`](../prelude/assert.md), and operations that fail at run time, such as an index out of
range or a division by zero, all start one. `Panic.catch` stops a panic on its way out of a function
and hands it back as a `Panic` record, in the `Err` of a `Result`. The `Panic` namespace holds the
functions that catch and raise panics, and the type of a trace's frames.

## Details

- A panic that nothing catches ends its goroutine, after the goroutine's deferred blocks run, and
  the runtime writes the message and trace to standard error. In the main goroutine the program then
  exits with status 1. In any other goroutine, that goroutine ends and the program goes on.
- The trace runs past the function `Panic.catch` called, to the root of the goroutine.
- A panic is for bugs. A failure that a caller is meant to handle is a
  [`Result`](../Result/index.md).

## Basic examples

Catch a panic and read its message:

```ant
let r = Panic.catch({ 10 / 0 })
if r is Err {
  println("caught: ${r.error.message}")
}
```

```text
caught: division by zero
```

A function that returns normally gives `Ok` of its value:

```ant
Panic.catch({ 10 / 2 }) // Ok{ value: 5 }
```

## Scope

`Panic.catch` returns once the deferred blocks inside the caught block have run, so cleanup happens
before the caller sees the panic:

```ant
let r = Panic.catch({
  defer { println("cleanup") }
  panic("boom")
})
if r is Err {
  println("caught: ${r.error.message}")
}
```

```text
cleanup
caught: boom
```

The trace starts at the function that panicked. Each later frame is the call that led to the one
before it, and the frames go on past the caught block:

```ant
func middle(n: int): int {
  inner(n)
}

func inner(n: int): int {
  100 / n
}

let r = Panic.catch({ middle(0) })
if r is Err {
  let trace = r.error.trace
  println("${trace[0].function} line ${trace[0].line}")
  println("${trace[1].function} line ${trace[1].line}")
  println(trace.length() > 2)
}
```

```text
inner line 6
middle line 2
true
```

A `Result[T, Panic]` works like any other `Result`, so `try` passes the `Err` on:

```ant
func half(n: int): Result[int, Panic] {
  let v = try Panic.catch({ 10 / n })
  Result.ok(v / 2)
}

half(2)        // Ok{ value: 2 }
half(0) is Err // true
```

A panic that leaves a `Parallel` callback panics again in the caller, so one `Panic.catch` around
the call sees it:

```ant
let r = Panic.catch({
  Parallel.map([1, 2, 0, 4], func(n: int, ctx: Context): int { 12 / n }, limit: 2)
})
if r is Err {
  println("caught: ${r.error.message}")
}
```

```text
caught: division by zero
```

## Functions by job

### Catching

- [Panic.catch](catch.md): call a function and return `Ok` of its value, or `Err` of the panic that
  left it

### Passing on

- [Panic.raise](raise.md): panic again with a caught panic's message and trace

### Types

- [Panic.Frame](types/Frame.md): one frame of a trace

### Starting a panic

- [panic](../prelude/panic.md): panic with a message
- [assert](../prelude/assert.md): panic when a condition is false

## Applications

A worker goroutine catches each job's panic, so that a job that hits a bug ends that job and the
worker goes on to the next. Here the handler reads the first item of a job, and an empty job panics:

```ant
let jobs: Channel[Array[int]] = Channel.new(3)
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  while true {
    let job = jobs.receive() ?? break
    let r = Panic.catch({ job[0] })
    if r is Err {
      println("skip: ${r.error.message}")
    } else {
      println("got ${r.value}")
    }
  }
}
for job in [[1], [], [3]] {
  jobs.send(job)
}
jobs.close()
wg.wait()
```

```text
got 1
skip: index 0 out of range for an array of length 0
got 3
```

Run a batch of checks over code under test, and report each failure's message and location. The
`assert` in `top` guards a precondition that one caller breaks:

```ant
func top(stack: Array[int]): int {
  assert(stack.length() > 0, "top of an empty stack")
  stack[stack.length() - 1]
}

let checks = [
  func(): int { top([1, 2]) },
  func(): int { top([]) },
  func(): int { top([7]) },
]
var failed = 0
for check in checks {
  let r = Panic.catch(check)
  if r is Err {
    failed += 1
    let at = r.error.trace[0]
    println("${at.path}:${at.line}: ${r.error.message}")
  }
}
println("${failed} failed")
```

```text
main.ant:2: assertion failed: stack.length() > 0 (0 > 0): top of an empty stack
1 failed
```

## Properties and relations

Catching a raised panic gives back an equal `Panic`:

```ant
let p = Panic{ message: "stop", trace: [] }
let r = Panic.catch({ Panic.raise(p) })
assert(r is Err && r.error == p)
```

Panics compare by their fields, as records do:

```ant
let p = Panic{ message: "boom", trace: [] }
assert(p == Panic{ ...p })
assert(p != Panic{ ...p, message: "bang" })
```

## Possible issues

A panic in the main goroutine that nothing catches ends the program with status 1. The lines after
it never run:

```ant
println("start")
let xs: Array[int] = []
let first = xs[0] // panic: index 0 out of range for an array of length 0
println("not printed: ${first}")
```

```text
start
```

`Panic` is a type. To start a panic, call `panic`:

```ant
Panic("boom") // error: Panic is a type, not a function
```

A `Panic` is not an `Error`, so `try` cannot pass a caught panic out of a function that returns
`Result[T, Error]`:

```ant
func load(): Result[int, Error] {
  let n = try Panic.catch({ 1 / 0 })
  // error: try would return Err[Panic], but the function returns Result[int, Error]
  Result.ok(n)
}
```

Turn it into an `Error` first, with `Result.mapErr`:

```ant
func load(): Result[int, Error] {
  let n = try (Panic.catch({ 1 / 0 }) | Result.mapErr({ Error{ message: x.message } }))
  Result.ok(n)
}

load() // Err{ error: Error{ message: "division by zero" } }
```

A goroutine started inside the caught block panics on its own, so `Panic.catch` does not see its
panic. The runtime reports it on standard error, as it does any uncaught panic in a goroutine, and
the caught block returns normally:

```ant
let wg = WaitGroup.new()
wg.add(1)
let r = Panic.catch({
  go {
    defer { wg.done() }
    panic("inner")
  }
  1
})
wg.wait()
println(r)
```

```text
Ok{ value: 1 }
```

Only a panic is caught: `Os.exit` inside the block ends the program as it would anywhere else.

A panic can leave shared data half changed. What the block wrote before the panic stays written, and
the code that catches the panic decides whether the data can still be used:

```ant
let log: Array[string] = []
let r = Panic.catch({
  log.push("step 1")
  panic("step 2 failed")
})
println(r is Err)
println(log)
```

```text
true
["step 1"]
```

## See also

[Panic.catch](catch.md) · [Panic.raise](raise.md) · [Panic.Frame](types/Frame.md) ·
[panic](../prelude/panic.md) · [assert](../prelude/assert.md) ·
[defer](../../language/keywords/defer.md) · [Result](../Result/index.md) ·
[Error](../Error/index.md) · [Errors](../guides/errors.md)
