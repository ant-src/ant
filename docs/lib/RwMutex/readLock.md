<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.readLock

```ant signature
func RwMutex.readLock(m: RwMutex): void
```

RwMutex.readLock takes one read hold on m, blocking while a writer holds m or waits in [`RwMutex.writeLock`](writeLock.md). A goroutine that holds m for reading must not read-lock it again: a writer waiting between the two calls blocks both.

<!-- end generated -->

## Details

- Any number of goroutines can hold the read lock at once. `RwMutex.readLock` waits while a writer
  holds `m` or waits in [`writeLock`](writeLock.md), so a stream of readers cannot starve a writer.
- Do not read-lock `m` again in a goroutine that already holds it for reading: a writer that starts
  waiting between the two calls blocks both, forever.
- Each `readLock` needs one [`readUnlock`](readUnlock.md). Put `defer { m.readUnlock() }` right
  after it, or use [`RwMutex.withRead`](withRead.md).
- What a writer changed before its `writeUnlock` is visible to a reader once its `readLock` returns.
- A goroutine that holds the read lock and calls `writeLock` on the same `m` waits for itself
  forever. Release the read lock first.

## Basic examples

Lock for reading, release with `defer`, and read the shared value in between:

```ant
let m = RwMutex.new()
let scores = Map{ "ann": 31, "bob": 27 }

func score(name: string): int | null {
  RwMutex.readLock(m)
  defer { m.readUnlock() }
  scores.get(name)
}

println(score("ann"))
println(score("cid"))
```

```text
31
null
```

A reader waits while a writer holds the lock. Here the goroutine's `readLock` returns only after the
main goroutine's `writeUnlock`, so its entry always comes second:

```ant
let m = RwMutex.new()
let log: Array[string] = []
m.writeLock()
let wg = WaitGroup.new()
wg.add(1)
go {
  defer { wg.done() }
  RwMutex.readLock(m)
  defer { m.readUnlock() }
  log.push("reader")
}
log.push("writer")
m.writeUnlock()
wg.wait()
println(log)
```

```text
["writer", "reader"]
```

## Scope

Readers do not exclude each other. Each of these three goroutines holds the read lock and waits
until all three hold it, which works only because the read locks are held at the same time:

```ant
let m = RwMutex.new()
let holding = WaitGroup.new()
holding.add(3)
let wg = WaitGroup.new()
for _ in Array.range(0, 3) {
  wg.add(1)
  go {
    defer { wg.done() }
    m.readLock()
    defer { m.readUnlock() }
    holding.done()
    holding.wait()
  }
}
wg.wait()
println("3 readers held the lock at once")
```

```text
3 readers held the lock at once
```

The lock is often a field of the record whose other fields it guards. Read the field into a fresh
local under the lock and narrow the local:

```ant
type Config = { lock: RwMutex, name: string | null }

func greeting(c: Config): string {
  c.lock.readLock()
  defer { c.lock.readUnlock() }
  let name = c.name
  if name == null { return "Hello" }
  "Hello, ${name}"
}

let config = Config{ lock: RwMutex.new(), name: null }
println(greeting(config))
config.lock.withWrite({ config.name = "Ada" })
println(greeting(config))
```

```text
Hello
Hello, Ada
```

## Applications

A price table that many goroutines read and few change. Lookups take the read lock and run side by
side; an update takes the write lock:

```ant
type Prices = { lock: RwMutex, cents: Map[string, int] }

func price(p: Prices, name: string): int | null {
  p.lock.readLock()
  defer { p.lock.readUnlock() }
  p.cents.get(name)
}

func setPrice(p: Prices, name: string, cents: int) {
  p.lock.writeLock()
  defer { p.lock.writeUnlock() }
  p.cents.set(name, cents)
}

let prices = Prices{ lock: RwMutex.new(), cents: Map{ "tea": 250, "cake": 400 } }
let names = ["tea", "cake", "soup", "tea"]
let found = Parallel.map(names, func(name: string, work: Context) { price(prices, name) }, limit: 4)
println(found)
setPrice(prices, "soup", 550)
println(price(prices, "soup"))
```

```text
[250, 400, null, 250]
550
```

A cache checks under the read lock and fills under the write lock. It releases the read lock before
it takes the write lock, and checks again, since another goroutine may have filled the entry in
between:

```ant
type Cache = { lock: RwMutex, squares: Map[int, int] }

func square(c: Cache, n: int): int {
  c.lock.readLock()
  let cached = c.squares.get(n)
  c.lock.readUnlock()
  if cached != null { return cached }

  c.lock.writeLock()
  defer { c.lock.writeUnlock() }
  let again = c.squares.get(n)
  if again != null { return again }
  let value = n * n
  c.squares.set(n, value)
  value
}

let cache = Cache{ lock: RwMutex.new(), squares: Map{} }
let results = Parallel.map([3, 4, 3, 4, 3], func(n: int, work: Context) { square(cache, n) }, limit: 5)
println(results)
println(cache.squares.size())
```

```text
[9, 16, 9, 16, 9]
2
```

## Properties and relations

`RwMutex.withRead` does what `readLock`, the call, and `readUnlock` do. Once the read lock is
released, a writer gets in at once:

```ant
let m = RwMutex.new()
let count = 3
m.readLock()
let manual = count * 2
m.readUnlock()
assert(m.withRead({ count * 2 }) == manual)
m.writeLock()
m.writeUnlock()
```

## Possible issues

A read lock is released with `readUnlock`, not `writeUnlock`:

```ant
let m = RwMutex.new()
m.readLock()
m.writeUnlock() // panic: RwMutex.writeUnlock: the RwMutex is not write-locked
```

One `readUnlock` too many panics. Pair each `readLock` with exactly one `readUnlock`, which `defer`
right after the lock does:

```ant
let m = RwMutex.new()
m.readLock()
m.readUnlock()
m.readUnlock() // panic: RwMutex.readUnlock: the RwMutex is not read-locked
```

A read lock that is never released, such as one skipped by an early `return`, keeps every writer
waiting for good.

A read lock does not protect writes, because other readers hold it at the same time. Code that
changes the guarded data, such as `c.squares.set(n, value)`, takes `writeLock`, as the cache in
Applications does.

## See also

[RwMutex.readUnlock](readUnlock.md) · [RwMutex.withRead](withRead.md) ·
[RwMutex.writeLock](writeLock.md) · [RwMutex.withWrite](withWrite.md) · [RwMutex](index.md) ·
[Mutex.lock](../Mutex/lock.md) · [WaitGroup](../WaitGroup/index.md) ·
[Concurrency guide](../../guide/concurrency.md)
