<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.writeLock

```ant signature
func RwMutex.writeLock(m: RwMutex): void
```

RwMutex.writeLock locks m for writing, blocking while any reader or writer holds it.

<!-- end generated -->

## Details

- A `readLock` on `m` waits while the write lock is held. Every write made before `writeUnlock` is
  visible to any goroutine that takes a read or write lock on `m` afterward.
- The lock is not reentrant. A goroutine that calls `writeLock` on a mutex it already holds, for
  reading or for writing, waits forever.
- Release it with [RwMutex.writeUnlock](writeUnlock.md), in a `defer` right after the lock, or let
  [RwMutex.withWrite](withWrite.md) do both.

## Basic examples

Take the write lock before changing shared state, and release it with `defer` right after:

```ant
type Scores = { lock: RwMutex, byName: Map[string, int] }

func Scores.add(s: Scores, name: string, points: int) {
  RwMutex.writeLock(s.lock)
  defer { s.lock.writeUnlock() }
  s.byName[name] = (s.byName.get(name) ?? 0) + points
}

let scores = Scores{ lock: RwMutex.new(), byName: Map{} }
let wg = WaitGroup.new()
for _ in Array.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    scores.add("ada", 1)
  }
}
wg.wait()
println(scores.byName)
```

```text
Map{ "ada": 100 }
```

Readers take `readLock` and writers `writeLock` on the same mutex. The readers share the lock with
each other, and the writer has it alone:

```ant
type Config = { lock: RwMutex, values: Map[string, string] }

func Config.lookup(c: Config, key: string): string | null {
  c.lock.readLock()
  defer { c.lock.readUnlock() }
  c.values.get(key)
}

func Config.put(c: Config, key: string, value: string) {
  RwMutex.writeLock(c.lock)
  defer { c.lock.writeUnlock() }
  c.values[key] = value
}

let config = Config{ lock: RwMutex.new(), values: Map{ mode: "dev" } }
config.put("mode", "prod")
println(config.lookup("mode"))
```

```text
prod
```

## Scope

`writeLock` waits while a reader holds the mutex. Here the writer cannot print until the main
goroutine releases its read lock:

```ant
let m = RwMutex.new()
let wg = WaitGroup.new()
m.readLock()
wg.add(1)
go {
  defer { wg.done() }
  m.writeLock()
  defer { m.writeUnlock() }
  println("writer")
}
Time.sleep(20ms)
println("reader done")
m.readUnlock()
wg.wait()
```

```text
reader done
writer
```

It also waits while another writer holds it, so two writers never overlap:

```ant
let m = RwMutex.new()
let wg = WaitGroup.new()
m.writeLock()
wg.add(1)
go {
  defer { wg.done() }
  m.writeLock()
  defer { m.writeUnlock() }
  println("second writer")
}
Time.sleep(20ms)
println("first writer done")
m.writeUnlock()
wg.wait()
```

```text
first writer done
second writer
```

A write lock held by one goroutine makes readers wait too. Once it is released, a reader sees every
write made under it:

```ant
type Counter = { lock: RwMutex, n: int }

let counter = Counter{ lock: RwMutex.new(), n: 0 }
let wg = WaitGroup.new()
counter.lock.writeLock()
wg.add(1)
go {
  defer { wg.done() }
  counter.lock.readLock()
  defer { counter.lock.readUnlock() }
  println("reader sees ${counter.n}")
}
Time.sleep(20ms)
counter.n = 42
counter.lock.writeUnlock()
wg.wait()
```

```text
reader sees 42
```

## Applications

A cache that is read far more often than it is filled. Lookups share the read lock. A miss takes the
write lock and checks again, since another goroutine may have filled the entry between the two
locks:

```ant
type Cache = { lock: RwMutex, entries: Map[string, int] }

func Cache.get(c: Cache, key: string, compute: func(string): int): int {
  let hit = c.lock.withRead({ c.entries.get(key) })
  if hit != null { return hit }

  c.lock.writeLock()
  defer { c.lock.writeUnlock() }
  let filled = c.entries.get(key)
  if filled != null { return filled }
  let value = compute(key)
  c.entries[key] = value
  value
}

let cache = Cache{ lock: RwMutex.new(), entries: Map{} }
let computed = Atomic.new(0)
let keys = ["ant", "bee", "ant", "cat", "bee", "ant"]
let lengths = Parallel.map(keys, func(key: string, work: Context): int {
  cache.get(key, func(k: string): int {
    computed.add(1)
    k.length()
  })
}, limit: 4)
println(lengths)
println("${computed.load()} computed")
```

```text
[3, 3, 3, 3, 3, 3]
3 computed
```

## Properties and relations

`writeLock` with a deferred `writeUnlock` around a block is what [RwMutex.withWrite](withWrite.md)
does; both leave the mutex free for the next writer:

```ant
let m = RwMutex.new()
let log: Array[string] = []

func appendLocked(m: RwMutex, log: Array[string], entry: string) {
  m.writeLock()
  defer { m.writeUnlock() }
  log.push(entry)
}

appendLocked(m, log, "a")
m.withWrite({ log.push("b") })
appendLocked(m, log, "c")
assert(log == ["a", "b", "c"])
```

## Possible issues

The lock is not reentrant. A second `writeLock` on the same goroutine waits for itself and never
returns. Here the goroutine is still blocked when the main goroutine stops waiting for it:

```ant
let m = RwMutex.new()
let acquired: Channel[bool] = Channel.new(1)
go {
  m.writeLock()
  m.writeLock()
  acquired.send(true)
}
println(acquired.receive(Context.new(100ms)) ?? "still waiting")
```

```text
still waiting
```

A read lock does not upgrade to a write lock. `writeLock` waits for every reader, the caller
included, so release the read lock first, and check the state again after taking the write lock, as
the cache above does:

```ant
let m = RwMutex.new()
let acquired: Channel[bool] = Channel.new(1)
go {
  m.readLock()
  m.writeLock()
  acquired.send(true)
}
println(acquired.receive(Context.new(100ms)) ?? "still waiting")
```

```text
still waiting
```

A read lock is released with `readUnlock`. Releasing it with `writeUnlock` panics:

```ant
let m = RwMutex.new()
m.readLock()
m.writeUnlock() // panic: RwMutex.writeUnlock: writeUnlock of an RwMutex that is not write-locked
```

## See also

[RwMutex.writeUnlock](writeUnlock.md) · [RwMutex.withWrite](withWrite.md) ·
[RwMutex.readLock](readLock.md) · [RwMutex.new](new.md) · [RwMutex](index.md) ·
[Mutex.lock](../Mutex/lock.md) · [Concurrency guide](../../guide/concurrency.md)
