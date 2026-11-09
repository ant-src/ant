<!-- status: reviewed -->

# Everything is an expression

In Ant, `if` gives a value and a block gives the value of its last expression, so a function rarely
needs `return`. Loops do work instead of giving a value, and `return`, `break`, and `continue` fit
anywhere a value does.

```ant
func describe(n: int): string {
  if n < 0 {
    "below zero"
  } else if n == 0 {
    "zero"
  } else {
    let parity = if n % 2 == 0 { "even" } else { "odd" }
    "${parity}, above zero"
  }
}

func reading(xs: Array[int], i: int): string {
  let n = Array.get(xs, i) ?? return "no reading at ${i}"
  describe(n)
}

let readings = [3, -2, 8, 0, 7]

var sum = 0
for n in readings {
  if n == 0 { break }
  if n < 0 { continue }
  sum += n
}

println(reading(readings, 0))
println(reading(readings, 1))
println(reading(readings, 3))
println(reading(readings, 9))
println("sum before the first zero: ${sum}")
```

```text
odd, above zero
below zero
zero
no reading at 9
sum before the first zero: 11
```

The body of `describe` is one `if`, and the `if` is the last expression of the function, so its
value is the result. Each branch is a block: the last branch first binds `parity` and then ends in
the string it gives back.

Without an `else`, an `if` gives `void` when its condition is false, so it cannot fill a place that
expects a `string`. Add an `else` to give a value on both paths:

```ant
let n = 3
let size: string = if n > 2 { "big" }
// error: an if without else has type void when the condition is false, which cannot be used where string is expected
```

`for` and `while` give no value, so a loop that builds a result updates a `var` declared before it,
as `sum` does. `break` leaves the loop, and `continue` skips to the next element.

`return` is an expression too. In `reading`, `??` falls back to its right side when `Array.get`
finds no element, and that right side leaves the function with a message. The rest of `reading` then
works with an `int`, never with `null`.

Try it: make `describe` say "large" for numbers above 5, turn the `break` into a `continue` and
watch the sum take in the 7, or rewrite the `for` loop over `readings` as a `while` loop with an
index.

The [control flow guide](../guide/control-flow.md) explains how these fit together. The reference
pages cover [`if`](../language/keywords/if.md), [`for`](../language/keywords/for.md),
[`while`](../language/keywords/while.md), [`return`](../language/keywords/return.md),
[`break`](../language/keywords/break.md), [`continue`](../language/keywords/continue.md), and
[`??`](../language/operators/null-coalescing.md).

Next: [sequences](09-sequences.md).
