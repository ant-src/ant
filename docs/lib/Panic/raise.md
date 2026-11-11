<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic.raise

```ant signature
func Panic.raise(p: Panic): never
```

Panic.raise panics with p's message and trace as they are, so a caught panic goes on as though it
had not been caught.

<!-- end generated -->

## Details

- The panic keeps its original trace, so the report and any outer [Panic.catch](catch.md) point to
  where the panic began, not to the `Panic.raise` call.
- `Panic.raise` does not return. After `if r is Err { Panic.raise(r.error) }`, the checker treats
  `r` as an `Ok`.

## Basic examples

Raise a panic, and catch it to see the record it carries:

```ant
let p = Panic{ message: "boom", trace: [] }
Panic.raise(p)                  // panic: boom
Panic.catch({ Panic.raise(p) }) // Err{ error: Panic{ message: "boom", trace: [] } }
```

Catch a panic, log it, and let it go on:

<!-- test: exit 1 -->

```ant
func parse(s: string): int {
  if s == "" { panic("empty input") }
  s.length()
}

let r = Panic.catch({ parse("") })
if r is Err {
  println("caught: ${r.error.message}")
  Panic.raise(r.error)
}
println("not reached")
```

```text
caught: empty input
```

## Scope

Call it through `.`:

```ant
let p = Panic{ message: "stop", trace: [] }
p.raise() // panic: stop
```

A caught panic from the runtime, such as an index out of range, raises with its own message and
trace. The report points to the line that indexed, not to the line that raised:

```ant
let r = Panic.catch({ [1, 2][5] }) // panic: index 5 out of range for an array of length 2
if r is Err { Panic.raise(r.error) }
```

A trace is plain data. A panic built by hand carries the frames you write, and a catch sees them
unchanged:

```ant
let frame = Panic.Frame{ function: "load", path: "lib.ant", line: 3, column: 5 }
let p = Panic{ message: "x", trace: [frame] }
let r = Panic.catch({ Panic.raise(p) })
assert(r is Err)
r.error.trace // [Panic.Frame{ function: "load", path: "lib.ant", line: 3, column: 5 }]
```

A call of `Panic.raise` has type `never`, so it fits where a value is expected, such as a branch of
an `if` expression:

```ant
func valueOf(r: Result[int, Panic]): int {
  let v = if r is Ok { r.value } else { Panic.raise(r.error) }
  v + 1
}

println(valueOf(Result.ok(1)))
```

```text
2
```

A raised panic in a goroutine ends that goroutine only, as any panic does:

```ant
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  let r = Panic.catch({ [1, 2][5] })
  if r is Err { Panic.raise(r.error) }
}
wg.wait()
println("main goes on")
```

```text
main goes on
```

## Applications

Run every job, even when one panics, and still fail at the end with the first panic and its original
trace:

<!-- test: exit 1 -->

```ant
func handle(job: string): int {
  if job == "" { panic("empty job name") }
  job.length()
}

let failures: Array[Panic] = []
for job in ["build", "", "deploy"] {
  let r = Panic.catch({ handle(job) })
  if r is Err {
    println("job \"${job}\": ${r.error.message}")
    failures.push(r.error)
  } else {
    println("job \"${job}\": ${r.value}")
  }
}
if failures.length() > 0 { Panic.raise(failures[0]) }
```

```text
job "build": 5
job "": empty job name
job "deploy": 6
```

Add context to the message and keep the trace, with a spread:

```ant
func load(): int {
  let xs = [1]
  xs[4] // panic: loading config: index 4 out of range for an array of length 1
}

let r = Panic.catch(load)
if r is Err {
  Panic.raise(Panic{ ...r.error, message: "loading config: ${r.error.message}" })
}
```

Clean up inside the caught function with `defer`, which runs before `Panic.catch` returns, then pass
the panic on:

<!-- test: exit 1 -->

```ant
func step(name: string) {
  let r = Panic.catch({
    defer { println("cleanup ${name}") }
    panic("${name} failed")
  })
  if r is Err {
    println("log: ${r.error.message}")
    Panic.raise(r.error)
  }
}

step("upload")
```

```text
cleanup upload
log: upload failed
```

## Properties and relations

Catching a raised panic gives back an equal panic:

```ant
let first = Panic.catch({ [1][3] })
if first is Err {
  let again = Panic.catch({ Panic.raise(first.error) })
  assert(again is Err && again.error == first.error)
}
```

## Possible issues

`panic(r.error.message)` passes the message on but starts a new trace, so the report points to the
line that panicked again and the original location is lost:

```ant
let r = Panic.catch({ [1][4] })
if r is Err {
  panic(r.error.message) // panic: index 4 out of range for an array of length 1
}
```

Use `Panic.raise(r.error)` to keep the trace.

`Panic.raise` takes the `Panic`, not the `Result` that holds it:

```ant
let r = Panic.catch({ 21 * 2 })
if r is Err { Panic.raise(r) } // error: cannot use Err[Panic] where Panic is expected
```

## See also

[Panic.catch](catch.md) · [Panic.Frame](types/Frame.md) · [panic](../prelude/panic.md) ·
[assert](../prelude/assert.md) · [defer](../../language/keywords/defer.md) ·
[Errors guide](../../guide/errors.md)
