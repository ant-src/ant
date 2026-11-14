<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.compareAndSwap

```ant signature
func Atomic.compareAndSwap[T](a: Atomic[T], expected: T, v: T): bool
```

Atomic.compareAndSwap sets the value in a to v if it equals expected, and reports whether it did. It
compares with ==.

<!-- end generated -->

## Details

- The comparison and the swap are one atomic step: no other operation on the cell happens between
  them.
- It compares with `==`, so records, arrays, and maps match by their contents, not by identity.
  `Mutex`, `Atomic`, `Channel`, and the other synchronization values match only themselves.
- When it returns `false`, the cell did not hold `expected` and stays unchanged. In a retry loop,
  load the current value and try again.

## Basic examples

Swap only when the cell holds the expected value:

```ant
let a = Atomic.new(1)
Atomic.compareAndSwap(a, 1, 2) // true
Atomic.compareAndSwap(a, 1, 3) // false
a.load()                       // 2
```

Let exactly one of many goroutines claim a job:

```ant
let claimed = Atomic.new(false)
let winners = Atomic.new(0)
Parallel.each(Array.range(0, 8), func(i: int, ctx: Context) {
  if Atomic.compareAndSwap(claimed, false, true) {
    winners.add(1)
  }
}, limit: 8)
println(winners.load())
```

```text
1
```

## Scope

Call it through `.` on the cell:

```ant
let state = Atomic.new("idle")
state.compareAndSwap("idle", "running") // true
state.compareAndSwap("idle", "running") // false
state.load()                            // "running"
```

Records compare by their fields, so an equal record built anew matches:

```ant
type Point = { x: int, y: int }
let a = Atomic.new(Point{ x: 1, y: 2 })
a.compareAndSwap(Point{ x: 1, y: 2 }, Point{ x: 3, y: 4 }) // true
a.load()                                                   // Point{ x: 3, y: 4 }
```

Arrays compare by length and elements:

```ant
let a = Atomic.new([1, 2])
a.compareAndSwap([1, 2], [3]) // true
a.compareAndSwap([1, 2], [4]) // false
a.load()                      // [3]
```

Maps compare by their entries:

```ant
let a = Atomic.new(Map{ "a": 1 })
a.compareAndSwap(Map{ "a": 1 }, Map{ "b": 2 }) // true
a.load()                                       // Map{ "b": 2 }
```

A nullable cell can start empty and be filled once. Annotate the type, since `Atomic.new(null)`
alone holds only `null`:

```ant
let owner: Atomic[string | null] = Atomic.new(null)
owner.compareAndSwap(null, "ann") // true
owner.compareAndSwap(null, "bob") // false
owner.load()                      // "ann"
```

A synchronization value, such as a `Mutex`, matches only itself:

```ant
let m = Mutex.new()
let a = Atomic.new(m)
a.compareAndSwap(Mutex.new(), m) // false
a.compareAndSwap(m, Mutex.new()) // true
```

Floats compare as IEEE 754 numbers:

```ant
let z = Atomic.new(-0.0)
z.compareAndSwap(0.0, 1.5) // true
z.load()                   // 1.5
```

## Applications

Keep the largest value seen by many goroutines. Load, decide, and retry when another goroutine got
in first:

```ant
let best = Atomic.new(0)
Parallel.each([4, 17, 9, 12, 3], func(n: int, ctx: Context) {
  while true {
    let current = best.load()
    if n <= current || best.compareAndSwap(current, n) {
      break
    }
  }
}, limit: 5)
println(best.load())
```

```text
17
```

Move a job through the states of an enumeration, where each step is allowed from one state only.
Empty records compare by tag, so `Queued{}` matches the `Queued{}` in the cell:

```ant
type Queued = {}
type Running = {}
type Done = {}
type State = Queued | Running | Done

let job: Atomic[State] = Atomic.new(Queued{})

func start(): bool {
  job.compareAndSwap(Queued{}, Running{})
}

func finish(): bool {
  job.compareAndSwap(Running{}, Done{})
}

println(finish())
println(start())
println(start())
println(finish())
println(job.load())
```

```text
false
true
false
true
Done{}
```

Update a record by replacing it whole. Build a new record from the loaded one and swap it in, so no
goroutine changes the shared record in place:

```ant
type Stats = { count: int, total: int }

let stats = Atomic.new(Stats{ count: 0, total: 0 })
Parallel.each([5, 10, 15, 20], func(n: int, ctx: Context) {
  while true {
    let old = stats.load()
    let next = Stats{ count: old.count + 1, total: old.total + n }
    if stats.compareAndSwap(old, next) {
      break
    }
  }
}, limit: 4)
println(stats.load())
```

```text
Stats{ count: 4, total: 50 }
```

## Properties and relations

It returns `true` exactly when the cell held `expected`, and then the cell holds `v`:

```ant
let a = Atomic.new(10)
let before = a.load()
let swapped = a.compareAndSwap(10, 20)
assert(swapped == (before == 10))
assert(a.load() == 20)
```

A failed swap leaves the cell as it was:

```ant
let a = Atomic.new("x")
assert(!a.compareAndSwap("y", "z"))
assert(a.load() == "x")
```

For a number, `Atomic.add` has the same effect as a compare-and-swap loop that adds:

```ant
func addByHand(a: Atomic[int], d: int): int {
  while true {
    let old = a.load()
    if a.compareAndSwap(old, old + d) {
      return old + d
    }
  }
}

let a = Atomic.new(1)
let b = Atomic.new(1)
assert(addByHand(a, 5) == Atomic.add(b, 5))
assert(a.load() == b.load())
```

## Possible issues

A single compare-and-swap is not a retry. When another goroutine changes the cell between the load
and the swap, the swap returns `false` and the update is lost unless you check the result. Loop
until it succeeds, as the applications above do, or use `Atomic.add` for a number.

A cell holding `NaN` never swaps, since `NaN == NaN` is `false`:

```ant
let a = Atomic.new(Math.NaN)
a.compareAndSwap(Math.NaN, 1.0) // false
```

Comparing by contents means a different value with equal contents matches. A goroutine that loaded
`Point{ x: 1, y: 2 }`, while others replaced it and then put back an equal record, still swaps.
Compare-and-swap checks the value, not whether anything changed in between.

The cell holds the record, not a copy. A field set on a loaded record changes the record in the
cell, with no atomic step and no ordering for other goroutines:

```ant
type Point = { x: int, y: int }
let a = Atomic.new(Point{ x: 1, y: 2 })
let p = a.load()
p.x = 9
a.load() // Point{ x: 9, y: 2 }
```

Build a new record and swap it in instead, as the `Stats` example does.

`expected` and `v` must have the cell's type:

```ant
let a = Atomic.new(1)
a.compareAndSwap(1, 2.5) // error: cannot use float64 where int is expected
```

## See also

[Atomic](index.md) · [Atomic.new](new.md) · [Atomic.load](load.md) · [Atomic.store](store.md) ·
[Atomic.add](add.md) · [Mutex.withLock](../Mutex/withLock.md) · [Once.do](../Once/do.md) ·
[Parallel.each](../Parallel/each.md) · [==](../../language/operators/equality.md)
