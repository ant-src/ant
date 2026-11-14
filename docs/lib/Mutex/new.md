<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Mutex.new

```ant signature
func Mutex.new(): Mutex
```

Mutex.new returns a new, unlocked mutex.

<!-- end generated -->

## Details

- Each call makes a distinct mutex: mutexes compare by identity. Goroutines exclude one another only
  when they lock the same mutex.

## Basic examples

Make a mutex:

```ant
Mutex.new() // Mutex
```

Guard a shared counter, so that the workers' increments do not race:

```ant
let lock = Mutex.new()
var count = 0
Parallel.each(Array.filled(100, 1), func(n: int, ctx: Context) {
  lock.withLock({ count += n })
}, limit: 8)
println(count)
```

```text
100
```

## Scope

A mutex is a value like any other: keep it in a `let`, an array, a map, or a record field. It
renders as its type's name:

```ant
let m = Mutex.new()
[m, Mutex.new()]  // [Mutex, Mutex]
Map{ "users": m } // Map{ "users": Mutex }
```

A record field can default to `Mutex.new()`. The default runs afresh for each record, so each record
gets its own lock:

```ant
type Account = { lock: Mutex = Mutex.new(), balance: int = 0 }
let a = Account{}
let b = Account{ balance: 10 }
a.lock == b.lock // false
```

A spread gives the field, so its default does not run: the copy shares the original's lock. Write
`Account{ ...a, lock: Mutex.new() }` for a lock of its own:

```ant
type Account = { lock: Mutex = Mutex.new(), balance: int = 0 }
let a = Account{}
let c = Account{ ...a }
c.lock == a.lock                                  // true
Account{ ...a, lock: Mutex.new() }.lock == a.lock // false
```

A [Cond](../Cond/new.md) takes the mutex it waits under:

```ant
let lock = Mutex.new()
let ready = Cond.new(lock)
var done = false
go {
  lock.withLock({ done = true })
  ready.signal()
}
lock.lock()
while !done {
  ready.wait()
}
lock.unlock()
println(done)
```

```text
true
```

## Applications

Keep a lock next to the data it guards, and touch the data only through functions that take the
lock:

```ant
type Account = { lock: Mutex = Mutex.new(), balance: int = 0 }

func Account.deposit(a: Account, amount: int): void {
  a.lock.withLock({ a.balance += amount })
}

func Account.read(a: Account): int {
  a.lock.withLock({ a.balance })
}

let account = Account{}
Parallel.each([10, 20, 30, 40], func(amount: int, ctx: Context) {
  account.deposit(amount)
}, limit: 4)
println(account.read())
```

```text
100
```

Guard a cache that several goroutines fill. `withLock` returns the value of its function, so a
lookup and an insert happen as one step:

```ant
let cache: Map[string, int] = Map{}
let lock = Mutex.new()

func lengthOf(word: string): int {
  lock.withLock({
    let hit = cache.get(word)
    if hit != null { return hit }
    let n = word.length()
    cache.set(word, n)
    n
  })
}

let words = ["ant", "bee", "ant", "wasp", "bee"]
let lengths = Parallel.map(words, func(w: string, ctx: Context) { lengthOf(w) }, limit: 3)
println(lengths)
println(cache.size())
```

```text
[3, 3, 3, 4, 3]
3
```

## Properties and relations

Two calls make two mutexes. A mutex equals only itself:

```ant
let a = Mutex.new()
let b = Mutex.new()
assert(a != b)
assert(a == a)
```

A new mutex is unlocked, so `lock` returns at once, and `unlock` then returns it to the unlocked
state:

```ant
let m = Mutex.new()
m.lock()
m.unlock()
m.lock()
m.unlock()
assert(m.withLock({ 42 }) == 42)
```

## Possible issues

A new mutex is unlocked, so unlocking it panics:

```ant
Mutex.new().unlock() // panic: Mutex.unlock: the Mutex is not locked
```

A mutex made inside the code it should guard guards nothing: each call gets its own lock, and no two
calls ever wait for each other. This form is wrong on purpose:

```ant
var count = 0

func bump(): void {
  Mutex.new().withLock({ count += 1 }) // a fresh lock each call: excludes nothing
}
```

Make the mutex once, next to the data, and share it:

```ant
let lock = Mutex.new()
var count = 0

func bump(): void {
  lock.withLock({ count += 1 })
}

Parallel.each(Array.filled(100, 0), func(_: int, ctx: Context) { bump() }, limit: 8)
println(count)
```

```text
100
```

A mutex has no JSON form, so a record that holds one cannot be passed to `Json.stringify` whole:

```ant
type Account = { lock: Mutex = Mutex.new(), balance: int = 0 }
Json.stringify(Account{}) // panic: Json.stringify: Mutex has no JSON form
```

Encode the guarded fields instead, such as `{ balance: account.balance }`.

## See also

[Mutex.lock](lock.md) · [Mutex.unlock](unlock.md) · [Mutex.withLock](withLock.md) ·
[RwMutex.new](../RwMutex/new.md) · [Cond.new](../Cond/new.md) · [Atomic.new](../Atomic/new.md) ·
[WaitGroup.new](../WaitGroup/new.md)
