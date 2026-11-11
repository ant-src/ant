<!-- status: reviewed -->

# `break`

`break` ends the innermost `for` or `while` loop. The program continues after the loop.

## Details

- `break` takes no label. In nested loops it ends only the innermost one.
- `break` has type `never`, so it can stand where a value is expected, as in `x ?? break`.
- `break` does not cross a function, a lambda, a `go` block, or a `defer` block. A `break` with no
  loop inside the same boundary is a syntax error.
- `break` runs the `defer` blocks of every block it leaves.
- In a `for` over a sequence, `break` makes `yield` return `false`, and the sequence must then
  return.
- A `while true` loop has type `never` only when its body has no `break`. With a `break`, it has
  type `void`.

## Basic examples

Stop at the first negative number:

```ant
let xs = [3, 8, -1, 5]
for x in xs {
  if x < 0 { break }
  println(x)
}
```

```text
3
8
```

Leave a `while true` loop once a condition holds:

```ant
var n = 1
while true {
  n *= 2
  if n > 100 { break }
}
println(n)
```

```text
128
```

## Scope

### As a value

`break` has type `never`, so it fits after `??`. The loop ends at the first name without an age:

```ant
let ages = Map{ "ann": 31, "bob": 27 }
let names = ["ann", "cy", "bob"]
for name in names {
  let age = ages[name] ?? break
  println("${name}: ${age}")
}
```

```text
ann: 31
```

### Strings

A `for` over a string yields one character at a time:

```ant
let text = "hello world"
var length = 0
for ch in text {
  if ch == " " { break }
  length += 1
}
println(length)
```

```text
5
```

### Sequences

In a `for` over a sequence, `break` makes `yield` return `false`. The sequence sees it and returns:

```ant
func countdown(from: int): Seq[int] {
  func(yield: func(int): bool) {
    var i = from
    while i > 0 {
      if !yield(i) {
        println("stopped at ${i}")
        return
      }
      i -= 1
    }
  }
}

for n in countdown(5) {
  if n == 3 { break }
  println(n)
}
```

```text
5
4
stopped at 3
```

### With `defer`

`break` runs the `defer` blocks of the loop body it leaves:

```ant
for x in [1, 2] {
  defer { println("closed ${x}") }
  if x == 1 { break }
}
```

```text
closed 1
```

### Inside `select`

A `break` in a `select` block ends the loop around the `select`. Here either a closed `lines`
channel or a value on `quit` ends the `while true`:

```ant
let lines: Channel[string] = Channel.new(3)
let quit: Channel[void] = Channel.new(1)
lines.send("a")
lines.send("b")
lines.close()
while true {
  select {
    let line = Channel.receive(lines) {
      if line == null { break }
      println(line)
    }
    let _ = Channel.receive(quit) { break }
  }
}
println("done")
```

```text
a
b
done
```

## Applications

### Read input up to a blank line

<!-- test: stdin
Ann
Bob

Cy
-->

```ant
let input = Os.stdin()
let names: Array[string] = []
while true {
  let line = try input.readLine() ?? break
  if line == "" { break }
  names.push(line)
}
println(names | String.join(", "))
```

```text
Ann, Bob
```

The first `break` handles the end of input, the second the blank line.

### Leave nested loops

`break` has no label. To leave several loops at once, put them in a function and `return`:

```ant
type Cell = { row: int, col: int }

func findZero(grid: Array[Array[int]]): Cell | null {
  for { index: row, value: cells } in grid.entries() {
    for { index: col, value } in cells.entries() {
      if value == 0 { return Cell{ row, col } }
    }
  }
  null
}

println(findZero([[1, 2], [3, 0], [0, 4]]))
```

```text
Cell{ row: 1, col: 1 }
```

## Properties and relations

A loop that keeps the first match and breaks gives what `find` gives. Prefer `find` when the first
match is all the code needs:

```ant
func firstNegative(xs: Array[int]): int | null {
  var found: int | null = null
  for x in xs {
    if x < 0 {
      found = x
      break
    }
  }
  found
}

let xs = [4, -2, -7]
assert(firstNegative(xs) == xs.find({ x < 0 }))
assert(firstNegative([]) == null)
```

## Possible issues

### `break` in a lambda

A lambda is a boundary, so a `break` in it cannot end a loop outside it. Use a `for` loop instead of
`forEach` when the code needs to stop early:

```ant
let rows = [[1, 2], [3, 4]]
for row in rows {
  row.forEach({
    if x > 2 { break } // error: 'break' outside a loop
  })
}
```

### Only the innermost loop ends

The `break` ends the loop over cells. The loop over rows goes on to the next row:

```ant
let grid = [[1, 2, 3], [4, 0, 6], [7, 8, 9]]
for row in grid {
  let seen: Array[string] = []
  for cell in row {
    if cell == 0 { break }
    seen.push("${cell}")
  }
  println(seen | String.join(" "))
}
```

```text
1 2 3
4
7 8 9
```

### A `while true` with `break` has type `void`

Without a `break`, `while true` never ends normally, and its type is `never`. A `break` makes its
type `void`, so it no longer fits as the last expression of a function that returns a value:

```ant
func firstPowerOver(limit: int): int {
  var n = 1
  while true { // error: cannot use void where int is expected
    n *= 2
    if n > limit { break }
  }
}
```

Return the value from inside the loop instead: `if n > limit { return n }`.

## See also

[continue](continue.md), [return](return.md), [for](for.md), [while](while.md), [select](select.md),
[Array.find](../../lib/Array/find.md), [Seq](../../lib/Iter/types/Seq.md)
