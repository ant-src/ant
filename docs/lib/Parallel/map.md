<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Parallel.map

```ant signature
func Parallel.map[T, U](xs: Array[T], f: func(T, Context): U, limit: int, ctx: Context | null = null): Array[U] | null
```

Parallel.map returns f applied to every element of xs and the scope context, in input order, running
at most limit callbacks at once. It admits elements in index order until the scope ends, and returns
null when some element was never admitted; otherwise it returns every result, even when ctx ended
after the last admission. It panics if limit \< 1, even when xs is empty.

<!-- end generated -->

## Details

- `f` gets the scope context as its second argument. Pass it to the waits inside `f`: it ends when
  `ctx` ends, when another callback panics, and when the call returns.
- Cancellation is cooperative: a callback that does not pass its context to its waits holds the call
  until it ends.
- A panic in `f` cancels the scope, waits for the other callbacks, and then panics in the caller
  with the same message.
- `limit` counts per call, so a `Parallel.map` inside another one runs up to the product of the two
  limits at once.

## Basic examples

Square each number, two at a time:

```ant
let squares = Parallel.map([1, 2, 3], func(n: int, _: Context) { n * n }, limit: 2)
println(squares)
```

```text
[1, 4, 9]
```

The result keeps the input order even when the callbacks finish in another order. Here the first job
takes the longest:

```ant
type Job = { name: string, delay: Time.Duration }

let jobs = [Job{ name: "slow", delay: 30ms }, Job{ name: "fast", delay: 10ms }]
let names = Parallel.map(jobs, func(job: Job, work: Context): string {
  Time.sleep(job.delay, ctx: work)
  job.name
}, limit: 2)
println(names)
```

```text
["slow", "fast"]
```

## Scope

An empty array gives an empty array, and no callback runs:

```ant
let empty: Array[int] = []
println(Parallel.map(empty, func(n: int, _: Context) { n }, limit: 4))
```

```text
[]
```

Pass a named function that takes the element and a `Context`:

```ant
type Order = { id: int, items: Array[int] }

func total(o: Order, _: Context): int {
  o.items.sum() ?? 0
}

let orders = [Order{ id: 1, items: [5, 7] }, Order{ id: 2, items: [] }]
println(Parallel.map(orders, total, limit: 2))
```

```text
[12, 0]
```

The callback can return any type, records and unions included. An `Err` is one result among the
others; it stops nothing:

```ant
let parsed = Parallel.map(["4", "x", "15"], func(s: string, _: Context): Result[int, Error] {
  let n = s.toInt() ?? return Result.fail("not a number: ${s}")
  Result.ok(n)
}, limit: 3)
println(parsed)
```

```text
[Ok{ value: 4 }, Err{ error: Error{ message: "not a number: x" } }, Ok{ value: 15 }]
```

Call it in a pipe. A `.` call on an array finds `Array.map`, not this function:

```ant
let shout = ["ant", "bee"] | Parallel.map(func(w: string, _: Context) { w.toUpper() }, limit: 2)
println(shout)
```

```text
["ANT", "BEE"]
```

## Named arguments

### limit

`limit` is the most callbacks running at once. With `limit: 1` the elements run one after another,
and each callback's deferred blocks end before the next element starts:

```ant
let upper = Parallel.map(["a", "b"], func(s: string, _: Context): string {
  println("start ${s}")
  defer { println("end ${s}") }
  s.toUpper()
}, limit: 1)
println(upper)
```

```text
start a
end a
start b
end b
["A", "B"]
```

For a goroutine per element, pass the length of the array:

```ant
let ids = [101, 102, 103]
let labels = Parallel.map(ids, func(id: int, _: Context) { "user-${id}" }, limit: ids.length())
println(labels)
```

```text
["user-101", "user-102", "user-103"]
```

### ctx

The scope context is a child of `ctx`. When `ctx` ends before every element is admitted, the call
waits for the callbacks it started and returns `null`. Here the deadline passes during the first
element, so the other two never start:

```ant
let deadline = Context.new(50ms)
let done = Parallel.map([1, 2, 3], func(n: int, work: Context): int {
  Time.sleep(1s, ctx: work)
  n
}, limit: 1, ctx: deadline)
println(done)
println(deadline.reason())
```

```text
null
timeout
```

When `ctx` ends after the last element is admitted, the call returns every result. The callbacks see
their context end and decide what to return:

```ant
let all = Context.new()
go {
  Time.sleep(50ms)
  all.cancel()
}
let reasons = Parallel.map([1, 2, 3], func(_: int, work: Context): string {
  work.wait()
  work.reason() ?? "running"
}, limit: 3, ctx: all)
println(reasons)
```

```text
["cancelled", "cancelled", "cancelled"]
```

A `ctx` that is done already admits nothing, so the result is `null`:

```ant
let stopped = Context.new()
stopped.cancel()
println(Parallel.map([1, 2], func(n: int, _: Context) { n }, limit: 2, ctx: stopped))
```

```text
null
```

## Applications

Read several files at once. Each callback returns a `Result`, and `Result.all` turns the array into
one `Result`:

<!-- test: file a.txt
one
two
-->
<!-- test: file b.txt
three
-->

```ant
func lineCounts(paths: Array[string]): Result[Array[int], Error] {
  let read = func(path: string, _: Context) { File.read(path) }
  let texts = Parallel.map(paths, read, limit: 4) ?? return Result.fail("cancelled")
  let all = try (texts | Result.all())
  Result.ok(all.map({ x.lines().length() }))
}

println(lineCounts(["a.txt", "b.txt"]))
println(lineCounts(["a.txt", "missing.txt"]))
```

```text
Ok{ value: [2, 1] }
Err{ error: Error{ message: "open missing.txt: no such file or directory" } }
```

Count the words of each document in parallel, then add the counts up:

```ant
func wordCount(text: string, _: Context): int {
  text.fields().length()
}

let docs = ["the quick fox", "jumps over", "the lazy dog"]
let counts = Parallel.map(docs, wordCount, limit: 2) ?? []
println(counts)
println(counts.sum())
```

```text
[3, 2, 3]
8
```

## Properties and relations

For a callback that ignores its context, the result equals `Array.map`:

```ant
let xs = [3, 1, 4, 1, 5]
let squares = Parallel.map(xs, func(n: int, _: Context) { n * n }, limit: 2)
assert(squares == xs.map({ x * x }))
```

The limit changes how many callbacks run at once, not the result:

```ant
func cube(n: int, _: Context): int { n * n * n }

let xs = [1, 2, 3, 4]
assert(Parallel.map(xs, cube, limit: 1) == Parallel.map(xs, cube, limit: xs.length()))
```

## Possible issues

The result is nullable, so use it as an array only after handling `null`:

```ant
let lengths = Parallel.map(["ant"], func(s: string, _: Context) { s.length() }, limit: 1)
println(lengths.length())
// error: Array[int] | null has no field or function length on every member: narrow the value first
```

`f` takes two parameters, so a lambda does not fit. Write a `func` with named parameters:

```ant
let doubled: Array[int] | null = Parallel.map([1, 2], { x * 2 }, limit: 2)
// error: a lambda has at most one parameter, x; func(int, Context): int expects 2: write func with named parameters
```

A `limit` below 1 panics, even when the array is empty:

```ant
let empty: Array[int] = []
let none = Parallel.map(empty, func(n: int, _: Context) { n }, limit: 0) // panic: Parallel.map needs a limit of at least 1, got 0
```

A panic in one callback cancels the scope context, so a callback that waits on it ends too. Then the
call panics with the original message, which `Panic.catch` can stop:

```ant
let caught = Panic.catch({
  Parallel.map([1, 2, 3], func(n: int, work: Context): int {
    if n == 2 { panic("bad input 2") }
    work.wait()
    n
  }, limit: 3)
})
if caught is Err {
  println(caught.error.message)
}
```

```text
bad input 2
```

Cancellation is cooperative. A callback that sleeps without its context holds the call for the full
sleep: the deadline passes during the first element, yet the call returns `null` only after that
callback wakes:

```ant
let deadline = Context.new(50ms)
let done = Parallel.map([1, 2, 3], func(n: int, _: Context): int {
  Time.sleep(200ms)
  println("woke ${n}")
  n
}, limit: 1, ctx: deadline)
println(done)
```

```text
woke 1
null
```

Callbacks that update the same variable need synchronization. An `Atomic` counts across them:

```ant
let failures = Atomic.new(0)
let numbers = Parallel.map(["4", "x", "15", "y"], func(s: string, _: Context): int {
  let n = s.toInt()
  if n == null { failures.add(1) }
  n ?? 0
}, limit: 4)
println(numbers)
println(failures.load())
```

```text
[4, 0, 15, 0]
2
```

## See also

[Parallel.each](each.md) · [Parallel.run](run.md) · [Parallel.race](race.md) ·
[Parallel.any](any.md) · [Array.map](../Array/map.md) · [Context.cancel](../Context/cancel.md) ·
[Result.all](../Result/all.md) · [Panic.catch](../Panic/catch.md)
