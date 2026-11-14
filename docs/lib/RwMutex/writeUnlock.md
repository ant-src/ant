<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.writeUnlock

```ant signature
func RwMutex.writeUnlock(m: RwMutex): void
```

RwMutex.writeUnlock releases the write lock on m. It panics if m is not write-locked.

<!-- end generated -->

## Details

- After the call, `m` is unlocked: every goroutine waiting in `readLock` may go on, or one waiting
  in `writeLock`.
- What the writer did while it held the lock is visible to every goroutine that locks `m` after the
  release, for reading or for writing.
- It panics when `m` is unlocked or only read-locked. A read lock is released with
  `RwMutex.readUnlock`.
- Call it in a `defer` right after `writeLock`, so that every return and every panic releases the
  lock, or use `RwMutex.withWrite`.

## Basic examples

Take the write lock, change the shared value, and release the lock:

```ant
let m = RwMutex.new()
var hits = 0
RwMutex.writeLock(m)
hits += 1
RwMutex.writeUnlock(m)
println(hits)
```

```text
1
```

In a function, defer the release right after the lock. The lock is released when the function
returns, whichever way it returns:

```ant
let lock = RwMutex.new()
let settings: Map[string, string] = Map{}

func set(key: string, value: string) {
  lock.writeLock()
  defer { lock.writeUnlock() }
  settings[key] = value
}

set("theme", "dark")
set("lang", "en")
println(settings)
```

```text
Map{ "theme": "dark", "lang": "en" }
```

## Scope

Once released, the lock can be taken again, by a writer or by any number of readers:

```ant
let m = RwMutex.new()
m.writeLock()
m.writeUnlock()
m.readLock()
m.readLock()
m.readUnlock()
m.readUnlock()
m.writeLock()
m.writeUnlock()
println("free")
```

```text
free
```

Each reader that takes the read lock after the release sees what the writer wrote before it let go.
A reader that arrives while the writer holds `m` waits in `readLock` until the release. Here the
main goroutine changes `version` while it holds the write lock, so all three readers see version 2:

```ant
let m = RwMutex.new()
var version = 1
let sawNew = Atomic.new(0)
let wg = WaitGroup.new()

m.writeLock()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.readLock()
    defer { m.readUnlock() }
    if version == 2 { sawNew.add(1) }
  }
}
version = 2
m.writeUnlock()
wg.wait()
println("${sawNew.load()} readers saw version 2")
```

```text
3 readers saw version 2
```

## Applications

A cache that many goroutines share keeps its lock in a field. Writers release it with a deferred
`writeUnlock`, so the early return in `put` releases it too:

```ant
type Cache = { lock: RwMutex, entries: Map[string, string] }

func Cache.put(c: Cache, key: string, value: string): bool {
  c.lock.writeLock()
  defer { c.lock.writeUnlock() }
  if c.entries.has(key) { return false }
  c.entries[key] = value
  true
}

func Cache.get(c: Cache, key: string): string | null {
  c.lock.readLock()
  defer { c.lock.readUnlock() }
  c.entries.get(key)
}

let cache = Cache{ lock: RwMutex.new(), entries: Map{} }
let added = Atomic.new(0)
let wg = WaitGroup.new()
for n in Array.range(0, 50) {
  wg.add(1)
  go {
    defer { wg.done() }
    if cache.put("k${n % 10}", "v${n}") { added.add(1) }
  }
}
wg.wait()
println(added.load())
println(cache.get("k7") != null)
```

```text
10
true
```

## Properties and relations

`RwMutex.withWrite(m, f)` is `writeLock`, then `f()` with a deferred `writeUnlock`. Both leave `m`
unlocked, so it can be locked again any number of times:

```ant
let m = RwMutex.new()
var count = 0

func bump(): int {
  m.writeLock()
  defer { m.writeUnlock() }
  count += 1
  count
}

assert(bump() == 1)
assert(m.withWrite({
  count += 1
  count
}) == 2)
for _ in Array.range(0, 3) {
  bump()
}
assert(count == 5)
```

## Possible issues

Releasing a lock that nobody holds panics, whether it was never taken or released already:

```ant
let m = RwMutex.new()
m.writeLock()
m.writeUnlock()
m.writeUnlock() // panic: RwMutex.writeUnlock: the RwMutex is not write-locked
```

A read lock is not a write lock. Release it with `RwMutex.readUnlock`:

```ant
let m = RwMutex.new()
m.readLock()
m.writeUnlock() // panic: RwMutex.writeUnlock: the RwMutex is not write-locked
```

A panic between `writeLock` and a plain `writeUnlock` skips the release, and every later lock of `m`
then waits forever. A panic ends its goroutine only after its deferred blocks run, so a deferred
`writeUnlock` still releases the lock:

```ant
let m = RwMutex.new()
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  m.writeLock()
  defer { m.writeUnlock() }
  panic("bad config")
}
wg.wait()
m.writeLock()
println("the lock is free")
m.writeUnlock()
```

```text
the lock is free
```

## See also

[RwMutex.writeLock](writeLock.md) · [RwMutex.withWrite](withWrite.md) ·
[RwMutex.readUnlock](readUnlock.md) · [RwMutex.readLock](readLock.md) ·
[RwMutex.withRead](withRead.md) · [Mutex.unlock](../Mutex/unlock.md) ·
[defer](../../language/keywords/defer.md) · [Concurrency guide](../../guide/concurrency.md)
