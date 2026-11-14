<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Cond.new

```ant signature
func Cond.new(m: Mutex): Cond
```

Cond.new returns a new condition variable over m.

<!-- end generated -->

## Details

- `Cond.wait` releases `m` while it waits and reacquires it before returning, so a goroutine locks
  `m` before it waits.
- Several condition variables can share one mutex, one for each condition the goroutines wait for.

## Basic examples

Make a condition variable over a mutex, and wait under that mutex until another goroutine changes
the state and signals:

```ant
let mu = Mutex.new()
let changed = Cond.new(mu)
var ready = false

go {
  mu.withLock({
    ready = true
    changed.signal()
  })
}

mu.withLock({
  while !ready {
    changed.wait()
  }
})
println("ready")
```

```text
ready
```

A new condition variable is a value of type `Cond`:

```ant
let mu = Mutex.new()
Cond.new(mu) // Cond
```

## Scope

A record can hold the mutex, the condition variable over it, and the state they guard:

```ant
type Counter = { mu: Mutex, reachedTen: Cond, count: int }

func newCounter(): Counter {
  let mu = Mutex.new()
  Counter{ mu, reachedTen: Cond.new(mu), count: 0 }
}

func Counter.increment(c: Counter): void {
  c.mu.lock()
  defer { c.mu.unlock() }
  c.count += 1
  if c.count == 10 { c.reachedTen.signal() }
}

let counter = newCounter()
for _ in Iter.range(0, 10) {
  go { counter.increment() }
}

counter.mu.withLock({
  while counter.count < 10 {
    counter.reachedTen.wait()
  }
})
println(counter.count)
```

```text
10
```

## Applications

A bounded queue: `put` waits while the queue is full, and `take` waits while it is empty. Each side
signals the condition the other side waits for:

```ant
type Queue = { mu: Mutex, notEmpty: Cond, notFull: Cond, items: Array[int], limit: int }

func newQueue(limit: int): Queue {
  let mu = Mutex.new()
  Queue{ mu, notEmpty: Cond.new(mu), notFull: Cond.new(mu), items: [], limit }
}

func Queue.put(q: Queue, n: int): void {
  q.mu.lock()
  defer { q.mu.unlock() }
  while q.items.length() == q.limit {
    q.notFull.wait()
  }
  q.items.push(n)
  q.notEmpty.signal()
}

func Queue.take(q: Queue): int {
  q.mu.lock()
  defer { q.mu.unlock() }
  while q.items.isEmpty() {
    q.notEmpty.wait()
  }
  let n = q.items.removeAt(0)
  q.notFull.signal()
  n
}

let queue = newQueue(2)
go {
  for n in Iter.range(1, 6) {
    queue.put(n)
  }
}

var total = 0
for _ in Iter.range(1, 6) {
  total += queue.take()
}
println(total)
```

```text
15
```

A start gate: workers wait until the gate opens, and `broadcast` releases all of them at once:

```ant
let mu = Mutex.new()
let opened = Cond.new(mu)
var open = false
let started = Atomic.new(0)
let wg = WaitGroup.new()

for _ in Iter.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    mu.withLock({
      while !open {
        opened.wait()
      }
    })
    started.add(1)
  }
}

mu.withLock({
  open = true
  opened.broadcast()
})
wg.wait()
println("${started.load()} workers started")
```

```text
3 workers started
```

## Properties and relations

A condition variable equals itself and no other, even one over the same mutex:

```ant
let mu = Mutex.new()
let a = Cond.new(mu)
let b = Cond.new(mu)
assert(a == a)
assert(a != b)
```

With no goroutine waiting, `signal` and `broadcast` do nothing:

```ant
let c = Cond.new(Mutex.new())
c.signal()
c.broadcast()
println("no waiters")
```

```text
no waiters
```

## Possible issues

A condition variable works only over a `Mutex`. `ant check` reports any other sync object, such as
an `RwMutex`, as a compile error:

```ant
let rw = RwMutex.new()
Cond.new(rw) // error: cannot use RwMutex where Mutex is expected
```

`wait` panics when the goroutine has not locked the condition variable's mutex:

```ant
let mu = Mutex.new()
let changed = Cond.new(mu)
changed.wait() // panic: Cond.wait: the Mutex is not locked
```

Wait in a `while` loop that checks the condition, not after an `if`. A signal is not remembered: one
sent before a goroutine waits wakes nobody, and a woken goroutine finds the state as the goroutines
that ran before it left it. The loop checks the state under the mutex before every wait and after
every wake.

## See also

[Cond.wait](wait.md) · [Cond.signal](signal.md) · [Cond.broadcast](broadcast.md) ·
[Mutex.new](../Mutex/new.md) · [Mutex.withLock](../Mutex/withLock.md) ·
[WaitGroup.new](../WaitGroup/new.md) · [Channel.new](../Channel/new.md)
