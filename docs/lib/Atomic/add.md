<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Atomic.add

```ant signature
func Atomic.add[T: Number](a: Atomic[T], d: T): T
```

Atomic.add adds d to the value in a and returns the new value. It panics on overflow.

<!-- end generated -->

## Details

- The read, the addition, and the write are one atomic operation, so adds from many goroutines at
  once never lose an update.
- On an integer cell, `add` panics when the new value does not fit the type. A float cell does not
  panic: its value goes to infinity, as `+` does.

## Basic examples

Add to a counter and get the new count:

```ant
let hits = Atomic.new(0)
Atomic.add(hits, 1) // 1
Atomic.add(hits, 1) // 2
Atomic.add(hits, 5) // 7
hits.load()         // 7
```

Count across goroutines. Each `add` happens as one step, so no increment is lost:

```ant
let hits = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 100) {
  wg.add(1)
  go {
    defer { wg.done() }
    hits.add(1)
  }
}
wg.wait()
println(hits.load())
```

```text
100
```

## Scope

A negative `d` subtracts:

```ant
let stock = Atomic.new(10)
stock.add(-3) // 7
stock.add(-7) // 0
```

A `d` of zero leaves the value as it is and returns it:

```ant
let n = Atomic.new(42)
n.add(0) // 42
```

Float cells add as `+` does on floats:

```ant
let seconds = Atomic.new(0.5)
seconds.add(0.25) // 0.75
seconds.add(-1.0) // -0.25
```

Every integer type works. Convert `d` to the cell's type:

```ant
let small = Atomic.new(int8(100))
small.add(int8(27)) // 127
let level = Atomic.new(uint16(1000))
level.add(uint16(24)) // 1024
```

## Applications

Hand out unique ticket numbers. `add` returns the new value, so each goroutine gets its own number
without reading the cell a second time:

```ant
let next = Atomic.new(0)
let tickets = Array.filled(5, 0)
let wg = WaitGroup.new()
for i in Array.range(0, 5) {
  wg.add(1)
  go {
    defer { wg.done() }
    tickets[i] = next.add(1)
  }
}
wg.wait()
println(tickets.sort())
```

```text
[1, 2, 3, 4, 5]
```

Total the bytes that parallel workers process:

```ant
let chunks = ["ant", "colony", "queen", "worker"]
let bytes = Atomic.new(0)
Parallel.each(chunks, func(chunk: string, work: Context) {
  bytes.add(chunk.length())
}, limit: 2)
println(bytes.load())
```

```text
20
```

Track how many jobs are running. Each job adds 1 when it starts and -1 when it ends, and the
deferred `add` runs even when the job returns early:

```ant
let running = Atomic.new(0)

func job(n: int) {
  running.add(1)
  defer { running.add(-1) }
  if n % 2 == 0 { return }
  Time.sleep(1ms)
}

let wg = WaitGroup.new()
for n in Array.range(0, 6) {
  wg.add(1)
  go {
    defer { wg.done() }
    job(n)
  }
}
wg.wait()
println(running.load())
```

```text
0
```

## Properties and relations

The value `add` returns is the value a `load` sees right after it, when nothing else changes the
cell:

```ant
let a = Atomic.new(3)
let after = a.add(4)
assert(after == a.load())
assert(after == 3 + 4)
```

Adding `d` and then `-d` gives back the start value:

```ant
let a = Atomic.new(50)
a.add(8)
a.add(-8)
assert(a.load() == 50)
```

Concurrent adds total the same as adding in sequence, in any order:

```ant
let amounts = [5, -2, 10, 7, -4, 1]
let sum = Atomic.new(0)
let wg = WaitGroup.new()
for d in amounts {
  wg.add(1)
  go {
    defer { wg.done() }
    sum.add(d)
  }
}
wg.wait()
assert(sum.load() == amounts.sum())
```

## Possible issues

An integer cell panics when the new value does not fit its type, in either direction:

```ant
let hits = Atomic.new(Math.maxInt)
hits.add(1) // panic: Atomic.add: integer overflow
```

```ant
let small = Atomic.new(int8(120))
small.add(int8(8)) // panic: Atomic.add: integer overflow
```

An integer literal is an `int`, so a cell of another integer type does not take it:

```ant
let small = Atomic.new(int8(1))
small.add(1) // error: cannot use int where int8 is expected
```

A float literal is a `float64`, so a `float32` cell needs a conversion too:

```ant
let ratio = Atomic.new(float32(1))
ratio.add(0.5) // error: cannot use float64 where float32 is expected
```

An unsigned cell cannot take a negative `d`, so it cannot count down with `add`. Use a signed cell
for a counter that goes both ways:

```ant
let free = Atomic.new(uint32(3))
free.add(uint32(-1)) // panic: -1 out of range for uint32
```

`add` needs a cell of a number type. A named type over a number is not one; keep the cell's type
numeric and convert at the edges:

```ant
type Count = int

let c = Atomic.new(Count(1))
c.add(Count(1))
// error: T would be Count, which is not a member of its bound Number; unwrap it with int(...)
```

A `load` followed by an `add` is two operations, and another goroutine can add in between. To cap a
counter, decide from the value `add` returns, and undo the add when it went past the cap:

```ant
let limit = 3
let slots = Atomic.new(0)

func tryTake(): bool {
  if slots.add(1) > limit {
    slots.add(-1)
    return false
  }
  true
}

let taken = Atomic.new(0)
let wg = WaitGroup.new()
for _ in Array.range(0, 10) {
  wg.add(1)
  go {
    defer { wg.done() }
    if tryTake() { taken.add(1) }
  }
}
wg.wait()
println(taken.load())
```

```text
3
```

## See also

[Atomic.new](new.md) · [Atomic.load](load.md) · [Atomic.store](store.md) ·
[Atomic.compareAndSwap](compareAndSwap.md) · [Atomic](index.md) · [Mutex.lock](../Mutex/lock.md) ·
[WaitGroup.add](../WaitGroup/add.md) · [Parallel.each](../Parallel/each.md) ·
[go](../../language/keywords/go.md) · [Concurrency guide](../../guide/concurrency.md)
