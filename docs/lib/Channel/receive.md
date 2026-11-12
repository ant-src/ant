<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.receive

```ant signature
func Channel.receive[T](c: Channel[T], ctx: Context | null = null): T | null
```

Channel.receive takes the next value from c, blocking until one is available. It returns null when c
is closed and drained, or when ctx is done first; [`Context.isDone`](../Context/isDone.md) tells the
two apart. When T includes null, a null value looks like the end, so wrap such values.

<!-- end generated -->

## Details

- Blocks until a value is available. On a rendezvous channel, made by `Channel.new()` with no
  capacity, that is until a sender hands one over.
- Returns `null` once `c` is closed and every buffered value has been received. Every later receive
  returns `null` at once.
- Without `ctx`, a receive on a channel that no goroutine sends on or closes waits until the run is
  cancelled.

## Basic examples

Receive the values a buffered channel holds:

```ant
let c: Channel[string] = Channel.new(2)
c.send("a")
c.send("b")
println(Channel.receive(c))
println(Channel.receive(c))
```

```text
a
b
```

Receive a value that another goroutine sends. On a rendezvous channel the receive waits for it:

```ant
let results: Channel[int] = Channel.new()
go { results.send(6 * 7) }
println(Channel.receive(results))
```

```text
42
```

A closed channel still gives what it buffered, and then `null`:

```ant
let done: Channel[bool] = Channel.new(1)
done.send(true)
done.close()
println(Channel.receive(done))
println(Channel.receive(done))
```

```text
true
null
```

## Scope

The `null` result marks a closed and drained channel. To receive every value until then, loop over
[`Channel.values`](values.md), which ends there:

```ant
let jobs: Channel[int] = Channel.new(3)
for n in [1, 2, 3] {
  jobs.send(n)
}
jobs.close()

var total = 0
for n in jobs.values() {
  total += n
}
println(total)
```

```text
6
```

A channel of a union carries several kinds of event. An `is` check tells them apart, and narrows the
value after it:

```ant
type Line = { text: string }
type Quit = {}
type Event = Line | Quit

let events: Channel[Event] = Channel.new()
go {
  events.send(Line{ text: "hello" })
  events.send(Line{ text: "world" })
  events.send(Quit{})
}
for event in events.values() {
  if event is Quit { break }
  println(event.text)
}
```

```text
hello
world
```

In a `select`, `Channel.receive(c)` is a head: the `select` waits until one of its heads can
proceed. `let` binds the received value, `T | null`, in the block:

```ant
let numbers: Channel[int] = Channel.new(1)
let names: Channel[string] = Channel.new(1)
names.send("ant")
select {
  let n = Channel.receive(numbers) { println("number ${n}") }
  let name = Channel.receive(names) { println("name ${name}") }
}
```

```text
name ant
```

A `select` with `else` receives only when a value is ready, and never waits:

```ant
let c: Channel[int] = Channel.new(1)
let got = select {
  let n = Channel.receive(c) { n }
  else { null }
}
println(got)
```

```text
null
```

## Named arguments

### ctx

A context bounds the wait. When it is done first, the receive returns `null`, and the context says
why:

```ant
let replies: Channel[string] = Channel.new()
let ctx = Context.new(timeout: Time.duration(milliseconds: 50))
println(replies.receive(ctx: ctx))
println(ctx.isDone())
println(ctx.reason())
```

```text
null
true
timeout
```

A closed channel gives `null` too, while the context is still running:

```ant
let closed: Channel[string] = Channel.new()
closed.close()
let ctx = Context.new()
println(closed.receive(ctx: ctx))
println(ctx.isDone())
```

```text
null
false
```

## Applications

### Collect results from goroutines

Start one goroutine per input, then receive one result for each:

```ant
let words = ["ant", "bee", "wasp"]
let lengths: Channel[int] = Channel.new()
for word in words {
  go { lengths.send(word.length()) }
}

var total = 0
for _ in words {
  total += lengths.receive() ?? 0
}
println(total)
```

```text
10
```

### Wait for a reply with a deadline

Turn a missing reply into an error, and use the context to say whether time ran out or the channel
closed:

```ant
func awaitReply(replies: Channel[string], wait: Time.Duration): Result[string, Error] {
  let ctx = Context.new(timeout: wait)
  let reply = replies.receive(ctx: ctx)
  if reply == null {
    let why = if ctx.isDone() { "no reply in time" } else { "replies closed" }
    return Result.fail(why)
  }
  Result.ok(reply)
}

let replies: Channel[string] = Channel.new(1)
let wait = Time.duration(milliseconds: 50)
replies.send("pong")
println(awaitReply(replies, wait))
println(awaitReply(replies, wait))
replies.close()
println(awaitReply(replies, wait))
```

```text
Ok{ value: "pong" }
Err{ error: Error{ message: "no reply in time" } }
Err{ error: Error{ message: "replies closed" } }
```

## Properties and relations

Each receive takes one buffered value, so the length drops by one. After a close and a drain, every
receive is `null`:

```ant
let c: Channel[int] = Channel.new(2)
c.send(1)
c.send(2)
c.close()
assert(c.length() == 2)
assert(c.receive() == 1)
assert(c.length() == 1)
assert(c.receive() == 2)
assert(c.receive() == null)
assert(c.receive() == null)
assert(c.isClosed())
```

## Possible issues

### A null value looks like the end

When `T` includes `null`, a received `null` may be a value or the end of the channel:

```ant
let maybe: Channel[int | null] = Channel.new(2)
maybe.send(null)
maybe.send(1)
println(maybe.receive())
```

```text
null
```

Wrap such values in a record, so that only the end is `null`:

```ant
type Slot = { value: int | null }

let slots: Channel[Slot] = Channel.new(1)
slots.send(Slot{ value: null })
slots.close()
println(slots.receive())
println(slots.receive())
```

```text
Slot{ value: null }
null
```

### The dot form is not a select head

`c.receive()` makes the same call elsewhere, but a `select` head takes only the namespace form:

```ant
let c: Channel[string] = Channel.new()
select {
  c.receive() { println("got one") }
  // error: a select clause's head must be a call of Channel.receive, Channel.send, Context.wait, WaitGroup.wait or Time.sleep
  else { println("none") }
}
```

### A closed channel is always ready in a select

A receive head on a closed channel can always proceed and binds `null`. A loop around the `select`
leaves on that `null`, or it spins:

```ant
let jobs: Channel[string] = Channel.new(2)
jobs.send("build")
jobs.send("test")
jobs.close()

let ticks: Channel[Time.Instant] = Channel.new()
while true {
  select {
    let job = Channel.receive(jobs) {
      if job == null { break }
      println(job)
    }
    Channel.receive(ticks) { println("tick") }
  }
}
```

```text
build
test
```

## See also

- [Channel.send](send.md)
- [Channel.values](values.md)
- [Channel.close](close.md)
- [Channel.new](new.md)
- [Context.new](../Context/new.md)
- [Context.isDone](../Context/isDone.md)
- [Parallel.each](../Parallel/each.md)
- [Time.ticker](../Time/ticker.md)
