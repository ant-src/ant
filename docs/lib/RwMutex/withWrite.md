<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# RwMutex.withWrite

```ant signature
func RwMutex.withWrite[T](m: RwMutex, f: func(): T): T
```

RwMutex.withWrite write-locks m, runs f, releases the lock, and returns f's result. It releases the
lock even if f panics, and panics if f returns with m not write-locked.

<!-- end generated -->

## Details

- Only one goroutine at a time runs a block under `withWrite`. It waits until no reader and no other
  writer holds the lock.
- A panic in `f` releases the lock and then goes on to the caller. What `f` wrote before the panic
  stays written. If `f` returns with the lock released, `withWrite` panics.
- Everything `f` writes happens before every later `withRead` or `withWrite` on the same lock, so a
  reader under that lock sees the whole change.
- The lock is not reentrant. Calling `withWrite`, `withRead`, `writeLock`, or `readLock` on the same
  `RwMutex` inside `f` blocks forever.

## Basic examples

Change shared data under the write lock:

```ant
let lock = RwMutex.new()
let stock = Map{ "apples": 3 }
RwMutex.withWrite(lock, { stock["apples"] = 5 })
println(stock)
```

```text
Map{ "apples": 5 }
```

`withWrite` returns the value of `f`:

```ant
let lock = RwMutex.new()
let stock = Map{ "apples": 3 }
let removed = RwMutex.withWrite(lock, { stock.remove("apples") })
println(removed)
println(stock)
```

```text
3
Map{}
```

## Scope

Call it with a dot on the lock:

```ant
let lock = RwMutex.new()
let log: Array[string] = []
lock.withWrite({ log.push("started") })
println(log)
```

```text
["started"]
```

Take a value out and reset the data in one step, so that no other goroutine sees the data between
the two:

```ant
type Queue = { lock: RwMutex, items: Array[string] }

func Queue.drain(q: Queue): Array[string] {
  q.lock.withWrite({
    let taken = q.items
    q.items = []
    taken
  })
}

let q = Queue{ lock: RwMutex.new(), items: ["a", "b"] }
println(q.drain())
println(q.items)
```

```text
["a", "b"]
[]
```

## Applications

Guard a record that many goroutines change at once. Writers take the write lock; a reader takes the
read lock with [RwMutex.withRead](withRead.md):

```ant
type Bank = { lock: RwMutex, balances: Map[string, int] }
type Transfer = { from: string, to: string, amount: int }

func Bank.transfer(b: Bank, t: Transfer) {
  b.lock.withWrite({
    b.balances[t.from] = (b.balances[t.from] ?? 0) - t.amount
    b.balances[t.to] = (b.balances[t.to] ?? 0) + t.amount
  })
}

func Bank.total(b: Bank): int {
  b.lock.withRead({ b.balances.values().sum() ?? 0 })
}

let bank = Bank{ lock: RwMutex.new(), balances: Map{ "ann": 100, "bob": 20 } }
let transfers = [
  Transfer{ from: "ann", to: "bob", amount: 30 },
  Transfer{ from: "bob", to: "ann", amount: 10 },
  Transfer{ from: "ann", to: "bob", amount: 5 },
]
Parallel.each(transfers, func(t: Transfer, work: Context) { bank.transfer(t) }, limit: 3)
println(bank.balances)
println(bank.total())
```

```text
Map{ "ann": 75, "bob": 45 }
120
```

A cache that many goroutines read. A hit needs only the read lock. A miss takes the write lock and
looks again, because another goroutine may have filled the entry between the two locks:

```ant
type Cache = { lock: RwMutex, lengths: Map[string, int] }

func Cache.length(c: Cache, word: string): int {
  let hit = c.lock.withRead({ c.lengths[word] })
  if hit != null { return hit }
  c.lock.withWrite({
    let stored = c.lengths[word]
    if stored != null { return stored }
    let n = word.length()
    c.lengths[word] = n
    n
  })
}

let cache = Cache{ lock: RwMutex.new(), lengths: Map{} }
let words = ["ant", "hill", "ant", "colony", "hill"]
let lengths = Parallel.map(words, func(w: string, work: Context) { cache.length(w) }, limit: 4)
println(lengths)
println(cache.lengths.size())
```

```text
[3, 4, 3, 6, 4]
3
```

The `return` inside the block leaves only the block; it works here because the `withWrite` call is
the function's last expression.

## Properties and relations

`withWrite` returns what `f` returns:

```ant
let lock = RwMutex.new()
assert(lock.withWrite({ 6 * 7 }) == 42)
assert(lock.withWrite({ null }) == null)
```

It does what [RwMutex.writeLock](writeLock.md) followed by a deferred
[RwMutex.writeUnlock](writeUnlock.md) does:

```ant
type Counter = { lock: RwMutex, n: int }

func Counter.bump(c: Counter) {
  c.lock.writeLock()
  defer { c.lock.writeUnlock() }
  c.n += 1
}

let a = Counter{ lock: RwMutex.new(), n: 0 }
let b = Counter{ lock: RwMutex.new(), n: 0 }
a.bump()
b.lock.withWrite({ b.n += 1 })
assert(a.n == b.n)
```

A panic inside `f` releases the lock, so the next `withWrite` does not block:

```ant
let lock = RwMutex.new()
let r = Panic.catch({ lock.withWrite({ panic("failed") }) })
assert(r is Err)
assert(lock.withWrite({ "free" }) == "free")
```

## Possible issues

`return` inside the block leaves the block, not the function around it. Here the guard's value is
dropped and `withdraw` goes on to return `"paid"`:

```ant
type Account = { lock: RwMutex, balance: int }

func withdraw(a: Account, amount: int): string {
  a.lock.withWrite({
    if a.balance < amount { return "declined" }
    a.balance -= amount
  })
  "paid"
}

println(withdraw(Account{ lock: RwMutex.new(), balance: 10 }, 50))
```

```text
paid
```

Make the block's value the function's value instead:

```ant
type Account = { lock: RwMutex, balance: int }

func withdraw(a: Account, amount: int): string {
  a.lock.withWrite({
    if a.balance < amount { return "declined" }
    a.balance -= amount
    "paid"
  })
}

println(withdraw(Account{ lock: RwMutex.new(), balance: 10 }, 50))
```

```text
declined
```

A panic releases the lock but does not undo what the block already wrote. The code that catches the
panic decides whether the data can still be used:

```ant
let lock = RwMutex.new()
let balances = Map{ "ann": 100, "bob": 20 }
let r = Panic.catch({
  lock.withWrite({
    balances["ann"] = 70
    panic("network down")
  })
})
println(r is Err)
println(balances)
```

```text
true
Map{ "ann": 70, "bob": 20 }
```

## See also

[RwMutex.withRead](withRead.md) · [RwMutex.writeLock](writeLock.md) ·
[RwMutex.writeUnlock](writeUnlock.md) · [RwMutex.new](new.md) · [RwMutex](index.md) ·
[Mutex.withLock](../Mutex/withLock.md) · [Panic.catch](../Panic/catch.md) ·
[Parallel.each](../Parallel/each.md) · [defer](../../language/keywords/defer.md)
