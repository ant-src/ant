<!-- status: reviewed -->

# Idiomatic Ant

Choose the form that makes the operation, data flow, and failure path clear.

## Function calls and pipelines

### Use dots for local operations

Keep short questions and actions close to their receiver. Keep compact chains that express one
thought.

<!-- test: before
type Parser = { tokens: Array[string], pos: int }
func Parser.next(p: Parser): string | null {
  let token = p.tokens.get(p.pos)
  p.pos += 1
  token
}
func accept(text: string, name: string, p: Parser): bool {
-->
<!-- test: after
true
}
-->

```ant
if !text.contains("ant") { return false }
if name.trim().isEmpty() { return false }
p.next()
```

Count operations, not dots. Field access such as `config.server.name` locates a value. Name the
value when the path becomes hard to follow.

### Show transformations with pipes

Use pipes to show successive transformations. Put substantial stages on separate lines. Show
namespace changes when they help readers follow the value.

<!-- test: before
type Order = { customer: string, paid: bool }
type Summary = { customer: string, orders: int }
func summarize(e: Entry[string, Array[Order]]): Summary {
  Summary{ customer: e.key, orders: e.value.length() }
}
let orders = [
  Order{ customer: "Bob", paid: true },
  Order{ customer: "Ann", paid: true },
  Order{ customer: "Bob", paid: false },
]
-->

```ant
let summaries = orders
  | Array.filter({ x.paid })
  | Array.groupBy({ x.customer })
  | Map.entries()
  | Array.map(summarize)
  | Array.sortBy({ x.customer })
```

Keep comparable stages in the same style. Use dots inside callbacks for local operations. An aligned
dot chain also works when the intermediate values remain clear.

### Keep useful namespaces

Write the namespace when it explains an unfamiliar operation or gives several inputs equal
attention.

<!-- test: before
type Parser = { tokens: Array[string], pos: int }
func Parser.next(p: Parser): string | null {
  let token = p.tokens.get(p.pos)
  p.pos += 1
  token
}
func Parser.cond(p: Parser, condition: string): bool { p.next() == condition }
type Point = { x: float64, y: float64 }
func Point.distance(a: Point, b: Point): float64 {
  Math.sqrt((a.x - b.x) * (a.x - b.x) + (a.y - b.y) * (a.y - b.y))
}
let p = Parser{ tokens: ["if", "x"], pos: 0 }
let condition = "if"
let start = Point{ x: 0.0, y: 0.0 }
let end = Point{ x: 3.0, y: 4.0 }
-->

```ant
Parser.cond(p, condition)
Point.distance(start, end)
```

Use `p.next()` when the surrounding code makes its meaning clear. Choose a more descriptive receiver
name, such as `parser`, as the scope grows.

A single pipe can also make a step clear:

<!-- test: before
type Config = { name: string }
func parseConfig(text: string): Result[Config, Error] { Result.ok(Config{ name: text }) }
let text = "app"
let path = "app.conf"
-->

```ant
parseConfig(text) | Result.wrapErr("parse ${path}")
```

## Control flow

### Handle exceptional cases early

Reject invalid input before continuing. Keep the successful path easy to follow. Use `return` to
leave early and the final expression for the ordinary result.

<!-- test: before
type Row = { id: string, quantity: string }
type Order = { id: string, quantity: int }
-->

```ant
func parseOrder(row: Row): Result[Order, Error] {
  let quantity = row.quantity.toInt()
  if quantity == null {
    return Result.err(Error{ message: "order ${row.id}: quantity must be an integer" })
  }
  if quantity <= 0 {
    return Result.err(Error{ message: "order ${row.id}: quantity must be positive" })
  }
  Result.ok(Order{ id: row.id, quantity })
}
```

### Choose values with expressions

Use an `if` expression for a choice between values.

<!-- test: before
type Summary = { orders: int }
let summary = Summary{ orders: 1 }
-->

```ant
let noun = if summary.orders == 1 { "order" } else { "orders" }
```

Keep conditions short. Prepare complex inputs before testing them, and keep `!` close to the
predicate. Use a guard when handling absence needs a long error message or several steps.

## Absence and errors

### Default missing values; validate supplied values

Use `null` for ordinary absence and `Result` for expected failure. Apply a default only when absence
means the default is appropriate.

<!-- test: before
type Setting = { value: string, line: int }
func boundedInt(text: string, low: int, high: int): Result[int, Error] {
  let n = text.toInt()
  if n == null || n < low || n > high {
    return Result.err(Error{ message: "want an integer from ${low} to ${high}" })
  }
  Result.ok(n)
}
-->

```ant
func intSetting(
  settings: Map[string, Setting],
  key: string,
  fallback: int,
  low: int,
  high: int,
): Result[int, Error] {
  let setting = settings[key]
  if setting == null { return Result.ok(fallback) }
  boundedInt(setting.value, low, high)
    | Result.wrapErr("line ${setting.line}, ${key}")
}
```

Here, a missing port may use `8080`; a supplied `port=oops` must report an error. A fallback after
parsing would hide that invalid input.

### Propagate failures and add useful context

Use `try` to propagate expected failures. Add context where the caller knows something the callee
does not. Avoid repeating information already in the error.

<!-- test: before
type Config = { name: string }
func parseConfig(text: string): Result[Config, Error] { Result.ok(Config{ name: text }) }
-->

```ant
func loadConfig(path: string): Result[Config, Error] {
  let text = try File.read(path)
  parseConfig(text) | Result.wrapErr("parse ${path}")
}
```

Use `panic` and `assert` for broken program assumptions. Make the scope of `try` clear when
combining it with pipes.

## Callbacks and helpers

### Name parameters when their roles matter

Use a short lambda for a simple operation. Name parameters when nested callbacks need both the outer
and inner values.

<!-- test: before
type Item = { product: string }
type Order = { customer: string, paid: bool, items: Array[Item] }
let orders = [Order{ customer: "Ann", paid: true, items: [Item{ product: "tea" }] }]
-->

```ant
let lines = orders
  | Array.filter({ x.paid })
  | Array.flatMap(func(order: Order): Array[string] {
    order.items.map({ "${order.customer}: ${x.product}" })
  })
```

Extract helpers when their jobs deserve names or their bodies hide the main flow. Keep small
callbacks beside the operations they explain.

## Collections and loops

### Ask for the result you need

Use `some` to ask whether a match exists, `every` to check all elements, `find` to get the first
match, and `filter` to collect the matches.

<!-- test: before
type Order = { paid: bool }
let orders = [Order{ paid: true }, Order{ paid: false }]
-->

```ant
let hasUnpaid = orders.some({ !x.paid })
```

Choose operations whose stopping behavior fits the task. Keep predicate effects in mind when
replacing a full scan with an operation that can stop early.

### Stop on failure with a loop

Keep per-item validation, related updates, and early exits together.

<!-- test: before
type Row = { id: string }
type Order = { id: string }
func parseOrder(row: Row): Result[Order, Error] { Result.ok(Order{ id: row.id }) }
-->

```ant
func parseOrders(rows: Array[Row]): Result[Array[Order], Error] {
  let orders: Array[Order] = []
  for row in rows {
    orders.push(try parseOrder(row))
  }
  Result.ok(orders)
}
```

This stops at the first error. An eager `Array.map` followed by `Result.all` evaluates every item
before selecting the first error.

Write mutations and repeated actions as statements. A pipeline may end with an output operation.

## Bindings and intermediate values

Prefer `let` for bindings that stay fixed. Use `var` for reassignment. A fixed binding can still
hold an array, map, or record whose contents change.

Split a pipeline when a result is reused, controls a branch, deserves a name, or needs inspection.
Name the value by its role, such as `activeUsers`.

Bind an intermediate result before printing it. `println` returns `void` and does not pass the value
onward.

## Resource lifetimes

### Register cleanup beside acquisition

Put `defer` immediately after acquiring a resource. Choose the block that should own its lifetime.

```ant
func firstLine(path: string): Result[string | null, Error] {
  let file = try File.open(path)
  defer { file.close() }
  file.readLine()
}
```

### Check completion before reporting success

Return errors that affect the promised result. Deferred results are discarded, so close an output
explicitly when closing may flush or fail.

```ant
func writeAndClose(output: File.Handle, lines: Array[string]): Result[void, Error] {
  let buffered = output.buffered()
  defer { buffered.close() }
  for line in lines { try buffered.writeLine(line) }
  buffered.close()
}
```

The deferred close covers early exits. The final close reports a completion error. Make ownership
clear when a function takes or returns an open resource.

## Changing style safely

Preserve optional chaining, interface calls, evaluation order, and failure behavior when changing
style.

See [Functions](functions.md), [Errors](errors.md), [Collections and sequences](collections.md), and
[File.buffered](../lib/File/buffered.md) for how these parts behave.
