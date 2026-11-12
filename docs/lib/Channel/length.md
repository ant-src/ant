<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Channel.length

```ant signature
func Channel.length[T](c: Channel[T]): int
```

Channel.length returns the number of values buffered in c now.

<!-- end generated -->

## Details

- The count is a snapshot: another goroutine may send or receive right after the call.
- An unbuffered channel holds nothing, so its length is always `0`, even while a sender waits.
- A closed channel keeps its buffered values until they are received, so its length counts them.
- `len(c)` is the same count.

## Basic examples

Count the values a buffered channel holds:

```ant
let c: Channel[string] = Channel.new(capacity: 3)
c.send("a")
c.send("b")
println(Channel.length(c))
c.receive()
println(Channel.length(c))
```

```text
2
1
```

## Scope

An unbuffered channel hands each value over directly, so its length stays `0` whether the sender has
started yet or not:

```ant
let c: Channel[int] = Channel.new()
go { c.send(1) }
println(c.length())
println(c.receive())
println(c.length())
```

```text
0
1
0
```

A closed channel still holds what was sent before the close. Each receive takes one:

```ant
let c: Channel[int] = Channel.new(capacity: 4)
for n in [1, 2, 3] {
  c.send(n)
}
c.close()
println("${c.isClosed()} ${c.length()}")
for n in c.values() {
  println("${n}, ${c.length()} left")
}
```

```text
true 3
1, 2 left
2, 1 left
3, 0 left
```

A ticker holds at most one tick, and none once its context is done:

```ant
let ctx = Context.new()
let ticks = Time.ticker(Time.duration(milliseconds: 10), ctx)
Time.sleep(Time.duration(milliseconds: 50))
println(ticks.length())
ctx.cancel()
println(ticks.length())
```

```text
1
0
```

## Applications

Take a batch of what is buffered without waiting for more. The loop is safe only because `drain` is
the one receiver, so no other goroutine can take a value between the check and the receive:

```ant
func drain(c: Channel[string]): Array[string] {
  let batch: Array[string] = []
  while c.length() > 0 {
    let event = c.receive()
    if event != null { batch.push(event) }
  }
  batch
}

let events: Channel[string] = Channel.new(capacity: 8)
for e in ["click", "scroll", "click"] {
  events.send(e)
}
println(drain(events))
println(drain(events))
```

```text
["click", "scroll", "click"]
[]
```

## Properties and relations

`len(c)` and `c.length()` give the same count. The length never exceeds the capacity, and a send
that finds room adds one:

```ant
let c: Channel[int] = Channel.new(capacity: 2)
c.send(1)
assert(len(c) == c.length())
assert(c.length() <= 2)
let before = c.length()
c.send(2)
assert(c.length() == before + 1)
```

## Possible issues

A check of the length followed by a send or a receive is a race when other goroutines share the
channel: the buffer can fill or empty between the two. To send only when there is room, write a
`select` with `else`, which decides and sends in one step:

```ant
let queue: Channel[int] = Channel.new(capacity: 2)
for job in [1, 2, 3] {
  select {
    Channel.send(queue, job) { println("queued ${job}") }
    else { println("dropped ${job}, ${queue.length()} waiting") }
  }
}
```

```text
queued 1
queued 2
dropped 3, 2 waiting
```

A length of `0` does not mean the channel is closed. Ask [Channel.isClosed](isClosed.md), or receive
until `null`.

## See also

[Channel.new](new.md) · [Channel.receive](receive.md) · [Channel.send](send.md) ·
[Channel.isClosed](isClosed.md) · [Channel.values](values.md) · [Time.ticker](../Time/ticker.md)
