<!-- status: reviewed -->

# Protect shared state

Let several goroutines update the same map, record, or counter without losing a write.

1. Put the data and its lock in one record, and create the lock with
   [Mutex.new](../../lib/Mutex/new.md).
2. Read and write the data only inside [Mutex.withLock](../../lib/Mutex/withLock.md). It unlocks the
   mutex when the block ends, even when the block panics.
3. Read and write the data only through functions on the record type, so no caller touches it
   without the lock.

## Guard a map with a mutex

This script counts page hits from many goroutines at once:

```ant
type Stats = { lock: Mutex, hits: Map[string, int] }

func Stats.record(s: Stats, page: string) {
  s.lock.withLock({ s.hits[page] = (s.hits[page] ?? 0) + 1 })
}

func Stats.count(s: Stats, page: string): int {
  s.lock.withLock({ s.hits[page] ?? 0 })
}

let stats = Stats{ lock: Mutex.new(), hits: Map{} }
let requests = ["/", "/docs", "/", "/blog", "/", "/docs"]
Parallel.each(requests, func(page: string, ctx: Context) { stats.record(page) }, limit: 4)
println(stats.count("/"))
println(stats.count("/docs"))
println(stats.count("/about"))
```

```text
3
2
0
```

[Parallel.each](../../lib/Parallel/each.md) runs `record` on up to four goroutines. The read and the
write of each count happen under one lock, so no increment is lost. The block's value is the result
of `withLock`, which `count` returns.

A mutex is not reentrant: a goroutine that locks a mutex it already holds waits forever. So `record`
must not call `count` inside its `withLock` block. Keep work that both need in a helper that expects
the lock to be held already, and take the lock only in the functions callers use.

## Read a snapshot

Copy what you need under the lock, and work on the copy outside it.
[Map.entries](../../lib/Map/entries.md) returns a new array, so sorting it and formatting it with
[Array.map](../../lib/Array/map.md) hold no lock:

```ant
type Stats = { lock: Mutex, hits: Map[string, int] }

func Stats.top(s: Stats): Array[string] {
  let entries = s.lock.withLock({ s.hits.entries() })
  entries
    | Array.sortBy({ -x.value })
    | Array.map({ "${x.key} ${x.value}" })
}

let stats = Stats{ lock: Mutex.new(), hits: Map{ "/": 3, "/docs": 2, "/blog": 1 } }
for line in stats.top() {
  println(line)
}
```

```text
/ 3
/docs 2
/blog 1
```

## Hold the lock across several steps

A `return` inside a `withLock` block leaves only the block, not the function around it. When the
whole function is the critical section, and a `return` must leave the function itself, call
[Mutex.lock](../../lib/Mutex/lock.md) and put [Mutex.unlock](../../lib/Mutex/unlock.md) in a
[`defer`](../../language/keywords/defer.md) right after it. Every `return` then releases the lock.
Keep the check and the update under the same lock; otherwise two goroutines can both pass the check.
Here `withdraw` returns [Result.fail](../../lib/Result/fail.md) early or
[Result.ok](../../lib/Result/ok.md) at the end:

```ant
type Account = { lock: Mutex, balance: int }

func Account.withdraw(a: Account, amount: int): Result[int, Error] {
  a.lock.lock()
  defer { a.lock.unlock() }
  if amount > a.balance {
    return Result.fail("insufficient funds: balance ${a.balance}")
  }
  a.balance -= amount
  Result.ok(a.balance)
}

let account = Account{ lock: Mutex.new(), balance: 100 }
let paid = Atomic.new(0)
Parallel.each([30, 30, 30, 30], func(amount: int, ctx: Context) {
  if account.withdraw(amount) is Ok { paid.add(1) }
}, limit: 4)
println("${paid.load()} paid, ${account.balance} left")
println(account.withdraw(50))
```

```text
3 paid, 10 left
Err{ error: Error{ message: "insufficient funds: balance 10" } }
```

Reading `account.balance` after `Parallel.each` needs no lock: the call returns only after every
callback has ended.

## Let readers share the lock

When reads far outnumber writes, guard the data with an [RwMutex](../../lib/RwMutex/index.md).
[RwMutex.withRead](../../lib/RwMutex/withRead.md) lets many readers in at once, and
[RwMutex.withWrite](../../lib/RwMutex/withWrite.md) waits until they have all left:

```ant
type Prices = { lock: RwMutex, table: Map[string, int] }

func Prices.get(p: Prices, item: string): int | null {
  p.lock.withRead({ p.table[item] })
}

func Prices.set(p: Prices, item: string, price: int) {
  p.lock.withWrite({ p.table[item] = price })
}

let prices = Prices{ lock: RwMutex.new(), table: Map{ tea: 3, cake: 5 } }
let total = Atomic.new(0)
let order = ["tea", "cake", "tea", "scone"]
Parallel.each(order, func(item: string, ctx: Context) {
  total.add(prices.get(item) ?? 0)
}, limit: 4)
println(total.load())
prices.set("scone", 4)
println(prices.get("scone"))
```

```text
11
4
```

Never write inside `withRead`: other readers hold the lock at the same time. Nor call a function
that takes the read lock again, such as `get` from inside another `withRead` block: a writer waiting
between the two calls blocks both.

## Count with an atomic

For a single number or flag, use an [Atomic](../../lib/Atomic/index.md) instead of a lock.
[Atomic.add](../../lib/Atomic/add.md) updates the value in one step and returns the new value;
[Atomic.load](../../lib/Atomic/load.md) reads it:

```ant
let processed = Atomic.new(0)
let failed = Atomic.new(0)
Parallel.each(Array.range(1, 21), func(n: int, ctx: Context) {
  if n % 7 == 0 { failed.add(1) } else { processed.add(1) }
}, limit: 4)
println("${processed.load()} processed, ${failed.load()} failed")
```

```text
18 processed, 2 failed
```

A `load` followed by a [store](../../lib/Atomic/store.md) is two steps, and another goroutine can
write between them. For any update other than adding, retry with
[Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md), which stores only when the value is
still the one you loaded:

```ant
func raiseTo(best: Atomic[int], n: int) {
  var current = best.load()
  while n > current && !best.compareAndSwap(current, n) {
    current = best.load()
  }
}

let longest = Atomic.new(0)
Parallel.each(["ant", "beetle", "bee", "mantis"], func(word: string, ctx: Context) {
  raiseTo(longest, word.length())
}, limit: 4)
println(longest.load())
```

```text
6
```

Use a mutex as soon as two values must change together.

## See also

[Mutex](../../lib/Mutex/index.md) · [Mutex.new](../../lib/Mutex/new.md) ·
[Mutex.withLock](../../lib/Mutex/withLock.md) · [Mutex.lock](../../lib/Mutex/lock.md) ·
[Mutex.unlock](../../lib/Mutex/unlock.md) · [RwMutex](../../lib/RwMutex/index.md) ·
[RwMutex.new](../../lib/RwMutex/new.md) · [RwMutex.withRead](../../lib/RwMutex/withRead.md) ·
[RwMutex.withWrite](../../lib/RwMutex/withWrite.md) · [Atomic](../../lib/Atomic/index.md) ·
[Atomic.new](../../lib/Atomic/new.md) · [Atomic.add](../../lib/Atomic/add.md) ·
[Atomic.load](../../lib/Atomic/load.md) · [Atomic.store](../../lib/Atomic/store.md) ·
[Atomic.compareAndSwap](../../lib/Atomic/compareAndSwap.md) ·
[Parallel.each](../../lib/Parallel/each.md) · [Array.range](../../lib/Array/range.md) ·
[Array.sortBy](../../lib/Array/sortBy.md) · [Array.map](../../lib/Array/map.md) ·
[Map.entries](../../lib/Map/entries.md) · [Result.fail](../../lib/Result/fail.md) ·
[Result.ok](../../lib/Result/ok.md) · [String.length](../../lib/String/length.md) ·
[Concurrency](../../guide/concurrency.md) · [Run a worker pool](worker-pool.md)
