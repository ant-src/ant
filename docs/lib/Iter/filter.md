<!-- status: reviewed -->
<!-- generated from internal/prelude/prelude.d.ant; ANT_DOCS_UPDATE=1 go test ./tests -run TestDocsReference rewrites it -->

# Iter.filter

```ant signature
func Iter.filter[T](s: Seq[T], f: func(T): bool): Seq[T]
```

Iter.filter returns a sequence of the elements of s that f accepts.

<!-- end generated -->

## Details

- Lazy: `Iter.filter` runs nothing when called. `f` runs only when the result is run, by `for` or
  `Iter.toArray`, once per element the source yields.
- When the consumer stops early, as `Iter.take` or a `break` does, the source stops too, and `f` is
  not called on the elements after.
- Each run of the result runs the source again and calls `f` again.

## Basic examples

Keep the multiples of 7 below 30:

```ant
Iter.range(0, 30) | Iter.filter({ x % 7 == 0 }) | Iter.toArray() // [0, 7, 14, 21, 28]
```

Loop over the elements a predicate accepts:

```ant
for n in Iter.range(1, 20) | Iter.filter({ x % 5 == 0 }) {
  println(n)
}
```

```text
5
10
15
```

## Scope

Filter the elements of an array as a sequence:

```ant
Iter.fromArray(["ant", "beetle", "fly"]) | Iter.filter({ String.length(x) > 3 }) | Iter.toArray() // ["beetle"]
```

An empty sequence gives an empty sequence, and so does a predicate that accepts nothing:

```ant
Iter.range(5, 5) | Iter.filter({ x > 0 }) | Iter.toArray() // []
Iter.range(0, 5) | Iter.filter({ false }) | Iter.toArray() // []
```

Test a record through its fields:

```ant
type Task = { title: string, done: bool }
var tasks = [Task{ title: "write", done: true }, Task{ title: "test", done: false }]
Iter.fromArray(tasks) | Iter.filter({ !x.done }) | Iter.toArray() // [Task{ title: "test", done: false }]
```

Pass a named function or a function on a type instead of a lambda:

```ant
func isOdd(n: int): bool { n % 2 == 1 }
Iter.range(0, 8) | Iter.filter(isOdd) | Iter.toArray()                        // [1, 3, 5, 7]
Iter.fromArray(["a", "", "b"]) | Iter.filter(String.isEmpty) | Iter.toArray() // [""]
```

Call it through `.` on a sequence:

```ant
Iter.range(0, 10).filter({ x > 7 }).toArray() // [8, 9]
```

Filter a sequence you wrote yourself, here an endless one, and stop it with `Iter.take`:

```ant
func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var n = 1
    while true {
      if !yield(n) { return }
      n += 1
    }
  }
}
naturals() | Iter.filter({ x % 3 == 0 }) | Iter.take(4) | Iter.toArray() // [3, 6, 9, 12]
```

## Applications

Find the first five primes without choosing an upper bound:

```ant
func naturals(): Seq[int] {
  func(yield: func(int): bool) {
    var n = 1
    while true {
      if !yield(n) { return }
      n += 1
    }
  }
}

func isPrime(n: int): bool {
  if n < 2 { return false }
  var d = 2
  while d * d <= n {
    if n % d == 0 { return false }
    d += 1
  }
  true
}

println(naturals() | Iter.filter(isPrime) | Iter.take(5) | Iter.toArray())
```

```text
[2, 3, 5, 7, 11]
```

Select and then transform, without building an array in between:

```ant
type User = { name: string, active: bool }
var users = [
  User{ name: "Ann", active: true },
  User{ name: "Bob", active: false },
  User{ name: "Cid", active: true },
]
Iter.fromArray(users).filter({ x.active }).map({ x.name }).toArray() // ["Ann", "Cid"]
```

Stop at the first match with `break`; the predicate sees nothing after it:

```ant
var checked: Array[int] = []
for n in Iter.range(10, 100) | Iter.filter({ Array.push(checked, x); x % 13 == 0 }) {
  println(n)
  break
}
println(checked)
```

```text
13
[10, 11, 12, 13]
```

## Properties and relations

On an array, filtering the sequence gives what `Array.filter` gives:

```ant
var xs = [5, 8, 1, 9, 4]
var f: func(int): bool = { x > 4 }
assert((Iter.fromArray(xs) | Iter.filter(f) | Iter.toArray()) == Array.filter(xs, f))
```

The predicate does not run until the result runs, and then only as far as the consumer asks:

```ant
var calls: Array[int] = []
var evens = Iter.range(0, 10) | Iter.filter({ Array.push(calls, x); x % 2 == 0 })
calls                                 // []
evens | Iter.take(2) | Iter.toArray() // [0, 2]
calls                                 // [0, 1, 2]
```

`filter` and `map` run element by element, interleaved, not one pass after the other:

```ant
var log: Array[string] = []
var kept = Iter.range(1, 5) | Iter.filter({ Array.push(log, "f${x}"); x % 2 == 0 })
kept | Iter.map({ Array.push(log, "m${x}"); x * 10 }) | Iter.toArray() // [20, 40]
log                                                                    // ["f1", "f2", "m2", "f3", "f4", "m4"]
```

## Possible issues

Running the result twice runs the predicate twice. Collect it once with `Iter.toArray` when `f` is
costly or has effects:

```ant
var n = 0
var evens = Iter.range(0, 4).filter({ n += 1; x % 2 == 0 })
evens.toArray() // [0, 2]
evens.toArray() // [0, 2]
n               // 8
```

On an array, `.filter` is `Array.filter`, which runs `f` at once and gives an array. On a sequence
it is `Iter.filter`, which runs nothing until the result runs:

```ant
[1, 2, 3].filter({ x > 1 })                           // [2, 3]
Iter.fromArray([1, 2, 3]).filter({ x > 1 }).toArray() // [2, 3]
```

`Iter.filter` takes a sequence, not an array. Wrap an array with `Iter.fromArray`, or call
`Array.filter`:

```ant
var big = Iter.filter([1, 2, 3], { x > 1 }) // error: cannot use Array[int] where Seq[T] is expected
```

The predicate must return a `bool`. Ant has no truthy values, so an `int` is a compile error:

```ant
var nonZero = Iter.range(0, 3).filter({ x }) // error: cannot use int where bool is expected
```

The result keeps the element type of the source. Filtering out `null` gives a `Seq[int | null]`, not
a `Seq[int]`:

```ant
var scores: Array[int | null] = [90, null, 75]
var ints: Array[int] = Iter.fromArray(scores).filter({ x != null }).toArray()
// error: cannot use Seq[int | null] where Seq[int] is expected
```

Narrow in a loop instead, where the test on the element narrows it to `int`:

```ant
var scores: Array[int | null] = [90, null, 75]
var total = 0
for s in Iter.fromArray(scores) {
  if s != null { total += s }
}
println(total)
```

```text
165
```

## See also

[Iter.map](map.md), [Iter.take](take.md), [Iter.range](range.md), [Iter.fromArray](fromArray.md),
[Iter.toArray](toArray.md), [Array.filter](../Array/filter.md), [Iter](index.md)
