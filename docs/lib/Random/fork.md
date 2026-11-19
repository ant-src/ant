<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.fork

```ant signature
func Random.fork(r: Random): Random
```

Random.fork returns a new generator seeded from four draws of r. Its values are unrelated to r's and
to every other fork's, a fork of a fork included, so each goroutine can take its own.

<!-- end generated -->

## Details

- Forking a seeded generator is as repeatable as drawing from it: the same seed and the same forks,
  in the same order, give the same generators on every run.
- A fork draws four values from `r`, so `r` goes on four draws later.
- Forks are unrelated to each other in any shape: a goroutine that got a fork may fork it again for
  its own workers.
- Fork in one goroutine, before the others start. Forking is a change to `r`, so two goroutines that
  fork one shared generator without a lock race, as two that draw from it do.

## Basic examples

Fork a generator. The new one gives values of its own, and `r` goes on:

```ant
let r = Random.new(42)
let child = Random.fork(r)
Array.range(0, 5) | Array.map({ child.int(0, 100) }) // [13, 43, 51, 86, 10]
Array.range(0, 5) | Array.map({ r.int(0, 100) })     // [56, 50, 7, 61, 54]
```

Give each worker a generator of its own, forked before the workers start:

```ant
let root = Random.new(7)
let gens = Array.range(0, 4) | Array.map({ Random.fork(root) })
let sixes = Parallel.map(gens, func(g: Random, ctx: Context): int {
  Array.range(0, 1000) | Array.countWhere({ g.int(1, 7) == 6 })
}, limit: 4)
println(sixes)
```

```text
[137, 177, 171, 166]
```

## Scope

A fork is a generator like any other. It renders as `Random` and compares by identity:

```ant
let r = Random.new(1)
let child = r.fork()
child      // Random
child == r // false
```

Each fork of one generator is seeded from different draws, so the forks differ from each other:

```ant
let r = Random.new(42)
let a = r.fork()
let b = r.fork()
a.int(0, 100) // 13
b.int(0, 100) // 97
```

A generator made without a seed forks too. Its forks differ from run to run, as it does:

```ant
let r = Random.new()
let gens = Array.range(0, 3) | Array.map({ r.fork() })
gens.length() // 3
```

## Applications

Keep one stream per part of a program, so a change to one part does not change the values of
another. Here the loot stays the same however many tiles the map draws:

```ant
func dungeon(seed: int, tiles: int): string {
  let world = Random.new(seed)
  let terrain = world.fork()
  let loot = world.fork()
  let map = Array.range(0, tiles) | Array.map({ if terrain.bool() { "#" } else { "." } })
  let items = Array.range(0, 3) | Array.map({ loot.pick(["sword", "shield", "potion"]) ?? "" })
  "${map | String.join("")} ${items | String.join(", ")}"
}

println(dungeon(2026, 5))
println(dungeon(2026, 12))
```

```text
.##.# sword, potion, sword
.##.###.###. sword, potion, sword
```

Run a seeded simulation in parallel and still get the same answer every time. The trials are split
into batches, and each batch draws from its own fork:

```ant
func estimatePi(seed: int, batches: int, size: int): float64 {
  let root = Random.new(seed)
  let gens = Array.range(0, batches) | Array.map({ root.fork() })
  let hits = Parallel.map(gens, func(g: Random, ctx: Context): int {
    Array.range(0, size) | Array.countWhere({
      let px = g.float()
      let py = g.float()
      px * px + py * py < 1.0
    })
  }, limit: Os.cpus()) ?? []
  4.0 * float64(hits.sum() ?? 0) / float64(batches * size)
}

println(estimatePi(1, 4, 10000))
println(estimatePi(1, 4, 10000))
```

```text
3.1292
3.1292
```

The forks are made before the workers start, so the answer is the same however many batches run at
once. `Os.cpus()` is the limit for CPU-bound work like this.

## Properties and relations

Forking draws four values from `r`: afterwards it gives what a generator with the same seed gives
after four draws:

```ant
let forked = Random.new(9)
forked.fork()
let drawn = Random.new(9)
for _ in Array.range(0, 4) {
  _ = drawn.float()
}
assert(forked.float() == drawn.float())
```

A fork of a fork does not repeat a stream its parent hands out later:

```ant
let r = Random.new(42)
let a = r.fork()
let b = r.fork()
a.fork()
println([a.int(0, 100), b.int(0, 100)])
```

```text
[86, 97]
```

Forking is repeatable: two generators with one seed, forked the same way, give the same values:

```ant
func draws(r: Random): Array[int] {
  Array.range(0, 10) | Array.map({ r.int(0, 1000) })
}

let a = Random.new(5)
let b = Random.new(5)
a.fork()
b.fork()
assert(draws(a) == draws(b))
```

## Possible issues

Forking moves `r`. To replay the values of a seed, make a new generator from the seed:

```ant
let r = Random.new(42)
let child = r.fork()
r.int(0, 100)              // 56
Random.new(42).int(0, 100) // 22
```

Assigning a generator to another name does not make a second one. Both names draw from one
generator, so neither gets the values the other skipped:

```ant
let r = Random.new(42)
let same = r
same.int(0, 100) // 22
r.int(0, 100)    // 33
```

## See also

[Random.new](new.md) · [Random.int](int.md) · [Random.float](float.md) · [Random.pick](pick.md) ·
[Random.shuffle](shuffle.md) · [Parallel.map](../Parallel/map.md) · [Os.cpus](../Os/cpus.md) ·
[Random](index.md)
