<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Mutex.unlock

```ant signature
func Mutex.unlock(m: Mutex): void
```

Mutex.unlock unlocks m. It panics if m is not locked.

<!-- end generated -->

## Details

- Everything written before an unlock is visible to every goroutine that locks the mutex after it.
- It panics when `m` is not locked: a mutex never locked, or one unlocked already.
- A mutex has no owner: any goroutine may unlock a mutex that another goroutine locked.
- A panic between `lock` and `unlock` skips the unlock. Call it in a `defer` right after the lock,
  or use `Mutex.withLock`.

## Basic examples

Lock, change the shared array, unlock:

```ant
let m = Mutex.new()
let names: Array[string] = []
Mutex.lock(m)
names.push("Ann")
Mutex.unlock(m)
println(names)
```

```text
["Ann"]
```

Put the unlock in a `defer` right after the lock, so it runs on every way out of the block. Here
three goroutines add to one counter:

```ant
type Counter = { hits: int }

let m = Mutex.new()
let counter = Counter{ hits: 0 }
let wg = WaitGroup.new()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.lock()
    defer { m.unlock() }
    counter.hits += 1
  }
}
wg.wait()
println(counter.hits)
```

```text
3
```

## Scope

A mutex can be locked again once it is unlocked:

```ant
let m = Mutex.new()
m.lock()
m.unlock()
m.lock()
m.unlock()
println("locked twice")
```

```text
locked twice
```

The unlock is what lets a waiting goroutine in. The worker blocks in `lock` until the main goroutine
unlocks, so `main` always prints first:

```ant
let m = Mutex.new()
let wg = WaitGroup.new()
m.lock()
wg.add(1)
go {
  defer { wg.done() }
  m.lock()
  defer { m.unlock() }
  println("worker")
}
println("main")
m.unlock()
wg.wait()
```

```text
main
worker
```

A mutex is often a field of the record it guards. A function on the record locks and defers the
unlock, so an early `return` releases the lock too:

```ant
type Stock = { lock: Mutex, items: Map[string, int] }

func Stock.take(s: Stock, item: string): bool {
  s.lock.lock()
  defer { s.lock.unlock() }
  let count = s.items.get(item) ?? 0
  if count == 0 { return false }
  s.items[item] = count - 1
  true
}

let stock = Stock{ lock: Mutex.new(), items: Map{ "apple": 1 } }
println(stock.take("apple"))
println(stock.take("apple"))
println(stock.take("pear"))
```

```text
true
false
false
```

## Applications

Hold the lock only while touching shared data. Copy what you need, unlock, and do the slow work
outside the lock:

```ant
type Log = { lock: Mutex, lines: Array[string] }

func Log.add(log: Log, line: string) {
  log.lock.lock()
  defer { log.lock.unlock() }
  log.lines.push(line)
}

func Log.snapshot(log: Log): Array[string] {
  log.lock.lock()
  defer { log.lock.unlock() }
  [...log.lines]
}

let log = Log{ lock: Mutex.new(), lines: [] }
log.add("start")
log.add("stop")
let lines = log.snapshot()
log.add("later")
println(lines | String.join("\n"))
```

```text
start
stop
```

## Properties and relations

`Mutex.withLock(m, f)` is a lock, a deferred unlock, and a call of `f`:

```ant
func withLockByHand[T](m: Mutex, f: func(): T): T {
  m.lock()
  defer { m.unlock() }
  f()
}

let m = Mutex.new()
let xs = [1, 2, 3]
assert(withLockByHand(m, { xs.length() }) == Mutex.withLock(m, { xs.length() }))
```

## Possible issues

Unlocking a mutex that is not locked panics, whether it was never locked or was unlocked already:

```ant
Mutex.unlock(Mutex.new()) // panic: Mutex.unlock: the Mutex is not locked
```

```ant
let m = Mutex.new()
m.lock()
m.unlock()
m.unlock() // panic: Mutex.unlock: the Mutex is not locked
```

A panic between `lock` and a plain `unlock` skips the unlock and leaves the mutex locked. The next
`lock` would then wait forever. Here `Panic.catch` stops the panic, and the unlock afterwards
succeeds only because the mutex is still locked:

```ant
let m = Mutex.new()
let r = Panic.catch({
  m.lock()
  panic("boom")
  m.unlock()
})
println(r.isOk())
m.unlock()
println("was still locked")
```

```text
false
was still locked
```

With `defer { m.unlock() }` right after the lock, or with `Mutex.withLock`, the panic unlocks the
mutex on its way out:

```ant
let m = Mutex.new()
let r = Panic.catch({
  m.lock()
  defer { m.unlock() }
  panic("boom")
})
println(r.isOk())
m.lock()
println("locked again")
m.unlock()
```

```text
false
locked again
```

## See also

[Mutex.lock](lock.md) · [Mutex.withLock](withLock.md) · [Mutex.new](new.md) · [Mutex](index.md) ·
[RwMutex.writeUnlock](../RwMutex/writeUnlock.md) · [Cond.wait](../Cond/wait.md) ·
[Panic.catch](../Panic/catch.md) · [defer](../../language/keywords/defer.md) ·
[Concurrency guide](../../guide/concurrency.md)
