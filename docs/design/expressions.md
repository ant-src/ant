<!-- status: reviewed -->

# Why everything is an expression

In Ant, a block gives the value of its last expression, `if` gives the value of the branch it runs,
and `return`, `break`, and `continue` fit anywhere a value does. What other languages call a
statement, such as an assignment or a loop, is an expression whose value is `void`. So the `if` that
chooses a value in code is the same `if` that chooses what a page shows in markup. This page
explains why Ant has no separate kind of statement, and what that costs.

## One way to choose a value

```ant
let stock = 3
let label = if stock == 0 { "sold out" } else if stock < 5 { "low stock" } else { "in stock" }
label // "low stock"
```

A language with an `if` statement needs a second construct to choose a value, such as the `?:`
operator of C and JavaScript, or it leaves you to declare a variable and assign it in each branch.
Ant keeps one construct for both jobs: the `if` you use to run code is the `if` that gives a value.

Because the choice is a value, the name that holds it is a `let`. The statement form needs a `var`
that starts with a placeholder, and nothing stops a later line from assigning it again. This form
works in Ant too, and it is shown here only for the contrast:

```ant
let stock = 3
var label = ""
if stock == 0 {
  label = "sold out"
} else if stock < 5 {
  label = "low stock"
} else {
  label = "in stock"
}
label // "low stock"
```

## Choices inside markup

Markup is an expression too: `<p>Hello</p>` is a value of type `Html`, and `{...}` puts any
expression inside it. Because `if` is one of those expressions, a component chooses what to show
with the `if` it already uses everywhere else:

```ant
type Task = { title: string, done: bool }

func summary(tasks: Array[Task]): Html {
  let left = tasks.filter({ !x.done }).length()
  <section>
    <h2>Tasks {if left > 0 { <span class="badge">{left}</span> }}</h2>
    {if tasks.isEmpty() {
      <p>Nothing to do.</p>
    } else if left == 0 {
      <p>All done.</p>
    } else {
      <ul>{tasks.map({ <li class={["task", if x.done { "done" }]}>{x.title}</li> })}</ul>
    }}
  </section>
}

println(summary([]))
println(summary([Task{ title: "Milk", done: true }]))
println(summary([Task{ title: "Milk", done: true }, Task{ title: "Eggs", done: false }]))
```

```text
<section><h2>Tasks </h2> <p>Nothing to do.</p></section>
<section><h2>Tasks </h2> <p>All done.</p></section>
<section><h2>Tasks <span class="badge">1</span></h2> <ul><li class="task done">Milk</li><li class="task">Eggs</li></ul></section>
```

An `if` without `else` gives `void` when its condition is false, and markup renders `void` as
nothing, so the badge and the `done` class need no second branch. An `else if` chain picks one of
three pieces of markup. The markup is the last expression of `summary`, so it is the result.

JSX, the markup of React, has the same braces, but JavaScript's `if` is a statement and cannot go
inside them. A choice in JSX takes a second construct, `?:` for two branches, nested `?:` for three,
and `&&` for one:

```jsx
function Summary({ tasks }) {
  const left = tasks.filter((t) => !t.done).length;
  return (
    <section>
      <h2>Tasks {left > 0 && <span className="badge">{left}</span>}</h2>
      {tasks.length === 0 ? (
        <p>Nothing to do.</p>
      ) : left === 0 ? (
        <p>All done.</p>
      ) : (
        <ul>
          {tasks.map((t) => (
            <li key={t.title} className={t.done ? "task done" : "task"}>
              {t.title}
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
```

The alternative is to assign a variable with `if` statements above the `return` and put the variable
in the markup, which moves the choice away from the place it shows up. `&&` has a trap of its own:
`{left && <span>...</span>}` gives `left` when it is `0`, and React renders the `0`. In Ant the slip
cannot happen. `&&` takes only `bool` operands, and markup rejects a `bool`, so the checker stops
both:

```ant
let left = 0
let badge = <h2>Tasks {left > 0 && <span>{left}</span>}</h2>
// error: markup takes a string, a number, void, Html, or an array of them, got bool
// error: cannot use Html where bool is expected
```

## The result is the last expression

A function body is a block, so its last expression is its result:

```ant
func shippingCost(weightKg: int, express: bool): int {
  if weightKg <= 0 {
    return 0
  }
  let base = 5 + weightKg * 2
  if express { base * 2 } else { base }
}

println(shippingCost(3, false))
println(shippingCost(3, true))
println(shippingCost(0, true))
```

```text
11
22
0
```

`return` is needed only to leave early. When every function ends in its result, a `return` in the
code always marks an early exit, such as the guard at the top of `shippingCost`, and the ordinary
result stays at the bottom where you look for it.

The same rule holds for every block: an `if` branch, a lambda, a `select` clause, and a function
body. A lambda needs no `return` for the same reason a function does not:

```ant
let prices = [12, 40, 7]
prices.map({ if x >= 20 { x - 5 } else { x } }) // [12, 35, 7]
```

## Exits fit anywhere a value does

`return`, `break`, and `continue` have the type `never`: they never produce a value, so they fit
where any value is expected. `panic` is an ordinary function whose result type is `never`. That lets
an exit sit inside the expression that needs it:

```ant
func port(text: string): Result[int, Error] {
  let n = if text.isEmpty() {
    8080
  } else {
    text.toInt() ?? return Result.fail("port ${text} is not a number")
  }
  if n < 1 || n > 65535 {
    return Result.fail("port ${n} is out of range")
  }
  Result.ok(n)
}

println(port(""))
println(port("80"))
println(port("http"))
```

```text
Ok{ value: 8080 }
Ok{ value: 80 }
Err{ error: Error{ message: "port http is not a number" } }
```

`??` gives its right side when the left is `null`, and here the right side leaves the function. The
right side of the `??` is `never`, so the `??` is an `int`, both branches of the `if` are `int`, and
so is `n`. The rest of `port` never sees `null`. Without exits as expressions, `port` would declare
`n` ahead, assign it in each branch, and test for `null` in a statement of its own.

`try` is a prefix operator, so it fits wherever a value goes, here on the right of `+=`. It unwraps
an `Ok` or returns the `Err` from the function:

```ant
func parseQuantity(text: string): Result[int, Error] {
  let n = text.toInt() ?? return Result.fail("${text} is not a number")
  Result.ok(n)
}

func total(lines: Array[string]): Result[int, Error] {
  var sum = 0
  for line in lines {
    let text = line.trim()
    if text.isEmpty() || text.startsWith("#") { continue }
    sum += try parseQuantity(text)
  }
  Result.ok(sum)
}

println(total(["# boxes", "3", "", "4"]))
println(total(["3", "three"]))
```

```text
Ok{ value: 7 }
Err{ error: Error{ message: "three is not a number" } }
```

## Statements are `void`

An assignment, a declaration, and a loop have the type `void`. They still run for their effect, as
statements do elsewhere, and the type checker reports one that lands where a value belongs:

```ant
var count = 0
if count = 5 { println("five") }
// error: cannot use void where bool is expected
```

A condition must be a `bool`, and an assignment is `void`, so the classic slip of `=` for `==` is a
compile error rather than a silent assignment. For the same reason, an assignment cannot be chained
as in `a = b = 0`.

## What it costs

- **An `if` without `else` can give `void`.** When the condition is false, the `if` has no branch to
  take, so it gives `void`. Its type is its branch's type or `void`, such as `int | void`. It works
  as a statement, but it cannot fill a place that expects a value:

  ```ant
  let stock = 3
  let note: string = if stock < 5 { "order more" }
  // error: an if without else has type void when the condition is false, which cannot be used where string is expected
  ```

  A chain of `is` tests that covers every member of a union needs no `else`.

- **A block stands only after a keyword.** Braces where a value is expected make a lambda, not a
  block, so Ant has no block expression of its own. Compute a multi-step value with `let` lines
  before it, or in a helper function:

  ```ant
  let base = 4
  let cost = { // error: a lambda needs an expected function type; annotate the variable or use func with typed parameters
    let doubled = base * 2
    doubled + 1
  }
  ```

- **Loops give no value.** A loop has no result: `break` takes no operand, and `for` and `while`
  have the type `void`. A loop that builds a result updates a `var` declared before it, as `sum`
  does in `total` above. Functions such as `map`, `filter`, and `find` cover the common cases
  without a loop. A `while true` that contains no `break` is the exception: only `return` or a panic
  can leave it, so its type is `never`.

- **A loop in markup renders nothing.** Markup takes `void`, so a `for` inside braces passes the
  checker, and its body's values are dropped. `map` is the loop in markup:

  ```ant
  let tasks = ["milk", "eggs"]
  println(<ul>{for t in tasks { <li>{t}</li> }}</ul>)
  println(<ul>{tasks.map({ <li>{x}</li> })}</ul>)
  ```

  ```text
  <ul></ul>
  <ul><li>milk</li><li>eggs</li></ul>
  ```

- **A body must end in its value.** A function that computes its result in a loop and forgets to
  name it at the end ends in the loop, which is `void`. The checker reports the loop:

  ```ant
  func sum(prices: Array[int]): int {
    var total = 0
    for p in prices { total += p }
    // error: cannot use void where int is expected
  }
  ```

- **Branches of different types make a union.** Without an expected type, an `if` has the union of
  its branch types. The mistake then shows up where the value is used, not where the branches
  disagree:

  ```ant
  let ready = true
  let limit = if ready { 10 } else { "none" }
  let doubled = limit * 2 // error: * needs both operands of one type, got int | string and int
  ```

  An annotation, `let limit: int = ...`, moves the error to the branch that does not fit.

- **A value where none is wanted.** A value of type `void` is always the value `void`, so a callback
  whose type returns `void` cannot end in a value. Discard the value with `_ = ...`, which is an
  assignment and so `void`:

  ```ant
  func forEach(xs: Array[string], f: func(string): void): void {
    for x in xs { f(x) }
  }

  let seen = Set.of(["ant"])
  forEach(["ant", "bee"], { seen.add(x) }) // error: cannot use bool where void is expected
  forEach(["ant", "bee"], { _ = seen.add(x) })
  ```

- **`return` takes everything to its right.** `return` takes the whole expression after it, so the
  `+ 1` in `count ?? return 0 + 1` belongs to the `return`. Parenthesize to end it sooner:

  ```ant
  func next(count: int | null): int {
    (count ?? return 0) + 1
  }

  println(next(null))
  println(next(4))
  ```

  ```text
  0
  5
  ```

## See also

The tour page [Everything is an expression](../tour/08-expressions.md) introduces the idea, and the
[control flow guide](../guide/control-flow.md) shows how the pieces fit together. Reference:
[`if`](../language/keywords/if.md), [`return`](../language/keywords/return.md),
[`try`](../language/keywords/try.md), [`??`](../language/operators/null-coalescing.md),
[`never`](../language/types/never.md), [lambdas](../language/literals/lambdas.md). The
[HTML guide](../guide/html.md) covers markup in full. The design note
[Why no exceptions](errors-are-values.md) covers a related decision.
