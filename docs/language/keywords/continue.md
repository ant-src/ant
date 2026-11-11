<!-- status: reviewed -->

# `continue`

`continue` ends the current iteration of the innermost loop and starts the next one. In a `for`
loop, the next iteration takes the next element; in a `while` loop, the condition is tested again.

## Details

- `continue` acts on the innermost `for` or `while` loop. Ant has no labels, so it cannot reach an
  outer loop.
- `continue` has type `never`, so it can stand where a value is expected, as in `x ?? continue`.
- `continue` does not cross a function, a lambda, a `go` block, or a `defer` block. A `continue`
  with no loop inside the same boundary is a syntax error.
- In a `for` over a sequence, `continue` ends the body and `yield` returns `true`, so the sequence
  goes on.
- The deferred blocks of every block that `continue` leaves run before the next iteration starts.

## Basic examples

Skip the multiples of three:

```ant
for n in Iter.range(1, 8) {
  if n % 3 == 0 { continue }
  println(n)
}
```

```text
1
2
4
5
7
```

In a `while` loop, update the counter before `continue`, so the condition sees the new value:

```ant
var i = 0
while i < 5 {
  i += 1
  if i == 2 { continue }
  println(i)
}
```

```text
1
3
4
5
```

## Scope

Skip map entries while destructuring each `Entry`:

```ant
let stock = Map{ "apples": 3, "pears": 0, "plums": 5 }
for { key, value } in stock {
  if value == 0 { continue }
  println("${key}: ${value}")
}
```

```text
apples: 3
plums: 5
```

Skip characters of a string, which a `for` loop visits one at a time:

```ant
for c in "héllo" {
  if c == "l" { continue }
  print(c)
}
println()
```

```text
héo
```

In a `for` over a sequence, `continue` makes `yield` return `true`, and the sequence goes on to the
next element:

```ant
func countdown(from: int): Seq[int] {
  func(yield: func(int): bool) {
    var i = from
    while i > 0 {
      if !yield(i) { return }
      i -= 1
    }
    println("liftoff")
  }
}

for n in countdown(4) {
  if n % 2 == 0 { continue }
  println(n)
}
```

```text
3
1
liftoff
```

Because `continue` has type `never`, `?? continue` skips a `null` and leaves a non-null value:

```ant
var total = 0
for line in ["10", "x", "32"] {
  let n = line.toInt() ?? continue
  total += n
}
println(total)
```

```text
42
```

It also works as one branch of an `if` expression:

```ant
for x in [1, 2, 3] {
  let doubled = if x > 1 { x * 2 } else { continue }
  println(doubled)
}
```

```text
4
6
```

In nested loops, `continue` restarts only the inner loop:

```ant
for row in [[1, -2], [3, 4]] {
  for x in row {
    if x < 0 { continue }
    println(x)
  }
}
```

```text
1
3
4
```

A deferred block in the loop body runs when `continue` leaves the body:

```ant
for name in ["draft", "final"] {
  defer { println("closed ${name}") }
  if name == "draft" { continue }
  println("published ${name}")
}
```

```text
closed draft
published final
closed final
```

A `continue` in a `select` block acts on the loop around the `select`:

```ant
let lines: Channel[string] = Channel.new(capacity: 3)
lines.send("first")
lines.send("")
lines.send("last")
lines.close()
while true {
  select {
    let line = Channel.receive(lines) {
      if line == null { break }
      if line.isEmpty() { continue }
      println(line)
    }
  }
}
```

```text
first
last
```

## Properties and relations

A loop that skips with `continue` and collects the rest keeps what `Array.filter` keeps:

```ant
let xs = [4, -1, 7, -3]
let kept: Array[int] = []
for x in xs {
  if x < 0 { continue }
  kept.push(x)
}
assert(kept == xs.filter({ x >= 0 }))
```

## Possible issues

A `continue` in a lambda has no loop to target, even when the function that calls the lambda loops.
To drop elements, filter before you map:

```ant
let doubled = [1, 2, 3].map({
  if x == 2 { continue } // error: 'continue' outside a loop
  x * 2
})
```

```ant
[1, 2, 3] | Array.filter({ x != 2 }) | Array.map({ x * 2 }) // [2, 6]
```

The same holds for a `defer` block inside a loop body:

```ant
for x in [1, 2] {
  defer {
    if x == 1 { continue } // error: 'continue' outside a loop
  }
}
```

In a `while` loop, a `continue` before the update skips the update too, and the condition never
changes. Put the update first, as in Basic examples.

## See also

- [break](break.md): leave the loop instead of starting the next iteration
- [for](for.md) and [while](while.md): the loops `continue` acts on
- [return](return.md): leave the enclosing function
- [defer](defer.md): blocks that run when `continue` leaves a block
- [select](select.md): waiting on several channels inside a loop
- [Array.filter](../../lib/Array/filter.md): keep elements without a loop
