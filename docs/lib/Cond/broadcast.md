<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond.broadcast

```ant signature
func Cond.broadcast(c: Cond): void
```

Cond.broadcast wakes every goroutine waiting on c.

<!-- end generated -->

## Details

- Only the goroutines waiting at the moment of the call wake. With nobody waiting, a broadcast does
  nothing, and a later `Cond.wait` does not see it.
- Each woken goroutine reacquires the mutex before its `wait` returns, so they resume one at a time.
- The broadcast orders nothing by itself; the mutex does. Change the shared state while holding the
  mutex, and have each waiter check it again in a `while` loop.

## Basic examples

Hold several goroutines at a start line. Each waits until `ready` is true; one broadcast under the
mutex releases them all:

```ant
let m = Mutex.new()
let start = Cond.new(m)
var ready = false
let ran = Atomic.new(0)
let wg = WaitGroup.new()

for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      while !ready { start.wait() }
    })
    ran.add(1)
  }
}

m.withLock({
  ready = true
  Cond.broadcast(start)
})
wg.wait()
println("${ran.load()} goroutines ran")
```

```text
3 goroutines ran
```

A broadcast with no goroutine waiting returns at once and changes nothing:

```ant
let c = Cond.new(Mutex.new())
Cond.broadcast(c)
println("nobody was waiting")
```

```text
nobody was waiting
```

## Scope

Waiters on one `Cond` can wait for different conditions. After each change, a broadcast wakes all of
them, and each checks its own condition: here the one that needs 1 item goes on after the first
delivery, the others wait for more:

```ant
let m = Mutex.new()
let restocked = Cond.new(m)
var stock = 0
let served: Array[int] = []
let wg = WaitGroup.new()

for need in [1, 3, 5] {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      while stock < need { restocked.wait() }
      served.push(need)
    })
  }
}

for _ in Array.range(0, 3) {
  m.withLock({
    stock += 2
    restocked.broadcast()
  })
}
wg.wait()
println(served | Array.sort())
```

```text
[1, 3, 5]
```

## Applications

### Closing a work queue

Workers take jobs from a shared list and sleep on a `Cond` while it is empty. One new job needs one
worker, so `Cond.signal` is enough for it. Closing the queue concerns every idle worker, so the
close broadcasts:

```ant
type Queue = { m: Mutex, changed: Cond, jobs: Array[int], closed: bool }

func Queue.make(): Queue {
  let m = Mutex.new()
  Queue{ m: m, changed: Cond.new(m), jobs: [], closed: false }
}

func Queue.put(q: Queue, job: int) {
  q.m.withLock({
    q.jobs.push(job)
    q.changed.signal()
  })
}

func Queue.close(q: Queue) {
  q.m.withLock({
    q.closed = true
    q.changed.broadcast()
  })
}

// Queue.take returns null once the queue is closed and empty.
func Queue.take(q: Queue): int | null {
  q.m.withLock({
    while q.jobs.isEmpty() && !q.closed { q.changed.wait() }
    q.jobs.pop()
  })
}

let q = Queue.make()
let total = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 4) {
  wg.add(1)
  go {
    defer { wg.done() }
    while true {
      let job = q.take() ?? break
      total.add(job)
    }
  }
}

for n in Array.range(1, 11) {
  q.put(n)
}
q.close()
wg.wait()
println(total.load())
```

```text
55
```

### A barrier

A barrier holds every goroutine of a group until the last one arrives, which then broadcasts. The
`round` counter lets the barrier be used again: a waiter leaves once the round it arrived in is
over. No goroutine starts phase 1 before all have finished phase 0:

```ant
type Barrier = { m: Mutex, done: Cond, size: int, arrived: int, round: int }

func Barrier.make(size: int): Barrier {
  let m = Mutex.new()
  Barrier{ m: m, done: Cond.new(m), size: size, arrived: 0, round: 0 }
}

func Barrier.await(b: Barrier) {
  b.m.lock()
  defer { b.m.unlock() }
  let round = b.round
  b.arrived += 1
  if b.arrived == b.size {
    b.arrived = 0
    b.round += 1
    b.done.broadcast()
    return
  }
  while b.round == round { b.done.wait() }
}

let barrier = Barrier.make(3)
let logM = Mutex.new()
let log: Array[int] = []
let wg = WaitGroup.new()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    for phase in Array.range(0, 2) {
      logM.withLock({ log.push(phase) })
      barrier.await()
    }
  }
}
wg.wait()
println(log)
```

```text
[0, 0, 0, 1, 1, 1]
```

## Properties and relations

Every goroutine waiting at the moment of the broadcast wakes, and each sees what was written under
the mutex before it. Here the main goroutine first waits until all three are waiting, on a second
`Cond` over the same mutex:

```ant
let m = Mutex.new()
let release = Cond.new(m)
let arrived = Cond.new(m)
var waiting = 0
var message = ""
let seen: Array[string] = []
let wg = WaitGroup.new()

for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      waiting += 1
      arrived.signal()
      while message == "" { release.wait() }
      seen.push(message)
    })
  }
}

m.withLock({
  while waiting < 3 { arrived.wait() }
  message = "go"
  release.broadcast()
})
wg.wait()
assert(seen == ["go", "go", "go"])
```

## Possible issues

`Cond.signal` where every waiter needs to know leaves all but one of them asleep. Here three
goroutines are waiting, the signal wakes one, and the other two are still blocked when the main
goroutine stops waiting for them:

```ant
let m = Mutex.new()
let open = Cond.new(m)
let arrived = Cond.new(m)
var waiting = 0
var isOpen = false
let passed = Atomic.new(0)
let wg = WaitGroup.new()

for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      waiting += 1
      arrived.signal()
      while !isOpen { open.wait() }
    })
    passed.add(1)
  }
}

m.withLock({
  while waiting < 3 { arrived.wait() }
  isOpen = true
  open.signal()
})
wg.wait(Context.new(200ms))
println("${passed.load()} of 3 passed")
```

```text
1 of 3 passed
```

With `open.broadcast()` all three pass.

A waiter that checks its condition with `if` instead of `while` goes on after any wakeup. A
broadcast wakes every waiter, but a change may satisfy only one of them. Here two consumers wait for
a job, the main goroutine adds one and broadcasts, and the consumer that wakes second finds the
queue empty again, so its `pop` returns `null`:

```ant
let m = Mutex.new()
let changed = Cond.new(m)
let arrived = Cond.new(m)
var waiting = 0
let jobs: Array[int] = []
let got: Array[int | null] = []
let wg = WaitGroup.new()

for _ in Array.range(0, 2) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      waiting += 1
      arrived.signal()
      if jobs.isEmpty() { changed.wait() }
      got.push(jobs.pop())
    })
  }
}

m.withLock({
  while waiting < 2 { arrived.wait() }
  jobs.push(1)
  changed.broadcast()
})
wg.wait()
println(got.contains(null))
```

```text
true
```

With `while jobs.isEmpty() { changed.wait() }` the second consumer checks again, finds no job, and
goes back to waiting:

```ant
let m = Mutex.new()
let changed = Cond.new(m)
let arrived = Cond.new(m)
var waiting = 0
let jobs: Array[int] = []
let got: Array[int | null] = []
let wg = WaitGroup.new()

for _ in Array.range(0, 2) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.withLock({
      waiting += 1
      arrived.signal()
      while jobs.isEmpty() { changed.wait() }
      got.push(jobs.pop())
    })
  }
}

m.withLock({
  while waiting < 2 { arrived.wait() }
  jobs.push(1)
  changed.broadcast()
})
wg.wait(Context.new(200ms))
m.withLock({ println(got) })
```

```text
[1]
```

A broadcast is not remembered. A goroutine that calls `Cond.wait` after the broadcast, without a
condition to check first, waits for a broadcast that already happened:

```ant
let m = Mutex.new()
let c = Cond.new(m)
c.broadcast()

let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  m.withLock({ c.wait() })
}
wg.wait(Context.new(200ms))
println("still waiting")
```

```text
still waiting
```

Keep the state in a variable guarded by the mutex, set it before the broadcast, and wait in a loop
such as `while !ready { c.wait() }`: a goroutine that comes late sees `ready` and does not wait.

## See also

[Cond.signal](signal.md) · [Cond.wait](wait.md) · [Cond.new](new.md) · [Cond](index.md) ·
[Mutex.withLock](../Mutex/withLock.md) · [WaitGroup](../WaitGroup/index.md) ·
[Channel.close](../Channel/close.md) · [Concurrency guide](../../guide/concurrency.md)
