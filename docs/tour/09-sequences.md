<!-- status: reviewed -->

# Sequences

A `for` loop walks an array, and it walks a sequence too: a function that hands out its elements one
at a time. You can write one yourself, and it may never run out.

```ant
for i in Iter.range(1, 4) {
  println("lap ${i}")
}

func countdown(from: int): Seq[int] {
  func(yield: func(int): bool) {
    var n = from
    while n > 0 {
      if !yield(n) { return }
      n -= 1
    }
  }
}

for n in countdown(3) {
  println(n)
  if n == 2 { break }
}

func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var n = 1
    while yield(n) { n += 1 }
  }
}

for n in naturals() {
  if n * n > 50 {
    println("first square over 50: ${n * n}")
    break
  }
}

let oddSquares = naturals()
  | Iter.map({ x * x })
  | Iter.filter({ x % 2 == 1 })
  | Iter.take(4)
println(oddSquares.toArray())
```

```text
lap 1
lap 2
lap 3
3
2
first square over 50: 64
[1, 9, 25, 49]
```

`Iter.range(1, 4)` is the sequence 1, 2, 3: it counts up to `4` and stops before it. The loop runs
its body once for each number, and no array of numbers is ever built.

`countdown` returns a `Seq[int]`, which is a function that takes a callback, `yield`. Each call of
`yield(n)` runs the body of the `for` loop once with `n`. When the body ends with `break`, `yield`
returns `false`, and the sequence must stop: that is what `if !yield(n) { return }` does. Calling
`yield` again after it returned `false` panics.

`naturals` counts forever. That is fine, because the loop decides how far to go: the `break` at 64
makes `yield` return `false`, and the `while` ends.

The `oddSquares` pipeline builds a chain with `Iter`. Building it runs nothing. `toArray` on the
last line runs it, and `Iter.take(4)` stops `naturals` once it has four odd squares, so `Iter.map`
squares only the seven numbers it needs. `Iter.fromArray(xs)` turns an array into a sequence, and a
chain like `oddSquares` runs again each time you loop over it.

Try changing `countdown(3)` to `countdown(10)`, or `Iter.take(4)` to `Iter.take(10)`.

The [collections chapter](../guide/collections.md) explains when to reach for a sequence instead of
an array. The reference has [`for`](../language/keywords/for.md), [Seq](../lib/Iter/types/Seq.md),
and the [Iter](../lib/Iter/index.md) functions: [Iter.range](../lib/Iter/range.md),
[Iter.map](../lib/Iter/map.md), [Iter.filter](../lib/Iter/filter.md),
[Iter.take](../lib/Iter/take.md), [Iter.fromArray](../lib/Iter/fromArray.md), and
[Iter.toArray](../lib/Iter/toArray.md).
