<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.new

```ant signature
func Channel.new[T](capacity: int = 0): Channel[T]
```

Channel.new returns a new open channel that buffers up to capacity values. A capacity of 0, the
default, makes a rendezvous channel, whose send waits for a receiver. It panics if capacity \< 0 or
is above 2^31 - 1.

<!-- end generated -->

## Details

- Nothing in the arguments fixes the element type `T`, so the call needs an expected type: an
  annotated `let` or `var`, a parameter, or a record field.
- With a capacity above 0, `send` returns at once while the channel holds fewer than `capacity`
  values.
- The new channel is open and empty. Each call makes a distinct channel: channels compare by
  identity.

## Basic examples

Make a channel of strings and receive a value that another goroutine sends:

```ant
let greetings: Channel[string] = Channel.new()
go { greetings.send("hello") }
println(greetings.receive())
```

```text
hello
```

Give a capacity to buffer values. Here two sends return without a receiver:

```ant
let queue: Channel[int] = Channel.new(2)
queue.send(1)
queue.send(2)
println(queue.length())
```

```text
2
```

## Scope

The element type can come from a parameter or a record field instead of an annotation:

```ant
type Job = { id: int }
type Pool = { jobs: Channel[Job], results: Channel[string] }

func fill(jobs: Channel[Job]): void {
  for id in [1, 2] {
    jobs.send(Job{ id })
  }
}

let pool = Pool{ jobs: Channel.new(4), results: Channel.new() }
fill(pool.jobs)
println(pool.jobs.length())
```

```text
2
```

The element type can be a union, so one channel carries several kinds of event. A receiver tells
them apart with `is`:

```ant
type Click = { x: int, y: int }
type Key = { name: string }

let events: Channel[Click | Key] = Channel.new(4)
events.send(Click{ x: 1, y: 2 })
events.send(Key{ name: "q" })
events.close()
for event in events.values() {
  if event is Click {
    println("click at ${event.x}, ${event.y}")
  } else {
    println("key ${event.name}")
  }
}
```

```text
click at 1, 2
key q
```

## Named arguments

### capacity

`capacity` is the number of values the channel holds before `send` waits. It can be passed by name:

```ant
let lines: Channel[string] = Channel.new(capacity: 3)
for line in ["a", "b", "c"] {
  lines.send(line)
}
println(lines.length())
```

```text
3
```

## Applications

Hand values from a producer to a consumer. The producer closes the channel when it is done, which
ends the consumer's loop:

```ant
func produce(out: Channel[string]): void {
  defer { out.close() }
  for word in ["ant", "bee", "wasp"] {
    out.send(word)
  }
}

let words: Channel[string] = Channel.new()
go { produce(words) }
for word in words.values() {
  println("${word}: ${word.length()}")
}
```

```text
ant: 3
bee: 3
wasp: 4
```

Collect one result from each goroutine. A capacity equal to the number of senders lets every sender
finish without waiting for the receiver:

```ant
let squares: Channel[int] = Channel.new(3)
for n in [1, 2, 3] {
  go { squares.send(n * n) }
}
var total = 0
for _ in Iter.range(0, 3) {
  total += squares.receive() ?? 0
}
println(total)
```

```text
14
```

Feed a worker pool. `Parallel.each` receives the jobs and runs at most `limit` of them at once:

```ant
type Job = { path: string }

let jobs: Channel[Job] = Channel.new(capacity: 8)
go {
  defer { jobs.close() }
  for path in ["a.txt", "b.txt", "c.txt"] {
    jobs.send(Job{ path })
  }
}

let total = Atomic.new(0)
Parallel.each(jobs, func(job: Job, ctx: Context) {
  total.add(job.path.length())
}, limit: 2)
println(total.load())
```

```text
15
```

## Properties and relations

A new channel is open and empty:

```ant
let c: Channel[int] = Channel.new(5)
assert(!c.isClosed())
assert(c.length() == 0)
```

Two calls make two channels, which are not equal even with the same element type and capacity:

```ant
let a: Channel[int] = Channel.new()
let b: Channel[int] = Channel.new()
assert(a != b)
assert(a == a)
```

A buffered channel never holds more than `capacity` values, and a receive frees one place:

```ant
let capacity = 3
let c: Channel[string] = Channel.new(capacity)
for s in ["x", "y", "z"] {
  c.send(s)
}
assert(c.length() <= capacity)
let first = c.receive()
assert(first == "x")
assert(c.length() == 2)
```

## Possible issues

Without an expected type, the checker cannot tell what the channel carries:

```ant
let c = Channel.new()
// error: cannot infer the type parameter T of Channel.new; annotate the result
```

A negative capacity panics:

```ant
let c: Channel[int] = Channel.new(-1) // panic: Channel.new needs a capacity of at least 0, got -1
```

So does a capacity above 2^31 − 1:

```ant
let c: Channel[int] = Channel.new(3000000000)
// panic: Channel.new needs a capacity of at most 2147483647, got 3000000000
```

On a channel with the default capacity of 0, a send with no receiver waits forever. Give the send a
context to bound the wait, or give the channel a capacity:

```ant
let c: Channel[int] = Channel.new()
println(c.send(1, Context.new(Time.duration(milliseconds: 20))))
let buffered: Channel[int] = Channel.new(1)
println(buffered.send(1, Context.new(Time.duration(milliseconds: 20))))
```

```text
false
true
```

When the element type includes `null`, a `null` value and the end of a closed channel both receive
as `null`:

```ant
let c: Channel[int | null] = Channel.new(1)
c.send(null)
c.close()
println(c.receive())
println(c.receive())
```

```text
null
null
```

Wrap the value in a record type to tell them apart.

## See also

[Channel.send](send.md) · [Channel.receive](receive.md) · [Channel.close](close.md) ·
[Channel.values](values.md) · [Channel.length](length.md) · [Parallel.each](../Parallel/each.md) ·
[Time.ticker](../Time/ticker.md)
