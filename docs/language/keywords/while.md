<!-- status: reviewed -->

# `while`

`while cond { ... }` evaluates `cond` and runs its block as long as it is `true`. Its type is
`void`, or `never` for a `while true` with no `break`.

## Details

- The condition is evaluated before each iteration, so a condition that starts `false` runs the
  block zero times. It must be a `bool`; there is no truthiness.
- `while true` with no `break` in its body has type `never`: only `return` or a panic leaves it.
  With a `break`, its type is `void`.
- After a `while` whose body has no `break`, the complement of the condition holds, so
  `while x == null { ... }` leaves `x` without `null` below the loop.
- An assignment anywhere in the body cancels a narrowing from before the loop, from the loop head
  on.
- A record or map literal in the condition must be parenthesized, since its `{` would otherwise
  begin the body.

## Basic examples

Count down while a counter is positive:

```ant
var n = 3
while n > 0 {
  println(n)
  n -= 1
}
println("liftoff")
```

```text
3
2
1
liftoff
```

Take work from a queue until it is empty:

```ant
let queue = ["fetch", "parse", "save"]
while !queue.isEmpty() {
  let job = queue.removeAt(0)
  println("running ${job}")
}
```

```text
running fetch
running parse
running save
```

## Scope

### A condition that starts false

The block runs zero times:

```ant
var n = 0
while n > 0 {
  println(n)
  n -= 1
}
println("done")
```

```text
done
```

### `while true`

A `while true` loop with no `break` has type `never`, so it can end a function that returns a value.
The function leaves through `return`:

```ant
func isPrime(n: int): bool {
  if n < 2 { return false }
  var d = 2
  while d * d <= n {
    if n % d == 0 { return false }
    d += 1
  }
  true
}

func nextPrime(after: int): int {
  var n = after + 1
  while true {
    if isPrime(n) { return n }
    n += 1
  }
}

println(nextPrime(10))
println(nextPrime(13))
```

```text
11
17
```

### Run the body at least once

Ant has no `do`, so a loop that runs its body at least once puts the test at the end of a
`while true` body. The `break` makes the loop's type `void`:

```ant
var n = 0
while true {
  n += 1
  println(n)
  if n >= 3 { break }
}
```

```text
1
2
3
```

### Narrowing after the loop

The loop ends only when its condition is `false`, so below a loop with no `break`, `reading` is an
`int`:

```ant
var reading: int | null = null
var tries = 0
while reading == null {
  tries += 1
  if tries == 3 { reading = 42 }
}
println(reading + 1)
```

```text
43
```

### Walking a linked structure

`while node != null` narrows `node` inside the body, and the body moves it along:

```ant
type Node = { value: int, next: Node | null }

func sum(head: Node | null): int {
  var node = head
  var total = 0
  while node != null {
    total += node.value
    node = node.next
  }
  total
}

let list = Node{ value: 1, next: Node{ value: 2, next: Node{ value: 3, next: null } } }
println(sum(list))
println(sum(null))
```

```text
6
0
```

### `continue` and `break`

`continue` goes back to the condition, and `break` leaves the loop. Advance the counter before a
`continue`, or the loop sees the same element again:

```ant
let xs = [4, -1, 0, 7, -3, 9]
var i = 0
while i < xs.length() {
  let x = xs[i]
  i += 1
  if x < 0 { continue }
  if x == 0 { break }
  println(x)
}
```

```text
4
```

## Applications

### Greatest common divisor

Euclid's algorithm runs until the remainder is zero:

```ant
func gcd(a: int, b: int): int {
  var x = a
  var y = b
  while y != 0 {
    let r = x % y
    x = y
    y = r
  }
  x
}

println(gcd(48, 18))
println(gcd(17, 5))
```

```text
6
1
```

### Binary search

The range to search shrinks until it is empty:

```ant
func search(xs: Array[int], target: int): int | null {
  var lo = 0
  var hi = xs.length()
  while lo < hi {
    let mid = (lo + hi) / 2
    if xs[mid] == target { return mid }
    if xs[mid] < target {
      lo = mid + 1
    } else {
      hi = mid
    }
  }
  null
}

let primes = [2, 3, 5, 7, 11, 13]
println(search(primes, 7))
println(search(primes, 4))
```

```text
3
null
```

### Read input to the end

`readLine` returns `null` at the end of input, and `?? break` leaves the loop there:

<!-- test: stdin
3
4
5
-->

```ant
let input = Os.stdin()
var total = 0
while true {
  let line = try input.readLine() ?? break
  total += line.toInt() ?? 0
}
println(total)
```

```text
12
```

## Properties and relations

A `while` with a counter visits the same numbers as a `for` over `Iter.range`:

```ant
let viaWhile: Array[int] = []
var i = 0
while i < 5 {
  viaWhile.push(i)
  i += 1
}
let viaFor: Array[int] = []
for j in Iter.range(0, 5) {
  viaFor.push(j)
}
assert(viaWhile == viaFor)
```

When the loop goes over a collection or a range, prefer `for`: it cannot forget the step.

After a loop with no `break`, its condition is `false`:

```ant
var n = 37
var steps = 0
while n != 1 {
  n = if n % 2 == 0 { n / 2 } else { 3 * n + 1 }
  steps += 1
}
assert(n == 1)
assert(steps == 21)
```

## Possible issues

### The condition must be a `bool`

An `int` is not a condition. Compare it:

```ant
var n = 3
while n { // error: cannot use int where bool is expected
  n -= 1
}
```

Write `while n > 0` instead.

### A record literal in the condition

The `{` of a record literal would begin the body, so the checker asks for parentheses:

```ant
type Point = { x: int, y: int }
var p = Point{ x: 0, y: 0 }
while p != Point{ x: 3, y: 0 } { // error: a record or map literal in the while header must be parenthesized
  p.x += 1
}
```

Write `while p != (Point{ x: 3, y: 0 })`.

### An assignment in the body cancels a narrowing

A guard narrows `item` to `int`, but the body assigns it a value that may be `null`, so the
narrowing does not reach the condition:

```ant
let stock = Map{ tea: 3 }
var item: int | null = stock.get("tea")
if item == null { panic("no tea") }
while item > 0 { // error: > needs both operands of one type, got int | null and int
  item = stock.get("rice")
}
```

### A `break` undoes the narrowing after the loop

With a `break`, the loop can end while the condition still holds, so nothing narrows below it:

```ant
var reading: int | null = null
var tries = 0
while reading == null {
  tries += 1
  if tries == 3 { reading = 42 }
  if tries > 10 { break }
}
println(reading + 1) // error: + needs both operands of one type, got int | null and int
```

Test `reading` again after the loop, or use `reading ?? 0`.

### A `while true` with `break` has type `void`

A `break` makes the loop's type `void`, so it no longer ends a function that returns a value:

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

### A loop that never ends

A `while` stops only when its condition changes. A `continue` placed before the step, or a step that
never reaches the end, keeps the loop running forever. Put the step first, as in the `continue`
example above, or use `for` over a collection or a range.

## See also

[for](for.md) · [break](break.md) · [continue](continue.md) · [return](return.md) · [var](var.md) ·
[if](if.md) · [Iter.range](../../lib/Iter/range.md)
