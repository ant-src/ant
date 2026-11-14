<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Mutex.withLock

```ant signature
func Mutex.withLock[T](m: Mutex, f: func(): T): T
```

Mutex.withLock locks m, runs f, unlocks m, and returns f's result. It unlocks m even if f panics, and panics if f returns with m unlocked.

<!-- end generated -->

## Details

- `Mutex.withLock` blocks until `m` is free, as `Mutex.lock` does, with no context to give up.
- A panic in `f` unlocks `m` and then goes on to the caller.
- `f` must leave `m` locked: if it returns with `m` unlocked, `withLock` panics.
- The lock is not reentrant: locking `m` again inside `f`, with `Mutex.withLock` or `Mutex.lock`,
  waits for the lock that `f` itself holds, forever.
- Everything `f` writes happens before the next goroutine that locks `m` gets the lock, so a reader
  under the same mutex sees the whole change.

## Basic examples

Guard a shared counter that many goroutines increment:

```ant
let m = Mutex.new()
var count = 0
Parallel.each(Array.range(0, 100), func(i: int, ctx: Context) {
  Mutex.withLock(m, { count += 1 })
}, limit: 8)
println(count)
```

```text
100
```

The result of the block is the result of the call:

```ant
let m = Mutex.new()
let stock = Map{ apples: 3, pears: 0 }
let apples = Mutex.withLock(m, { stock["apples"] ?? 0 })
println(apples)
```

```text
3
```

## Scope

Call it through `.` on the mutex, here a field of a record:

```ant
type Hits = { lock: Mutex, pages: Map[string, int] }

func Hits.add(h: Hits, page: string) {
  h.lock.withLock({ h.pages[page] = (h.pages[page] ?? 0) + 1 })
}

let hits = Hits{ lock: Mutex.new(), pages: Map{} }
for page in ["/", "/docs", "/"] {
  hits.add(page)
}
println(hits.pages)
```

```text
Map{ "/": 2, "/docs": 1 }
```

A function with no parameters can be passed by name:

```ant
let m = Mutex.new()
var next = 0

func takeId(): int {
  next += 1
  next
}

println(m.withLock(takeId))
println(m.withLock(takeId))
```

```text
1
2
```

A block that returns nothing gives `void`:

```ant
let m = Mutex.new()
let log: Array[string] = []
let r = m.withLock({ log.push("started") })
println(r)
println(log)
```

```text
void
["started"]
```

Read shared state into a fresh `let` under the lock, then work with that snapshot outside it:

```ant
let m = Mutex.new()
let sessions: Map[string, string] = Map{ abc: "ann" }

func userOf(token: string): string {
  let user = m.withLock({ sessions[token] })
  if user == null {
    return "guest"
  }
  user.toUpper()
}

println(userOf("abc"))
println(userOf("xyz"))
```

```text
ANN
guest
```

## Applications

Compute a value once per key and share it between goroutines. The lookup and the insert happen in
one critical section, so no two goroutines insert the same key:

```ant
let m = Mutex.new()
let lengths: Map[string, int] = Map{}
var computed = 0

func lengthOf(word: string): int {
  m.withLock({
    let known = lengths[word]
    if known != null {
      return known
    }
    computed += 1
    let n = word.length()
    lengths[word] = n
    n
  })
}

Parallel.each(["ant", "bee", "ant", "ant", "bee"], func(w: string, ctx: Context) {
  lengthOf(w)
}, limit: 5)
println(lengths["ant"])
println(computed)
```

```text
3
2
```

Move money between accounts so that no reader sees one side changed without the other:

```ant
let m = Mutex.new()
let balances = Map{ ann: 100, bob: 100 }

func transfer(from: string, to: string, amount: int) {
  m.withLock({
    balances[from] = (balances[from] ?? 0) - amount
    balances[to] = (balances[to] ?? 0) + amount
  })
}

func total(): int {
  m.withLock({ balances.values().sum() ?? 0 })
}

Parallel.each(Array.range(0, 50), func(i: int, ctx: Context) {
  if i % 2 == 0 {
    transfer("ann", "bob", 3)
  } else {
    transfer("bob", "ann", 2)
  }
  assert(total() == 200)
}, limit: 4)
println(balances)
```

```text
Map{ "ann": 75, "bob": 125 }
```

Collect results from several goroutines into one array:

```ant
let m = Mutex.new()
let squares: Array[int] = []
Parallel.each([1, 2, 3, 4], func(n: int, ctx: Context) {
  let square = n * n
  m.withLock({ squares.push(square) })
}, limit: 4)
println(squares.sort())
```

```text
[1, 4, 9, 16]
```

## Properties and relations

`Mutex.withLock(m, f)` does what `Mutex.lock`, a deferred `Mutex.unlock`, and a call of `f` do:

```ant
func byHand(m: Mutex, f: func(): int): int {
  m.lock()
  defer { m.unlock() }
  f()
}

let m = Mutex.new()
assert(Mutex.withLock(m, { 6 * 7 }) == byHand(m, { 6 * 7 }))
```

After a panic leaves the block, the mutex is free again:

```ant
let m = Mutex.new()
let r = Panic.catch({ m.withLock({ panic("boom") }) })
assert(r.mapErr({ x.message }) == Err{ error: "boom" })
m.lock()
m.unlock()
```

## Possible issues

`Mutex.withLock` releases the lock after a panic, but what the block wrote before it stays written.
The code that catches the panic decides whether the data can still be used:

```ant
let m = Mutex.new()
let balances = Map{ ann: 100, bob: 0 }
let r = Panic.catch({
  m.withLock({
    balances["ann"] = 50
    panic("transfer interrupted")
  })
})
println(r is Err)
println(balances)
```

```text
true
Map{ "ann": 50, "bob": 0 }
```

A `return` inside the block returns from the block, not from the function around it. Its value
becomes the result of `Mutex.withLock`, as `lengthOf` above uses.

An outer `var` does not narrow inside the block, since another goroutine may assign it at any
moment:

```ant
type User = { name: string }

let lock = Mutex.new()
var current: User | null = User{ name: "Ann" }
lock.withLock({
  if current != null { println(current.name) }
  // error: User | null has no field name: narrow the value first; current is a var of an enclosing body, which does not narrow here, so bind it to a let first
})
```

Read it into a `let` inside the block instead: `let user = current`, then test `user`.

A function that takes arguments does not fit. Wrap the call in a block:

```ant
let m = Mutex.new()
var total = 0

func add(n: int) {
  total += n
}

m.withLock(add) // error: cannot use func(int): void where func(): T is expected
```

`m.withLock({ add(5) })` compiles.

## See also

[Mutex](index.md) · [Mutex.new](new.md) · [Mutex.lock](lock.md) · [Mutex.unlock](unlock.md) ·
[RwMutex.withWrite](../RwMutex/withWrite.md) · [RwMutex.withRead](../RwMutex/withRead.md) ·
[Atomic](../Atomic/index.md) · [Parallel.each](../Parallel/each.md) ·
[Panic.catch](../Panic/catch.md) · [defer](../../language/keywords/defer.md)
