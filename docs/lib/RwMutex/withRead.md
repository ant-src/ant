<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.withRead

```ant signature
func RwMutex.withRead[T](m: RwMutex, f: func(): T): T
```

RwMutex.withRead read-locks m, runs f, releases the lock, and returns f's result. It releases the
lock even if f panics.

<!-- end generated -->

## Details

- Many goroutines can be inside `withRead` on the same lock at once. A writer waits until all of
  them have left.
- `f` runs while the read lock is held; anything it waits on keeps writers waiting too.

## Basic examples

Read a value under the read lock:

```ant
let m = RwMutex.new()
let prices = Map{ "tea": 3, "cake": 5 }
println(RwMutex.withRead(m, { prices["tea"] }))
println(RwMutex.withRead(m, { prices.size() }))
```

```text
3
2
```

The lock is free again once `withRead` returns, so a writer can take it:

```ant
let m = RwMutex.new()
let names = ["Ann", "Bob"]
let first = RwMutex.withRead(m, { names[0] })
m.withWrite({ names.push("Cid") })
println(first)
println(names)
```

```text
Ann
["Ann", "Bob", "Cid"]
```

## Scope

Call it through the lock with a dot:

```ant
let lock = RwMutex.new()
let seen = Map{ "a.txt": true }
println(lock.withRead({ seen.has("a.txt") }))
println(lock.withRead({ seen.has("b.txt") }))
```

```text
true
false
```

A function with no parameters can be passed by name:

```ant
let lock = RwMutex.new()
var version = 3

func current(): int {
  version
}

println(lock.withRead(current))
version = 4
println(lock.withRead(current))
```

```text
3
4
```

The result can be any type, such as a record that reads several fields together, so that no writer
changes one of them in between:

```ant
type Account = { owner: string, balance: int }
type Bank = { lock: RwMutex, account: Account }

func Bank.snapshot(b: Bank): Account {
  b.lock.withRead({ Account{ owner: b.account.owner, balance: b.account.balance } })
}

let bank = Bank{ lock: RwMutex.new(), account: Account{ owner: "Ann", balance: 100 } }
println(bank.snapshot())
```

```text
Account{ owner: "Ann", balance: 100 }
```

A `return` inside the lambda leaves the lambda, and its value becomes the result:

```ant
let lock = RwMutex.new()
let stock = Map{ "tea": 4 }

func count(item: string): int {
  lock.withRead({
    if item.isEmpty() { return 0 }
    stock[item] ?? -1
  })
}

println([count(""), count("tea"), count("cake")])
```

```text
[0, 4, -1]
```

## Applications

Guard a map that many goroutines read and few write. Reads go through `withRead`, writes through
[RwMutex.withWrite](withWrite.md):

```ant
type Store = { lock: RwMutex, data: Map[string, string] }

func Store.get(s: Store, key: string): string | null {
  s.lock.withRead({ s.data[key] })
}

func Store.set(s: Store, key: string, value: string) {
  s.lock.withWrite({ s.data[key] = value })
}

let store = Store{ lock: RwMutex.new(), data: Map{ "lang": "Ant" } }
store.set("version", "0.1")
let keys = ["lang", "version", "missing"]
let found = Parallel.map(keys, func(key: string, ctx: Context) { store.get(key) ?? "-" }, limit: 3)
println(found)
```

```text
["Ant", "0.1", "-"]
```

## Properties and relations

`withRead` returns what `f` returns, and leaves the lock free, even after a panic in `f`:

```ant
let m = RwMutex.new()
assert(m.withRead({ 6 * 7 }) == 42)

let r = Panic.catch({ m.withRead({ panic("bad read") }) })
assert(r is Err)

// The lock was released, so a write lock does not wait.
m.writeLock()
m.writeUnlock()
println("free")
```

```text
free
```

Readers share the lock: a second reader enters while the first still holds it:

```ant
let m = RwMutex.new()
let entered: Channel[string] = Channel.new()
let message = m.withRead({
  go { m.withRead({ entered.send("second reader is in") }) }
  entered.receive()
})
println(message)
```

```text
second reader is in
```

It is the same as taking the read lock with [RwMutex.readLock](readLock.md) and releasing it with
`defer` and [RwMutex.readUnlock](readUnlock.md):

```ant
let m = RwMutex.new()
let xs = [1, 2, 3]

func count(): int {
  m.readLock()
  defer { m.readUnlock() }
  xs.length()
}

assert(m.withRead({ xs.length() }) == count())
println(count())
```

```text
3
```

## Possible issues

A read lock does not upgrade to a write lock. A write lock taken inside `f` waits for every reader
to leave, its own goroutine included, so it waits forever:

```ant
let m = RwMutex.new()
let upgraded: Channel[bool] = Channel.new()
go {
  m.withRead({
    // Blocks for good: this goroutine still holds the read lock.
    m.writeLock()
    upgraded.send(true)
  })
}
println(upgraded.receive(ctx: Context.new(100ms)))
```

```text
null
```

Read what you need under the read lock, release it, and then take the write lock with
[RwMutex.withWrite](withWrite.md). Check again under the write lock, since another writer may have
run in between:

```ant
let m = RwMutex.new()
let cache = Map{ "a": 1 }

func ensure(key: string, value: int) {
  if m.withRead({ cache.has(key) }) { return }
  m.withWrite({
    if !cache.has(key) { cache[key] = value }
  })
}

ensure("a", 10)
ensure("b", 2)
println(cache)
```

```text
Map{ "a": 1, "b": 2 }
```

The read lock does not keep other readers out, so a change made inside `f` races with them. Here
increments can be lost, and the count can end below 100:

<!-- test: check-only the lost increments depend on goroutine order -->

```ant
let m = RwMutex.new()
var hits = 0
Parallel.each(Array.filled(100, 1), func(n: int, work: Context) { m.withRead({ hits += n }) }, limit: 8)
println(hits)
```

Change shared data under [RwMutex.withWrite](withWrite.md):

```ant
let m = RwMutex.new()
var hits = 0
Parallel.each(Array.filled(100, 1), func(n: int, work: Context) { m.withWrite({ hits += n }) }, limit: 8)
println(hits)
```

```text
100
```

## See also

[RwMutex.withWrite](withWrite.md) · [RwMutex.readLock](readLock.md) ·
[RwMutex.readUnlock](readUnlock.md) · [Mutex.withLock](../Mutex/withLock.md) ·
[Panic.catch](../Panic/catch.md)
