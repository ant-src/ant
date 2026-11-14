<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Random.fork

```ant signature
func Random.fork(r: Random): Random
```

Random.fork returns a generator with r's state, and moves r 2^128 draws ahead, so neither sequence
reaches the other's in any run. The new generator goes on with the values r would have given.

<!-- end generated -->

## Details

- Forking a seeded generator is as repeatable as drawing from it: the same seed and the same forks,
  in the same order, give the same generators on every run.
- Fork in one goroutine, before the others start. Forking is a change to `r`, so two goroutines that
  fork one shared generator without a lock race, as two that draw from it do.

## Basic examples

Fork a generator. The new one gives the values `r` would have given:

```ant
let r = Random.new(42)
let child = Random.fork(r)
Array.range(0, 5) | Array.map({ child.int(0, 100) }) // [22, 33, 33, 56, 50]
Array.range(0, 5) | Array.map({ r.int(0, 100) })     // [74, 3, 42, 49, 11]
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
[153, 172, 164, 161]
```

## Scope

A fork is a generator like any other. It renders as `Random` and compares by identity:

```ant
let r = Random.new(1)
let child = r.fork()
child      // Random
child == r // false
```

Each fork of one generator starts where the previous one moved it, so the forks differ from each
other:

```ant
let r = Random.new(42)
let a = r.fork()
let b = r.fork()
a.int(0, 100) // 22
b.int(0, 100) // 74
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
#.### sword, potion, shield
#.#######.#. sword, potion, shield
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
  }, limit: batches) ?? []
  4.0 * float64(hits.sum() ?? 0) / float64(batches * size)
}

println(estimatePi(1, 4, 10000))
println(estimatePi(1, 4, 10000))
```

```text
3.144
3.144
```

## Properties and relations

A fork of a fresh generator gives the same values as a fresh generator with the same seed:

```ant
func draws(r: Random): Array[int] {
  Array.range(0, 10) | Array.map({ r.int(0, 1000) })
}

assert(draws(Random.fork(Random.new(9))) == draws(Random.new(9)))
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
r.int(0, 100)              // 74
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

The streams of `r` and its fork never meet, but a fork of a fork can repeat a stream that `r` hands
out later. Here forking `a` moves it onto `b`'s stream, so `a` gives the values of `b`:

```ant
let r = Random.new(42)
let a = r.fork()
let b = r.fork()
a.fork()
println([a.int(0, 100), b.int(0, 100)])
```

```text
[74, 74]
```

## See also

[Random.new](new.md) · [Random.int](int.md) · [Random.float](float.md) · [Random.pick](pick.md) ·
[Random.shuffle](shuffle.md) · [Parallel.map](../Parallel/map.md) · [Random](index.md)
