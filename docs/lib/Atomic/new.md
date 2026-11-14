<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.new

```ant signature
func Atomic.new[T](v: T): Atomic[T]
```

Atomic.new returns a new atomic cell holding v.

<!-- end generated -->

## Details

- Each call makes a distinct cell. Atomics compare by identity and print as `Atomic`.
- The cell makes its loads and stores atomic, not changes inside the value it holds. An array in a
  cell is the same array that `load` returns, and pushing to it from several goroutines needs a
  lock.

## Basic examples

Make a cell holding 0, and add to it from many goroutines at once. No increment is lost:

```ant
let count = Atomic.new(0)
Parallel.each(Array.range(0, 100), func(i: int, ctx: Context) {
  count.add(1)
}, limit: 8)
println(count.load())
```

```text
100
```

The new cell holds the value given:

```ant
let greeting = Atomic.new("hello")
greeting.load() // "hello"
```

## Scope

The cell can hold a value of any type. A union type, annotated, lets the cell start as `null`:

```ant
let latest: Atomic[string | null] = Atomic.new(null)
latest.load() // null
latest.store("ready")
latest.load() // "ready"
```

A floating-point cell adds floating-point values:

```ant
let total = Atomic.new(1.5)
total.add(2.25) // 3.75
```

A field default makes a new cell for each record:

```ant
type Stats = { hits: Atomic[int] = Atomic.new(0), misses: Atomic[int] = Atomic.new(0) }

let a = Stats{}
let b = Stats{}
a.hits.add(1)
println("${a.hits.load()} ${b.hits.load()}")
```

```text
1 0
```

## Applications

### Hand a result back from a goroutine

A goroutine stores its result in a cell, and the parent loads it after the wait into a fresh `let`,
which narrows:

```ant
func fetch(id: int): Result[string, Error] {
  if id == 1 { Result.ok("Ann") } else { Result.fail("no user ${id}") }
}

let wg = WaitGroup.new()
let latest: Atomic[Result[string, Error] | null] = Atomic.new(null)
wg.add(1)
go {
  defer { wg.done() }
  latest.store(fetch(1))
}
wg.wait()
let r = latest.load()
if r != null && r is Ok { println(r.value) }
```

```text
Ann
```

### Let the first goroutine win

`compareAndSwap` from `null` succeeds once, so exactly one goroutine claims the cell:

```ant
let winner: Atomic[string | null] = Atomic.new(null)
let claims = Atomic.new(0)
Parallel.each(["ann", "bob", "cy"], func(name: string, ctx: Context) {
  if winner.compareAndSwap(null, name) { claims.add(1) }
}, limit: 3)
println(claims.load())
```

```text
1
```

## Properties and relations

A new cell loads as the value it was made with:

```ant
let v = [1, 2, 3]
let a = Atomic.new(v)
assert(a.load() == v)
```

Two calls make two cells, even with equal values:

```ant
let a = Atomic.new(0)
let b = Atomic.new(0)
assert(a != b)
assert(a == a)
```

## Possible issues

Without an annotation, `Atomic.new(null)` makes an `Atomic[null]`, which holds nothing else:

```ant
let latest = Atomic.new(null)
latest.store("ready") // error: cannot use string where null is expected
```

Annotate the cell with the union it holds, as in [Scope](#scope).

Likewise `Atomic.new(0)` makes an `Atomic[int]`, which takes no floating-point values:

```ant
let total = Atomic.new(0)
total.add(1.5) // error: cannot use float64 where int is expected
```

Start it with `0.0` instead.

The cell holds the array itself, not a copy, so a change through `load` changes what the cell holds,
outside any atomic operation:

```ant
let xs = Atomic.new([1, 2])
xs.load().push(3)
println(xs.load())
```

```text
[1, 2, 3]
```

Guard such changes from several goroutines with a [Mutex](../Mutex/index.md), or build a new array
and swap it in with [Atomic.compareAndSwap](compareAndSwap.md), retrying when it returns false.

## See also

[Atomic.load](load.md) · [Atomic.store](store.md) · [Atomic.add](add.md) ·
[Atomic.compareAndSwap](compareAndSwap.md) · [Mutex](../Mutex/index.md) ·
[WaitGroup](../WaitGroup/index.md) · [Parallel.each](../Parallel/each.md)
