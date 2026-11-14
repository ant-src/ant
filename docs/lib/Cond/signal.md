<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond.signal

```ant signature
func Cond.signal(c: Cond): void
```

Cond.signal wakes one goroutine waiting on c, if any.

<!-- end generated -->

## Details

- A signal with no goroutine waiting does nothing, and a later `Cond.wait` does not see it. Keep the
  condition itself in shared state, guarded by the mutex.
- The woken goroutine returns from `Cond.wait` only once it holds the mutex again. Make the change
  under the mutex, and have the waiter check the condition in a `while` loop. To wake every waiter,
  call `Cond.broadcast`.

## Basic examples

A goroutine opens a gate under the mutex and signals. The main goroutine waits until the gate is
open:

```ant
type Gate = { open: bool }

let m = Mutex.new()
let opened = Cond.new(m)
let gate = Gate{ open: false }

go {
  m.withLock({ gate.open = true })
  Cond.signal(opened)
}

m.lock()
while !gate.open {
  opened.wait()
}
m.unlock()
println("the gate is open")
```

```text
the gate is open
```

The `while` loop also covers a goroutine that opens the gate before the main goroutine starts to
wait: the loop then never calls `Cond.wait`, and the lost signal does no harm.

## Scope

A signal with nobody waiting returns at once, outside the lock or inside it:

```ant
let m = Mutex.new()
let changed = Cond.new(m)
changed.signal()
m.withLock({ changed.signal() })
println("no one was waiting")
```

```text
no one was waiting
```

One signal per unit of work wakes one waiter per unit. Here each ticket releases one customer:

```ant
type Desk = { tickets: int, served: Array[string] }

let m = Mutex.new()
let ticketReady = Cond.new(m)
let desk = Desk{ tickets: 0, served: [] }
let wg = WaitGroup.new()

for name in ["Ann", "Bob", "Cy"] {
  wg.add(1)
  go {
    defer { wg.done() }
    m.lock()
    defer { m.unlock() }
    while desk.tickets == 0 {
      ticketReady.wait()
    }
    desk.tickets -= 1
    desk.served.push(name)
  }
}

for _ in Iter.range(0, 3) {
  m.withLock({ desk.tickets += 1 })
  ticketReady.signal()
}
wg.wait()
println("${desk.served.length()} served")
```

```text
3 served
```

## Applications

A blocking queue: `put` adds an item and signals, and `take` waits while the queue is empty. Each
item wakes at most one taker, the one that will remove it:

```ant
type Queue = { items: Array[string], lock: Mutex, nonEmpty: Cond }

func newQueue(): Queue {
  let lock = Mutex.new()
  Queue{ items: [], lock: lock, nonEmpty: Cond.new(lock) }
}

func put(q: Queue, item: string) {
  q.lock.withLock({ q.items.push(item) })
  q.nonEmpty.signal()
}

func take(q: Queue): string {
  q.lock.lock()
  defer { q.lock.unlock() }
  while q.items.isEmpty() {
    q.nonEmpty.wait()
  }
  q.items.removeAt(0)
}

let q = newQueue()
go {
  for word in ["ant", "bee", "cat"] {
    put(q, word)
  }
}
for _ in Iter.range(0, 3) {
  println(take(q))
}
```

```text
ant
bee
cat
```

## Possible issues

A signal wakes only one of the goroutines waiting. When every waiter waits for the same change, the
others sleep on. Here three runners wait for the start, and one signal starts one of them:

```ant
type Race = { waiting: int, started: bool, running: int }

let m = Mutex.new()
let arrived = Cond.new(m)
let start = Cond.new(m)
let ran = Cond.new(m)
let race = Race{ waiting: 0, started: false, running: 0 }

for _ in Iter.range(0, 3) {
  go {
    m.lock()
    defer { m.unlock() }
    race.waiting += 1
    arrived.signal()
    while !race.started {
      start.wait()
    }
    race.running += 1
    ran.signal()
  }
}

m.lock()
while race.waiting < 3 {
  arrived.wait()
}
race.started = true
start.signal()
while race.running == 0 {
  ran.wait()
}
println("${race.running} of 3 running")
m.unlock()
```

```text
1 of 3 running
```

Call `Cond.broadcast(start)` for a change that every waiter is waiting for.

A signal can also reach the wrong waiter. This queue holds one item. Consumers wait for an item and
the producer waits for room, all on one cond. When the first consumer takes an item and signals, the
signal can wake the second consumer instead of the producer. The queue is still empty, so that
consumer waits again, the signal is used up, and the producer sleeps on with nothing left to wake
it. The program then stalls on some runs and not on others:

<!-- test: check-only which waiter a signal wakes changes between runs -->

```ant
type Queue = { items: Array[int], limit: int }

let m = Mutex.new()
let changed = Cond.new(m)
let q = Queue{ items: [], limit: 1 }
let wg = WaitGroup.new()

for _ in Iter.range(0, 2) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      while q.items.isEmpty() {
        changed.wait()
      }
      q.items.removeAt(0)
      changed.signal()
    })
  }
}

wg.add(1)
go {
  defer { wg.done() }
  for n in Iter.range(0, 2) {
    m.withLock({
      while q.items.length() == q.limit {
        changed.wait()
      }
      q.items.push(n)
      changed.signal()
    })
  }
}

let outcome = select {
  WaitGroup.wait(wg) { "all done" }
  Time.sleep(1s) { "stalled" }
}
println(outcome)
```

Give each condition its own cond, so that a signal reaches only goroutines that wait for that
change, as the runners above use `arrived`, `start` and `ran`. Here consumers wait on `notEmpty` and
the producer on `notFull`:

```ant
type Queue = { items: Array[int], limit: int }

let m = Mutex.new()
let notEmpty = Cond.new(m)
let notFull = Cond.new(m)
let q = Queue{ items: [], limit: 1 }
let wg = WaitGroup.new()

for _ in Iter.range(0, 2) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      while q.items.isEmpty() {
        notEmpty.wait()
      }
      q.items.removeAt(0)
      notFull.signal()
    })
  }
}

wg.add(1)
go {
  defer { wg.done() }
  for n in Iter.range(0, 2) {
    m.withLock({
      while q.items.length() == q.limit {
        notFull.wait()
      }
      q.items.push(n)
      notEmpty.signal()
    })
  }
}

let outcome = select {
  WaitGroup.wait(wg) { "all done" }
  Time.sleep(1s) { "stalled" }
}
println(outcome)
```

```text
all done
```

Where waiters for different conditions must share one cond, call `Cond.broadcast` instead: every
waiter checks its own condition, and the one that can go on does.

Waiting without checking a condition first can wait forever: a signal sent before the wait is lost,
and `Cond.wait` takes no context to give up with. Always pair the cond with state that the waiter
checks under the mutex, as the examples above do.

## See also

[Cond.wait](wait.md) · [Cond.broadcast](broadcast.md) · [Cond.new](new.md) · [Cond](index.md) ·
[Mutex](../Mutex/index.md) · [Channel.send](../Channel/send.md) ·
[Concurrency guide](../../guide/concurrency.md)
