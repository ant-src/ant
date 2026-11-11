<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# panic

```ant signature
func panic(message: string): never
```

panic stops the current goroutine with message, running its deferred blocks on the way out, unless a
[`Panic.catch`](../Panic/catch.md) on the way stops it. A panic in the main goroutine ends the
program with exit status 1.

<!-- end generated -->

## Details

- `panic` never returns. Its type is `never`, so a call fits wherever a value of any type is
  expected: a branch of an `if`, the right side of `??`, the end of a function body.
- As the panic unwinds, every block it leaves runs its deferred blocks, innermost first, up to the
  goroutine's root or the `Panic.catch` that stops it. A panic inside a deferred block replaces the
  one in progress.
- When nothing catches it, the runtime writes the message and a stack trace to standard error. A
  panic in a goroutine other than the main one ends only that goroutine; the program goes on.
- `panic` is for bugs and impossible states. For a failure a caller is meant to handle, return a
  `Result`.

## Basic examples

Stop the program when a value breaks a rule the code depends on:

```ant
func area(width: int, height: int): int {
  if width < 0 || height < 0 {
    panic("negative size: ${width}x${height}")
  }
  width * height
}

println(area(2, 3))
let bad = area(-1, 3) // panic: negative size: -1x3
```

```text
6
```

Stand in for a missing value on the right of `??`:

```ant
let config: string | null = null
let path = config ?? panic("no config file given") // panic: no config file given
```

Deferred blocks run on the way out:

```ant
func run() {
  defer { println("cleanup") }
  println("start")
  panic("boom") // panic: boom
}

run()
```

```text
start
cleanup
```

## Scope

`panic` has type `never`, so it can end a function that returns a value:

```ant
func todo(): int {
  panic("not implemented")
}

let r = Panic.catch({ todo() })
println(r is Err)
```

```text
true
```

After an `if` with no `else` whose branch panics, the rest of the block knows the condition was
false. Here `user` is a `User` after the check:

```ant
type User = { id: int, name: string }

func find(id: int): User | null {
  if id == 1 { User{ id: 1, name: "Ann" } } else { null }
}

let user: User | null = find(1)
if user == null {
  panic("no user 1")
}
println(user.name)
```

```text
Ann
```

It can be one branch of an `if` whose other branches give a value:

```ant
func hex(color: string): string {
  if color == "red" {
    "#f00"
  } else if color == "green" {
    "#0f0"
  } else {
    panic("unknown color ${color}")
  }
}

hex("red")  // "#f00"
hex("pink") // panic: unknown color pink
```

The deferred blocks of every function the panic leaves run, from the innermost out:

```ant
func inner() {
  defer { println("inner deferred") }
  panic("deep")
}

func outer() {
  defer { println("outer deferred") }
  inner()
}

let r = Panic.catch({ outer() })
println(r is Err)
```

```text
inner deferred
outer deferred
true
```

A panic in a deferred block replaces the one in progress:

```ant
let r = Panic.catch({
  defer { panic("second") }
  panic("first")
})
if r is Err {
  println(r.error.message)
}
```

```text
second
```

A panic in a goroutine ends that goroutine only. Deferring `WaitGroup.done` lets the main goroutine
stop waiting for it:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  panic("worker failed")
}
wg.wait()
println("main goes on")
```

```text
main goes on
```

## Applications

Run each job through `Panic.catch`, so a bug in one job does not stop the others:

```ant
type Job = { id: int, input: string }

func handle(job: Job): string {
  if job.input == "" {
    panic("empty input")
  }
  job.input.toUpper()
}

let jobs = [Job{ id: 1, input: "a" }, Job{ id: 2, input: "" }, Job{ id: 3, input: "c" }]
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
job 1: A
job 2: empty input
job 3: C
```

Read where a panic happened from its trace. The first frame is the innermost, the function that
called `panic`:

```ant
func check(n: int) {
  if n > 10 {
    panic("too big: ${n}")
  }
}

let r = Panic.catch({ check(42) })
if r is Err {
  let frame = r.error.trace[0]
  println("${frame.function} ${frame.path}:${frame.line}")
}
```

```text
check main.ant:3
```

## Properties and relations

`Panic.catch` returns the message given to `panic`, and the first frame of the trace is the `panic`
call:

```ant
let r = Panic.catch({ panic("lost") })
assert(r is Err)
assert(r.error.message == "lost")
assert(r.error.trace[0].line == 1)
```

`Panic.raise` panics again with a caught panic, its message and trace unchanged:

```ant
let first = Panic.catch({ panic("once") })
assert(first is Err)
let again = Panic.catch({ Panic.raise(first.error) })
assert(again is Err && again.error == first.error)
```

`assert(cond, message)` panics as `if !cond { panic(message) }` does, and narrows the same way. Its
message also quotes the condition and its values, such as `assertion failed: n > 0 (0 > 0)`, so
prefer it for a precondition:

```ant
let n = 0
let a = Panic.catch({ assert(n > 0, "n must be positive") })
let b = Panic.catch({
  if !(n > 0) {
    panic("n must be positive")
  }
})
assert(a is Err && b is Err)
assert(a.error.message == "assertion failed: n > 0 (0 > 0): n must be positive")
assert(b.error.message == "n must be positive")
```

## Possible issues

The message must be a `string`. Interpolate other values:

```ant
let code = 42
panic(code) // error: cannot use int where string is expected
```

```ant
let code = 42
panic("code ${code}") // panic: code 42
```

A panic skips the rest of the goroutine's code, so a `wg.done()` call at the end of a `go` block
never runs when the block panics, and `WaitGroup.wait` waits forever. Put the call in a `defer`, as
in the goroutine example under Scope.

Do not panic on input from a user, a file, or the network. Return a `Result`, or a value the caller
can check, instead:

```ant
func port(text: string): Result[int, Error] {
  let n = text.toInt()
  if n == null {
    return Result.err(Error{ message: "not a port: ${text}" })
  }
  Result.ok(n)
}

println(port("80x"))
```

```text
Err{ error: Error{ message: "not a port: 80x" } }
```

Catching a panic does not undo what ran before it. Data the code changed stays changed:

```ant
let log: Array[string] = []
let r = Panic.catch({
  log.push("step 1")
  panic("halfway")
})
println(log)
```

```text
["step 1"]
```

## See also

[assert](assert.md) · [Panic.catch](../Panic/catch.md) · [Panic.raise](../Panic/raise.md) ·
[Panic](../Panic/index.md) · [defer](../../language/keywords/defer.md) ·
[go](../../language/keywords/go.md) · [never](../../language/types/never.md) ·
[Result](../Result/index.md) · [Errors](../guides/errors.md)
