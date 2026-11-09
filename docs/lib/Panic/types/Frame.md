<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Panic.Frame

```ant signature
type Panic.Frame = { function: string, path: string, line: int, column: int }
```

Panic.Frame is one frame of a panic's trace: the function, the path of its module, and the line and
column, counted from 1, of the panic in the first frame and of a call in the others.

<!-- end generated -->

## Details

- [`Panic.catch`](../catch.md) gives a `Panic` whose `trace` is an array of frames, innermost first:
  `trace[0]` is where the panic began, and each later frame is the call that led to the one before
  it.
- The trace does not stop at the function `Panic.catch` called. It goes on past that call to the
  goroutine's root, so it also holds the frames of the code around the `catch`.
- [`Panic.raise`](../raise.md) keeps the trace as it is, so the frames still point at the original
  panic, not at the `raise`.

## Basic examples

Catch a panic and read where it began:

```ant
func half(n: int): int {
  if n % 2 != 0 { panic("odd: ${n}") }
  n / 2
}

let r = Panic.catch({ half(7) })
if r is Err {
  let top = r.error.trace[0]
  println("${r.error.message} in ${top.function}, ${top.path} line ${top.line}")
}
```

```text
odd: 7 in half, main.ant line 2
```

Write one with a literal, and read its fields:

```ant
let f = Panic.Frame{ function: "load", path: "config.ant", line: 12, column: 5 }
f          // Panic.Frame{ function: "load", path: "config.ant", line: 12, column: 5 }
f.function // "load"
f.line     // 12
```

## Scope

The frame after the first is the caller, at the line of its call:

```ant
func check(age: int): int {
  assert(age >= 0, "age is never negative")
  age
}

func register(age: int): int {
  check(age)
}

let r = Panic.catch({ register(-1) })
if r is Err {
  let trace = r.error.trace
  println("${trace[0].function} line ${trace[0].line}")
  println("${trace[1].function} line ${trace[1].line}")
}
```

```text
check line 2
register line 7
```

A frame in an imported module names that module's path:

```ant [geo.ant]
export func area(width: int, height: int): int {
  if width < 0 || height < 0 { panic("negative size") }
  width * height
}
```

```ant
import "./geo.ant" { area }

let r = Panic.catch({ area(-1, 2) })
if r is Err {
  let top = r.error.trace[0]
  println("${top.function} in ${top.path}")
}
```

```text
area in geo.ant
```

Annotate an array of frames with `Array[Panic.Frame]`:

```ant
let f = Panic.Frame{ function: "parse", path: "json.ant", line: 40, column: 9 }
let frames: Array[Panic.Frame] = [f, Panic.Frame{ ...f, line: 41 }]
len(frames) // 2
```

`String.from` and `Json.stringify` show every field:

```ant
let f = Panic.Frame{ function: "f", path: "a.ant", line: 3, column: 1 }
String.from(f)    // "Panic.Frame{ function: \"f\", path: \"a.ant\", line: 3, column: 1 }"
Json.stringify(f) // "{\"function\":\"f\",\"path\":\"a.ant\",\"line\":3,\"column\":1}"
```

## Applications

Format a frame the way `file:line:column` locations are written:

```ant
func location(f: Panic.Frame): string {
  "${f.path}:${f.line}:${f.column}"
}

location(Panic.Frame{ function: "run", path: "jobs.ant", line: 8, column: 3 }) // "jobs.ant:8:3"
```

Run each job under `Panic.catch` and report where a failing job panicked, then go on with the next:

```ant
type Job = { id: int, items: Array[int] }

func firstItem(job: Job): int {
  job.items[0]
}

let jobs = [Job{ id: 1, items: [5] }, Job{ id: 2, items: [] }, Job{ id: 3, items: [7] }]
for job in jobs {
  let r = Panic.catch({ firstItem(job) })
  if r is Err {
    let top = r.error.trace[0]
    println("job ${job.id} failed in ${top.function} at line ${top.line}")
  } else {
    println("job ${job.id}: ${r.value}")
  }
}
```

```text
job 1: 5
job 2 failed in firstItem at line 4
job 3: 7
```

Log the innermost frame as JSON, for a log collector:

```ant
type LogLine = { message: string, function: string, line: int }

func fail() {
  panic("disk full")
}

let r = Panic.catch(fail)
if r is Err {
  let top = r.error.trace[0]
  println(Json.stringify(LogLine{ message: r.error.message, function: top.function, line: top.line }))
}
```

```text
{"message":"disk full","function":"fail","line":4}
```

## Properties and relations

Frames compare by their fields:

```ant
let f = Panic.Frame{ function: "a", path: "a.ant", line: 1, column: 1 }
assert(f == Panic.Frame{ ...f })
assert(f != Panic.Frame{ ...f, line: 2 })
```

Every frame of a caught trace counts its line and column from 1:

```ant
func deep(n: int): int {
  if n == 0 { panic("bottom") }
  deep(n - 1)
}

let r = Panic.catch({ deep(3) })
if r is Err {
  assert(r.error.trace.every({ x.line >= 1 && x.column >= 1 }))
}
```

`Panic.raise` keeps the trace, so catching a raised panic gives the same frames:

```ant
func boom() {
  panic("boom")
}

let first = Panic.catch(boom)
if first is Err {
  let again = Panic.catch({ Panic.raise(first.error) })
  if again is Err {
    assert(again.error.trace == first.error.trace)
  }
}
```

## Possible issues

`Frame` alone is not a name. Write `Panic.Frame`:

```ant
let f = Frame{ function: "a", path: "a.ant", line: 1, column: 1 }
// error: unknown type Frame; did you mean Panic.Frame?
```

A literal gives every field:

```ant
let f = Panic.Frame{ function: "a", path: "a.ant", line: 1 }
// error: Panic.Frame literal is missing the field column
```

The trace is innermost first. `trace[0]` is the function that panicked, not the one that called it:

```ant
func inner() {
  panic("deep")
}

func outer() {
  inner()
}

let r = Panic.catch(outer)
if r is Err {
  println(r.error.trace[0].function)
  println(r.error.trace[1].function)
}
```

```text
inner
outer
```

## See also

[Panic](../index.md) · [Panic.catch](../catch.md) · [Panic.raise](../raise.md) ·
[panic](../../prelude/panic.md) · [assert](../../prelude/assert.md)
